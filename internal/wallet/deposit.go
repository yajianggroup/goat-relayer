package wallet

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/goatnetwork/goat-relayer/internal/config"
	"github.com/goatnetwork/goat-relayer/internal/types"

	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/goatnetwork/goat-relayer/internal/state"
	bitcointypes "github.com/goatnetwork/goat/x/bitcoin/types"
	log "github.com/sirupsen/logrus"
)

func (w *WalletServer) depositLoop(ctx context.Context) {
	w.state.EventBus.Subscribe(state.SigFailed, w.depositSigFailChan)
	w.state.EventBus.Subscribe(state.SigFinish, w.depositSigFinishChan)
	w.state.EventBus.Subscribe(state.SigTimeout, w.depositSigTimeoutChan)

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case deposit := <-w.depositCh:
			depositData, ok := deposit.(types.MsgUtxoDeposit)
			if !ok {
				log.Errorf("Invalid deposit data type")
				continue
			}
			err := w.state.AddUnconfirmDeposit(depositData.TxId, depositData.RawTx, depositData.EvmAddr, depositData.SignVersion, depositData.OutputIndex)
			if err != nil {
				log.Errorf("Failed to add unconfirmed deposit: %v", err)
				continue
			}
		case sigFail := <-w.depositSigFailChan:
			w.handleDepositSigFailed(sigFail, "failed")
		case sigTimeout := <-w.depositSigTimeoutChan:
			w.handleDepositSigFailed(sigTimeout, "timeout")
		case sigFinish := <-w.depositSigFinishChan:
			w.handleDepositSigFinish(sigFinish)
		case <-ticker.C:
			w.initDepositSig()
		}
	}
}

func (w *WalletServer) handleDepositSigFailed(event interface{}, reason string) {
	w.sigDepositMu.Lock()
	defer w.sigDepositMu.Unlock()

	if !w.sigDepositStatus {
		log.Debug("Event handleDepositSigFailed ignore, sigDepositStatus is false")
		return
	}

	switch e := event.(type) {
	case types.MsgSignDeposit:
		log.Infof("Event handleDepositSigFailed is of type MsgSignDeposit, request id %s, reason: %s", e.RequestId, reason)
		w.sigDepositStatus = false
	default:
		log.Debug("WalletServer depositLoop ignore unsupport type")
	}
}

func (w *WalletServer) handleDepositSigFinish(event interface{}) {
	w.sigDepositMu.Lock()
	defer w.sigDepositMu.Unlock()

	if !w.sigDepositStatus {
		log.Debug("Event handleDepositSigFinish ignore, sigDepositStatus is false")
		return
	}

	switch e := event.(type) {
	case types.MsgSignDeposit:
		log.Infof("Event handleDepositSigFinish is of type MsgSignDeposit, request id %s", e.RequestId)
		w.sigDepositStatus = false
		w.sigDepositFinishHeight = w.state.GetL2Info().Height
	default:
		log.Debug("WalletServer depositLoop ignore unsupport type")
	}
}

func (w *WalletServer) initDepositSig() {
	log.Debug("WalletServer initDepositSig")

	// 1. check catching up, self is proposer
	l2Info := w.state.GetL2Info()
	if l2Info.Syncing {
		log.Debug("WalletServer initDepositSig ignore, layer2 is catching up")
		return
	}

	if w.state.GetBtcHead().Syncing {
		log.Debug("WalletServer initDepositSig ignore, btc is catching up")
		return
	}

	w.sigDepositMu.Lock()
	defer w.sigDepositMu.Unlock()

	epochVoter := w.state.GetEpochVoter()
	if epochVoter.Proposer != config.AppConfig.RelayerAddress {
		if w.sigDepositStatus && l2Info.Height > epochVoter.Height+1 {
			w.sigDepositStatus = false
			// clean process, role changed, remove all status "create", "aggregating"
			w.cleanDepositProcess()
		}
		log.Debugf("WalletServer initDepositSig ignore, self is not proposer, epoch: %d, proposer: %s", epochVoter.Epoch, epochVoter.Proposer)
		return
	}

	// 2. check if there is a sig in progress
	if w.sigDepositStatus {
		log.Debug("WalletServer initDepositSig ignore, there is a sig")
		return
	}
	if l2Info.Height <= w.sigDepositFinishHeight+2 {
		log.Debug("WalletServer initDepositSig ignore, last finish sig in 2 blocks")
		return
	}

	// 3. find confirmed deposits for signing
	deposits, err := w.state.GetDepositForSign(16)
	if err != nil {
		log.Errorf("WalletServer initDepositSig error: %v", err)
		return
	}
	if len(deposits) == 0 {
		log.Debug("WalletServer initDepositSig ignore, no deposit for sign")
		return
	}

	// 4. spv verify
	blockHashes := make([]string, 0)
	for _, deposit := range deposits {
		txhash, err := chainhash.NewHashFromStr(deposit.TxHash)
		if err != nil {
			log.Errorf("NewHashFromStr TxHash err: %v", err)
			continue
		}
		txIndex := uint32(deposit.TxIndex)
		if bitcointypes.VerifyMerkelProof(txhash[:], deposit.MerkleRoot, deposit.Proof, txIndex) {
			blockHashes = append(blockHashes, deposit.BlockHash)
		}
	}

	if len(blockHashes) == 0 {
		log.Debug("WalletServer initDepositSig ignore, no valid block hash")
		return
	}

	// 5. build sign msg
	pubKey, err := hex.DecodeString(l2Info.DepositKey)
	if err != nil {
		log.Errorf("DecodeString DepositKey err: %v", err)
		return
	}

	// 6. get block headers
	blockData, err := w.state.QueryBtcBlockDataByBlockHashes(blockHashes)
	if err != nil {
		log.Errorf("QueryBtcBlockDataByBlockHashes error: %v", err)
		return
	}
	blockHeaders := make(map[uint64][]byte)
	for _, blockData := range blockData {
		blockHeaders[blockData.BlockHeight] = blockData.Header
	}
	proposer := w.state.GetEpochVoter().Proposer
	headersBytes, err := json.Marshal(blockHeaders)
	if err != nil {
		log.Errorf("Failed to marshal headers: %v", err)
		return
	}

	requestId := fmt.Sprintf("DEPOSIT:%s:%s", config.AppConfig.RelayerAddress, deposits[0].TxHash)
	var msgDepositTXs []*types.DepositTX
	for _, deposit := range deposits {
		txHash, err := chainhash.NewHashFromStr(deposit.TxHash)
		if err != nil {
			log.Errorf("NewHashFromStr err: %v", err)
			continue
		}

		// verify merkle proof
		success := bitcointypes.VerifyMerkelProof(txHash.CloneBytes(), []byte(deposit.MerkleRoot), []byte(deposit.Proof), uint32(deposit.TxIndex))
		if !success {
			log.Errorf("VerifyMerkelProof failed, txHash: %s", txHash.String())
			continue
		}

		noWitnessTx, err := types.SerializeNoWitnessTx([]byte(deposit.RawTx))
		if err != nil {
			log.Errorf("SerializeNoWitnessTx err: %v", err)
			continue
		}
		msgDepositTX := types.DepositTX{
			Version:           deposit.SignVersion,
			BlockNumber:       deposit.BlockHeight,
			TxHash:            []byte(deposit.TxHash),
			TxIndex:           uint32(deposit.TxIndex),
			NoWitnessTx:       noWitnessTx,
			MerkleRoot:        []byte(deposit.MerkleRoot),
			OutputIndex:       deposit.OutputIndex,
			IntermediateProof: []byte(deposit.Proof),
			EvmAddress:        []byte(deposit.EvmAddr),
		}
		msgDepositTXs = append(msgDepositTXs, &msgDepositTX)
	}
	msgSignDeposit := types.MsgSignDeposit{
		MsgSign: types.MsgSign{
			RequestId: requestId,
		},
		BlockHeader:   headersBytes,
		DepositTX:     msgDepositTXs,
		Proposer:      proposer,
		RelayerPubkey: pubKey,
	}
	w.state.EventBus.Publish(state.SigStart, msgSignDeposit)
	w.sigDepositStatus = true
	log.Infof("P2P publish msgSignDeposit success, request id: %s", requestId)
}

func (w *WalletServer) cleanDepositProcess() {
	// TODO remove all status "create", "aggregating"
}
