package btc

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/goatnetwork/goat-relayer/internal/db"
	"gorm.io/gorm"

	"github.com/btcsuite/btcd/wire"
	log "github.com/sirupsen/logrus"
)

type BTCCache struct {
	db        *gorm.DB
	blockChan chan BlockWithHeight
}

type BlockWithHeight struct {
	Block  *wire.MsgBlock
	Height uint64
}

func NewBTCCache(db *gorm.DB) *BTCCache {
	return &BTCCache{
		db:        db,
		blockChan: make(chan BlockWithHeight, 100),
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

func (bc *BTCCache) cacheBlockData(blockWithHeight BlockWithHeight) {
	block := blockWithHeight.Block
	blockHash := block.BlockHash().String()
	header := block.Header
	difficulty := header.Bits
	randomNumber := header.Nonce
	merkleRoot := header.MerkleRoot.String()
	blockTime := header.Timestamp.Unix()

	headerBuffer := new(bytes.Buffer)
	err := header.Serialize(headerBuffer)
	if err != nil {
		log.Errorf("Failed to serialize block header: %v", err)
		return
	}
	headerBytes := headerBuffer.Bytes()

	txHashes, _ := block.TxHashes()
	txHashesJSON, _ := json.Marshal(txHashes)

	blockData := db.BtcBlockData{
		BlockHeight:  blockWithHeight.Height,
		BlockHash:    blockHash,
		Header:       headerBytes,
		Difficulty:   difficulty,
		RandomNumber: randomNumber,
		MerkleRoot:   merkleRoot,
		BlockTime:    blockTime,
		TxHashes:     string(txHashesJSON),
	}
	bc.db.Save(&blockData)
	for _, tx := range block.Transactions {
		for _, txOut := range tx.TxOut {
			txOutput := db.BtcTXOutput{
				BlockID:  blockData.ID,
				TxHash:   tx.TxID(),
				Value:    uint64(txOut.Value),
				PkScript: txOut.PkScript,
			}
			bc.db.Save(&txOutput)
		}
	}

	log.Debugf("Caching block %s data: header %x, difficulty %d, random number %d, Merkle root %s, block time %d",
		blockHash, headerBytes, difficulty, randomNumber, merkleRoot, blockTime)
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
