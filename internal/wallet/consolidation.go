package wallet

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

func SpentP2wsh(netwk *chaincfg.Params, tssGroupKey *btcec.PrivateKey, evmAddresses [][]byte, prevTxIds []string,
	prevTxouts []int, amounts []int64, fee int64) (string, error) {
	if len(prevTxIds) != len(prevTxouts) || len(prevTxouts) != len(amounts) || len(evmAddresses) != len(prevTxIds) {
		return "", fmt.Errorf("mismatched input lengths")
	}

	// Create a new transaction
	newtx := wire.NewMsgTx(2)

	totalInputAmount := int64(0)
	for i, evmAddress := range evmAddresses {
		// Create redeemScript for each evmAddress
		redeemScript, err := txscript.NewScriptBuilder().
			AddData(evmAddress).
			AddOp(txscript.OP_DROP).
			AddData(tssGroupKey.PubKey().SerializeCompressed()).
			AddOp(txscript.OP_CHECKSIG).Script()
		if err != nil {
			return "", fmt.Errorf("failed to build redeemScript: %v", err)
		}

		witnessProg := sha256.Sum256(redeemScript)

		// Add the UTXO inputs
		prevTxid, err := chainhash.NewHashFromStr(prevTxIds[i])
		if err != nil {
			return "", fmt.Errorf("failed to build prevTxid: %v", err)
		}

		// Script for previous UTXO
		prevPkScript, err := txscript.NewScriptBuilder().AddOp(txscript.OP_0).AddData(witnessProg[:]).Script()
		if err != nil {
			return "", fmt.Errorf("failed to build prevPkScript: %v", err)
		}

		txin := wire.NewTxIn(wire.NewOutPoint(prevTxid, uint32(prevTxouts[i])), nil, nil)
		newtx.AddTxIn(txin)

		// Update total input amount
		totalInputAmount += amounts[i]

		// Create the witness signature for each input
		sigHashes := txscript.NewTxSigHashes(newtx,
			txscript.NewCannedPrevOutputFetcher(prevPkScript, amounts[i]))

		witSig, err := txscript.RawTxInWitnessSignature(
			newtx, sigHashes, i,
			amounts[i], redeemScript,
			txscript.SigHashAll, tssGroupKey,
		)
		if err != nil {
			return "", fmt.Errorf("failed to build witSig: %v", err)
		}

		// Add witness data
		txin.Witness = wire.TxWitness{witSig, redeemScript}
	}

	// Output to the internal TSS group address
	tssGroupAddress, err := btcutil.NewAddressWitnessPubKeyHash(btcutil.Hash160(tssGroupKey.PubKey().SerializeCompressed()), netwk)
	if err != nil {
		return "", fmt.Errorf("failed to build tssGroupAddress: %v", err)
	}

	outputAddress, err := txscript.PayToAddrScript(tssGroupAddress)
	if err != nil {
		return "", fmt.Errorf("failed to build outputAddress: %v", err)
	}

	// Deduct fee and create output
	totalOutputAmount := totalInputAmount - fee
	if totalOutputAmount <= 0 {
		return "", fmt.Errorf("output amount is too small")
	}

	txout := wire.NewTxOut(totalOutputAmount, outputAddress)
	newtx.AddTxOut(txout)

	// Serialize the transaction
	txbuf := bytes.NewBuffer(nil)
	if err := newtx.Serialize(txbuf); err != nil {
		return "", fmt.Errorf("failed to build txbuf: %v", err)
	}

	return hex.EncodeToString(txbuf.Bytes()), nil
}
