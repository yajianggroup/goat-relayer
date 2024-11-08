// handle_block.go handle btc new block bls sig
package bls

import (
	"context"
	"fmt"
	"time"

	"github.com/goatnetwork/goat-relayer/internal/config"
	"github.com/goatnetwork/goat-relayer/internal/layer2"
	"github.com/goatnetwork/goat-relayer/internal/p2p"
	"github.com/goatnetwork/goat-relayer/internal/state"
	"github.com/goatnetwork/goat-relayer/internal/types"
	relayertypes "github.com/goatnetwork/goat/x/relayer/types"
	log "github.com/sirupsen/logrus"
)

func (s *Signer) handleSigStartNewVoter(ctx context.Context, e types.MsgSignNewVoter) error {
	canSign := s.CanSign()
	if !canSign {
		log.Debugf("Ignore SigStart request id %s, canSign: %v", e.RequestId, canSign)
		return fmt.Errorf("cannot start sig %s in current l2 context, catching up: %v", e.RequestId, !canSign)
	}

	// request id format: BtcHead:VoterAddr:StartBlockNumber
	// check map
	_, ok := s.sigExists(e.RequestId)
	if ok {
		return fmt.Errorf("sig exists: %s", e.RequestId)
	}

	// build sign
	newSign := &types.MsgSignNewVoter{
		MsgSign: types.MsgSign{
			RequestId:    e.RequestId,
			Sequence:     e.Sequence,
			Epoch:        e.Epoch,
			IsProposer:   false,
			VoterAddress: s.address, // proposer address
			SigData:      e.SigData,
			CreateTime:   time.Now().Unix(),
		},
		Proposer:         e.Proposer,
		VoterTxKey:       e.VoterTxKey,
		VoterTxKeyProof:  e.VoterTxKeyProof,
		VoterBlsKey:      e.VoterBlsKey,
		VoterBlsKeyProof: e.VoterBlsKeyProof,
	}

	// p2p broadcast
	p2pMsg := p2p.Message[any]{
		MessageType: p2p.MessageTypeSigReq,
		RequestId:   e.RequestId,
		DataType:    "MsgSignNewVoter",
		Data:        *newSign,
	}
	if err := p2p.PublishMessage(ctx, p2pMsg); err != nil {
		log.Errorf("SigStart public NewVoter to p2p error, request id: %s, err: %v", e.RequestId, err)
		return err
	}

	s.sigMu.Lock()
	s.sigMap[e.RequestId] = make(map[string]interface{})
	s.sigMap[e.RequestId][s.address] = *newSign
	timeoutDuration := config.AppConfig.BlsSigTimeout
	s.sigTimeoutMap[e.RequestId] = time.Now().Add(timeoutDuration)
	s.sigMu.Unlock()
	log.Infof("SigStart broadcast ok, request id: %s", e.RequestId)

	return nil
}

func (s *Signer) handleSigReceiveNewVoter(ctx context.Context, e types.MsgSignNewVoter) error {
	canSign := s.CanSign()
	isProposer := s.IsProposer()
	if !canSign || !isProposer {
		log.Debugf("Ignore SigReceive request id %s, canSign: %v, isProposer: %v", e.RequestId, canSign, isProposer)
		return fmt.Errorf("cannot handle receive sig %s in current l2 context, catching up: %v, is proposer: %v", e.RequestId, !canSign, isProposer)
	}

	// collect voter sig
	if e.IsProposer {
		return nil
	}
	if e.Proposer != s.address {
		// not expected proposer
		return fmt.Errorf("not expected proposer: %s, request id: %s", e.Proposer, e.RequestId)
	}

	rpcMsg := &relayertypes.MsgNewVoterRequest{
		Proposer:         e.Proposer,
		VoterTxKey:       e.VoterTxKey,
		VoterTxKeyProof:  e.VoterTxKeyProof,
		VoterBlsKey:      e.VoterBlsKey,
		VoterBlsKeyProof: e.VoterBlsKeyProof,
	}

	newProposal := layer2.NewProposal[*relayertypes.MsgNewVoterRequest](s.layer2Listener)
	err := newProposal.RetrySubmit(ctx, e.RequestId, rpcMsg, config.AppConfig.L2SubmitRetry)
	if err != nil {
		log.Errorf("SigReceive proposer submit NewBlock to RPC error, request id: %s, err: %v", e.RequestId, err)
		s.removeSigMap(e.RequestId, false)
		return err
	}

	s.removeSigMap(e.RequestId, false)

	// feedback SigFinish
	s.state.EventBus.Publish(state.SigFinish, e)

	log.Infof("SigReceive proposer submit NewVoter to RPC ok, request id: %s", e.RequestId)
	return nil
}
