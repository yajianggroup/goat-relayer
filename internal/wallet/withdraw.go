package wallet

import (
	"context"

	"github.com/goatnetwork/goat-relayer/internal/types"
	log "github.com/sirupsen/logrus"
)

// withdrawLoop handles unconfirmed withdrawals
func (w *WalletServer) withdrawLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Info("UnConfirm withdraw query stopping...")
			return
		case withdraw := <-w.withdrawCh:
			withdrawData, ok := withdraw.(types.MsgUtxoWithdraw)
			if !ok {
				log.Errorf("Invalid withdraw data type")
				continue
			}

			// TODO: get btc address from evm address
			err := w.state.CreateWithdrawal(withdrawData.TxId, withdrawData.EvmAddr, "", 0, 0)
			if err != nil {
				log.Errorf("Failed to add unconfirmed withdraw: %v", err)
				continue
			}
			// Process the withdraw
			// TODO: Add logic to process the withdraw
		}
	}
}

// ProcessConfirmedWithdraw processes confirmed withdrawals
func (w *WalletServer) processConfirmedWithdraw(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Info("Confirmed withdraw query stopping...")
			return
		case withdraw := <-w.withdrawCh:
			withdrawData, ok := withdraw.(types.MsgUtxoWithdraw)
			if !ok {
				log.Errorf("Invalid withdraw data type")
				continue
			}
			err := w.state.SaveConfirmWithdraw(withdrawData.TxId)
			if err != nil {
				log.Errorf("Failed to save confirmed withdraw: %v", err)
				continue
			}
			// Process the confirmed withdraw
			// TODO: Add logic to process the confirmed withdraw
		}
	}
}
