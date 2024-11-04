package wallet

import (
	"fmt"

	"github.com/goatnetwork/goat-relayer/internal/config"
	"github.com/goatnetwork/goat-relayer/internal/state"
	"github.com/goatnetwork/goat-relayer/internal/types"
	log "github.com/sirupsen/logrus"
)

func (w *WalletServer) cancelWithdrawSig() {
	log.Debug("WalletServer cancelWithdrawSig")

	// 1. check catching up, self is proposer
	l2Info := w.state.GetL2Info()
	if l2Info.Syncing {
		log.Infof("WalletServer cancelWithdrawSig ignore, layer2 is catching up")
		return
	}

	btcState := w.state.GetBtcHead()
	if btcState.Syncing {
		log.Infof("WalletServer cancelWithdrawSig ignore, btc is catching up")
		return
	}

	w.sigMu.Lock()
	defer w.sigMu.Unlock()

	epochVoter := w.state.GetEpochVoter()
	if epochVoter.Proposer != config.AppConfig.RelayerAddress {
		// do not clean immediately
		if w.cancelWithdrawStatus && l2Info.Height > epochVoter.Height+1 {
			w.cancelWithdrawStatus = false
			// clean process, role changed, remove all status "create", "aggregating"
		}
		log.Debugf("WalletServer cancelWithdrawSig ignore, self is not proposer, epoch: %d, proposer: %s", epochVoter.Epoch, epochVoter.Proposer)
		return
	}

	// 2. check if there is a sig in progress
	if w.cancelWithdrawStatus {
		log.Debug("WalletServer cancelWithdrawSig ignore, there is cancel in progress")
		return
	}
	if l2Info.Height <= w.cancelWithdrawFinishHeight+2 {
		log.Debug("WalletServer cancelWithdrawSig ignore, last finish cancel in 2 blocks")
		return
	}

	withdraws, err := w.state.GetWithdrawsCanceling()
	if err != nil {
		log.Errorf("WalletServer cancelWithdrawSig GetWithdrawsCanceling error: %v", err)
		return
	}
	if len(withdraws) == 0 {
		log.Infof("WalletServer cancelWithdrawSig no withdraw to cancel")
		return
	}

	withdrawIds := make([]uint64, len(withdraws))
	for i, withdraw := range withdraws {
		withdrawIds[i] = uint64(withdraw.ID)
	}
	requestId := fmt.Sprintf("WITHDRAWCANCEL:%s:%d", config.AppConfig.RelayerAddress, l2Info.Height)
	msgSignCancel := types.MsgSignCancelWithdraw{
		MsgSign: types.MsgSign{
			RequestId:    requestId,
			VoterAddress: epochVoter.Proposer,
		},
		WithdrawIds: withdrawIds,
	}
	w.state.EventBus.Publish(state.SigStart, msgSignCancel)
	w.cancelWithdrawStatus = true
	log.Infof("P2P publish msgSignCancel success, request id: %s", requestId)
}
