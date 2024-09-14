package layer2

import (
	"context"

	relayertypes "github.com/goatnetwork/goat/x/relayer/types"
	log "github.com/sirupsen/logrus"
)

func (lis *Layer2Listener) QueryRelayer(ctx context.Context) (*relayertypes.QueryRelayerResponse, error) {
	if err := lis.checkAndReconnect(); err != nil {
		log.Errorf("check and reconnect goat client faild: %v", err)
		return nil, err
	}
	client := relayertypes.NewQueryClient(lis.goatGrpcConn)
	response, err := client.Relayer(ctx, &relayertypes.QueryRelayerRequest{})
	if err != nil {
		log.Errorf("Error while querying relayer status: %v", err)
		return nil, err
	}

	return response, nil
}

func (lis *Layer2Listener) QueryVotersOfRelayer(ctx context.Context) (*relayertypes.QueryVotersResponse, error) {
	if err := lis.checkAndReconnect(); err != nil {
		log.Errorf("check and reconnect goat client faild: %v", err)
		return nil, err
	}
	client := relayertypes.NewQueryClient(lis.goatGrpcConn)
	response, err := client.Voters(ctx, &relayertypes.QueryVotersRequest{})
	if err != nil {
		log.Errorf("Error while querying voters status: %v", err)
		return nil, err
	}

	return response, nil
}
