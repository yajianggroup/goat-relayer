package btc

import (
	"github.com/goatnetwork/goat-relayer/internal/db"
	"github.com/goatnetwork/goat-relayer/internal/p2p"
)

type Option func(*BTCListenerImpl)

func SetLibP2PService(p2psrv p2p.LibP2PService) Option {
	return func(bl *BTCListenerImpl) {
		bl.LibP2PService = p2psrv
	}
}

func SetDatabaseManager(dbm db.DatabaseManager) Option {
	return func(bl *BTCListenerImpl) {
		bl.DatabaseManager = dbm
	}
}
