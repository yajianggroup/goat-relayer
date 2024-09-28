package wallet

import (
	"context"
	"time"

	"github.com/goatnetwork/goat-relayer/internal/config"
	"github.com/goatnetwork/goat-relayer/internal/state"
	log "github.com/sirupsen/logrus"
)

func (w *WalletServer) withdrawLoop(ctx context.Context) {
	w.state.EventBus.Subscribe(state.SigFailed, w.withdrawSigFailChan)
	w.state.EventBus.Subscribe(state.SigFinish, w.withdrawSigFinishChan)
	w.state.EventBus.Subscribe(state.SigTimeout, w.withdrawSigTimeoutChan)

	// init status process, if restart && layer2 status is up to date, remove all status "create", "aggregating"
	if !w.state.GetBtcHead().Syncing {
		w.cleanWithdrawProcess()
	}

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case sigFail := <-w.withdrawSigFailChan:
			w.handleWithdrawSigFailed(sigFail, "failed")
		case sigTimeout := <-w.withdrawSigTimeoutChan:
			w.handleWithdrawSigFailed(sigTimeout, "timeout")
		case sigFinish := <-w.withdrawSigFinishChan:
			w.handleWithdrawSigFinish(sigFinish)
		case <-ticker.C:
			w.initWithdrawSig()
		}
	}
}

func (w *WalletServer) handleWithdrawSigFailed(sigFail interface{}, reason string) {
	log.Infof("WalletServer handleWithdrawSigFailed, reason: %s", reason)
}

func (w *WalletServer) handleWithdrawSigFinish(sigFinish interface{}) {
	log.Info("WalletServer handleWithdrawSigFinish")
}

func (w *WalletServer) initWithdrawSig() {
	log.Debug("WalletServer initWithdrawSig")

	// 1. check catching up, self is proposer
	if w.state.GetL2Info().Syncing {
		log.Debug("WalletServer initWithdrawSig ignore, layer2 is catching up")
		return
	}

	if w.state.GetBtcHead().Syncing {
		log.Debug("WalletServer initWithdrawSig ignore, btc is catching up")
		return
	}

	w.sigMu.Lock()
	defer w.sigMu.Unlock()

	epochVoter := w.state.GetEpochVoter()
	if epochVoter.Proposer != config.AppConfig.RelayerAddress {
		if w.sigStatus {
			w.sigStatus = false
			// clean process, role changed, remove all status "create", "aggregating"
			w.cleanWithdrawProcess()
		}
		log.Debugf("WalletServer initWithdrawSig ignore, self is not proposer, epoch: %d, proposer: %s", epochVoter.Epoch, epochVoter.Proposer)
		return
	}

	// 2. check if there is a sig in progress
	if w.sigStatus {
		log.Debug("WalletServer initWithdrawSig ignore, there is a sig")
		return
	}
	// clean process, become proposer again, remove all status "create", "aggregating"
	w.cleanWithdrawProcess()

	// 3. check queue send order in db,
	// has status 'aggregating', start bls again [6]
	// there is no status 'aggregating', go to 4

	// 4. query withraw list from db, status 'create'
	// if count > 150, built soon
	// else if count > 50, check oldest one, if than 2 hours (optional), built
	// else if check oldest one, if status 'pending', skip
	// else go to 5

	// 5. do consolidation

	// 6. start bls sig

	// w.sigStatus should update to false after layer2 InitalWithdraw callback
}

func (w *WalletServer) cleanWithdrawProcess() {
	// TODO remove all status "create", "aggregating"
}
