package utxo

import (
	"context"
	"encoding/hex"
	"errors"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/goatnetwork/goat-relayer/internal/btc"
	dbmodule "github.com/goatnetwork/goat-relayer/internal/db"
	"github.com/goatnetwork/goat-relayer/internal/state"
	bitcointypes "github.com/goatnetwork/goat/x/bitcoin/types"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"
)

var UnconfirmedChannel = make(chan DepositTransaction)
var ConfirmedChannel = make(chan DepositTransaction)

func (d *Deposit) QueryUnconfirmedDeposit(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Info("UnConfirm deposit query stopping...")
			return
		default:
			deposits, err := d.state.QueryUnConfirmDeposit()
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Errorf("QueryUnConfirmDeposit failed: %v", err)
				continue
			} else if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
				// TODO sleep how long?
				time.Sleep(5 * time.Second)
				continue
			}

			for _, deposit := range deposits {
				UnconfirmedChannel <- DepositTransaction{
					TxHash:     deposit.TxHash,
					RawTx:      deposit.RawTx,
					EvmAddress: deposit.EvmAddr,
				}
			}
		}
	}
}
func (d *Deposit) ProcessOnceConfirmedDeposit(ctx context.Context) {
	for tx := range UnconfirmedChannel {
		go onceConfirmedDeposit(tx, 0, d.cacheDb)
	}
}

func (d *Deposit) ProcessSixConfirmedDeposit(ctx context.Context) {
	for tx := range ConfirmedChannel {
		go sixConfirmedDeposit(tx, 0, d.lightDb, d.state)
	}
}

func onceConfirmedDeposit(tx DepositTransaction, attempt int, db *gorm.DB) {
	if attempt >= 3 {
		log.Errorf("onceConfirmedDeposit discarded after 3 attempts, txHahs: %s", tx.TxHash)
		return
	}

	found := true
	block, err := queryBtcCacheDatabaseForBlock(tx.TxHash, db)
	if err != nil {
		log.Errorf("onceConfirmedDeposit failed: %v", err)
		found = false
	}

	if found {
		tx.BlockHash = block.BlockHash
		tx.BlockHeight = block.BlockHeight
		tx.BlockHeader = block.Header
		// send to confirm channel
		ConfirmedChannel <- tx
	} else {
		log.Info("onceConfirmedDeposit not found, retrying...")
		// TODO sleep how long?
		time.Sleep(5 * time.Minute)             // wait 5 minutes
		onceConfirmedDeposit(tx, attempt+1, db) // recursive retry
	}
}

func sixConfirmedDeposit(tx DepositTransaction, attempt int, db *gorm.DB, state *state.State) {
	if attempt >= 7 {
		log.Errorf("sixConfirmedDeposit discarded after 3 attempts, blockHeight: %v", tx.BlockHeight)
		return
	}

	found := true
	_, err := queryBtcLightDatabaseForBlock(tx.BlockHeight, db)
	if err != nil {
		log.Errorf("sixConfirmedDeposit failed: %v", err)
		found = false
	}

	if found {
		err = state.SaveConfirmDeposit(tx.TxHash, tx.RawTx, tx.EvmAddress)
		if err != nil {
			log.Errorf("SaveConfirmDeposit err: %v", err)
			return
		}

		proposer := state.GetEpochVoter().Proposer
		_, err := newDeposit(tx, proposer)
		if err != nil {
			log.Errorf("newDeposit err: %v", err)
			return
		}

	} else {
		log.Info("sixConfirmedDeposit not found, retrying...")
		// TODO sleep how long?
		time.Sleep(10 * time.Minute)                  // wait 10 minutes
		sixConfirmedDeposit(tx, attempt+1, db, state) // recursive retry
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

	blockHashBytes := btcTxOutput.PkScript[:32] // Assuming the block hash is the first 32 bytes of PkScript
	blockHash, err := chainhash.NewHash(blockHashBytes)
	if err != nil {
		return nil, err
	}

	// checkout block
	if btcBlockData.BlockHash != blockHash.String() {
		return nil, errors.New("block hash mismatch")
	}

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

func newDeposit(tx DepositTransaction, proposer string) (*bitcointypes.MsgNewDeposits, error) {
	address, err := hex.DecodeString(tx.EvmAddress)
	if err != nil {
		return nil, err
	}

	decodeString, err := hex.DecodeString(tx.RawTx)
	if err != nil {
		return nil, err
	}

	noWitnessTx, err := btc.SerializeNoWitnessTx(decodeString)
	if err != nil {
		return nil, err
	}

	headers := make(map[uint64][]byte)
	// TODO block headers type
	headers[tx.BlockHeight] = []byte(tx.BlockHeader)

	Deposits := make([]*bitcointypes.Deposit, 1)
	Deposits[0] = &bitcointypes.Deposit{
		Version:           1,
		BlockNumber:       tx.BlockHeight,
		TxIndex:           0,
		NoWitnessTx:       noWitnessTx,
		OutputIndex:       0,
		IntermediateProof: nil,
		EvmAddress:        address,
		RelayerPubkey:     nil,
	}

	deposit := &bitcointypes.MsgNewDeposits{
		Proposer:     proposer,
		BlockHeaders: headers,
		Deposits:     Deposits,
	}
	return deposit, nil
}

type DepositTransaction struct {
	TxHash      string
	RawTx       string
	EvmAddress  string
	BlockHash   string
	BlockHeight uint64
	BlockHeader string
}
