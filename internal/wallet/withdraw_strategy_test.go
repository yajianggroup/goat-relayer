package wallet_test

import (
	"fmt"
	"testing"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/goatnetwork/goat-relayer/internal/db"
	"github.com/goatnetwork/goat-relayer/internal/types"
	"github.com/goatnetwork/goat-relayer/internal/wallet"
	"github.com/stretchr/testify/assert"
)

// Test ConsolidateSmallUTXOs function
func TestConsolidateSmallUTXOs(t *testing.T) {
	// mock UTXOs for testing
	utxos := []*db.Utxo{
		{Amount: 20000000, ReceiverType: types.WALLET_TYPE_P2PKH},
		{Amount: 15000000, ReceiverType: types.WALLET_TYPE_P2WPKH},
		{Amount: 10000000, ReceiverType: types.WALLET_TYPE_P2WPKH},
		{Amount: 10000000, ReceiverType: types.WALLET_TYPE_P2WPKH},
		{Amount: 10000000, ReceiverType: types.WALLET_TYPE_P2WPKH},
		{Amount: 10000000, ReceiverType: types.WALLET_TYPE_P2WPKH},
		{Amount: 10000000, ReceiverType: types.WALLET_TYPE_P2WPKH},
		{Amount: 10000000, ReceiverType: types.WALLET_TYPE_P2WPKH},
		{Amount: 10000000, ReceiverType: types.WALLET_TYPE_P2WPKH},
		{Amount: 15000000, ReceiverType: types.WALLET_TYPE_P2WSH},
		{Amount: 5000000, ReceiverType: types.WALLET_TYPE_P2PKH},
	}

	// when network fee is too high
	selectedUtxos, totalAmount, finalAmount, err := wallet.ConsolidateSmallUTXOs(utxos, 300, 10000000, 10, 10)
	assert.Error(t, err)
	assert.EqualError(t, err, "network fee is too high, cannot consolidate")
	assert.Nil(t, selectedUtxos)
	assert.Equal(t, int64(0), totalAmount)
	assert.Equal(t, int64(0), finalAmount)

	// when there are not enough UTXOs to consolidate
	selectedUtxos, totalAmount, finalAmount, err = wallet.ConsolidateSmallUTXOs(utxos, 100, 10000000, 10, 10)
	assert.Error(t, err)
	assert.EqualError(t, err, "not enough small utxos to consolidate")
	assert.Nil(t, selectedUtxos)
	assert.Equal(t, int64(0), totalAmount)
	assert.Equal(t, int64(0), finalAmount)

	// valid consolidation
	selectedUtxos, totalAmount, finalAmount, err = wallet.ConsolidateSmallUTXOs(utxos, 100, 20000000, 10, 10)
	assert.NoError(t, err)
	assert.NotNil(t, selectedUtxos)
	assert.Greater(t, totalAmount, int64(0))
	assert.Greater(t, finalAmount, int64(0))
}

// Test SelectOptimalUTXOs function
func TestSelectOptimalUTXOs(t *testing.T) {
	// mock UTXOs for testing
	utxos := []*db.Utxo{
		{Amount: 50000000, ReceiverType: types.WALLET_TYPE_P2PKH},
		{Amount: 30000000, ReceiverType: types.WALLET_TYPE_P2WPKH},
		{Amount: 20000000, ReceiverType: types.WALLET_TYPE_P2WSH},
		{Amount: 15000000, ReceiverType: types.WALLET_TYPE_P2PKH},
	}

	// valid selection of UTXOs
	selectedUtxos, totalSelectedAmount, withdrawAmount, changeAmount, estimatedFee, err := wallet.SelectOptimalUTXOs(utxos, 40000000, 100, 1)
	assert.NoError(t, err)
	assert.NotNil(t, selectedUtxos)
	assert.Greater(t, totalSelectedAmount, int64(0))
	assert.Equal(t, int64(40000000), withdrawAmount)
	assert.Greater(t, changeAmount, int64(0))
	assert.Greater(t, estimatedFee, int64(0))

	// when not enough UTXOs
	_, _, _, _, _, err = wallet.SelectOptimalUTXOs(utxos, 150000000, 100, 1)
	assert.Error(t, err)
	assert.EqualError(t, err, "not enough utxos to satisfy the withdrawal amount and network fee, withdraw amount: 150000000, selected amount: 115000000, estimated fee: 300")
}

// Test SelectWithdrawals function
func TestSelectWithdrawals(t *testing.T) {
	// mock Withdrawals for testing
	withdrawals := []*db.Withdraw{
		{Amount: 70000000, TxPrice: 300},
		{Amount: 50000000, TxPrice: 200},
		{Amount: 30000000, TxPrice: 100},
		{Amount: 15000000, TxPrice: 50},
	}

	// valid withdrawal selection
	selectedWithdrawals, withdrawAmount, minTxFee, err := wallet.SelectWithdrawals(withdrawals, 100, 2)
	t.Logf("SelectWithdrawals returns selectedWithdrawals len %d, withdrawAmount %d, minTxFee %d, err %v", len(selectedWithdrawals), withdrawAmount, minTxFee, err)
	assert.NoError(t, err)
	assert.NotNil(t, selectedWithdrawals)
	assert.Greater(t, withdrawAmount, int64(0))
	assert.Greater(t, minTxFee, int64(0))

	// when network fee is too high
	_, _, _, err = wallet.SelectWithdrawals(withdrawals, 600, 2)
	assert.Error(t, err)
	assert.EqualError(t, err, "network fee too high, no withdrawals allowed")
}

// Test CreateRawTransaction function
func TestCreateRawTransaction(t *testing.T) {
	// mock UTXOs for testing
	utxos := []*db.Utxo{
		{Txid: "a3e0d1c3ff01", OutIndex: 0, Amount: 50000000, ReceiverType: types.WALLET_TYPE_P2PKH, Sender: "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"},
		{Txid: "b4f1e2d4ff02", OutIndex: 1, Amount: 30000000, ReceiverType: types.WALLET_TYPE_P2WPKH, Sender: "bc1q254g9ax097y04erwjesrrce8nv3t7k0ajwylwu"},
	}

	// mock Withdrawals for testing
	withdrawals := []*db.Withdraw{
		{Amount: 20000000, To: "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", ID: 1},
		{Amount: 10000000, To: "bc1q254g9ax097y04erwjesrrce8nv3t7k0ajwylwu", ID: 2},
	}

	// bitcoin mainnet params
	net := &chaincfg.MainNetParams

	// valid transaction creation
	tx, dustWithdraw, err := wallet.CreateRawTransaction(utxos, withdrawals, "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", 5000000, 1000, net)
	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, uint(0), dustWithdraw)

	// validate transaction inputs and outputs
	assert.Equal(t, 2, len(tx.TxIn))  // 2 inputs
	assert.Equal(t, 3, len(tx.TxOut)) // 2 withdrawals + 1 change

	// check if inputs are added correctly
	for i, utxo := range utxos {
		expectedHash, _ := chainhash.NewHashFromStr(utxo.Txid)
		assert.Equal(t, expectedHash, &tx.TxIn[i].PreviousOutPoint.Hash)
		assert.Equal(t, uint32(utxo.OutIndex), tx.TxIn[i].PreviousOutPoint.Index)
	}

	// check if outputs are added correctly
	for i, withdrawal := range withdrawals {
		addr, err := btcutil.DecodeAddress(withdrawal.To, net)
		assert.NoError(t, err)
		expectedScript, err := txscript.PayToAddrScript(addr)
		assert.NoError(t, err)
		assert.Equal(t, expectedScript, tx.TxOut[i].PkScript)
		assert.Equal(t, int64(withdrawal.Amount)-1000/3, tx.TxOut[i].Value) // minus fee
	}

	// when withdrawal amount is too small (dust)
	withdrawals[0].Amount = 500 // less than dust limit
	_, dustWithdraw, err = wallet.CreateRawTransaction(utxos, withdrawals, "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", 5000000, 1000, net)
	assert.Error(t, err)
	assert.Equal(t, uint(1), dustWithdraw) // the ID of the withdrawal with the small amount
	assert.EqualError(t, err, fmt.Sprintf("withdrawal amount too small after fee deduction: %d", withdrawals[0].Amount-1000/3))
}
