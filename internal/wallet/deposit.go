package wallet

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

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

func (w *WalletServer) depositLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Info("UnConfirm deposit query stopping...")
			return
		case deposit := <-w.depositCh:
			depositData, ok := deposit.(types.MsgUtxoDeposit)
			if !ok {
				log.Errorf("Invalid deposit data type")
				continue
			}
			err := w.state.AddUnconfirmDeposit(depositData.TxId, depositData.RawTx, depositData.EvmAddr, depositData.SignVersion)
			if err != nil {
				log.Errorf("Failed to add unconfirmed deposit: %v", err)
				continue
			}
		}
	}
}

func (w *WalletServer) processConfirmedDeposit(ctx context.Context) {
	for {
		queues := w.state.GetDepositState().UnconfirmQueue
		if len(queues) == 0 {
			time.Sleep(5 * time.Second)
			continue
		}

		deposit := queues[0]
		queues = queues[1:] // remove the first element
		w.state.UpdateDepositState(queues)

		tx := DepositTransaction{
			TxHash:      deposit.TxHash,
			RawTx:       deposit.RawTx,
			EvmAddress:  deposit.EvmAddr,
			SignVersion: deposit.SignVersion,
		}
		go w.confirmingDeposit(ctx, tx, 0)
	}
}

func (w *WalletServer) confirmingDeposit(ctx context.Context, tx DepositTransaction, attempt int) {
	if attempt > 7 {
		log.Errorf("Confirmed deposit discarded after 7 attempts, txHahs: %s", tx.TxHash)
		return
	}

	block, err := w.state.QueryBlockByTxHash(tx.TxHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// if tx not found, retry
			log.Info("Confirmed deposit not found, retrying...")
		} else {
			// if tx validate or other error found, add attempt
			log.Infof("Confirmed deposit error: %v, attempt %d retrying...", err, attempt)
			attempt++
		}
		// TODO sleep how long?
		time.Sleep(5 * time.Second)
		w.confirmingDeposit(ctx, tx, attempt)
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

	isProposer := w.signer.IsProposer()
	if isProposer {
		l2Info := w.state.GetL2Info()
		depositKey, err := hex.DecodeString(l2Info.DepositKey)
		if err != nil {
			log.Errorf("DecodeString DepositKey err: %v", err)
			return
		}

		proposer := w.state.GetEpochVoter().Proposer

		requestId := fmt.Sprintf("DEPOSIT:%s:%s", config.AppConfig.RelayerAddress, tx.TxHash)

		msgSignDeposit, err := newMsgSignDeposit(tx, proposer, depositKey, merkleRoot, proof, txIndex)
		if err != nil {
			log.Errorf("NewMsgSignDeposit err: %v", err)
			return
		}
		w.state.EventBus.Publish(internalstate.SigStart, *msgSignDeposit)

		log.Infof("P2P publish msgSignDeposit success, request id: %s", requestId)
	}
	// update Deposit status to confirmed
	err = w.state.SaveConfirmDeposit(tx.TxHash, tx.RawTx, tx.EvmAddress)
	if err != nil {
		log.Errorf("SaveConfirmDeposit err: %v", err)
		return
	}

	log.Infof("Confirmed deposit success, blockHeight: %v", tx.BlockHeight)
}

func newMsgSignDeposit(tx DepositTransaction, proposer string, pubKey []byte, merkleRoot []byte, proof []byte, txIndex uint32) (*types.MsgSignDeposit, error) {
	address := common.HexToAddress(tx.EvmAddress).Bytes()

	txHash, err := chainhash.NewHashFromStr(tx.TxHash)
	if err != nil {
		return nil, fmt.Errorf("newHashFromStr err: %v", err)
	}

	decodeString, err := hex.DecodeString(tx.RawTx)
	if err != nil {
		return nil, fmt.Errorf("decodeString err: %v", err)
	}

	noWitnessTx, err := btc.SerializeNoWitnessTx(decodeString)
	if err != nil {
		return nil, fmt.Errorf("serializeNoWitnessTx err: %v", err)
	}

	headers := make(map[uint64][]byte)
	headers[tx.BlockHeight] = tx.BlockHeader

	return &types.MsgSignDeposit{
		Proposer:          proposer,
		Version:           tx.SignVersion,
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
