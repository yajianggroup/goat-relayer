package types

import (
	"bytes"
	"errors"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/ethereum/go-ethereum/common"
	log "github.com/sirupsen/logrus"
)

var (
	GOAT_MAGIC_BYTES = []byte{0x47, 0x54, 0x54, 0x30} // "GTT0"
)

// MsgUtxoDeposit defines deposit UTXO broadcast to p2p which received in relayer rpc
// TODO columns
type MsgUtxoDeposit struct {
	RawTx     string `json:"raw_tx"`
	TxId      string `json:"tx_id"`
	EvmAddr   string `json:"evm_addr"`
	Timestamp int64  `json:"timestamp"`
}

type BtcBlockExt struct {
	wire.MsgBlock

	BlockNumber uint64
}

func isOpReturn(txOut *wire.TxOut) bool {
	return len(txOut.PkScript) > 0 && txOut.PkScript[0] == txscript.OP_RETURN
}

func parseOpReturnGoatJson(data []byte) (common.Address, error) {
	if !bytes.HasPrefix(data, GOAT_MAGIC_BYTES) {
		return common.Address{}, errors.New("data does not start with magic bytes")
	}
	log.Debugf("Parsed OP_RETURN as GTT0: %v", data)
	remainingBytes := data[len(GOAT_MAGIC_BYTES):]
	if len(remainingBytes) == 20 {
		evmAddr := common.BytesToAddress(remainingBytes)
		log.Debugf("Parsed OP_RETURN evm address: %s", evmAddr.Hex())
		return evmAddr, nil
	}
	return common.Address{}, errors.New("data does not match as evm address")
}

func IsUtxoGoatDepositV1(tx *wire.MsgTx, tssAddress []btcutil.Address, net *chaincfg.Params) (bool, string) {
	if len(tx.TxOut) < 2 {
		return false, ""
	}
	if !isOpReturn(tx.TxOut[1]) || isOpReturn(tx.TxOut[0]) {
		return false, ""
	}
	_, addresses, requireSigs, err := txscript.ExtractPkScriptAddrs(tx.TxOut[0].PkScript, net)
	if err != nil || addresses == nil || requireSigs > 1 {
		return false, ""
	}
	for _, address := range tssAddress {
		if address.EncodeAddress() == addresses[0].EncodeAddress() {
			// check if tx.TxOut[1] OP_RETURN rule: https://www.goat.network/docs/deposit/v1
			// extract evm_addr from tx.TxOut[1] OP_RETURN
			data := tx.TxOut[1].PkScript[1:]
			evmAddr, err := parseOpReturnGoatJson(data)
			if err != nil {
				return false, ""
			}
			return true, evmAddr.Hex()
		}
	}
	return false, ""
}
