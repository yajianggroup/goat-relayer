package btc

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/goatnetwork/goat-relayer/internal/config"
	"github.com/goatnetwork/goat-relayer/internal/db"
	"github.com/goatnetwork/goat-relayer/internal/p2p"
	"github.com/syndtr/goleveldb/leveldb"
)

type BTCListener interface {
	StartBTCListener(p2p.LibP2PService, db.DatabaseManager)
}

type BTCListenerImpl struct {
	p2p.LibP2PService
	db.DatabaseManager
}

func (bl *BTCListenerImpl) StartBTCListener(libp2p p2p.LibP2PService, dbman db.DatabaseManager) {
	connConfig := &rpcclient.ConnConfig{
		Host:         config.AppConfig.BTCRPC, // RPC server address
		HTTPPostMode: true,                    // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true,                    // Disable TLS for simplicity (only if not using TLS)
	}

	client, err := rpcclient.New(connConfig, nil)
	if err != nil {
		log.Fatalf("Failed to start Bitcoin client: %v", err)
	}
	defer client.Shutdown()

	// Open or create the local storage (LevelDB)
	dbPath := filepath.Join(config.AppConfig.DbDir, "btc_cache.db")
	db, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		log.Fatalf("Failed to open local storage: %v", err)
	}
	defer db.Close()

	listenAndCacheBTCBlocks(client, db)
	go func() {
		for range time.NewTicker(24 * time.Hour).C {
			purgeOldData(db)
		}
	}()
}

func listenAndCacheBTCBlocks(client *rpcclient.Client, db *leveldb.DB) {
	currentHeight := config.AppConfig.BTCStartHeight
	for {
		// Get the current block hash
		blockHash, err := client.GetBlockHash(int64(currentHeight))
		if err != nil {
			log.Printf("Error getting block hash: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		// Get the block
		msgBlock, err := client.GetBlock(blockHash)
		if err != nil {
			log.Printf("Error getting block: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		// Convert to *btcutil.Block
		block := btcutil.NewBlock(msgBlock)

		// Cache block data
		CacheBlockData(db, block)
		SendBlockData(block)

		log.Printf("Cached block height: %d", currentHeight)

		// Move to the next block
		currentHeight++

		// Check if the latest block has been reached
		bestHeight, err := client.GetBlockCount()
		if err != nil {
			log.Printf("Error getting latest block height: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		if int64(currentHeight) > bestHeight {
			log.Printf("Reached the latest block, waiting for new blocks...")
			time.Sleep(10 * time.Second)
			currentHeight = int(bestHeight)
		}
	}
}

func CacheBlockData(db *leveldb.DB, block *btcutil.Block) {
	blockHash := block.Hash().String()
	header := block.MsgBlock().Header
	difficulty := header.Bits
	randomNumber := header.Nonce
	merkleRoot := header.MerkleRoot.String()
	blockTime := header.Timestamp.Unix()

	// Manual formatting of header fields
	headerStr := fmt.Sprintf("Version: %d, PrevBlock: %s, MerkleRoot: %s, Timestamp: %d, Bits: %d, Nonce: %d",
		header.Version, header.PrevBlock, header.MerkleRoot, header.Timestamp.Unix(), header.Bits, header.Nonce)

	// Convert to little-endian and store
	difficultyLE := make([]byte, 4)
	randomNumberLE := make([]byte, 4)
	blockTimeLE := make([]byte, 8)

	binary.LittleEndian.PutUint32(difficultyLE, difficulty)
	binary.LittleEndian.PutUint32(randomNumberLE, randomNumber)
	binary.LittleEndian.PutUint64(blockTimeLE, uint64(blockTime))

	// Cache block header
	db.Put([]byte("header:"+blockHash), []byte(headerStr), nil)

	// Cache difficulty in little-endian
	db.Put([]byte("difficulty:"+blockHash), difficultyLE, nil)

	// Cache random number in little-endian
	db.Put([]byte("random:"+blockHash), randomNumberLE, nil)

	// Cache Merkle root
	db.Put([]byte("merkleroot:"+blockHash), []byte(merkleRoot), nil)

	// Cache block time in little-endian
	db.Put([]byte("blocktime:"+blockHash), blockTimeLE, nil)

	// Cache block hash
	db.Put([]byte("blockhash:"+blockHash), []byte(blockHash), nil)

	// Cache UTXOs (Simplified example, more details should be stored in real case)
	for _, tx := range block.Transactions() {
		for _, txOut := range tx.MsgTx().TxOut {
			utxoKey := fmt.Sprintf("utxo:%s:%d", blockHash, txOut.Value)
			db.Put([]byte(utxoKey), txOut.PkScript, nil)
		}
	}

	log.Printf("Cached block %s with header %s, difficulty %d, random number %d, Merkle root %s, and block time %d",
		blockHash, headerStr, difficulty, randomNumber, merkleRoot, blockTime)
}

func purgeOldData(db *leveldb.DB) {
	thresholdTime := time.Now().AddDate(0, 0, -3).Unix()

	iter := db.NewIterator(nil, nil)
	defer iter.Release()

	for iter.Next() {
		key := iter.Key()
		blockTimeBytes, err := db.Get([]byte("blocktime:"+string(key[len("blockhash:"):])), nil)
		if err != nil {
			log.Printf("Error getting block time: %v", err)
			continue
		}

		blockTime := int64(binary.LittleEndian.Uint64(blockTimeBytes))
		if blockTime < thresholdTime {
			db.Delete(key, nil)
			log.Printf("Deleted block with key: %s", key)
		}
	}
	if err := iter.Error(); err != nil {
		log.Printf("Error during data purge: %v", err)
	}
}

func GenerateSPVProof(msgTx *wire.MsgTx) (string, error) {
	// Open or create the local storage
	dbPath := filepath.Join(config.AppConfig.DbDir, "btc_cache.db")
	db, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		log.Fatalf("Failed to open local storage: %v", err)
	}
	defer db.Close()

	txHash := msgTx.TxHash()

	// Get block hash
	var blockHashBytes []byte
	iter := db.NewIterator(nil, nil)
	defer iter.Release()

	for iter.Next() {
		key := iter.Key()
		if bytes.HasPrefix(key, []byte("utxo:")) {
			value := iter.Value()
			if bytes.Equal(value, txHash[:]) {
				blockHashBytes = key[len("utxo:") : len("utxo:")+64]
				break
			}
		}
	}
	if blockHashBytes == nil {
		return "", fmt.Errorf("failed to find block hash for tx: %v", txHash)
	}

	blockHash, err := chainhash.NewHash(blockHashBytes)
	if err != nil {
		return "", fmt.Errorf("invalid block hash: %v", err)
	}

	// Get block header
	headerBytes, err := db.Get([]byte("header:"+blockHash.String()), nil)
	if err != nil {
		return "", fmt.Errorf("failed to get block header from db: %v", err)
	}
	var header wire.BlockHeader
	err = header.Deserialize(bytes.NewReader(headerBytes))
	if err != nil {
		return "", fmt.Errorf("failed to deserialize block header: %v", err)
	}

	// Get transaction hash list
	txHashesBytes, err := db.Get([]byte("txhashes:"+blockHash.String()), nil)
	if err != nil {
		return "", fmt.Errorf("failed to get tx hashes from db: %v", err)
	}
	var txHashes []chainhash.Hash
	err = json.Unmarshal(txHashesBytes, &txHashes)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal tx hashes: %v", err)
	}

	// Find the transaction's position in the block
	var txIndex int
	for i, hash := range txHashes {
		if hash == txHash {
			txIndex = i
			break
		}
	}

	// Generate Merkle proof
	txHashesPtrs := make([]*chainhash.Hash, len(txHashes))
	for i := range txHashes {
		txHashesPtrs[i] = &txHashes[i]
	}
	var proof []*chainhash.Hash
	merkleRoot := ComputeMerkleRootAndProof(txHashesPtrs, txIndex, &proof)

	// Serialize Merkle proof
	var buf bytes.Buffer
	buf.Write(txHash[:])
	for _, p := range proof {
		buf.Write(p[:])
	}
	buf.Write(merkleRoot[:])

	return hex.EncodeToString(buf.Bytes()), nil
}
