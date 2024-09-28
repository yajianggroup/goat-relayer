package wallet

import (
	"context"
	"sync"

	"github.com/goatnetwork/goat-relayer/internal/bls"
	"github.com/goatnetwork/goat-relayer/internal/p2p"
	"github.com/goatnetwork/goat-relayer/internal/state"
	log "github.com/sirupsen/logrus"
)

type WalletServer struct {
	libp2p    *p2p.LibP2PService
	state     *state.State
	signer    *bls.Signer
	once      sync.Once
	sigMu     sync.Mutex
	sigStatus bool

	depositCh chan interface{}
	blockCh   chan interface{}

	withdrawSigFailChan    chan interface{}
	withdrawSigFinishChan  chan interface{}
	withdrawSigTimeoutChan chan interface{}
}

func NewWalletServer(libp2p *p2p.LibP2PService, st *state.State, signer *bls.Signer) *WalletServer {

	return &WalletServer{
		libp2p:    libp2p,
		state:     st,
		signer:    signer,
		depositCh: make(chan interface{}, 100),
		blockCh:   make(chan interface{}, state.BTC_BLOCK_CHAN_LENGTH),

		withdrawSigFailChan:    make(chan interface{}, 10),
		withdrawSigFinishChan:  make(chan interface{}, 10),
		withdrawSigTimeoutChan: make(chan interface{}, 10),
	}
}

func (w *WalletServer) Start(ctx context.Context) {
	w.state.EventBus.Subscribe(state.BlockScanned, w.blockCh)

	go w.blockScanLoop(ctx)
	go w.depositLoop(ctx)
	go w.processConfirmedDeposit(ctx)

	go w.withdrawLoop(ctx)

	log.Info("WalletServer started.")

	<-ctx.Done()
	w.Stop()

	log.Info("WalletServer stopped.")
}

func (w *WalletServer) Stop() {
	w.once.Do(func() {
		close(w.blockCh)
		close(w.depositCh)

		close(w.withdrawSigFailChan)
		close(w.withdrawSigFinishChan)
		close(w.withdrawSigTimeoutChan)
	})
}
