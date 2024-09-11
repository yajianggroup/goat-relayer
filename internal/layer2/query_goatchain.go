package layer2

import (
	"context"
	relayertypes "github.com/goatnetwork/goat/x/relayer/types"
	log "github.com/sirupsen/logrus"
)

func (lis *Layer2Listener) QueryRelayer(ctx context.Context) *relayertypes.QueryRelayerResponse {
	client := relayertypes.NewQueryClient(lis.goatGrpcConn)
	response, err := client.Relayer(ctx, &relayertypes.QueryRelayerRequest{})
	if err != nil {
		log.Errorf("Error while querying relayer status: %v", err)
	}

	return response
}
