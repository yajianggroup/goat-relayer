package types

import (
	"encoding/hex"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/goatnetwork/goat-relayer/internal/config"
	bitcointype "github.com/goatnetwork/goat/x/bitcoin/types"
	log "github.com/sirupsen/logrus"
)

func ReverseHash(hash string) (string, error) {
	data, err := hex.DecodeString(hash)
	if err != nil {
		return "", err
	}
	return bitcointype.BtcTxid(data), nil
}

func PrivateKeyToGethAddress(privateKeyHex string) (string, error) {
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		log.Errorf("Failed to decode private key: %v", err)
		return "", err
	}

	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		log.Errorf("Failed to parse private key: %v", err)
		return "", err
	}

	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	return address.Hex(), nil
}

func PrivateKeyToGoatAddress(privateKeyHex string) (string, error) {
	sdkConfig := sdk.GetConfig()
	sdkConfig.SetBech32PrefixForAccount(config.AppConfig.GoatChainAccountPrefix, config.AppConfig.GoatChainAccountPrefix+sdk.PrefixPublic)

	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		log.Errorf("Failed to decode private key: %v", err)
		return "", err
	}
	privateKey := &secp256k1.PrivKey{Key: privateKeyBytes}
	return sdk.AccAddress(privateKey.PubKey().Address().Bytes()).String(), nil
}
