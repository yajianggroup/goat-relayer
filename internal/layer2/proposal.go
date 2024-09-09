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
	"github.com/goatnetwork/goat-relayer/internal/bls"
	"github.com/goatnetwork/goat-relayer/internal/db"
	"github.com/goatnetwork/goat-relayer/internal/p2p"
	"github.com/goatnetwork/goat-relayer/internal/state"
	bitcointypes "github.com/goatnetwork/goat/x/bitcoin/types"
	relayertypes "github.com/goatnetwork/goat/x/relayer/types"
	"github.com/kelindar/bitmap"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type Proposal struct {
	blsHelper  *bls.SignatureHelper
	state      *state.State
	rpcClient  *rpchttp.HTTP
	grpcClient *grpc.ClientConn
	authClient authtypes.QueryClient
}

func NewProposal(state *state.State, blsHelper *bls.SignatureHelper, p2pService *p2p.LibP2PService) *Proposal {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var err error
	var rpcClient *rpchttp.HTTP
	var grpcClient *grpc.ClientConn
	var authClient authtypes.QueryClient

	rpcURI := config.AppConfig.GoatChainRPCURI
	grpcURI := config.AppConfig.GoatChainGRPCURI

	if rpcClient, err = rpchttp.New(rpcURI, "/websocket"); err != nil {
		log.Errorf("unable to connect to goat chain rpc server: %v", err)
	}

	if grpcClient, err = grpc.NewClient(grpcURI, grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
		log.Errorf("unable to connect to goat chain grpc server: %v", err)
	}

	authClient = authtypes.NewQueryClient(grpcClient)

	p := &Proposal{
		blsHelper:  blsHelper,
		state:      state,
		rpcClient:  rpcClient,
		grpcClient: grpcClient,
		authClient: authClient,
	}

	btcHeadChan := make(chan interface{}, 100)
	state.GetEventBus().Subscribe("btcHeadStateUpdated", btcHeadChan)

	go p.handleBtcBlocks(ctx, btcHeadChan)

	return p
}

func (p *Proposal) handleBtcBlocks(ctx context.Context, btcHeadChan chan interface{}) {
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-btcHeadChan:
			block, ok := data.(db.BtcBlock)
			if !ok {
				continue
			}
			p.sendTxMsgNewBlockHashes(ctx, &block)
		}
	}
}

func (p *Proposal) sendTxMsgNewBlockHashes(ctx context.Context, block *db.BtcBlock) {
	voters := make(bitmap.Bitmap, 256)

	votes := &relayertypes.Votes{
		Sequence:  0,
		Epoch:     0,
		Voters:    voters.ToBytes(),
		Signature: nil,
	}

	msgBlock := bitcointypes.MsgNewBlockHashes{
		Proposer:         "",
		Vote:             votes,
		StartBlockNumber: block.Height,
		BlockHash:        [][]byte{[]byte(block.Hash)},
	}

	signature := p.blsHelper.SignDoc(ctx, msgBlock.VoteSigDoc())

	votes.Signature = signature.Compress()
	msgBlock.Vote = votes

	p.submitToConsensus(ctx, &msgBlock)
}

func (p *Proposal) submitToConsensus(ctx context.Context, msg interface{}) {
	var err error
	accountPrefix := config.AppConfig.GoatChainAccountPrefix
	chainID := config.AppConfig.GoatChainID
	privKeyStr := config.AppConfig.RelayerPriKey
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

	privKeyBytes, err := hex.DecodeString(privKeyStr)
	if err != nil {
		log.Error(err)
	}
	privKey := &secp256k1.PrivKey{Key: privKeyBytes}
	address := sdk.AccAddress(privKey.PubKey().Address().Bytes()).String()

	accountReq := &authtypes.QueryAccountRequest{Address: address}
	accountResp, err := p.authClient.Account(ctx, accountReq)
	if err != nil {
		log.Error(err)
	}

	var account sdk.AccountI
	if err = protoCodec.UnpackAny(accountResp.GetAccount(), &account); err != nil {
		log.Error(err)
	}

	sequence := account.GetSequence()
	accountNumber := account.GetAccountNumber()

	txConfig := txtypes.NewTxConfig(protoCodec, txtypes.DefaultSignModes)
	txBuilder := txConfig.NewTxBuilder()

	if msgNewDeposits, msgNewBlockHashes, err := convertToTypes(msg); err != nil {
		log.Error(err)
	} else if msgNewDeposits != nil {
		err = txBuilder.SetMsgs(msgNewDeposits)
		if err != nil {
			log.Error(err)
		}
	} else if msgNewBlockHashes != nil {
		err = txBuilder.SetMsgs(msgNewBlockHashes)
		if err != nil {
			log.Error(err)
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
		log.Error(err)
	}

	signerData := authsigning.SignerData{
		ChainID:       chainID,
		AccountNumber: accountNumber,
		Sequence:      sequence,
	}
	sigV2, err := tx.SignWithPrivKey(ctx, signing.SignMode(txConfig.SignModeHandler().DefaultMode()), signerData, txBuilder, privKey, txConfig, sequence)
	if err != nil {
		log.Error(err)
	}
	if err = txBuilder.SetSignatures(sigV2); err != nil {
		log.Error(err)
	}

	tx := txBuilder.GetTx()
	txBytes, err := txConfig.TxEncoder()(tx)
	if err != nil {
		log.Error(err)
	}

	serviceClient := sdktx.NewServiceClient(p.grpcClient)

	txResp, err := serviceClient.BroadcastTx(ctx, &sdktx.BroadcastTxRequest{
		Mode:    sdktx.BroadcastMode_BROADCAST_MODE_SYNC,
		TxBytes: txBytes},
	)
	if err != nil {
		log.Error(err)
	}

	// wait tx to be included in a block
	time.Sleep(5 * time.Second)

	hashBytes, err := hex.DecodeString(txResp.TxResponse.TxHash)
	if err != nil {
		log.Error(err)
	}

	resultTx, err := p.rpcClient.Tx(ctx, hashBytes, false)
	if err != nil {
		log.Error(err)
	}

	// if code = 0, tx success
	if resultTx.TxResult.Code != 0 {
		log.Errorf("submit tx error: %v", resultTx.TxResult)
	}
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
