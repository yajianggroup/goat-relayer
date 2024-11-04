package voter

import (
	"context"

	"github.com/goatnetwork/goat-relayer/internal/bls"
	"github.com/goatnetwork/goat-relayer/internal/p2p"
	"github.com/goatnetwork/goat-relayer/internal/state"
	"github.com/goatnetwork/goat-relayer/internal/types"
	log "github.com/sirupsen/logrus"
)

type VoterProcessor struct {
	libp2p *p2p.LibP2PService
	state  *state.State
	signer *bls.Signer

	relayerSignProcessor types.BlsSignProcessor
}

func NewVoterProcessor(libp2p *p2p.LibP2PService, state *state.State, signer *bls.Signer) *VoterProcessor {
	return &VoterProcessor{
		libp2p: libp2p,
		state:  state,
		signer: signer,

		relayerSignProcessor: NewRelayerSignProcessor(libp2p, state, signer),
	}
}

func (p *VoterProcessor) Start(ctx context.Context) {
	go p.relayerSignProcessor.Start(ctx)

	log.Info("VoterProcessor started.")

	<-ctx.Done()

	log.Info("VoterProcessor stopped.")
}
