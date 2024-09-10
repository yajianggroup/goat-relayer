package btc

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/goatnetwork/goat-relayer/internal/db"
	"gorm.io/gorm"
	"time"

	"github.com/btcsuite/btcd/wire"
	log "github.com/sirupsen/logrus"
)

type BTCCache struct {
	db        *gorm.DB
	blockChan chan *wire.MsgBlock
}

func NewBTCCache(db *gorm.DB) *BTCCache {
	return &BTCCache{
		db:        db,
		blockChan: make(chan *wire.MsgBlock, 100),
	}
}

func (bc *BTCCache) Start(ctx context.Context) {
	go bc.processBlocks(ctx)
	go bc.startPeriodicPurge(ctx)
}

func (bc *BTCCache) processBlocks(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Info("Stop processing blocks...")
			return
		case block := <-bc.blockChan:
			bc.cacheBlockData(block)
		}
	}
}

func (bc *BTCCache) startPeriodicPurge(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("Stop purging old cache task...")
			return
		case <-ticker.C:
			bc.purgeOldData()
		}
	}
}

func (bc *BTCCache) cacheBlockData(block *wire.MsgBlock) {
	blockHash := block.BlockHash().String()
	header := block.Header
	difficulty := header.Bits
	randomNumber := header.Nonce
	merkleRoot := header.MerkleRoot.String()
	blockTime := header.Timestamp.Unix()

	headerStr := fmt.Sprintf("Version: %d, PrevBlock: %s, MerkleRoot: %s, Timestamp: %d, Bits: %d, Nonce: %d",
		header.Version, header.PrevBlock, header.MerkleRoot, header.Timestamp.Unix(), header.Bits, header.Nonce)

	difficultyLE := make([]byte, 4)
	randomNumberLE := make([]byte, 4)
	blockTimeLE := make([]byte, 8)

	binary.LittleEndian.PutUint32(difficultyLE, difficulty)
	binary.LittleEndian.PutUint32(randomNumberLE, randomNumber)
	binary.LittleEndian.PutUint64(blockTimeLE, uint64(blockTime))

	txHashes, _ := block.TxHashes()
	txHashesJSON, _ := json.Marshal(txHashes)

	blockData := db.BtcBlockData{
		BlockHash:    blockHash,
		Header:       headerStr,
		Difficulty:   difficultyLE,
		RandomNumber: randomNumberLE,
		MerkleRoot:   merkleRoot,
		BlockTime:    blockTimeLE,
		TxHashes:     string(txHashesJSON),
	}
	bc.db.Save(&blockData)
	for _, tx := range block.Transactions {
		for _, txOut := range tx.TxOut {
			txOutput := db.BtcTXOutput{
				BlockID:  blockData.ID,
				TxHash:   tx.TxHash().String(),
				Value:    uint64(txOut.Value),
				PkScript: txOut.PkScript,
			}
			bc.db.Save(&txOutput)
		}
	}

	log.Printf("Caching block %s data: header %s, difficulty %d, random number %d, Merkle root %s, block time %d",
		blockHash, headerStr, difficulty, randomNumber, merkleRoot, blockTime)
}

func (c *BTCCache) purgeOldData() {
	thresholdTime := time.Now().AddDate(0, 0, -3).Unix()

	var blockDatas []db.BtcBlockData
	if err := c.db.Where("block_time < ?", thresholdTime).Find(&blockDatas).Error; err != nil {
		log.Printf("Error querying block data: %v", err)
		return
	}

	for _, blockData := range blockDatas {
		if err := c.db.Delete(&blockData).Error; err != nil {
			log.Printf("Error deleting block data: %v", err)
		} else {
			log.Printf("Deleted block with block hash: %s", blockData.BlockHash)
		}
	}
}
