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

func (w *WalletServer) handleDepositSigFailed(sigFail interface{}, reason string) {
	log.Infof("WalletServer handleDepositSigFailed, reason: %s", reason)
}

func (w *WalletServer) handleDepositSigFinish(sigFinish interface{}) {
	log.Info("WalletServer handleDepositSigFinish")
}

func (w *WalletServer) initDepositSig() {
	log.Debug("WalletServer initDepositSig")

	// 1. check catching up, self is proposer
	if w.state.GetL2Info().Syncing {
		log.Debug("WalletServer initDepositSig ignore, layer2 is catching up")
		return
	}

	if w.state.GetBtcHead().Syncing {
		log.Debug("WalletServer initDepositSig ignore, btc is catching up")
		return
	}

	w.sigMu.Lock()
	defer w.sigMu.Unlock()

	epochVoter := w.state.GetEpochVoter()
	if epochVoter.Proposer != config.AppConfig.RelayerAddress {
		if w.sigStatus {
			w.sigStatus = false
			// clean process, role changed, remove all status "create", "aggregating"
			w.cleanDepositProcess()
		}
		log.Debugf("WalletServer initDepositSig ignore, self is not proposer, epoch: %d, proposer: %s", epochVoter.Epoch, epochVoter.Proposer)
		return
	}

	// 2. check if there is a sig in progress
	if w.sigStatus {
		log.Debug("WalletServer initDepositSig ignore, there is a sig")
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
		merkleRoot := []byte(deposit.MerkleRoot)
		proofBytes := []byte(deposit.Proof)
		txIndex := uint32(deposit.TxIndex)
		txhash := []byte(deposit.TxHash)
		if bitcointypes.VerifyMerkelProof(txhash, merkleRoot, proofBytes, txIndex) {
			blockHashes = append(blockHashes, deposit.BlockHash)
		}
	}

	if len(blockHashes) == 0 {
		log.Debug("WalletServer initDepositSig ignore, no valid block hash")
		return
	}

	// 5. build sign msg
	l2Info := w.state.GetL2Info()
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

	log.Infof("P2P publish msgSignDeposit success, request id: %s", requestId)
}

func (w *WalletServer) cleanDepositProcess() {
	// TODO remove all status "create", "aggregating"
}
