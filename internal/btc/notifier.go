package btc

import (
	"context"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/goatnetwork/goat-relayer/internal/db"
	"github.com/goatnetwork/goat-relayer/internal/types"
	"gorm.io/gorm"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/goatnetwork/goat-relayer/internal/config"
	log "github.com/sirupsen/logrus"
)

type BTCNotifier struct {
	client         *rpcclient.Client
	confirmations  int64
	currentHeight  uint64
	catchingStatus bool
	catchingMu     sync.Mutex
	syncStatus     *db.BtcSyncStatus
	syncMu         sync.Mutex

	cache  *BTCCache
	poller *BTCPoller
}

func NewBTCNotifier(client *rpcclient.Client, cache *BTCCache, poller *BTCPoller) *BTCNotifier {
	var maxBlockHeight int64 = -1
	var lastBlock db.BtcBlockData
	result := cache.db.Order("block_height desc").First(&lastBlock)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		log.Fatalf("Failed to get max block height from db, %v", result.Error)
	}
	if result.Error == nil {
		maxBlockHeight = int64(lastBlock.BlockHeight) - 1
	}

	if maxBlockHeight < int64(config.AppConfig.BTCStartHeight) {
		maxBlockHeight = int64(config.AppConfig.BTCStartHeight) - 1
	}
	log.Infof("New btc notify at max block height is %d", maxBlockHeight)

	var syncStatus db.BtcSyncStatus
	resultQuery := cache.db.First(&syncStatus)
	if resultQuery.Error != nil && resultQuery.Error == gorm.ErrRecordNotFound {
		syncStatus.ConfirmedHeight = int64(config.AppConfig.BTCStartHeight - 1)
		syncStatus.UnconfirmHeight = int64(config.AppConfig.BTCStartHeight - 1)
		syncStatus.UpdatedAt = time.Now()
		cache.db.Create(&syncStatus)
		log.Info("New btc notify sync status not found, create one")
	}
	log.Infof("New btc notify sync status confirmed height is %d", syncStatus.ConfirmedHeight)

	return &BTCNotifier{
		client:         client,
		confirmations:  int64(config.AppConfig.BTCConfirmations),
		currentHeight:  uint64(maxBlockHeight + 1),
		catchingStatus: true,
		syncStatus:     &syncStatus,
		cache:          cache,
		poller:         poller,
	}
}

func (bn *BTCNotifier) Start(ctx context.Context) {
	bn.cache.Start(ctx)
	go bn.poller.Start(ctx)
	go bn.checkConfirmations(ctx)
}

func (bn *BTCNotifier) checkConfirmations(ctx context.Context) {
	checkInterval := 30 * time.Second
	catchUpInterval := 1 * time.Second

	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("Stopping confirmation checks...")
			return
		case <-ticker.C:
			bestHeight, err := bn.client.GetBlockCount()
			if err != nil {
				log.Errorf("Error getting the latest block height: %v", err)
				continue
			}

			catchingStatus := int64(bn.currentHeight)+bn.confirmations < bestHeight
			bn.catchingMu.Lock()
			bn.catchingStatus = catchingStatus
			bn.catchingMu.Unlock()

			confirmedHeight := bestHeight - bn.confirmations
			log.Debugf("Btc check block confirmation best height: %d, confirmed height: %d, current height: %d, catching status: %v", bestHeight, confirmedHeight, bn.currentHeight, catchingStatus)

			bn.syncMu.Lock()
			syncConfirmedHeight := bn.syncStatus.ConfirmedHeight
			bn.syncMu.Unlock()

			if syncConfirmedHeight >= confirmedHeight {
				// important: check utxo deposit result need fetch sub script
				depositResults, err := bn.poller.state.GetDepositResultsNeedFetchSubScript()
				if err != nil {
					log.Errorf("Btc check block confirmation get deposit results need fetch sub script error: %v", err)
					continue
				}
				if len(depositResults) > 0 {
					log.Warnf("Btc check block confirmation found %d deposit results need fetch sub script", len(depositResults))
					for _, depositResult := range depositResults {
						log.Infof("Btc check block confirmation deposit result need fetch sub script: %s, %d, %s", depositResult.Txid, depositResult.TxOut, depositResult.Address)

						time.Sleep(catchUpInterval)

						txHash, err := chainhash.NewHashFromStr(depositResult.Txid)
						if err != nil {
							log.Errorf("Failed to new chain hash from string %s, %v", depositResult.Txid, err)
							break
						}
						tx, err := bn.client.GetRawTransactionVerbose(txHash)
						if err != nil {
							log.Errorf("Failed to fetch transaction details for Txid: %s, error: %v", depositResult.Txid, err)
							break
						}

						log.Infof("Transaction %s details: Confirmations: %d, BlockHash: %s, Inputs: %d, Outputs: %d",
							tx.Txid, tx.Confirmations, tx.BlockHash, len(tx.Vin), len(tx.Vout))

						if len(tx.Vout) > int(depositResult.TxOut) {
							txBlockNumber := uint64(bestHeight) - tx.Confirmations - 1
							pubkey, err := bn.poller.state.GetDepositKeyByBtcBlock(txBlockNumber)
							if err != nil {
								log.Errorf("Get current deposit key by btc height %d err %v", txBlockNumber, err)
								break
							}
							pubkeyBytes, err := base64.StdEncoding.DecodeString(pubkey.PubKey)
							if err != nil {
								log.Errorf("Base64 decode pubkey %s err %v", pubkey.PubKey, err)
								break
							}
							// trust goat layer2 result, just build sub script
							err = bn.poller.state.UpdateUtxoSubScript(depositResult.Txid, depositResult.TxOut, depositResult.Address, pubkeyBytes)
							if err != nil {
								log.Errorf("Failed to update utxo sub script for p2wsh, %v", err)
								break
							}

						} else {
							log.Errorf("Tx vout length not match: %s, tx vout length: %d, expect: %d", depositResult.Txid, len(tx.Vout), depositResult.TxOut+1)
							break
						}
					}
					continue
				}
				bn.poller.state.UpdateBtcSyncing(false)
				log.Debugf("Btc check block confirmation ignored by up to confirmed height, best height: %d, synced: %d, confirmed: %d", bestHeight, syncConfirmedHeight, confirmedHeight)
				continue
			}

			bn.poller.state.UpdateBtcSyncing(true)

			// get network fee
			feeEstimate, err := bn.client.EstimateSmartFee(1, nil)
			if err != nil {
				log.Errorf("Error estimating smart fee: %v", err)
				continue
			}
			var satoshiPerVByte uint64
			if feeEstimate == nil || feeEstimate.FeeRate == nil {
				log.Warnf("Fee estimate or fee rate is nil. Using default fee.")
				satoshiPerVByte = uint64(3) // use 3 sat/vbyte as default fee
			} else {
				satoshiPerVByte = uint64((*feeEstimate.FeeRate * 1e8) / 1000)
			}
			log.Infof("Btc network fee is %d sat/vbyte", satoshiPerVByte)
			bn.poller.state.UpdateBtcNetworkFee(satoshiPerVByte)

			newSyncHeight := syncConfirmedHeight
			grows := false
			log.Infof("Btc check block confirmation fired, best height: %d, from: %d, to: %d", bestHeight, syncConfirmedHeight+1, confirmedHeight)
			for height := syncConfirmedHeight + 1; height <= confirmedHeight; height++ {
				// blockInDb, err := bn.poller.GetBlock(uint64(height))
				// if err != nil {
				// 	log.Errorf("Error getting block at height %d from cache: %v", height, err)
				// 	break
				// }
				block, err := bn.getBlockAtHeight(height)
				if err != nil {
					log.Errorf("Btc getting block at %d error: %v", height, err)
					break
				}
				blockWithHeight := BlockWithHeight{
					Block:  block,
					Height: bn.currentHeight,
				}
				bn.cache.blockChan <- blockWithHeight
				// if blockInDb.BlockHash != block.BlockHash().String() {
				// 	// block change detect
				// 	log.Warnf("Btc block hash changed, height: %d, old hash: %s, new hash: %s,", height, blockInDb.BlockHash, block.BlockHash().String())
				// 	// TODO save to cache again
				// }
				grows = true
				newSyncHeight = height
				log.Debugf("Pushing block at height %d to confirmChan", height)
				bn.poller.confirmChan <- &types.BtcBlockExt{
					MsgBlock:    *block,
					BlockNumber: uint64(height),
				}
				log.Debugf("Pushed block at height %d to confirmChan", height)
				bn.currentHeight++

				if bn.syncStatus.ConfirmedHeight+10 < newSyncHeight {
					// save every 10 when catching up
					bn.syncMu.Lock()
					bn.syncStatus.ConfirmedHeight = newSyncHeight
					bn.syncStatus.UpdatedAt = time.Now()
					bn.cache.db.Save(bn.syncStatus)
					bn.syncMu.Unlock()
				}

				time.Sleep(catchUpInterval)
			}
			if grows {
				// Save sync status
				log.Infof("Btc sync status saving, new confirmed height: %d ", newSyncHeight)

				bn.syncMu.Lock()
				bn.syncStatus.ConfirmedHeight = newSyncHeight
				bn.syncStatus.UpdatedAt = time.Now()
				bn.cache.db.Save(bn.syncStatus)
				bn.syncMu.Unlock()
			}
		}
	}
}

func (bn *BTCNotifier) getBlockAtHeight(height int64) (*wire.MsgBlock, error) {
	blockHash, err := bn.client.GetBlockHash(height)
	if err != nil {
		return nil, fmt.Errorf("error getting block hash at height %d: %v", height, err)
	}

	block, err := bn.client.GetBlock(blockHash)
	if err != nil {
		return nil, fmt.Errorf("error getting block at height %d: %v", height, err)
	}

	return block, nil
}
