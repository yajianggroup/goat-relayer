package btc

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/goatnetwork/goat-relayer/internal/config"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
)

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

func SerializeNoWitnessTx(rawTransaction []byte) ([]byte, error) {
	// Parse the raw transaction
	rawTx := wire.NewMsgTx(wire.TxVersion)
	err := rawTx.Deserialize(bytes.NewReader(rawTransaction))
	if err != nil {
		return nil, fmt.Errorf("failed to parse raw transaction: %v", err)
	}

	// Create a new transaction without witness data
	noWitnessTx := wire.NewMsgTx(rawTx.Version)

	// Copy transaction inputs, excluding witness data
	for _, txIn := range rawTx.TxIn {
		newTxIn := wire.NewTxIn(&txIn.PreviousOutPoint, nil, nil)
		newTxIn.Sequence = txIn.Sequence
		noWitnessTx.AddTxIn(newTxIn)
	}

	// Copy transaction outputs
	for _, txOut := range rawTx.TxOut {
		noWitnessTx.AddTxOut(txOut)
	}

	// Set lock time
	noWitnessTx.LockTime = rawTx.LockTime

	// Serialize the transaction without witness data
	var buf bytes.Buffer
	err = noWitnessTx.Serialize(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction without witness data: %v", err)
	}

	return buf.Bytes(), nil
}
