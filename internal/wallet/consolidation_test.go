package wallet

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/assert"
)

func TestSpentP2wsh(t *testing.T) {
	privKeyBytes, _ := hex.DecodeString("e9ccd0ec6bb77c263dc46c0f81962c0b378a67befe089e90ef81e96a4a4c5bc5")
	privKey, _ := btcec.PrivKeyFromBytes(privKeyBytes)
	evmAddress := "0x70997970c51812dc3a010c7d01b50e0d17dc79c8"
	if strings.HasPrefix(evmAddress, "0x") {
		evmAddress = evmAddress[2:]
	}

	evmAddress_, _ := hex.DecodeString(evmAddress)

	evmAddresses := [][]byte{
		evmAddress_,
	}

	prevTxIds := []string{
		"1a4ecdb32ca38287e862de3d7e21e551a0d76645d72cb7229058600b7a817553",
	}

	prevTxouts := []int{
		0,
	}
	amounts := []int64{
		1e8,
	}

	rawTxHex, err := SpentP2wsh(&chaincfg.RegressionNetParams, privKey, evmAddresses, prevTxIds, prevTxouts, amounts,
		1e3)
	assert.Nil(t, err)
	assert.NotEmpty(t, rawTxHex)

	txBytes, err := hex.DecodeString(rawTxHex)
	assert.Nil(t, err)

	tx := wire.NewMsgTx(2)
	err = tx.Deserialize(bytes.NewReader(txBytes))
	assert.Nil(t, err)

	assert.NotNil(t, tx.TxOut)
	assert.NotNil(t, tx.TxIn)
	assert.Equal(t, len(tx.TxIn), len(evmAddresses))
}
