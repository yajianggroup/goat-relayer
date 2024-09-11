package common

import (
	"encoding/hex"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/goatnetwork/goat-relayer/internal/config"
	log "github.com/sirupsen/logrus"
)

func PrivateKeyToGethAddress(privateKeyHex string) string {
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		log.Errorf("Failed to decode private key: %v", err)
	}

	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		log.Errorf("Failed to parse private key: %v", err)
	}

	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	return address.Hex()
}

func PrivateKeyToGoatAddress(privateKeyHex string) string {
	sdkConfig := sdk.GetConfig()
	sdkConfig.SetBech32PrefixForAccount(config.AppConfig.GoatChainAccountPrefix, config.AppConfig.GoatChainAccountPrefix+sdk.PrefixPublic)

	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		log.Errorf("Failed to decode private key: %v", err)
	}
	privateKey := &secp256k1.PrivKey{Key: privateKeyBytes}

	address := sdk.AccAddress(privateKey.PubKey().Address().Bytes()).String()

	return address

}
