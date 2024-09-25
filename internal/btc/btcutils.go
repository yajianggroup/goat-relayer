package btc

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

func GenerateSPVProof(txHash string, txHashes []string) ([]byte, []byte, uint32, error) {
	// Find the transaction's position in the block
	var txIndex int
	for i, hash := range txHashes {
		if hash == txHash {
			txIndex = i
			break
		}
	}

	// Generate merkle root and proof
	txHashesPtrs := make([]*chainhash.Hash, len(txHashes))
	for i, hashStr := range txHashes {
		hash, err := chainhash.NewHashFromStr(hashStr)
		if err != nil {
			return nil, nil, 0, fmt.Errorf("failed to parse transaction hash: %v", err)
		}
		txHashesPtrs[i] = hash
	}
	var proof []*chainhash.Hash
	merkleRoot := ComputeMerkleRootAndProof(txHashesPtrs, txIndex, &proof)

	// Serialize immediate proof
	var buf bytes.Buffer
	for _, p := range proof {
		buf.Write(p[:])
	}

	return merkleRoot.CloneBytes(), buf.Bytes(), uint32(txIndex), nil
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

func VerifyTransaction(tx wire.MsgTx, txHash string, evmAddress string) error {
	// TODO: Implement more detailed verification
	// type 0:
	// |-- extract receiver, txout=0?, get receiver, evm_address, goat
	// |-- compare public key, find txout
	// type 1
	// |-- find by receiver, txout=0
	// |-- find OP_RETURN, txout=1, get evm_address

	return nil
}

func P2wshDeposit(netwk *chaincfg.Params, tssGroupKey *btcec.PrivateKey, evmAddress []byte,
	prevTxId string, prevTxout int) (string, error) {
	redeemScript, err := txscript.NewScriptBuilder().
		AddData(evmAddress).
		AddOp(txscript.OP_DROP).
		AddData(tssGroupKey.PubKey().SerializeCompressed()).
		AddOp(txscript.OP_CHECKSIG).Script()
	if err != nil {
		return "", fmt.Errorf("failed to build redeemScript: %v", err)
	}

	witnessProg := sha256.Sum256(redeemScript)
	//depositAddress, err := btcutil.NewAddressWitnessScriptHash(witnessProg[:], netwk)
	//if err != nil {
	//	return "", fmt.Errorf("failed to build depositAddress: %v", err)
	//}
	//fmt.Println("depositAddress", depositAddress.EncodeAddress())

	tssGroupAddress, err := btcutil.NewAddressWitnessPubKeyHash(btcutil.Hash160(tssGroupKey.PubKey().SerializeCompressed()), netwk)
	if err != nil {
		return "", fmt.Errorf("failed to build tssGroupAddress: %v", err)
	}

	{
		newtx := wire.NewMsgTx(2)

		const prevAmountSat = 1e8
		const curAmountSat = prevAmountSat - 1e3

		prevTxid, err := chainhash.NewHashFromStr(prevTxId)
		if err != nil {
			return "", fmt.Errorf("failed to build prevTxid: %v", err)
		}

		prevPkScript, err := txscript.NewScriptBuilder().AddOp(txscript.OP_0).AddData(witnessProg[:]).Script()
		if err != nil {
			return "", fmt.Errorf("failed to build prevPkScript: %v", err)
		}

		// txout to the internal tss group address
		{
			outputAddress, err := txscript.PayToAddrScript(tssGroupAddress)
			if err != nil {
				return "", fmt.Errorf("failed to build outputAddress: %v", err)
			}
			txout := wire.NewTxOut(curAmountSat, outputAddress)
			newtx.AddTxOut(txout)
		}

		// txin, spend the user deposit
		{
			const txIdx = 0
			txin := wire.NewTxIn(wire.NewOutPoint(prevTxid, uint32(prevTxout)), nil, nil)
			newtx.AddTxIn(txin)

			sigHashes := txscript.NewTxSigHashes(newtx,
				txscript.NewCannedPrevOutputFetcher(prevPkScript, prevAmountSat))

			witSig, err := txscript.RawTxInWitnessSignature(
				newtx, sigHashes, txIdx,
				prevAmountSat, redeemScript,
				txscript.SigHashAll, tssGroupKey,
			)
			if err != nil {
				return "", fmt.Errorf("failed to build witSig: %v", err)
			}
			// witness = <unlock script> + <redeem script>
			// unlock script should be splited, refer to bip141 for the details
			txin.Witness = wire.TxWitness{witSig, redeemScript}
		}

		txbuf := bytes.NewBuffer(nil)
		if err := newtx.Serialize(txbuf); err != nil {
			return "", fmt.Errorf("failed to build txbuf: %v", err)
		}
		return hex.EncodeToString(txbuf.Bytes()), nil
	}
}
