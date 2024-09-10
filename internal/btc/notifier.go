package btc

import (
	"context"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/goatnetwork/goat-relayer/internal/config"
	log "github.com/sirupsen/logrus"
)

type BTCNotifier struct {
	client        *rpcclient.Client
	currentHeight uint64

	cache  *BTCCache
	poller *BTCPoller
}

func NewBTCNotifier(client *rpcclient.Client, cache *BTCCache, poller *BTCPoller) *BTCNotifier {
	return &BTCNotifier{
		client:        client,
		currentHeight: uint64(config.AppConfig.BTCStartHeight),
		cache:         cache,
		poller:        poller,
	}
}

func (bn *BTCNotifier) Start(ctx context.Context) {
	go bn.listenForBTCBlocks(ctx)
	go bn.checkConfirmations(ctx)
}

func (bn *BTCNotifier) listenForBTCBlocks(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Info("Stopping BTC block listening...")
			return
		default:

			bn.currentHeight++

			bestHeight, err := bn.client.GetBlockCount()
			if err != nil {
				log.Printf("Error getting the latest block height: %v", err)
				continue
			}

			if int64(bn.currentHeight) > bestHeight {
				log.Printf("Reached the latest block, waiting for new blocks...")
				bn.currentHeight = uint64(bestHeight)
				continue
			}

			blockHash, err := bn.client.GetBlockHash(int64(bn.currentHeight))
			if err != nil {
				log.Printf("Error getting block hash: %v", err)
				continue
			}

			block, err := bn.client.GetBlock(blockHash)
			if err != nil {
				log.Printf("Error getting block: %v", err)
				continue
			}

			bn.cache.blockChan <- block
			log.Printf("Starting to cache block at height %d", bn.currentHeight)
		}
	}
}

func (bn *BTCNotifier) checkConfirmations(ctx context.Context) {
	checkInterval := 10 * time.Minute
	requiredConfirmations := uint64(6) // Can be adjusted as needed

	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("Stopping confirmation checks...")
			return
		case <-ticker.C:
			bestHeight, err := bn.client.GetBlockCount()
			if err != nil {
				log.Printf("Error getting the latest block height: %v", err)
				continue
			}

			confirmedHeight := uint64(bestHeight) - requiredConfirmations + 1
			for height := bn.currentHeight - requiredConfirmations; height <= confirmedHeight; height++ {
				block, err := bn.poller.GetBlock(height)
				if err != nil {
					log.Printf("Error getting block at height %d from cache: %v", height, err)
					continue
				}

				bn.poller.confirmChan <- block
				log.Infof("Starting to submit block at height %d to consensus", height)
			}
		}
	}
}
