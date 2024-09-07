package btc

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/wire"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
)

type BTCCache struct {
	db        *leveldb.DB
	blockChan chan *wire.MsgBlock
}

func NewBTCCache(db *leveldb.DB) *BTCCache {
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

	bc.db.Put([]byte("header:"+blockHash), []byte(headerStr), nil)
	bc.db.Put([]byte("difficulty:"+blockHash), difficultyLE, nil)
	bc.db.Put([]byte("random:"+blockHash), randomNumberLE, nil)
	bc.db.Put([]byte("merkleroot:"+blockHash), []byte(merkleRoot), nil)
	bc.db.Put([]byte("blocktime:"+blockHash), blockTimeLE, nil)
	bc.db.Put([]byte("blockhash:"+blockHash), []byte(blockHash), nil)

	for _, tx := range block.Transactions {
		for _, txOut := range tx.TxOut {
			utxoKey := fmt.Sprintf("utxo:%s:%d", blockHash, txOut.Value)
			bc.db.Put([]byte(utxoKey), txOut.PkScript, nil)
		}
	}

	log.Printf("Caching block %s data: header %s, difficulty %d, random number %d, Merkle root %s, block time %d",
		blockHash, headerStr, difficulty, randomNumber, merkleRoot, blockTime)
}

func (c *BTCCache) purgeOldData() {
	thresholdTime := time.Now().AddDate(0, 0, -3).Unix()

	iter := c.db.NewIterator(nil, nil)
	defer iter.Release()

	for iter.Next() {
		key := iter.Key()
		blockTimeBytes, err := c.db.Get([]byte("blocktime:"+string(key[len("blockhash:"):])), nil)
		if err != nil {
			log.Printf("Error getting block time: %v", err)
			continue
		}

		blockTime := int64(binary.LittleEndian.Uint64(blockTimeBytes))
		if blockTime < thresholdTime {
			c.db.Delete(key, nil)
			log.Printf("Deleted block with key: %s", key)
		}
	}
	if err := iter.Error(); err != nil {
		log.Printf("Error during data purge: %v", err)
	}
}
