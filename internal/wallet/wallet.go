package wallet

import (
	"context"

	"github.com/goatnetwork/goat-relayer/internal/bls"
	"github.com/goatnetwork/goat-relayer/internal/p2p"
	"github.com/goatnetwork/goat-relayer/internal/state"
	log "github.com/sirupsen/logrus"
)

type WalletServer struct {
	libp2p *p2p.LibP2PService
	state  *state.State
	signer *bls.Signer

	depositCh  chan interface{}
	withdrawCh chan interface{}
	blockCh    chan interface{}
}

func NewWalletServer(libp2p *p2p.LibP2PService, st *state.State, signer *bls.Signer) *WalletServer {

	return &WalletServer{
		libp2p:     libp2p,
		state:      st,
		signer:     signer,
		depositCh:  make(chan interface{}, 100),
		withdrawCh: make(chan interface{}, 100),
		blockCh:    make(chan interface{}, state.BTC_BLOCK_CHAN_LENGTH),
	}
}

func (w *WalletServer) Start(ctx context.Context) {
	w.state.EventBus.Subscribe(state.BlockScanned, w.blockCh)

	go w.blockScanLoop(ctx)
	go w.depositLoop(ctx)
	go w.withdrawLoop(ctx)
	go w.processConfirmedDeposit(ctx)
	go w.processConfirmedWithdraw(ctx)

	log.Info("WalletServer started.")

	<-ctx.Done()
	w.Stop()

	log.Info("WalletServer stopped.")
}

func (w *WalletServer) Stop() {
	close(w.blockCh)
}
