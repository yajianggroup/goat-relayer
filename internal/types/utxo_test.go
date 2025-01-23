package types

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

var (
	TEST_GOAT_MAGIC_BYTES = []byte{0x47, 0x4F, 0x41, 0x54} // "GOAT"

	// Actual witness data from blockchain
	TEST_P2WPKH_SIGNATURE = "3045022100fca1b651552cf416f08cda21018a5c6c9405ab1b94cc694b52026a90cdcb769702200543a5d3d1b33369607fc1354f2a8af6bb0e08b9fe5445047c4d77c81c0f2b8101"
	TEST_P2WPKH_PUBKEY    = "02b46f2e2e387cbc2bfb541da34d5149256f593a3c175b18004ba21db23d2b8c24"
)

func encodeWitnessStack(items ...[]byte) []byte {
	// Calculate total size
	size := 1 // varint for number of items
	for _, item := range items {
		size += 1 // varint for item length
		size += len(item)
	}

	// Encode witness stack
	witness := make([]byte, 0, size)
	witness = append(witness, byte(len(items))) // Number of items (varint)
	for _, item := range items {
		witness = append(witness, byte(len(item))) // Item length (varint)
		witness = append(witness, item...)         // Item data
	}
	return witness
}

func TestIsUtxoGoatDepositV1WithRawTx(t *testing.T) {
	// Decode the provided raw transaction
	rawTx := "020000000001013b16081931db7633dde8e0475385e607de9d8f3b372a219115179858787517a30200000000fdffffff0380969800000000001600149759ed6aae6ade43ae6628a943a39974cd21c5df00000000000000001a6a184754543070997970c51812dc3a010c7d01b50e0d17dc79c8ca649b1c00000000160014059ce0647de86cf966dfa4656a08530eb8f26772024730440220183ac1ff33c5ad358069ec79a030c3329078d02d77f3ae19d0226810bad74d6d02203c61ba76f590380cf61b4cfc188a0de718c801b4eb24d58dbbf9ad928879449901210361e82e71277ea205814b1cb69777abe5fc417c03d4d39829cefb8f92da08b1fc00000000"
	txBytes, err := hex.DecodeString(rawTx)
	if err != nil {
		t.Fatalf("Failed to decode raw transaction: %v", err)
	}

	// Deserialize the transaction
	tx := wire.NewMsgTx(wire.TxVersion)
	err = tx.Deserialize(bytes.NewReader(txBytes))
	if err != nil {
		t.Fatalf("Failed to deserialize raw transaction: %v", err)
	}

	// Decode the provided public key (address)
	pubKeyHex := "0383560def84048edefe637d0119a4428dd12a42765a118b2bf77984057633c50e"
	pubKeyBytes, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		t.Fatalf("Failed to decode public key: %v", err)
	}

	network := &chaincfg.RegressionNetParams
	p2pkhAddress, err := GenerateP2PKHAddress(pubKeyBytes, network)
	if err != nil {
		t.Fatalf("Gen P2PKH address err %v", err)
	}
	p2wpkhAddress, err := GenerateP2WPKHAddress(pubKeyBytes, network)
	if err != nil {
		t.Fatalf("Gen P2WPKH address err %v", err)
	}
	t.Logf("p2pkhAddress: %s", p2pkhAddress.EncodeAddress())
	t.Logf("p2wpkhAddress: %s", p2wpkhAddress.EncodeAddress())

	// Mock tssAddress with the same address as TxOut[0]
	tssAddress := []btcutil.Address{p2pkhAddress, p2wpkhAddress}

	// Test the transaction using IsUtxoGoatDepositV1
	isDepositV1, evmAddress, _ := IsUtxoGoatDepositV1(tx, tssAddress, network, 0, TEST_GOAT_MAGIC_BYTES)

	if !isDepositV1 {
		t.Errorf("Expected transaction to be a valid GOAT deposit, got invalid")
	}
	t.Logf("Extract evmAddress: %s", evmAddress)
	assert.NotEqual(t, evmAddress, (common.Address{}).Hex())
}

func TestTransactionSizeEstimateV2(t *testing.T) {
	// Convert witness data from hex
	signature, _ := hex.DecodeString(TEST_P2WPKH_SIGNATURE)
	pubkey, _ := hex.DecodeString(TEST_P2WPKH_PUBKEY)

	// Create complete witness stack
	witnessStack := encodeWitnessStack(signature, pubkey)

	tests := []struct {
		name          string
		numInputs     int
		receiverTypes []string
		numOutputs    int
		utxoTypes     []string
		wantVSize     int64
		wantWitSize   int64
		witnessData   []byte // actual witness data
	}{
		{
			name:          "Single P2WPKH input with 5 P2WPKH outputs",
			numInputs:     1,
			receiverTypes: []string{WALLET_TYPE_P2WPKH, WALLET_TYPE_P2WPKH, WALLET_TYPE_P2WPKH, WALLET_TYPE_P2WPKH, WALLET_TYPE_P2WPKH},
			numOutputs:    5,
			utxoTypes:     []string{WALLET_TYPE_P2WPKH},
			wantVSize:     234, // (baseSize*4 + witSize)/4 = 233.5, truncated to 234
			wantWitSize:   108, // witness: stack_items(1) + sig_len(1) + sig(72) + pubkey_len(1) + pubkey(33) = 110
			witnessData:   witnessStack,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVSize, gotWitSize := TransactionSizeEstimateV2(tt.numInputs, tt.receiverTypes, tt.numOutputs, tt.utxoTypes)
			if gotVSize != tt.wantVSize {
				t.Errorf("TransactionSizeEstimateV2() virtual size = %v, want %v", gotVSize, tt.wantVSize)
			}
			if gotWitSize != tt.wantWitSize {
				t.Errorf("TransactionSizeEstimateV2() witness size = %v, want %v", gotWitSize, tt.wantWitSize)
			}

			// Verify witness data size matches our calculation
			if tt.witnessData != nil {
				actualWitnessSize := len(tt.witnessData)
				if int64(actualWitnessSize) != tt.wantWitSize {
					t.Errorf("Actual witness data size = %v, want %v", actualWitnessSize, tt.wantWitSize)
				}
			}
		})
	}
}
