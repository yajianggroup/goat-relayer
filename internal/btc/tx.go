package btc

import (
	"context"
	"log"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/goatnetwork/goat-relayer/internal/config"
	bitcointypes "github.com/goatnetwork/goat/x/bitcoin/types"
	relayertypes "github.com/goatnetwork/goat/x/relayer/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func SendBlockData(block *btcutil.Block) {
	// Connect to gRPC server
	grpcConn, err := grpc.NewClient(
		config.AppConfig.L2RPC,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Printf("Unable to connect to gRPC server: %v", err)
		return
	}
	defer grpcConn.Close()

	// Initialize TxConfig
	encodingConfig := makeEncodingConfig()

	// Initialize keyring
	kr, err := keyring.New(
		"goat",
		keyring.BackendTest,
		config.AppConfig.DbDir,
		nil,
		encodingConfig.Codec,
	)
	if err != nil {
		log.Printf("Unable to initialize keyring: %v", err)
		return
	}

	clientCtx := client.Context{}.
		WithGRPCClient(grpcConn).
		WithCodec(encodingConfig.Codec).
		WithTxConfig(encodingConfig.TxConfig).
		WithKeyring(kr).
		WithChainID(config.AppConfig.L2ChainId.String())

	// Create MsgNewBlockHashes message
	msg := &bitcointypes.MsgNewBlockHashes{
		Proposer:         "goat1...", // TODO: get from tss
		StartBlockNumber: uint64(block.Height()),
		BlockHash:        [][]byte{block.Hash()[:]},
		Vote: &relayertypes.Votes{
			Signature: []byte("signature"), // TODO: get from tss
		},
	}

	// Create transaction
	txFactory := tx.Factory{}.
		WithChainID(clientCtx.ChainID).
		WithGas(200000).
		WithFees("1000ugoat").
		WithKeybase(kr).
		WithAccountRetriever(clientCtx.AccountRetriever).
		WithTxConfig(encodingConfig.TxConfig)

	// Build and sign transaction
	txBuilder, err := txFactory.BuildUnsignedTx(msg)
	if err != nil {
		log.Printf("Unable to build transaction: %v", err)
		return
	}

	err = tx.Sign(context.Background(), txFactory, kr.Backend(), txBuilder, true)
	if err != nil {
		log.Printf("Unable to sign transaction: %v", err)
		return
	}

	// Broadcast transaction
	txBytes, err := clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		log.Printf("Unable to encode transaction: %v", err)
		return
	}

	res, err := clientCtx.BroadcastTx(txBytes)
	if err != nil {
		log.Printf("Unable to broadcast transaction: %v", err)
		return
	}

	log.Printf("Transaction successfully broadcasted: %s", res.TxHash)
}

func SendDepositData(deposit *bitcointypes.Deposit) {
	// Connect to gRPC server
	grpcConn, err := grpc.NewClient(
		config.AppConfig.L2RPC,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Printf("Unable to connect to gRPC server: %v", err)
		return
	}
	defer grpcConn.Close()

	// Initialize TxConfig
	encodingConfig := makeEncodingConfig()

	// Initialize keyring
	kr, err := keyring.New(
		"goat",
		keyring.BackendTest,
		config.AppConfig.DbDir,
		nil,
		encodingConfig.Codec,
	)
	if err != nil {
		log.Printf("Unable to initialize keyring: %v", err)
		return
	}

	clientCtx := client.Context{}.
		WithGRPCClient(grpcConn).
		WithCodec(encodingConfig.Codec).
		WithTxConfig(encodingConfig.TxConfig).
		WithKeyring(kr).
		WithChainID(config.AppConfig.L2ChainId.String())

	// Create MsgNewDeposit message
	msg := &bitcointypes.MsgNewDeposits{
		Proposer: "goat1...", // TODO: get from tss
		Deposits: []*bitcointypes.Deposit{deposit},
		BlockHeaders: map[uint64][]byte{
			uint64(deposit.OutputIndex): deposit.NoWitnessTx,
		},
	}

	// Create transaction
	txFactory := tx.Factory{}.
		WithChainID(clientCtx.ChainID).
		WithGas(200000).
		WithFees("1000ugoat").
		WithKeybase(kr).
		WithAccountRetriever(clientCtx.AccountRetriever).
		WithTxConfig(encodingConfig.TxConfig)

	// Build and sign transaction
	txBuilder, err := txFactory.BuildUnsignedTx(msg)
	if err != nil {
		log.Printf("Unable to build transaction: %v", err)
		return
	}

	err = tx.Sign(context.Background(), txFactory, kr.Backend(), txBuilder, true)
	if err != nil {
		log.Printf("Unable to sign transaction: %v", err)
		return
	}

	// Broadcast transaction
	txBytes, err := clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		log.Printf("Unable to encode transaction: %v", err)
		return
	}

	res, err := clientCtx.BroadcastTx(txBytes)
	if err != nil {
		log.Printf("Unable to broadcast transaction: %v", err)
		return
	}

	log.Printf("Transaction successfully broadcasted: %s", res.TxHash)
}

func makeEncodingConfig() EncodingConfig {
	interfaceRegistry := types.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)
	return EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             marshaler,
		TxConfig:          authtx.NewTxConfig(marshaler, authtx.DefaultSignModes),
	}
}

type EncodingConfig struct {
	InterfaceRegistry types.InterfaceRegistry
	Codec             codec.Codec
	TxConfig          client.TxConfig
}
