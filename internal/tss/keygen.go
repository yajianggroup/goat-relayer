package tss

import (
	"context"
	"encoding/json"

	"github.com/bnb-chain/tss-lib/v2/ecdsa/keygen"
	tsslib "github.com/bnb-chain/tss-lib/v2/tss"
	log "github.com/sirupsen/logrus"
)

func (ts *TSSServiceImpl) HandleKeygenMessages(ctx context.Context, inCh chan KeygenMessage, outCh chan tsslib.Message, endCh chan *keygen.LocalPartySaveData) {
	parties := 3
	threshold := 2
	partyIDs := createPartyIDs(parties)
	peerCtx := tsslib.NewPeerContext(partyIDs)
	params := tsslib.NewParameters(tsslib.S256(), peerCtx, partyIDs[0], parties, threshold)

	party := keygen.NewLocalParty(params, outCh, endCh)
	if err := party.Start(); err != nil {
		log.Errorf("TSS keygen process failed to start: %v", err)
		return
	}

	for {
		select {
		case msg := <-inCh:
			var tssMsg tsslib.Message
			if err := json.Unmarshal([]byte(msg.Content), &tssMsg); err != nil {
				log.Errorf("Failed to unmarshal TSS keygen message: %v", err)
				continue
			}
			// TODO: handle the message
		case <-ctx.Done():
			return
		}
	}
}
