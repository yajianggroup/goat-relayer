package rpc

import (
	"encoding/base64"
	"errors"
	"strings"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/wire"
	"github.com/goatnetwork/goat-relayer/internal/config"
	"github.com/goatnetwork/goat-relayer/internal/types"
)

func (s *UtxoServer) VerifyDeposit(tx wire.MsgTx, evmAddress string) (isTrue bool, signVersion uint32, outputIndex int, err error) {
	network := types.GetBTCNetwork(config.AppConfig.BTCNetworkType)

	pubKey, err := s.getPubKey()
	if err != nil {
		return false, 100, -1, err
	}

	p2wpkh, err := types.GenerateP2WPKHAddress(pubKey, network)
	if err != nil {
		return false, 100, -1, err
	}

	isTrue, _ = types.IsUtxoGoatDepositV1(&tx, []btcutil.Address{p2wpkh}, network)
	if isTrue {
		return true, 1, 0, nil
	}

	p2wsh, err := types.GenerateV0P2WSHAddress(pubKey, evmAddress, network)
	if err != nil {
		return false, 100, -1, err
	}

	isTrue, outputIndex = types.IsUtxoGoatDepositV0(&tx, []btcutil.Address{p2wsh}, network)
	if isTrue {
		return true, 0, outputIndex, nil
	}

	return false, 100, -1, errors.New("invalid deposit address")
}

func (s *UtxoServer) getPubKey() ([]byte, error) {
	l2Info := s.state.GetL2Info()

	var err error
	var pubKey []byte
	if l2Info.DepositKey == "" {
		depositPubKey, err := s.state.GetDepositKeyByBtcBlock(0)
		if err != nil {
			return nil, err
		}
		pubKey, err = base64.StdEncoding.DecodeString(depositPubKey.PubKey)
		if err != nil {
			return nil, err
		}
	} else {
		pubKeyStr := strings.Split(l2Info.DepositKey, ",")[1]
		pubKey, err = base64.StdEncoding.DecodeString(pubKeyStr)
		if err != nil {
			return nil, err
		}
	}

	return pubKey, nil
}
