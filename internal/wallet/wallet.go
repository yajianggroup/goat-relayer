package wallet

import (
	"context"

	"github.com/goatnetwork/goat-relayer/internal/db"
	"github.com/goatnetwork/goat-relayer/internal/p2p"
	"github.com/goatnetwork/goat-relayer/internal/state"
	log "github.com/sirupsen/logrus"
)

type WalletServer struct {
	libp2p *p2p.LibP2PService
	db     *db.DatabaseManager
	state  *state.State

	blockCh chan interface{}
}

func NewWalletServer(libp2p *p2p.LibP2PService, st *state.State, db *db.DatabaseManager) *WalletServer {
	return &WalletServer{
		libp2p: libp2p,
		db:     db,
		state:  st,

		blockCh: make(chan interface{}, state.BTC_BLOCK_CHAN_LENGTH),
	}
}

func (w *WalletServer) Start(ctx context.Context) {
	w.state.EventBus.Subscribe(state.BlockScanned, w.blockCh)

	go w.blockScanLoop(ctx)

	log.Info("WalletServer started.")

	<-ctx.Done()
	w.Stop()

	log.Info("WalletServer stopped.")
}

func (w *WalletServer) Stop() {
	close(w.blockCh)
}
