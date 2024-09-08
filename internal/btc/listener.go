package btc

import (
	"context"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/goatnetwork/goat-relayer/internal/config"
	"github.com/goatnetwork/goat-relayer/internal/db"
	"github.com/goatnetwork/goat-relayer/internal/p2p"
	log "github.com/sirupsen/logrus"
)

type BTCListener struct {
	libp2p *p2p.LibP2PService
	dbm    *db.DatabaseManager

	notifier *BTCNotifier
}

func NewBTCListener(libp2p *p2p.LibP2PService, dbm *db.DatabaseManager) *BTCListener {
	db := dbm.GetCacheDB()
	cache := NewBTCCache(db)
	poller := NewBTCPoller(db)

	connConfig := &rpcclient.ConnConfig{
		Host:         config.AppConfig.BTCRPC,
		HTTPPostMode: true,
		DisableTLS:   true,
	}
	client, err := rpcclient.New(connConfig, nil)
	if err != nil {
		log.Fatalf("failed to start bitcoin client: %v", err)
	}
	notifier := NewBTCNotifier(client, cache, poller)

	return &BTCListener{
		libp2p:   libp2p,
		dbm:      dbm,
		notifier: notifier,
	}
}

func (bl *BTCListener) Start(ctx context.Context) {
	go bl.notifier.Start(ctx)

	log.Info("BTCListener started all modules")

	<-ctx.Done()
	log.Info("BTCListener is stopping...")
}
