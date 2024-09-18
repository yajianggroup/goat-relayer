package utxo

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/goatnetwork/goat-relayer/internal/p2p"
	"github.com/goatnetwork/goat-relayer/internal/types"
	relayertypes "github.com/goatnetwork/goat/x/relayer/types"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/goatnetwork/goat-relayer/internal/btc"
	dbmodule "github.com/goatnetwork/goat-relayer/internal/db"
	"github.com/goatnetwork/goat-relayer/internal/state"
	bitcointypes "github.com/goatnetwork/goat/x/bitcoin/types"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"
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
	if err != nil {
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
		log.Info("onceConfirmedDeposit not found, retrying...")
		// TODO sleep how long?
		time.Sleep(5 * time.Second) // wait 5 minutes
		log.Info("-------------------------------")
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
	if err != nil {
		log.Errorf("sixConfirmedDeposit failed: %v", err)
		found = false
	}

	if found {
		// generate spv proof
		merkleRoot, proof, txIndex, err := btc.GenerateSPVProof(tx.TxHash, tx.TxHashList)
		if err != nil {
			log.Errorf("GenerateSPVProof err: %v", err)
			return
		}

		l2Info := state.GetL2Info()

		depositKey, err := hex.DecodeString(l2Info.DepositKey)
		if err != nil {
			log.Errorf("DecodeString DepositKey err: %v", err)
			return
		}

		pubKey := relayertypes.DecodePublicKey(depositKey)

		proposer := state.GetEpochVoter().Proposer
		deposit, err := newDeposit(tx, proposer, txIndex, proof, pubKey)
		if err != nil {
			log.Errorf("newDeposit err: %v", err)
			return
		}

		txHash, err := chainhash.NewHashFromStr(tx.TxHash)
		if err != nil {
			log.Errorf("NewHashFromStr err: %v", err)
		}

		signDeposit := &types.MsgSignDeposit{
			MsgSign:    types.MsgSign{},
			Deposit:    deposit,
			TxHash:     txHash.CloneBytes(),
			MerkleRoot: merkleRoot,
			Proof:      proof,
			TxIndex:    txIndex,
		}

		// p2p broadcast
		p2pMsg := p2p.Message{
			MessageType: p2p.MessageTypeDepositReceive,
			RequestId:   "",
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

	} else {
		log.Info("sixConfirmedDeposit not found, retrying...")
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

func newDeposit(tx DepositTransaction, proposer string, txIndex uint32, proof []byte, pubKey *relayertypes.PublicKey) (*bitcointypes.MsgNewDeposits, error) {
	address := common.HexToAddress(tx.EvmAddress).Bytes()

	decodeString, err := hex.DecodeString(tx.RawTx)
	if err != nil {
		return nil, err
	}

	// TODO use wire.MsgTx to NoWitnessTx
	noWitnessTx, err := btc.SerializeNoWitnessTx(decodeString)
	if err != nil {
		return nil, err
	}

	headers := make(map[uint64][]byte)
	headers[tx.BlockHeight] = tx.BlockHeader

	deposits := make([]*bitcointypes.Deposit, 1)
	deposits[0] = &bitcointypes.Deposit{
		Version:           1,
		BlockNumber:       tx.BlockHeight,
		TxIndex:           txIndex,
		NoWitnessTx:       noWitnessTx,
		OutputIndex:       0,
		IntermediateProof: proof,
		EvmAddress:        address,
		RelayerPubkey:     pubKey,
	}

	deposit := &bitcointypes.MsgNewDeposits{
		Proposer:     proposer,
		BlockHeaders: headers,
		Deposits:     deposits,
	}
	return deposit, nil
}
