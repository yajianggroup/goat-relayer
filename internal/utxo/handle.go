package utxo

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/goatnetwork/goat-relayer/internal/bls"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goatnetwork/goat-relayer/internal/config"
	"github.com/goatnetwork/goat-relayer/internal/types"

	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/goatnetwork/goat-relayer/internal/btc"
	internalstate "github.com/goatnetwork/goat-relayer/internal/state"
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

var deduplicateChannel = make(chan DepositTransaction)
var confirmedChannel = make(chan DepositTransaction)

func (d *Deposit) AddUnconfirmedDeposit(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Info("UnConfirm deposit query stopping...")
			return
		case deposit := <-d.confirmDepositCh:
			depositData, ok := deposit.(types.MsgUtxoDeposit)
			if !ok {
				log.Errorf("Invalid deposit data type")
				continue
			}
			err := d.state.AddUnconfirmDeposit(depositData.TxId, depositData.RawTx, depositData.EvmAddr)
			if err != nil {
				log.Errorf("Failed to add unconfirmed deposit: %v", err)
				continue
			}
			deduplicateChannel <- DepositTransaction{
				TxHash:     depositData.TxId,
				RawTx:      depositData.RawTx,
				EvmAddress: depositData.EvmAddr,
			}
		}
	}
}

func deduplicate(dedupChan <-chan DepositTransaction, confirmChan chan<- DepositTransaction) {
	seen := make(map[string]struct{})

	for input := range dedupChan {
		if _, exists := seen[input.TxHash]; !exists {
			seen[input.TxHash] = struct{}{}
			confirmChan <- input
		}
	}
}

func (d *Deposit) ProcessConfirmedDeposit(ctx context.Context) {
	go deduplicate(deduplicateChannel, confirmedChannel)

	for tx := range confirmedChannel {
		go confirmingDeposit(ctx, tx, 0, d.state, d.signer)
	}
}

func confirmingDeposit(ctx context.Context, tx DepositTransaction, attempt int, state *internalstate.State, signer *bls.Signer) {
	if attempt > 7 {
		log.Errorf("Confirmed deposit discarded after 7 attempts, txHahs: %s", tx.TxHash)
		return
	}

	block, err := state.QueryBlockByTxHash(tx.TxHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// if tx not found, retry
			log.Info("Confirmed deposit not found, retrying...")
		} else {
			// if tx validate or other error found, add attempt
			log.Infof("Confirmed deposit error, attempt %d retrying...", attempt)
			attempt++
		}
		// TODO sleep how long?
		time.Sleep(5 * time.Second)
		confirmingDeposit(ctx, tx, attempt, state, signer)
		return
	}

	tx.BlockHash = block.BlockHash
	tx.BlockHeight = block.BlockHeight
	tx.BlockHeader = block.Header

	var txHashList []string
	var parsedHashes []chainhash.Hash
	err = json.Unmarshal([]byte(block.TxHashes), &parsedHashes)
	if err != nil {
		log.Errorf("Unmarshal TxHashes error: %v", err)
		return
	}

	for _, hash := range parsedHashes {
		txHashList = append(txHashList, hash.String())
	}

	tx.TxHashList = txHashList

	// generate spv proof
	merkleRoot, proof, txIndex, err := btc.GenerateSPVProof(tx.TxHash, tx.TxHashList)
	if err != nil {
		log.Errorf("GenerateSPVProof err: %v", err)
		return
	}

	isProposer := signer.IsProposer()
	if isProposer {
		l2Info := state.GetL2Info()
		depositKey, err := hex.DecodeString(l2Info.DepositKey)
		if err != nil {
			log.Errorf("DecodeString DepositKey err: %v", err)
			return
		}

		proposer := state.GetEpochVoter().Proposer

		requestId := fmt.Sprintf("DEPOSIT:%s:%s", config.AppConfig.RelayerAddress, tx.TxHash)

		msgSignDeposit, err := newMsgSignDeposit(tx, proposer, depositKey, merkleRoot, proof, txIndex)
		if err != nil {
			log.Errorf("newMsgSignDeposit err: %v", err)
			return
		}
		state.EventBus.Publish(internalstate.SigStart, *msgSignDeposit)

		log.Infof("proposer submit MsgNewDeposits to consensus success, request id: %s", requestId)
	}
	// update Deposit status to confirmed
	err = state.SaveConfirmDeposit(tx.TxHash, tx.RawTx, tx.EvmAddress)
	if err != nil {
		log.Errorf("SaveConfirmDeposit err: %v", err)
		return
	}

	log.Infof("ConfirmedDeposit success, blockHeight: %v", tx.BlockHeight)
}

func newMsgSignDeposit(tx DepositTransaction, proposer string, pubKey []byte, merkleRoot []byte, proof []byte, txIndex uint32) (*types.MsgSignDeposit, error) {
	address := common.HexToAddress(tx.EvmAddress).Bytes()

	txHash, err := chainhash.NewHashFromStr(tx.TxHash)
	if err != nil {
		return nil, fmt.Errorf("NewHashFromStr err: %v", err)
	}

	decodeString, err := hex.DecodeString(tx.RawTx)
	if err != nil {
		return nil, fmt.Errorf("DecodeString err: %v", err)
	}

	noWitnessTx, err := btc.SerializeNoWitnessTx(decodeString)
	if err != nil {
		return nil, fmt.Errorf("SerializeNoWitnessTx err: %v", err)
	}

	headers := make(map[uint64][]byte)
	headers[tx.BlockHeight] = tx.BlockHeader

	return &types.MsgSignDeposit{
		Proposer:          proposer,
		Version:           1,
		BlockNumber:       tx.BlockHeight,
		BlockHeader:       tx.BlockHeader,
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
