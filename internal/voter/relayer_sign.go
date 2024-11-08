package voter

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cosmossdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/goatnetwork/goat-relayer/internal/bls"
	"github.com/goatnetwork/goat-relayer/internal/config"
	"github.com/goatnetwork/goat-relayer/internal/db"
	"github.com/goatnetwork/goat-relayer/internal/p2p"
	"github.com/goatnetwork/goat-relayer/internal/state"
	"github.com/goatnetwork/goat-relayer/internal/types"
	goatcryp "github.com/goatnetwork/goat/pkg/crypto"
	relayertypes "github.com/goatnetwork/goat/x/relayer/types"
	log "github.com/sirupsen/logrus"
)

var (
	_ types.BlsSignProcessor = (*RelayerSignProcessor)(nil)
)

type RelayerSignProcessor struct {
	libp2p *p2p.LibP2PService
	state  *state.State
	signer *bls.Signer
	once   sync.Once

	sigMu          sync.Mutex
	sigStatus      bool
	sigFailChan    chan interface{}
	sigFinishChan  chan interface{}
	sigTimeoutChan chan interface{}
}

func NewRelayerSignProcessor(libp2p *p2p.LibP2PService, state *state.State, signer *bls.Signer) types.BlsSignProcessor {
	return &RelayerSignProcessor{
		libp2p: libp2p,
		state:  state,
		signer: signer,

		sigFailChan:    make(chan interface{}, 10),
		sigFinishChan:  make(chan interface{}, 10),
		sigTimeoutChan: make(chan interface{}, 10),
	}
}

func (p *RelayerSignProcessor) Start(ctx context.Context) {
	p.state.EventBus.Subscribe(state.SigFailed, p.sigFailChan)
	p.state.EventBus.Subscribe(state.SigFinish, p.sigFinishChan)
	p.state.EventBus.Subscribe(state.SigTimeout, p.sigTimeoutChan)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.Stop()
			return
		case sigFail := <-p.sigFailChan:
			p.HandleSigFailed(sigFail)
		case sigTimeout := <-p.sigTimeoutChan:
			p.HandleSigTimeout(sigTimeout)
		case sigFinish := <-p.sigFinishChan:
			p.HandleSigFinish(sigFinish)
		case <-ticker.C:
			p.BeginSig()
		}
	}
}

func (p *RelayerSignProcessor) Stop() {
	p.once.Do(func() {
		close(p.sigFailChan)
		close(p.sigFinishChan)
		close(p.sigTimeoutChan)
	})
}

func (p *RelayerSignProcessor) BeginSig() {
	p.beginSigNewVoter()
}

func (p *RelayerSignProcessor) beginSigNewVoter() {
	p.sigMu.Lock()
	defer p.sigMu.Unlock()

	if p.sigStatus {
		log.Debug("RelayerSignProcessor BeginSig ignore, already in sig")
		return
	}

	epochVoter := p.state.GetEpochVoter()
	voterAll := strings.Split(epochVoter.VoteAddrList, ",")
	if len(voterAll) > 0 && types.IndexOfSlice(voterAll, config.AppConfig.RelayerAddress) >= 0 {
		log.Debug("RelayerSignProcessor BeginSig ignore, self is not new voter")
		return
	}
	// check if voter queue contains self
	voterQueue := p.state.GetL2VoterQueue()
	if len(voterQueue) == 0 {
		log.Debug("RelayerSignProcessor BeginSig ignore, voter queue is empty")
		return
	}
	found := false
	var foundVoter *db.VoterQueue
	for _, voter := range voterQueue {
		if strings.EqualFold(voter.VoteAddr, config.AppConfig.RelayerAddress) {
			found = true
			foundVoter = voter
			break
		}
	}
	if !found {
		log.Debug("RelayerSignProcessor BeginSig ignore, self is not in voter queue")
		return
	}
	// start sig
	requestId := fmt.Sprintf("NEWVOTER:%s", config.AppConfig.RelayerAddress)
	privKeyBytes, err := hex.DecodeString(config.AppConfig.RelayerPriKey)
	if err != nil {
		log.Fatalf("RelayerSignProcessor BeginSig decode private key failed: %v", err)
	}
	privKey := &secp256k1.PrivKey{Key: privKeyBytes}
	addrRaw := cosmossdktypes.AccAddress(privKey.PubKey().Address().Bytes())
	blsBytes, err := hex.DecodeString(config.AppConfig.RelayerBlsSk)
	if err != nil {
		log.Fatalf("Decode bls sk error: %v", err)
	}

	blsSk := new(goatcryp.PrivateKey).Deserialize(blsBytes)
	blsPk := new(goatcryp.PublicKey).From(blsSk).Compress()
	reqMsg := relayertypes.NewOnBoardingVoterRequest(foundVoter.Epoch, addrRaw, blsPk)
	sigMsg := relayertypes.VoteSignDoc(
		reqMsg.MethodName(), config.AppConfig.GoatChainID, epochVoter.Proposer, 0 /* sequence */, epochVoter.Epoch, reqMsg.SignDoc())
	txKeyProof, err := privKey.Sign(sigMsg)
	if err != nil {
		log.Fatalf("RelayerSignProcessor BeginSig sign tx key proof failed: %v", err)
	}
	blsKeyProof := goatcryp.Sign(blsSk, sigMsg)
	msgSignNewVoter := types.MsgSignNewVoter{
		MsgSign: types.MsgSign{
			RequestId:  requestId,
			Epoch:      epochVoter.Epoch,
			SigData:    sigMsg,
			CreateTime: time.Now().Unix(),
		},

		Proposer:         epochVoter.Proposer,
		VoterTxKey:       privKey.PubKey().Bytes(),
		VoterBlsKey:      blsPk,
		VoterBlsKeyProof: blsKeyProof,
		VoterTxKeyProof:  txKeyProof,
	}
	p.state.EventBus.Publish(state.SigStart, msgSignNewVoter)
	p.sigStatus = true
}

func (p *RelayerSignProcessor) HandleSigFinish(msgSign interface{}) {
	// never happen for new voter
}

func (p *RelayerSignProcessor) HandleSigFailed(msgSign interface{}) {
	p.sigMu.Lock()
	defer p.sigMu.Unlock()

	if !p.sigStatus {
		log.Debug("Event handleSigFailed ignore, sigStatus is false")
		return
	}

	switch e := msgSign.(type) {
	case types.MsgSignNewVoter:
		log.Infof("Event handleSigFailed is of type MsgSignNewVoter, request id %s", e.RequestId)
		p.sigStatus = false
	default:
		log.Debug("RelayerSignProcessor ignore unsupport type")
	}
}

func (p *RelayerSignProcessor) HandleSigTimeout(msgSign interface{}) {
	p.sigMu.Lock()
	defer p.sigMu.Unlock()

	if !p.sigStatus {
		log.Debug("Event handleSigTimeout ignore, sigStatus is false")
		return
	}

	switch e := msgSign.(type) {
	case types.MsgSignNewVoter:
		log.Infof("Event handleSigTimeout is of type MsgSignNewVoter, request id %s", e.RequestId)
		p.sigStatus = false
	default:
		log.Debug("RelayerSignProcessor ignore unsupport type")
	}
}
