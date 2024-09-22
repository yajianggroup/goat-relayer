package utxo

import (
	"context"

	"github.com/goatnetwork/goat-relayer/internal/bls"

	"github.com/goatnetwork/goat-relayer/internal/db"
	"github.com/goatnetwork/goat-relayer/internal/state"
	"gorm.io/gorm"
)

// Deposit struct
type Deposit struct {
	state   *state.State
	signer  *bls.Signer
	cacheDb *gorm.DB
	lightDb *gorm.DB

	confirmDepositCh chan interface{}
}

func NewDeposit(state *state.State, signer *bls.Signer, dbm *db.DatabaseManager) *Deposit {
	cacheDb := dbm.GetBtcCacheDB()
	lightDb := dbm.GetBtcLightDB()
	return &Deposit{
		state:   state,
		signer:  signer,
		cacheDb: cacheDb,
		lightDb: lightDb,
	}
}

func (d *Deposit) Start(ctx context.Context) {
	d.state.EventBus.Subscribe(state.DepositReceive, d.confirmDepositCh)
	// TODO wait for btc & goat chain complete blocks sync
	go d.AddUnconfirmedDeposit(ctx)
	go d.ProcessConfirmedDeposit(ctx)
}
