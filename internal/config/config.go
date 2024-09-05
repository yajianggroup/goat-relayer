package config

import (
	"crypto/ecdsa"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var AppConfig Config

type ConfigManager interface {
	InitConfig()
}

type ConfigManagerImpl struct{}

func (cm *ConfigManagerImpl) InitConfig() {
	viper.AutomaticEnv()

	// Default config
	viper.SetDefault("HTTP_PORT", "8080")
	viper.SetDefault("RPC_PORT", "50051")
	viper.SetDefault("LIBP2P_PORT", 4001)
	viper.SetDefault("LIBP2P_BOOT_NODES", "")
	viper.SetDefault("BTC_RPC", "http://localhost:8332")
	viper.SetDefault("BTC_START_HEIGHT", 0)
	viper.SetDefault("L2_RPC", "http://localhost:8545")
	viper.SetDefault("L2_START_HEIGHT", 0)
	viper.SetDefault("L2_CONFIRMATIONS", 3)
	viper.SetDefault("L2_MAX_BLOCK_RANGE", 500)
	viper.SetDefault("L2_REQUEST_INTERVAL", "10s")
	viper.SetDefault("L2_PRIVATE_KEY", "")
	viper.SetDefault("L2_CHAIN_ID", "2345")
	viper.SetDefault("ENABLE_WEBHOOK", true)
	viper.SetDefault("ENABLE_RELAYER", true)
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("DB_DIR", "/app/db")
	viper.SetDefault("VOTING_CONTRACT", "")
	viper.SetDefault("WITHDRAW_CONTRACT", "")
	viper.SetDefault("FIREBLOCKS_PUBKEY", "")
	viper.SetDefault("FIREBLOCKS_PRIVKEY", "")

	logLevel, err := logrus.ParseLevel(strings.ToLower(viper.GetString("LOG_LEVEL")))
	if err != nil {
		logrus.Fatalf("Invalid log level: %v", err)
	}

	l2PrivateKey, err := crypto.HexToECDSA(viper.GetString("L2_PRIVATE_KEY"))
	if err != nil {
		logrus.Fatalf("Failed to load l2 private key: %v, given length %d", err, len(viper.GetString("L2_PRIVATE_KEY")))
	}

	l2ChainId, err := strconv.ParseInt(viper.GetString("L2_CHAIN_ID"), 10, 64)
	if err != nil {
		logrus.Fatalf("Failed to parse l2 chain id: %v", err)
	}

	AppConfig = Config{
		HTTPPort:          viper.GetString("HTTP_PORT"),
		RPCPort:           viper.GetString("RPC_PORT"),
		Libp2pPort:        viper.GetInt("LIBP2P_PORT"),
		Libp2pBootNodes:   viper.GetString("LIBP2P_BOOT_NODES"),
		BTCRPC:            viper.GetString("BTC_RPC"),
		BTCStartHeight:    viper.GetInt("BTC_START_HEIGHT"),
		L2RPC:             viper.GetString("L2_RPC"),
		L2StartHeight:     viper.GetInt("L2_START_HEIGHT"),
		L2Confirmations:   viper.GetInt("L2_CONFIRMATIONS"),
		L2MaxBlockRange:   viper.GetInt("L2_MAX_BLOCK_RANGE"),
		L2RequestInterval: viper.GetDuration("L2_REQUEST_INTERVAL"),
		FireblocksPubKey:  viper.GetString("FIREBLOCKS_PUBKEY"),
		FireblocksPrivKey: viper.GetString("FIREBLOCKS_PRIVKEY"),
		EnableWebhook:     viper.GetBool("ENABLE_WEBHOOK"),
		EnableRelayer:     viper.GetBool("ENABLE_RELAYER"),
		DbDir:             viper.GetString("DB_DIR"),
		LogLevel:          logLevel,
		VotingContract:    viper.GetString("VOTING_CONTRACT"),
		WithdrawContract:  viper.GetString("WITHDRAW_CONTRACT"),
		L2PrivateKey:      l2PrivateKey,
		L2ChainId:         big.NewInt(l2ChainId),
	}

	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(AppConfig.LogLevel)
}

type Config struct {
	HTTPPort          string
	RPCPort           string
	Libp2pPort        int
	Libp2pBootNodes   string
	BTCRPC            string
	BTCStartHeight    int
	L2RPC             string
	L2StartHeight     int
	L2Confirmations   int
	L2MaxBlockRange   int
	L2RequestInterval time.Duration
	FireblocksPubKey  string
	FireblocksPrivKey string
	EnableWebhook     bool
	EnableRelayer     bool
	DbDir             string
	LogLevel          logrus.Level
	VotingContract    string
	WithdrawContract  string
	L2PrivateKey      *ecdsa.PrivateKey
	L2ChainId         *big.Int
}
