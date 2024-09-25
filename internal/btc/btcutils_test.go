package btc

import (
	"bytes"
	"encoding/hex"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/chaincfg"
	"strings"
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	bitcointypes "github.com/goatnetwork/goat/x/bitcoin/types"
	"github.com/stretchr/testify/assert"
)

func TestCorrectTxHash(t *testing.T) {
	txHashes := []string{
		"6f33de3f5347f832f0f5ad39b0bc4309ec6a9de586d6763b733e1fbecbd9c8d8",
		"43a434c639ab3884361f168870b658d331e8dbc9dfbf05af093ee07c20ab766f",
		"915cf91cef8a56c1284616ad149b6ee0674360ed09c51e10b2de5b9ec36b24d4",
	}

	// correct tx hash
	rawTxHex := "020000000001010000000000000000000000000000000000000000000000000000000000000000ffffffff03016f00ffffffff02706e0a2a01000000160014f890a762934d98c2f4cabbb829863de03fd7e7b00000000000000000266a24aa21a9edc1c1fdca70dc22518c456c4125655712e481c401137aad186c2eac5851c077b80120000000000000000000000000000000000000000000000000000000000000000000000000"
	assert.True(t, processTransaction(rawTxHex, txHashes))
}

func TestWrongTxHash(t *testing.T) {
	txHashes := []string{
		"6f33de3f5347f832f0f5ad39b0bc4309ec6a9de586d6763b733e1fbecbd9c8d8",
		"43a434c639ab3884361f168870b658d331e8dbc9dfbf05af093ee07c20ab766f",
		"915cf91cef8a56c1284616ad149b6ee0674360ed09c51e10b2de5b9ec36b24d4",
	}
	// wrong tx hash which modifies the raw tx
	rawTxHex := "021011111101010000000000000000000000000000000000000000000000000000000000000000ffffffff03016f001111111102706e0a2a01000000160014f890a762934d98c2f4cabbb829863de03fd7e7b00000000000000000266a24aa21a9edc1c1fdca70dc22518c456c4125655712e481c401137aad186c2eac5851c077b80120000000000000000000000000000000000000000000000000000000000000000000000000"
	assert.False(t, processTransaction(rawTxHex, txHashes))
}

func processTransaction(rawTxHex string, txHashes []string) bool {
	rawTxBytes, err := hex.DecodeString(rawTxHex)
	if err != nil {
		return false
	}

	if err != nil {
		return false
	}

	var tx wire.MsgTx
	err = tx.Deserialize(bytes.NewReader(rawTxBytes))
	if err != nil {
		return false
	}

	merkleRoot, proofBytes, txIndex, err := GenerateSPVProof(tx.TxHash().String(), txHashes)
	if err != nil {
		return false
	}

	// Verify the generated proof
	targetTxHash, _ := chainhash.NewHashFromStr(txHashes[txIndex])

	return bitcointypes.VerifyMerkelProof(targetTxHash[:], merkleRoot[:], proofBytes, txIndex)
}

func TestP2wshDeposit(t *testing.T) {
	privKeyBytes, _ := hex.DecodeString("e9ccd0ec6bb77c263dc46c0f81962c0b378a67befe089e90ef81e96a4a4c5bc5")
	privKey, _ := btcec.PrivKeyFromBytes(privKeyBytes)
	evmAddress := "0x70997970c51812dc3a010c7d01b50e0d17dc79c8"
	if strings.HasPrefix(evmAddress, "0x") {
		evmAddress = evmAddress[2:]
	}

	evmAddress_, _ := hex.DecodeString(evmAddress)
	const prevTxId = "1a4ecdb32ca38287e862de3d7e21e551a0d76645d72cb7229058600b7a817553"
	const prevTxout = 0
	const amount = 1e8
	const fee = 1e3
	rawTxHex, err := P2wshDeposit(&chaincfg.RegressionNetParams, privKey, evmAddress_, prevTxId, prevTxout, amount, fee)
	assert.Nil(t, err)
	assert.NotEmpty(t, rawTxHex)

	txBytes, err := hex.DecodeString(rawTxHex)
	assert.Nil(t, err)

	tx := wire.NewMsgTx(2)
	err = tx.Deserialize(bytes.NewReader(txBytes))
	assert.Nil(t, err)

	assert.NotNil(t, tx.TxOut)
	assert.NotNil(t, tx.TxIn)
}
