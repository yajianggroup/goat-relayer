package utxo

import (
	"context"

	"github.com/goatnetwork/goat-relayer/internal/bls"
	"github.com/goatnetwork/goat-relayer/internal/db"
	"github.com/goatnetwork/goat-relayer/internal/state"
	"github.com/goatnetwork/goat-relayer/internal/types"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Withdraw struct
type Withdraw struct {
	state   *state.State
	signer  *bls.Signer
	cacheDb *gorm.DB
	lightDb *gorm.DB

	confirmWithdrawCh chan interface{}
}

// NewWithdraw creates a new Withdraw instance
func NewWithdraw(st *state.State, signer *bls.Signer, dbm *db.DatabaseManager) *Withdraw {
	cacheDb := dbm.GetBtcCacheDB()
	lightDb := dbm.GetBtcLightDB()
	return &Withdraw{
		state:             st,
		signer:            signer,
		cacheDb:           cacheDb,
		lightDb:           lightDb,
		confirmWithdrawCh: make(chan interface{}, 100),
	}
}

// Start starts the withdraw process
func (w *Withdraw) Start(ctx context.Context) {
	w.state.EventBus.Subscribe(state.WithdrawRequest, w.confirmWithdrawCh)
	go w.withdrawLoop(ctx)
	go w.ProcessConfirmedWithdraw(ctx)
}

// withdrawLoop handles unconfirmed withdrawals
func (w *Withdraw) withdrawLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Info("UnConfirm withdraw query stopping...")
			return
		case withdraw := <-w.confirmWithdrawCh:
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
func (w *Withdraw) ProcessConfirmedWithdraw(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Info("Confirmed withdraw query stopping...")
			return
		case withdraw := <-w.confirmWithdrawCh:
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
