package btc

import (
	"context"
	"time"

	"github.com/goatnetwork/goat-relayer/internal/db"
	"gorm.io/gorm"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/goatnetwork/goat-relayer/internal/config"
	log "github.com/sirupsen/logrus"
)

type BTCNotifier struct {
	client        *rpcclient.Client
	currentHeight uint64

	cache  *BTCCache
	poller *BTCPoller
}

func NewBTCNotifier(client *rpcclient.Client, cache *BTCCache, poller *BTCPoller) *BTCNotifier {
	var maxBlockHeight int64 = -1
	result := cache.db.Model(&db.BtcBlockData{}).Select("MAX(block_height)").Row()
	result.Scan(&maxBlockHeight)
	if maxBlockHeight < int64(config.AppConfig.BTCStartHeight) {
		maxBlockHeight = int64(config.AppConfig.BTCStartHeight) - 1
	}

	return &BTCNotifier{
		client:        client,
		currentHeight: uint64(maxBlockHeight + 1),
		cache:         cache,
		poller:        poller,
	}
}

func (bn *BTCNotifier) Start(ctx context.Context) {
	bn.cache.Start(ctx)
	bn.poller.Start(ctx)
	go bn.listenForBTCBlocks(ctx)
	go bn.checkConfirmations(ctx)
}

func (bn *BTCNotifier) listenForBTCBlocks(ctx context.Context) {
	normalInterval := 30 * time.Second
	catchUpInterval := 1 * time.Second
	errInterval := 5 * time.Second
	for {
		select {
		case <-ctx.Done():
			log.Info("Stopping BTC block listening...")
			return
		default:
			bestHeight, err := bn.client.GetBlockCount()
			if err != nil {
				log.Errorf("Btc getting the latest block height error: %v", err)
				time.Sleep(errInterval)
				continue
			}
			log.Infof("Btc starts to cache block at height %d, best height %d,", bn.currentHeight, bestHeight)

			if int64(bn.currentHeight) >= bestHeight {
				log.Infof("Btc reached the latest block, waiting for new blocks...")
				time.Sleep(normalInterval)
				continue
			}

			blockHash, err := bn.client.GetBlockHash(int64(bn.currentHeight))
			if err != nil {
				log.Errorf("Btc getting block hash error: %v", err)
				time.Sleep(errInterval)
				continue
			}

			block, err := bn.client.GetBlock(blockHash)
			if err != nil {
				log.Errorf("Btc getting block error: %v", err)
				time.Sleep(errInterval)
				continue
			}

			blockWithHeight := BlockWithHeight{
				Block:  block,
				Height: bn.currentHeight,
			}
			bn.cache.blockChan <- blockWithHeight
			bn.currentHeight++
			time.Sleep(catchUpInterval)
		}
	}
}

func (bn *BTCNotifier) checkConfirmations(ctx context.Context) {
	checkInterval := 30 * time.Second
	catchUpInterval := 1 * time.Second
	requiredConfirmations := int64(6) // Can be adjusted as needed

	var syncStatus db.BtcSyncStatus
	db := bn.cache.db
	result := db.First(&syncStatus)
	if result.Error != nil && result.Error == gorm.ErrRecordNotFound {
		syncStatus.ConfirmedHeight = int64(config.AppConfig.BTCStartHeight - 1)
		syncStatus.UnconfirmHeight = int64(config.AppConfig.BTCStartHeight - 1)
		syncStatus.UpdatedAt = time.Now()
		db.Create(&syncStatus)
	}

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
				log.Error("Error getting the latest block height: %v", err)
				continue
			}

			grows := false
			confirmedHeight := bestHeight - requiredConfirmations
			log.Infof("Btc check block confirmation fired, best height: %d, from: %d, to: %d", bestHeight, syncStatus.ConfirmedHeight+1, confirmedHeight)
			for height := syncStatus.ConfirmedHeight + 1; height <= confirmedHeight; height++ {
				blockInDb, err := bn.poller.GetBlock(uint64(height))
				if err != nil {
					log.Errorf("Error getting block at height %d from cache: %v", height, err)
					break
				}
				blockHash, err := bn.client.GetBlockHash(height)
				if err != nil {
					log.Errorf("Error getting block at height %d from client: %v", height, err)
					break
				}
				block, err := bn.client.GetBlock(blockHash)
				if err != nil {
					log.Errorf("Btc getting block at %d error: %v", height, err)
					break
				}
				if blockInDb.BlockHash != blockHash.String() {
					// block change detect
					log.Warnf("Btc block hash changed, height: %d, old hash: %s, new hash: %s,", height, blockInDb.BlockHash, blockHash.String())
					// TODO save to cache again
				}
				// TODO insert utxo db
				grows = true
				syncStatus.ConfirmedHeight = height
				log.Debugf("Pushing block at height %d to confirmChan", height)
				bn.poller.confirmChan <- &BtcBlockExt{
					MsgBlock:    *block,
					blockNumber: uint64(height),
				}
				log.Debugf("Pushed block at height %d to confirmChan", height)
				time.Sleep(catchUpInterval)
			}
			if grows {
				// Save sync status
				log.Infof("Btc sync status saving, new confirmed height: %d ", syncStatus.ConfirmedHeight)
				syncStatus.UpdatedAt = time.Now()
				db.Save(&syncStatus)
			}
		}
	}
}
