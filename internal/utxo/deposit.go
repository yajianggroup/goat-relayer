package utxo

import (
	"context"
	"github.com/goatnetwork/goat-relayer/internal/db"
	"github.com/goatnetwork/goat-relayer/internal/state"
	"gorm.io/gorm"
)

// Deposit struct
type Deposit struct {
	state   *state.State
	cacheDb *gorm.DB
	lightDb *gorm.DB
}

func NewDeposit(state *state.State, dbm *db.DatabaseManager) *Deposit {
	cacheDb := dbm.GetBtcCacheDB()
	lightDb := dbm.GetBtcLightDB()
	return &Deposit{
		state:   state,
		cacheDb: cacheDb,
		lightDb: lightDb,
	}
}

func (d *Deposit) Start(ctx context.Context) {
	go d.QueryUnconfirmedDeposit(ctx)
	go d.ProcessOnceConfirmedDeposit(ctx)
	go d.ProcessSixConfirmedDeposit(ctx)
}
