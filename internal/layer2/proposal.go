package layer2

import (
	"context"
	"math/big"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/goatnetwork/goat-relayer/internal/config"
)

var (
	globalNonce uint64
	nonceLock   sync.Mutex
)

func initNonce(client *ethclient.Client, address common.Address) error {
	nonce, err := client.PendingNonceAt(context.Background(), address)
	if err != nil {
		return err
	}
	globalNonce = nonce
	return nil
}

func submitDepositInfo(targetAddress common.Address, amount *big.Int, txInfo, extraInfo, signature []byte, chainId *big.Int) error {
	client, parsedABI, votingManagerAddress := getClientAndAbi()

	auth, err := bind.NewKeyedTransactorWithChainID(config.AppConfig.L2PrivateKey, config.AppConfig.L2ChainId)
	if err != nil {
		return err
	}

	callData, err := parsedABI.Pack("submitDepositInfo", targetAddress, amount, txInfo, chainId, extraInfo, signature)
	if err != nil {
		return err
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}

	// TODO check max gas price

	msg := ethereum.CallMsg{
		From:     auth.From,
		To:       &votingManagerAddress,
		GasPrice: gasPrice,
		Data:     callData,
	}

	gasLimit, err := client.EstimateGas(context.Background(), msg)
	if err != nil {
		return err
	}

	nonceLock.Lock()
	if globalNonce == 0 {
		err = initNonce(client, auth.From)
		if err != nil {
			return err
		}
	}
	nonce := globalNonce
	nonceLock.Unlock()

	tx := types.NewTransaction(nonce, votingManagerAddress, big.NewInt(0), gasLimit, gasPrice, callData)

	signedTx, err := auth.Signer(auth.From, tx)
	if err != nil {
		return err
	}

	err = client.SendTransaction(context.Background(), signedTx)

	if err != nil {
		if strings.Contains(err.Error(), "nonce too low") {
			log.Error("Nonce too low error, retrying with updated nonce")
			latestNonce, err := client.PendingNonceAt(context.Background(), auth.From)
			if err != nil {
				return err
			}

			nonceLock.Lock()
			globalNonce = latestNonce
			nonceLock.Unlock()

			return submitDepositInfo(targetAddress, amount, txInfo, extraInfo, signature, chainId)
		}
		return err
	}

	nonceLock.Lock()
	globalNonce++
	nonceLock.Unlock()

	log.Infof("Transaction sent: %s", signedTx.Hash().Hex())
	return nil
}
