package types

import (
	"encoding/hex"

	bitcointype "github.com/goatnetwork/goat/x/bitcoin/types"
)

func ReverseHash(hash string) (string, error) {
	data, err := hex.DecodeString(hash)
	if err != nil {
		return "", err
	}
	return bitcointype.BtcTxid(data), nil
}
