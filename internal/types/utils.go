package types

import (
	"encoding/hex"
	"slices"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/crypto"
	log "github.com/sirupsen/logrus"
)

func DecodeBtcHash(hash string) ([]byte, error) {
	data, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}
	txid := slices.Clone(data)
	slices.Reverse(txid)
	return txid, nil
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

func PrivateKeyToGoatAddress(privateKeyHex string, accountPrefix string) (string, error) {
	sdkConfig := sdk.GetConfig()
	sdkConfig.SetBech32PrefixForAccount(accountPrefix, accountPrefix+sdk.PrefixPublic)

	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		log.Errorf("Failed to decode private key: %v", err)
		return "", err
	}
	privateKey := &secp256k1.PrivKey{Key: privateKeyBytes}
	return sdk.AccAddress(privateKey.PubKey().Address().Bytes()).String(), nil
}
