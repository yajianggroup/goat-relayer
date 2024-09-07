package http

import (
	"context"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/goatnetwork/goat-relayer/internal/config"
	"github.com/goatnetwork/goat-relayer/internal/db"
	"github.com/goatnetwork/goat-relayer/internal/p2p"
)

type HTTPServer struct {
	libp2p *p2p.LibP2PService
	db     *db.DatabaseManager
}

func NewHTTPServer(libp2p *p2p.LibP2PService, db *db.DatabaseManager) *HTTPServer {
	return &HTTPServer{
		libp2p: libp2p,
		db:     db,
	}
}

func (s *HTTPServer) Start(ctx context.Context) {
	r := gin.Default()

	if gin.IsDebugging() {
		r.GET("/api/v1/helloworld", s.handleHelloWorld)
	}
	if config.AppConfig.EnableWebhook {
		r.POST("/api/fireblocks/webhook", s.handleFireblocksWebhook)
	}
	if config.AppConfig.EnableRelayer {
		r.POST("/api/fireblocks/cosigner/v2/tx_sign_request", s.handleFireblocksCosignerTxSign)
	}

	// Use configuration port
	addr := ":" + config.AppConfig.HTTPPort
	log.Infof("HTTP server is running on port %s", config.AppConfig.HTTPPort)

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	<-ctx.Done()

	log.Info("HTTP server is stopping...")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Fatalf("HTTP server forced to shutdown: %v", err)
	}

	log.Info("HTTP server has stopped.")
}

// a demo handler
func (s *HTTPServer) handleHelloWorld(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "data": "hello world."})
}
