package wallet

import (
	"encoding/json"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/goatnetwork/goat-relayer/internal/config"
	"github.com/goatnetwork/goat-relayer/internal/state"
	"github.com/goatnetwork/goat-relayer/internal/types"
	log "github.com/sirupsen/logrus"
)

func (w *WalletServer) finalizeWithdrawSig() {
	log.Debug("WalletServer finalizeWithdrawSig")

	// 1. check catching up, self is proposer
	l2Info := w.state.GetL2Info()
	if l2Info.Syncing {
		log.Infof("WalletServer finalizeWithdrawSig ignore, layer2 is catching up")
		return
	}

	btcState := w.state.GetBtcHead()
	if btcState.Syncing {
		log.Infof("WalletServer finalizeWithdrawSig ignore, btc is catching up")
		return
	}
	if btcState.NetworkFee > 500 {
		log.Infof("WalletServer finalizeWithdrawSig ignore, btc network fee too high: %d", btcState.NetworkFee)
		return
	}

	w.sigMu.Lock()
	defer w.sigMu.Unlock()

	epochVoter := w.state.GetEpochVoter()
	if epochVoter.Proposer != config.AppConfig.RelayerAddress {
		// do not clean immediately
		if w.sigStatus && l2Info.Height > epochVoter.Height+1 {
			w.sigStatus = false
			// clean process, role changed, remove all status "create", "aggregating"
			w.cleanWithdrawProcess()
		}
		log.Debugf("WalletServer finalizeWithdrawSig ignore, self is not proposer, epoch: %d, proposer: %s", epochVoter.Epoch, epochVoter.Proposer)
		return
	}

	// 2. check if there is a sig in progress
	if !w.sigStatus {
		log.Debug("WalletServer finalizeWithdrawSig ignore, there is no sig")
		return
	}
	if l2Info.Height <= w.sigFinishHeight+2 {
		log.Debug("WalletServer finalizeWithdrawSig ignore, last finish sig in 2 blocks")
		return
	}

	sendOrders, err := w.state.GetSendOrderProcessed(1)
	if err != nil {
		log.Errorf("WalletServer finalizeWithdrawSig error: %v", err)
		return
	}
	if len(sendOrders) == 0 {
		log.Debug("WalletServer finalizeWithdrawSig ignore, no withdraw for finalize")
		return
	}

	// assemble msg tx
	for _, sendOrder := range sendOrders {
		log.Debugf("WalletServer finalizeWithdrawSig sendOrder: %+v", sendOrder)

		btcBlockData, err := w.state.QueryBtcBlockDataByHeight(sendOrder.BtcBlock)
		if err != nil {
			log.Errorf("WalletServer finalizeWithdrawSig query btc block data by height error: %v", err)
			return
		}
		txhash, err := chainhash.NewHashFromStr(sendOrder.Txid)
		if err != nil {
			log.Errorf("WalletServer finalizeWithdrawSig new hash from str error: %v", err)
			return
		}
		txHashes := make([]string, 0)
		err = json.Unmarshal([]byte(btcBlockData.TxHashes), &txHashes)
		if err != nil {
			return
		}
		_, proof, txIndex, err := types.GenerateSPVProof(sendOrder.Txid, txHashes)
		if err != nil {
			log.Errorf("WalletServer finalizeWithdrawSig generate spv proof error: %v", err)
			return
		}
		msgSignFinalize := types.MsgSignFinalizeWithdraw{
			Txid:              txhash[:],
			BlockNumber:       uint64(sendOrder.BtcBlock),
			TxIndex:           uint32(txIndex),
			IntermediateProof: proof,
			BlockHeader:       btcBlockData.Header,
		}
		w.state.EventBus.Publish(state.WithdrawFinalize, msgSignFinalize)
	}

}
