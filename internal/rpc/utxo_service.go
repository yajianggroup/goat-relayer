package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"google.golang.org/grpc/credentials/insecure"
	"net"

	"github.com/btcsuite/btcd/wire"
	"github.com/goatnetwork/goat-relayer/internal/btc"
	"github.com/goatnetwork/goat-relayer/internal/config"
	pb "github.com/goatnetwork/goat-relayer/proto"
	bitcointypes "github.com/goatnetwork/goat/x/bitcoin/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	log "github.com/sirupsen/logrus"
)

type UTXOService interface {
	StartUTXOService(btc.BTCListener)
}

type UTXOServiceImpl struct {
	btc.BTCListener
}

func (us *UTXOServiceImpl) StartUTXOService(btcListener btc.BTCListener) {
	us.BTCListener = btcListener

	addr := ":" + config.AppConfig.HTTPPort
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterBitcoinLightWalletServer(s, &UtxoServer{})
	reflection.Register(s)

	log.Infof("gRPC server is running on port %s", config.AppConfig.HTTPPort)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

type UtxoServer struct {
	pb.UnimplementedBitcoinLightWalletServer
}

func NewUtxoServer() *UtxoServer {
	return &UtxoServer{}
}

func (s *UtxoServer) NewTransaction(ctx context.Context, req *pb.NewTransactionRequest) (*pb.NewTransactionResponse, error) {
	var tx wire.MsgTx
	if err := json.NewDecoder(bytes.NewReader(req.RawTransaction)).Decode(&tx); err != nil {
		log.Errorf("Failed to decode transaction: %v", err)
		return nil, err
	}

	_, err := btc.GenerateSPVProof(&tx)
	if err != nil {
		log.Errorf("Failed to generate SPV proof: %v", err)
		return nil, err
	}

	// TODO: Send the transaction to the consensus layer

	return &pb.NewTransactionResponse{
		TransactionId: "txhash",
		ErrorMessage:  "",
	}, nil
}

func (s *UtxoServer) QueryDepositAddress(ctx context.Context, req *pb.QueryDepositAddressRequest) (*pb.QueryDepositAddressResponse, error) {
	// Connect to grpc server
	grpcConn, err := grpc.NewClient(
		config.AppConfig.GoatChainGRPCURI,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Errorf("unable to connect to goat chain grpc server: %v", err)
		return nil, err
	}
	defer grpcConn.Close()

	pubKeyRequest := bitcointypes.QueryPubkeyRequest{}
	client := bitcointypes.NewQueryClient(grpcConn)
	_, err = client.Pubkey(ctx, &pubKeyRequest)
	if err != nil {
		log.Errorf("query pubkey error: %v", err)
		return nil, err
	}

	return &pb.QueryDepositAddressResponse{
		// todo return the btc deposit address
		DepositAddress: "address",
	}, nil
}
