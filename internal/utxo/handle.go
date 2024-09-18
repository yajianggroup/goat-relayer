package utxo

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goatnetwork/goat-relayer/internal/config"
	"github.com/goatnetwork/goat-relayer/internal/p2p"
	"github.com/goatnetwork/goat-relayer/internal/types"

	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/goatnetwork/goat-relayer/internal/btc"
	dbmodule "github.com/goatnetwork/goat-relayer/internal/db"
	"github.com/goatnetwork/goat-relayer/internal/state"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type DepositTransaction struct {
	TxHash      string
	RawTx       string
	EvmAddress  string
	BlockHash   string
	BlockHeight uint64
	BlockHeader []byte
	TxHashList  []string
}

var TempUnconfirmedChannel = make(chan DepositTransaction)
var UnconfirmedChannel = make(chan DepositTransaction)
var ConfirmedChannel = make(chan DepositTransaction)

func (d *Deposit) QueryUnconfirmedDeposit(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Info("UnConfirm deposit query stopping...")
			return
		default:
			// TODO sleep how long?
			time.Sleep(5 * time.Second)

			deposits, err := d.state.QueryUnConfirmDeposit()
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Errorf("QueryUnConfirmDeposit failed: %v", err)
				continue
			} else if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}

			// TODO if need send to unConfirm stata queue

			for _, deposit := range deposits {
				TempUnconfirmedChannel <- DepositTransaction{
					TxHash:     deposit.TxHash,
					RawTx:      deposit.RawTx,
					EvmAddress: deposit.EvmAddr,
				}
			}
		}
	}
}

func deduplicate(TempUnconfirmedChannel <-chan DepositTransaction, UnconfirmedChannel chan<- DepositTransaction) {
	seen := make(map[string]struct{})

	for input := range TempUnconfirmedChannel {
		if _, exists := seen[input.TxHash]; !exists {
			seen[input.TxHash] = struct{}{}
			UnconfirmedChannel <- input
		}
	}
}

func (d *Deposit) ProcessOnceConfirmedDeposit(ctx context.Context) {
	go deduplicate(TempUnconfirmedChannel, UnconfirmedChannel)

	for tx := range UnconfirmedChannel {
		go onceConfirmedDeposit(tx, 0, d.cacheDb)
	}
}

func (d *Deposit) ProcessSixConfirmedDeposit(ctx context.Context) {
	for tx := range ConfirmedChannel {
		go sixConfirmedDeposit(ctx, tx, 0, d.lightDb, d.state)
	}
}

func onceConfirmedDeposit(tx DepositTransaction, attempt int, db *gorm.DB) {
	if attempt > 3 {
		log.Errorf("onceConfirmedDeposit discarded after 3 attempts, txHahs: %s", tx.TxHash)
		return
	}

	found := true
	block, err := queryBtcCacheDatabaseForBlock(tx.TxHash, db)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		log.Info("onceConfirmedDeposit not found, retrying...")
		// TODO sleep how long?
		time.Sleep(5 * time.Second) // wait 5 minutes

		onceConfirmedDeposit(tx, 0, db)
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Errorf("onceConfirmedDeposit failed: %v", err)
		found = false
	}

	if found {
		tx.BlockHash = block.BlockHash
		tx.BlockHeight = block.BlockHeight
		tx.BlockHeader = block.Header

		var txHashList []string
		var parsedHashes []chainhash.Hash
		err := json.Unmarshal([]byte(block.TxHashes), &parsedHashes)
		if err != nil {
			log.Errorf("Unmarshal TxHashes error: %v", err)
			return
		}

		for _, hash := range parsedHashes {
			txHashList = append(txHashList, hash.String())
		}

		tx.TxHashList = txHashList

		// send to confirm channel
		ConfirmedChannel <- tx
	} else {
		log.Info("onceConfirmedDeposit err, retrying...")
		// TODO sleep how long?
		time.Sleep(5 * time.Second)             // wait 5 minutes
		onceConfirmedDeposit(tx, attempt+1, db) // recursive retry
	}
}

func sixConfirmedDeposit(ctx context.Context, tx DepositTransaction, attempt int, db *gorm.DB, state *state.State) {
	if attempt > 7 {
		log.Errorf("sixConfirmedDeposit discarded after 7 attempts, blockHeight: %v", tx.BlockHeight)
		return
	}

	found := true
	_, err := queryBtcLightDatabaseForBlock(tx.BlockHeight, db)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		log.Info("sixConfirmedDeposit not found, retrying...")
		// TODO sleep how long?
		time.Sleep(5 * time.Second) // wait 5 minutes

		sixConfirmedDeposit(ctx, tx, attempt+1, db, state)
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Errorf("sixConfirmedDeposit failed: %v", err)
		found = false
	}

	if found {
		l2Info := state.GetL2Info()
		depositKey, err := hex.DecodeString(l2Info.DepositKey)
		if err != nil {
			log.Errorf("DecodeString DepositKey err: %v", err)
			return
		}

		proposer := state.GetEpochVoter().Proposer
		signDeposit, err := newMsgSignDeposit(tx, proposer, depositKey)
		if err != nil {
			log.Errorf("newMsgSignDeposit err: %v", err)
			return
		}

		requestId := fmt.Sprintf("DEPOSIT:%s:%s", config.AppConfig.RelayerAddress, tx.TxHash)
		// p2p broadcast
		p2pMsg := p2p.Message{
			MessageType: p2p.MessageTypeDepositReceive,
			RequestId:   requestId,
			DataType:    "MsgSignDeposit",
			Data:        *signDeposit,
		}
		if err := p2p.PublishMessage(ctx, p2pMsg); err != nil {
			log.Errorf("public MsgSignDeposit to p2p error: %v", err)
			return
		}

		// update Deposit status to confirmed
		err = state.SaveConfirmDeposit(tx.TxHash, tx.RawTx, tx.EvmAddress)
		if err != nil {
			log.Errorf("SaveConfirmDeposit err: %v", err)
			return
		}

		log.Infof("sixConfirmedDeposit success, blockHeight: %v", tx.BlockHeight)
	} else {
		log.Infof("sixConfirmedDeposit not found, retrying... blockHeight: %v", tx.BlockHeight)
		// TODO sleep how long?
		time.Sleep(5 * time.Second)                        // wait 10 minutes
		sixConfirmedDeposit(ctx, tx, attempt+1, db, state) // recursive retry
	}
}

func queryBtcCacheDatabaseForBlock(txHash string, db *gorm.DB) (block *dbmodule.BtcBlockData, err error) {
	var btcTxOutput dbmodule.BtcTXOutput
	if err := db.Where("tx_hash = ?", txHash).First(&btcTxOutput).Error; err != nil {
		return nil, err
	}

	var btcBlockData dbmodule.BtcBlockData
	if err := db.Where("id = ?", btcTxOutput.BlockID).First(&btcBlockData).Error; err != nil {
		return nil, err
	}

	// TODO check block hash by pkscript
	// blockHashBytes := btcTxOutput.PkScript[:32] // Assuming the block hash is the first 32 bytes of PkScript
	// blockHash, err := chainhash.NewHash(blockHashBytes)
	// if err != nil {
	// 	return nil, err
	// }

	// // checkout block
	// if btcBlockData.BlockHash != blockHash.String() {
	// 	return nil, errors.New("block hash mismatch")
	// }

	return &btcBlockData, nil
}

func queryBtcLightDatabaseForBlock(blockHeight uint64, db *gorm.DB) (block *dbmodule.BtcBlock, err error) {
	var btcBlock dbmodule.BtcBlock
	if err := db.Where("height = ? and status != ?", blockHeight, "unconfirm").First(&btcBlock).Error; err != nil {
		return nil, err
	}

	// TODO check block

	return &btcBlock, nil
}

func newMsgSignDeposit(tx DepositTransaction, proposer string, pubKey []byte) (*types.MsgSignDeposit, error) {
	address := common.HexToAddress(tx.EvmAddress).Bytes()

	txHash, err := chainhash.NewHashFromStr(tx.TxHash)
	if err != nil {
		log.Errorf("NewHashFromStr err: %v", err)
	}
	// generate spv proof
	merkleRoot, proof, txIndex, err := btc.GenerateSPVProof(tx.TxHash, tx.TxHashList)
	if err != nil {
		log.Errorf("GenerateSPVProof err: %v", err)
		return nil, err
	}
	decodeString, err := hex.DecodeString(tx.RawTx)
	if err != nil {
		log.Errorf("DecodeString err: %v", err)
		return nil, err
	}

	// TODO use wire.MsgTx to NoWitnessTx
	noWitnessTx, err := btc.SerializeNoWitnessTx(decodeString)
	if err != nil {
		log.Errorf("SerializeNoWitnessTx err: %v", err)
		return nil, err
	}

	headers := make(map[uint64][]byte)
	headers[tx.BlockHeight] = tx.BlockHeader

	return &types.MsgSignDeposit{
		Proposer:          proposer,
		Version:           1,
		BlockNumber:       tx.BlockHeight,
		TxHash:            txHash.CloneBytes(),
		TxIndex:           txIndex,
		NoWitnessTx:       noWitnessTx,
		MerkleRoot:        merkleRoot,
		OutputIndex:       0,
		IntermediateProof: proof,
		EvmAddress:        address,
		RelayerPubkey:     pubKey,
	}, nil
}
