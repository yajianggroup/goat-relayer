package types

import "github.com/btcsuite/btcd/wire"

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
