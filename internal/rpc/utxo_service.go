package rpc

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/goatnetwork/goat-relayer/internal/btc"
	"github.com/goatnetwork/goat-relayer/internal/types"
	relayertypes "github.com/goatnetwork/goat/x/relayer/types"

	"net"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	"github.com/goatnetwork/goat-relayer/internal/config"
	"github.com/goatnetwork/goat-relayer/internal/layer2"
	"github.com/goatnetwork/goat-relayer/internal/p2p"
	"github.com/goatnetwork/goat-relayer/internal/state"
	pb "github.com/goatnetwork/goat-relayer/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	log "github.com/sirupsen/logrus"
)

type UtxoServer struct {
	pb.UnimplementedBitcoinLightWalletServer
	state          *state.State
	layer2Listener *layer2.Layer2Listener
}

func (s *UtxoServer) Start(ctx context.Context) {
	addr := ":" + config.AppConfig.RPCPort
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterBitcoinLightWalletServer(server, s)
	reflection.Register(server)

	log.Infof("gRPC server is running on port %s", config.AppConfig.RPCPort)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func NewUtxoServer(state *state.State, layer2Listener *layer2.Layer2Listener) *UtxoServer {
	return &UtxoServer{
		state:          state,
		layer2Listener: layer2Listener,
	}
}

func (s *UtxoServer) NewTransaction(ctx context.Context, req *pb.NewTransactionRequest) (*pb.NewTransactionResponse, error) {
	rawTxBytes, err := hex.DecodeString(req.RawTransaction)
	if err != nil {
		return nil, err
	}

	var tx wire.MsgTx
	if err := tx.Deserialize(bytes.NewReader(rawTxBytes)); err != nil {
		log.Errorf("Failed to decode transaction: %v", err)
		return nil, err
	}

	if err := btc.VerifyTransaction(tx, req.TransactionId, req.EvmAddress); err != nil {
		log.Errorf("Failed to verify transaction: %v", err)
		return nil, err
	}

	err = s.state.AddUnconfirmDeposit(req.TransactionId, req.RawTransaction, req.EvmAddress)
	if err != nil {
		log.Errorf("Failed to add unconfirmed deposit: %v", err)
		return nil, err
	}

	deposit := types.MsgUtxoDeposit{
		RawTx:     req.RawTransaction,
		TxId:      req.TransactionId,
		EvmAddr:   req.EvmAddress,
		Timestamp: time.Now().Unix(),
	}

	p2p.PublishMessage(context.Background(), p2p.Message{
		MessageType: p2p.MessageTypeDepositReceive,
		RequestId:   fmt.Sprintf("DEPOSIT:%s:%s", config.AppConfig.RelayerAddress, deposit.TxId),
		DataType:    "MsgUtxoDeposit",
		Data:        deposit,
	})

	return &pb.NewTransactionResponse{
		ErrorMessage: "Confirming transaction",
	}, nil
}

func (s *UtxoServer) QueryDepositAddress(ctx context.Context, req *pb.QueryDepositAddressRequest) (*pb.QueryDepositAddressResponse, error) {
	l2Info := s.state.GetL2Info()

	var err error
	var pubKey []byte
	if l2Info.DepositKey == "" {
		// query from layer2 goat chain
		pubkeyResponse := s.layer2Listener.QueryPubKey(ctx)
		pubKey = relayertypes.EncodePublicKey(&pubkeyResponse.PublicKey)
	} else {
		pubKey, err = hex.DecodeString(l2Info.DepositKey)
		if err != nil {
			return nil, err
		}
	}

	var newPubKey []byte
	if len(pubKey)-1 == btcec.PubKeyBytesLenCompressed {
		newPubKey = relayertypes.DecodePublicKey(pubKey).GetSecp256K1()
	} else if len(pubKey)-1 == schnorr.PubKeyBytesLen {
		newPubKey = relayertypes.DecodePublicKey(pubKey).GetSchnorr()
	} else {
		return nil, errors.New("invalid public key")
	}

	var network *chaincfg.Params
	switch config.AppConfig.BTCNetworkType {
	case "":
		network = &chaincfg.MainNetParams
	case "mainnet":
		network = &chaincfg.MainNetParams
	case "regtest":
		network = &chaincfg.RegressionNetParams
	case "testnet3":
		network = &chaincfg.TestNet3Params
	}

	p2wpkh, err := btcutil.NewAddressWitnessPubKeyHash(btcutil.Hash160(newPubKey), network)
	if err != nil {
		return nil, err
	}

	return &pb.QueryDepositAddressResponse{
		DepositAddress: p2wpkh.EncodeAddress(),
	}, nil
}
