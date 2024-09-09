package layer2

import (
	"context"
	"math/big"
	"strings"
	"time"

	"github.com/go-errors/errors"
	log "github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/goatnetwork/goat-relayer/internal/config"
	"github.com/goatnetwork/goat-relayer/internal/db"
	"github.com/goatnetwork/goat-relayer/internal/layer2/abis"
	"github.com/goatnetwork/goat-relayer/internal/p2p"
	"github.com/goatnetwork/goat-relayer/internal/state"
	"gorm.io/gorm"
)

// TODO cosmos client
type Layer2Listener struct {
	libp2p    *p2p.LibP2PService
	db        *db.DatabaseManager
	state     *state.State
	ethClient *ethclient.Client

	contractBitcoin *abis.BitcoinContract
	contractBridge  *abis.BridgeContract
	contractRelayer *abis.RelayerContract
}

func NewLayer2Listener(libp2p *p2p.LibP2PService, state *state.State, db *db.DatabaseManager) *Layer2Listener {
	ethClient, err := DialEthClient()
	if err != nil {
		log.Fatalf("Error creating Layer2 EVM RPC client: %v", err)
	}

	contractRelayer, err := abis.NewRelayerContract(abis.RelayerAddress, ethClient)
	if err != nil {
		log.Fatalf("Failed to instantiate contract relayer: %v", err)
	}
	contractBitcoin, err := abis.NewBitcoinContract(abis.BitcoinAddress, ethClient)
	if err != nil {
		log.Fatalf("Failed to instantiate contract bitcoin: %v", err)
	}
	contractBridge, err := abis.NewBridgeContract(abis.BridgeAddress, ethClient)
	if err != nil {
		log.Fatalf("Failed to instantiate contract bridge: %v", err)
	}

	return &Layer2Listener{
		libp2p:    libp2p,
		db:        db,
		state:     state,
		ethClient: ethClient,

		contractBitcoin: contractBitcoin,
		contractBridge:  contractBridge,
		contractRelayer: contractRelayer,
	}
}

func DialEthClient() (*ethclient.Client, error) {
	var opts []rpc.ClientOption

	if config.AppConfig.L2JwtSecret != "" {
		jwtSecret := common.FromHex(strings.TrimSpace(config.AppConfig.L2JwtSecret))
		if len(jwtSecret) != 32 {
			return nil, errors.New("jwt secret is not a 32 bytes hex string")
		}
		var jwtKey [32]byte
		copy(jwtKey[:], jwtSecret)
		opts = append(opts, rpc.WithHTTPAuth(node.NewJWTAuth(jwtKey)))
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	// Dial the Ethereum node with optional JWT authentication
	client, err := rpc.DialOptions(ctx, config.AppConfig.L2RPC, opts...)
	if err != nil {
		return nil, err
	}
	return ethclient.NewClient(client), nil
}

func (lis *Layer2Listener) Start(ctx context.Context) {
	// Get latest sync height
	var syncStatus db.L2SyncStatus
	db := lis.db.GetL2SyncDB()
	result := db.First(&syncStatus)
	if result.Error != nil && result.Error == gorm.ErrRecordNotFound {
		syncStatus.LastSyncBlock = uint64(config.AppConfig.L2StartHeight)
		syncStatus.UpdatedAt = time.Now()
		db.Create(&syncStatus)
	}

	l2RequestInterval := config.AppConfig.L2RequestInterval
	l2Confirmations := uint64(config.AppConfig.L2Confirmations)
	l2MaxBlockRange := uint64(config.AppConfig.L2MaxBlockRange)
	clientTimeout := time.Second * 10

	for {
		select {
		case <-ctx.Done():
			log.Info("Layer2Listener stoping...")
			return
		default:
			ctx1, cancel1 := context.WithTimeout(ctx, clientTimeout)
			latestBlock, err := lis.ethClient.BlockNumber(ctx1)
			cancel1()
			if err != nil {
				log.Errorf("Error getting latest block number: %v", err)
				time.Sleep(l2RequestInterval)
				continue
			}

			targetBlock := latestBlock - l2Confirmations
			if syncStatus.LastSyncBlock < targetBlock {
				fromBlock := syncStatus.LastSyncBlock + 1
				toBlock := min(fromBlock+l2MaxBlockRange-1, targetBlock)

				log.WithFields(log.Fields{
					"fromBlock": fromBlock,
					"toBlock":   toBlock,
				}).Info("Syncing L2 goat events")

				filterQuery := ethereum.FilterQuery{
					FromBlock: big.NewInt(int64(fromBlock)),
					ToBlock:   big.NewInt(int64(toBlock)),
					Addresses: []common.Address{abis.BridgeAddress, abis.BitcoinAddress, abis.RelayerAddress},
				}

				ctx2, cancel2 := context.WithTimeout(ctx, clientTimeout)
				logs, err := lis.ethClient.FilterLogs(ctx2, filterQuery)
				cancel2()
				if err != nil {
					log.Errorf("Failed to filter logs: %v", err)
					time.Sleep(l2RequestInterval)
					continue
				}

				for _, vLog := range logs {
					lis.processGoatLogs(vLog)
					if syncStatus.LastSyncBlock < vLog.BlockNumber {
						syncStatus.LastSyncBlock = vLog.BlockNumber
					}
				}

				// Save sync status
				syncStatus.UpdatedAt = time.Now()
				db.Save(&syncStatus)
			} else {
				log.Debugf("Sync to tip, target block: %d", targetBlock)
			}

			time.Sleep(l2RequestInterval)
		}
	}
}

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}
