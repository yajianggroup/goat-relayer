package layer2

import (
	"context"
	"encoding/hex"
	"errors"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/goatnetwork/goat-relayer/internal/config"
	"google.golang.org/grpc/credentials/insecure"
	"time"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	txtypes "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bitcointypes "github.com/goatnetwork/goat/x/bitcoin/types"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func SubmitInfoToChain(msg interface{}) error {
	var err error
	var rpcClient *rpchttp.HTTP
	var grpcClient *grpc.ClientConn
	var authClient authtypes.QueryClient
	ctx := context.Background()
	accountPrefix := config.AppConfig.GoatChainAccountPrefix
	chainID := config.AppConfig.GoatChainID
	privKeyStr := config.AppConfig.RelayerPriKey
	rpcURI := config.AppConfig.GoatChainRPCURI
	grpcURI := config.AppConfig.GoatChainGRPCURI
	denom := config.AppConfig.GoatChainDenom

	sdkConfig := sdk.GetConfig()
	sdkConfig.SetBech32PrefixForAccount(accountPrefix, accountPrefix+sdk.PrefixPublic)
	sdkConfig.Seal()

	amino := codec.NewLegacyAmino()
	std.RegisterLegacyAminoCodec(amino)
	authtypes.RegisterLegacyAminoCodec(amino)

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	protoCodec := codec.NewProtoCodec(interfaceRegistry)
	authtypes.RegisterInterfaces(interfaceRegistry)
	std.RegisterInterfaces(interfaceRegistry)

	if rpcClient, err = rpchttp.New(rpcURI, "/websocket"); err != nil {
		log.Errorf("unable to connect to goat chain rpc server: %v", err)
		return err
	}

	if grpcClient, err = grpc.NewClient(grpcURI, grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
		log.Errorf("unable to connect to goat chain grpc server: %v", err)
		return err
	}

	authClient = authtypes.NewQueryClient(grpcClient)

	privKeyBytes, err := hex.DecodeString(privKeyStr)
	if err != nil {
		return err
	}
	privKey := &secp256k1.PrivKey{Key: privKeyBytes}
	address := sdk.AccAddress(privKey.PubKey().Address().Bytes()).String()

	accountReq := &authtypes.QueryAccountRequest{Address: address}
	accountResp, err := authClient.Account(ctx, accountReq)
	if err != nil {
		return err
	}

	var account sdk.AccountI
	if err = protoCodec.UnpackAny(accountResp.GetAccount(), &account); err != nil {
		return err
	}

	sequence := account.GetSequence()
	accountNumber := account.GetAccountNumber()

	txConfig := txtypes.NewTxConfig(protoCodec, txtypes.DefaultSignModes)
	txBuilder := txConfig.NewTxBuilder()

	if msgNewDeposits, msgNewBlockHashes, err := convertToTypes(msg); err != nil {
		return err
	} else if msgNewDeposits != nil {
		err = txBuilder.SetMsgs(msgNewDeposits)
		if err != nil {
			return err
		}
	} else if msgNewBlockHashes != nil {
		err = txBuilder.SetMsgs(msgNewBlockHashes)
		if err != nil {
			return err
		}
	}

	// set fee
	fees := sdk.NewCoins(sdk.NewInt64Coin(denom, 100))
	txBuilder.SetGasLimit(uint64(flags.DefaultGasLimit))
	txBuilder.SetFeeAmount(fees)

	if err = txBuilder.SetSignatures(signing.SignatureV2{
		PubKey: privKey.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode(txConfig.SignModeHandler().DefaultMode()),
			Signature: nil,
		},
		Sequence: sequence,
	}); err != nil {
		return err
	}

	signerData := authsigning.SignerData{
		ChainID:       chainID,
		AccountNumber: accountNumber,
		Sequence:      sequence,
	}
	sigV2, err := tx.SignWithPrivKey(ctx, signing.SignMode(txConfig.SignModeHandler().DefaultMode()), signerData, txBuilder, privKey, txConfig, sequence)
	if err != nil {
		return err
	}
	if err = txBuilder.SetSignatures(sigV2); err != nil {
		return err
	}

	tx := txBuilder.GetTx()
	txBytes, err := txConfig.TxEncoder()(tx)
	if err != nil {
		return err
	}

	serviceClient := sdktx.NewServiceClient(grpcClient)

	txResp, err := serviceClient.BroadcastTx(ctx, &sdktx.BroadcastTxRequest{
		Mode:    sdktx.BroadcastMode_BROADCAST_MODE_SYNC,
		TxBytes: txBytes},
	)
	if err != nil {
		return err
	}

	// wait tx to be included in a block
	time.Sleep(5 * time.Second)

	hashBytes, err := hex.DecodeString(txResp.TxResponse.TxHash)
	if err != nil {
		return err
	}

	resultTx, err := rpcClient.Tx(ctx, hashBytes, false)
	if err != nil {
		return err
	}

	// if code = 0, tx success
	if resultTx.TxResult.Code != 0 {
		log.Errorf("submit tx error: %v", resultTx.TxResult)
		return errors.New(resultTx.TxResult.Log)
	}

	return nil
}

func convertToTypes(msg interface{}) (*bitcointypes.MsgNewDeposits, *bitcointypes.MsgNewBlockHashes, error) {
	if msgNewDeposits, ok := msg.(*bitcointypes.MsgNewDeposits); ok {
		return msgNewDeposits, nil, nil
	}
	if msgNewBlockHashes, ok := msg.(*bitcointypes.MsgNewBlockHashes); ok {
		return nil, msgNewBlockHashes, nil
	}
	return nil, nil, errors.New("type assertion failed: not type MsgNewDeposits or MsgNewBlockHashes")
}
