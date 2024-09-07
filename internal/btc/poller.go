package btc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
)

type BTCPoller struct {
	db          *leveldb.DB
	confirmChan chan *wire.MsgBlock
}

func NewBTCPoller(db *leveldb.DB) *BTCPoller {
	return &BTCPoller{
		db:          db,
		confirmChan: make(chan *wire.MsgBlock),
	}
}

func (p *BTCPoller) Start(ctx context.Context) {
	go p.pollLoop(ctx)
}

func (p *BTCPoller) Stop() {
	p.db.Close()
}
func (p *BTCPoller) pollLoop(ctx context.Context) {
	for {
		select {
		case block := <-p.confirmChan:
			p.handleConfirmedBlock(block)
		case <-ctx.Done():
			log.Info("Stopping the polling of confirmed blocks...")
			return
		}
	}
}

func (p *BTCPoller) GetBlockHashForTx(txHash chainhash.Hash) (*chainhash.Hash, error) {
	iter := p.db.NewIterator(nil, nil)
	defer iter.Release()

	for iter.Next() {
		key := iter.Key()
		if bytes.HasPrefix(key, []byte("utxo:")) {
			value := iter.Value()
			if bytes.Equal(value, txHash[:]) {
				blockHashBytes := key[len("utxo:") : len("utxo:")+64]
				return chainhash.NewHash(blockHashBytes)
			}
		}
	}
	return nil, fmt.Errorf("failed to find the block hash for the transaction")
}

func (p *BTCPoller) GetBlockHeader(blockHash *chainhash.Hash) (*wire.BlockHeader, error) {
	headerBytes, err := p.db.Get([]byte("header:"+blockHash.String()), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve block header from database: %v", err)
	}
	var header wire.BlockHeader
	err = header.Deserialize(bytes.NewReader(headerBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize block header: %v", err)
	}
	return &header, nil
}

func (p *BTCPoller) GetTxHashes(blockHash *chainhash.Hash) ([]chainhash.Hash, error) {
	txHashesBytes, err := p.db.Get([]byte("txhashes:"+blockHash.String()), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve transaction hash list from database: %v", err)
	}
	var txHashes []chainhash.Hash
	err = json.Unmarshal(txHashesBytes, &txHashes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction hash list: %v", err)
	}
	return txHashes, nil
}

func (p *BTCPoller) GetBlock(height uint64) (*wire.MsgBlock, error) {
	blockBytes, err := p.db.Get([]byte(fmt.Sprintf("block:%d", height)), nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving block from database: %v", err)
	}

	var block wire.MsgBlock
	err = block.Deserialize(bytes.NewReader(blockBytes))
	if err != nil {
		return nil, fmt.Errorf("error deserializing block: %v", err)
	}

	return &block, nil
}

func (p *BTCPoller) handleConfirmedBlock(block *wire.MsgBlock) {
	// Logic for handling confirmed blocks
	blockHash := block.BlockHash()
	log.Infof("Handling confirmed block: %s", blockHash.String())

	// TODO: Submit the confirmed block to the consensus layer

}
