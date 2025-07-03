package p2p

import (
	"context"
	"crypto/rand"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/libp2p/go-libp2p/core/crypto"
	log "github.com/sirupsen/logrus"

	"github.com/goatnetwork/goat-relayer/internal/config"
	"github.com/goatnetwork/goat-relayer/internal/state"
)

const (
	privKeyFile = "node_private_key.pem"
)

var libp2pNetwork *Network

type LibP2PService struct {
	state   *state.State
	network *Network
}

func NewLibP2PService(state *state.State) *LibP2PService {
	network, err := NewNetwork(state)
	if err != nil {
		log.Fatalf("Failed to create libp2p network: %v", err)
	}
	libp2pNetwork = network
	return &LibP2PService{
		state:   state,
		network: network,
	}
}

func (lp *LibP2PService) Start(ctx context.Context) {
	log.Info("Starting LibP2PService...")
	lp.network.Start()

	<-ctx.Done()

	log.Info("LibP2PService is stopping...")
	lp.network.Close()
	log.Info("LibP2PService has stopped.")
}

func loadOrCreatePrivateKey(fileName string) (crypto.PrivKey, error) {
	dbDir := config.AppConfig.DbDir
	if err := os.MkdirAll(dbDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create database directory: %v", err)
	}

	pemPath := filepath.Join(dbDir, fileName)
	if _, err := os.Stat(pemPath); err == nil {
		privKeyBytes, err := ioutil.ReadFile(pemPath)
		if err != nil {
			return nil, err
		}
		privKey, err := crypto.UnmarshalPrivateKey(privKeyBytes)
		if err != nil {
			return nil, err
		}
		return privKey, nil
	}

	privKey, _, err := crypto.GenerateKeyPairWithReader(crypto.Ed25519, 2048, rand.Reader)
	if err != nil {
		return nil, err
	}

	privKeyBytes, err := crypto.MarshalPrivateKey(privKey)
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(pemPath, privKeyBytes, 0600); err != nil {
		return nil, err
	}

	return privKey, nil
}
