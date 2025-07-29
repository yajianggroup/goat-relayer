package voter

import (
	"encoding/hex"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cosmossdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/crypto"
	goatcryp "github.com/goatnetwork/goat/pkg/crypto"
	relayertypes "github.com/goatnetwork/goat/x/relayer/types"
)

// TestValidateNewVoterSignatures Test signature validation
func TestValidateNewVoterSignatures(t *testing.T) {
	// Ramdom generated data
	relayerPriKeyHex := "70c967688aa247bf6ec842b5eef981401da6135b93dc712bb69e8fb75f0f13d2"
	relayerBlsSkHex := "023fdc2a776ac542d615864b73fe2875f1ee299c643cd7d6adf0eed8f7211854"

	// Data queried from the database
	epoch := uint64(6162878) // foundVoter.Epoch
	chainID := "goat-test-1"
	proposer := "goat1slhs4ylhzuc2vsq45n33drtkdr27zeytf25yq8"
	sequence := uint64(0)

	// Parse private key
	privKeyBytes, err := hex.DecodeString(relayerPriKeyHex)
	if err != nil {
		t.Fatalf("Failed to decode private key: %v", err)
	}
	privKey := &secp256k1.PrivKey{Key: privKeyBytes}

	// Parse BLS private key
	blsBytes, err := hex.DecodeString(relayerBlsSkHex)
	if err != nil {
		t.Fatalf("Failed to decode BLS private key: %v", err)
	}
	blsSk := new(goatcryp.PrivateKey).Deserialize(blsBytes)
	blsPk := new(goatcryp.PublicKey).From(blsSk).Compress()

	// Calculate address
	addrRaw := cosmossdktypes.AccAddress(goatcryp.Hash160Sum(privKey.PubKey().Bytes()))

	t.Logf("=== Real Production Data ===")
	t.Logf("Relayer Address: %s", addrRaw.String())
	t.Logf("Relayer Public Key: %s", hex.EncodeToString(privKey.PubKey().Bytes()))
	t.Logf("BLS Public Key: %s", hex.EncodeToString(blsPk))
	t.Logf("Epoch: %d", epoch)

	// Construct OnBoardingVoterRequest
	blsKeyHash := goatcryp.SHA256Sum(blsPk)
	reqMsg := relayertypes.NewOnBoardingVoterRequest(epoch, addrRaw, blsKeyHash)

	// Construct signature document
	sigMsg := relayertypes.VoteSignDoc(
		reqMsg.MethodName(), chainID, proposer, sequence, epoch, reqMsg.SignDoc())

	// Generate signature
	// Use go-ethereum's signing function to generate a 65-byte signature and take the first 64 bytes for verification
	ecdsaPriv, err := crypto.ToECDSA(privKeyBytes)
	if err != nil {
		t.Fatalf("Failed to convert private key: %v", err)
	}
	sig65, err := crypto.Sign(sigMsg, ecdsaPriv)
	if err != nil {
		t.Fatalf("Failed to sign tx key proof: %v", err)
	}
	txKeyProof := sig65[:64]

	blsKeyProof := goatcryp.Sign(blsSk, sigMsg)

	t.Logf("=== Signature Data ===")
	t.Logf("Signature Document: %s", hex.EncodeToString(sigMsg))
	t.Logf("TX Key Proof: %s", hex.EncodeToString(txKeyProof))
	t.Logf("BLS Key Proof: %s", hex.EncodeToString(blsKeyProof))

	// Create a test RelayerSignProcessor
	processor := &RelayerSignProcessor{}

	// Test validation function
	t.Run("ValidateSignatures", func(t *testing.T) {
		valid := processor.validateNewVoterSignatures(
			sigMsg, txKeyProof, blsKeyProof,
			privKey.PubKey().Bytes(), blsPk, addrRaw)

		if !valid {
			t.Errorf("Signature validation failed")
		} else {
			t.Logf("✅ Signature validation successful")
		}
	})

	// Test individual validation steps
	t.Run("VerifyTXKeyProof", func(t *testing.T) {
		valid := crypto.VerifySignature(privKey.PubKey().Bytes(), sigMsg, txKeyProof)
		if !valid {
			t.Errorf("TX Key Proof validation failed")
		} else {
			t.Logf("✅ TX Key Proof validation successful")
		}
	})

	t.Run("VerifyBLSKeyProof", func(t *testing.T) {
		valid := goatcryp.Verify(blsPk, sigMsg, blsKeyProof)
		if !valid {
			t.Errorf("BLS Key Proof validation failed")
		} else {
			t.Logf("✅ BLS Key Proof validation successful")
		}
	})

	// Test address calculation
	t.Run("VerifyAddressCalculation", func(t *testing.T) {
		calculatedAddr := cosmossdktypes.AccAddress(goatcryp.Hash160Sum(privKey.PubKey().Bytes()))
		if calculatedAddr.String() != addrRaw.String() {
			t.Errorf("Address calculation error: expected %s, got %s", addrRaw.String(), calculatedAddr.String())
		} else {
			t.Logf("✅ Address calculation correct: %s", calculatedAddr.String())
		}
	})

	// Test parameter format
	t.Run("VerifyParameterFormat", func(t *testing.T) {
		if len(privKey.PubKey().Bytes()) != Secp256k1CompressedPubKeyLength {
			t.Errorf("TX Key length error: expected %d, got %d", Secp256k1CompressedPubKeyLength, len(privKey.PubKey().Bytes()))
		} else {
			t.Logf("✅ TX Key length correct: %d bytes", len(privKey.PubKey().Bytes()))
		}

		if len(blsPk) != BLSCompressedPubKeyLength {
			t.Errorf("BLS Key length error: expected %d, got %d", BLSCompressedPubKeyLength, len(blsPk))
		} else {
			t.Logf("✅ BLS Key length correct: %d bytes", len(blsPk))
		}
	})
}
