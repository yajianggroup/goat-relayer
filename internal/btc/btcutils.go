package btc

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/goatnetwork/goat-relayer/internal/db"
)

func GenerateSPVProof(msgTx *wire.MsgTx) (string, error) {
	dbm := db.NewDatabaseManager()
	btcCacheDb := dbm.GetBtcCacheDB()

	txHash := msgTx.TxHash()

	var btcTXOutput db.BtcTXOutput
	if err := btcCacheDb.Where("tx_hash = ?", txHash).First(&btcTXOutput).Error; err != nil {
		return "", fmt.Errorf("failed to retrieve block data from database: %v", err)
	}

	blockHashBytes := btcTXOutput.PkScript[:32] // Assuming the block hash is the first 32 bytes of PkScript
	blockHash, err := chainhash.NewHash(blockHashBytes)
	if err != nil {
		return "", fmt.Errorf("failed to create hash from block hash bytes: %v", err)
	}

	var blockData db.BtcBlockData
	if err := btcCacheDb.Where("block_hash = ?", blockHash.String()).First(&blockData).Error; err != nil {
		return "", fmt.Errorf("failed to retrieve block header from database: %v", err)
	}

	var txHashes []chainhash.Hash
	if err := json.Unmarshal([]byte(blockData.TxHashes), &txHashes); err != nil {
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
