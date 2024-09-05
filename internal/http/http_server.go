package http

import (
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/goatnetwork/goat-relayer/internal/config"
)

type HTTPServer interface {
	StartHTTPServer()
}

type HTTPServerImpl struct{}

func (hs *HTTPServerImpl) StartHTTPServer() {
	r := gin.Default()

	r.GET("/api/v1/helloworld", handleHelloWorld)

	if config.AppConfig.EnableWebhook {
		r.POST("/api/fireblocks/webhook", handleFireblocksWebhook)
	}
	if config.AppConfig.EnableRelayer {
		r.POST("/api/fireblocks/cosigner/v2/tx_sign_request", handleFireblocksCosignerTxSign)
	}

	// Use configuration port
	addr := ":" + config.AppConfig.HTTPPort
	log.Infof("HTTP server is running on port %s", config.AppConfig.HTTPPort)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

// a demo handler
func handleHelloWorld(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "data": "hello world."})
}
