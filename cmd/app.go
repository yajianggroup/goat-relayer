package main

import (
	"github.com/goatnetwork/goat-relayer/internal/btc"
	"github.com/goatnetwork/goat-relayer/internal/config"
	"github.com/goatnetwork/goat-relayer/internal/db"
	"github.com/goatnetwork/goat-relayer/internal/http"
	"github.com/goatnetwork/goat-relayer/internal/layer2"
	"github.com/goatnetwork/goat-relayer/internal/p2p"
	"github.com/goatnetwork/goat-relayer/internal/rpc"
	"github.com/goatnetwork/goat-relayer/internal/tss"
)

type Application struct {
	ConfigManager   config.ConfigManager
	DatabaseManager db.DatabaseManager
	Layer2Monitor   layer2.Layer2Monitor
	HTTPServer      http.HTTPServer
	LibP2PService   p2p.LibP2PService
	TSSService      tss.TSSService
	BTCListener     btc.BTCListener
	UTXOService     rpc.UTXOService
}

func NewApplication() *Application {
	return &Application{
		ConfigManager:   &config.ConfigManagerImpl{},
		DatabaseManager: &db.DatabaseManagerImpl{},
		Layer2Monitor:   &layer2.Layer2MonitorImpl{},
		HTTPServer:      &http.HTTPServerImpl{},
		LibP2PService:   &p2p.LibP2PServiceImpl{},
		TSSService:      &tss.TSSServiceImpl{},
		BTCListener:     &btc.BTCListenerImpl{},
		UTXOService:     &rpc.UTXOServiceImpl{},
	}
}

func (app *Application) Run() {
	app.ConfigManager.InitConfig()
	app.DatabaseManager.InitDB()

	if config.AppConfig.EnableRelayer {
		go app.Layer2Monitor.StartLayer2Monitor()
	}
	app.HTTPServer.StartHTTPServer()
	app.LibP2PService.StartLibp2p(app.TSSService, app.HTTPServer)
	app.BTCListener.StartBTCListener(app.LibP2PService, app.DatabaseManager)
	app.UTXOService.StartUTXOService(app.BTCListener)

	select {}
}
