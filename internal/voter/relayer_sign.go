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
	"github.com/ethereum/go-ethereum/crypto"
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

const (
	// Secp256k1CompressedPubKeyLength is the length of a compressed secp256k1 public key
	Secp256k1CompressedPubKeyLength = 33
	// BLSCompressedPubKeyLength is the length of a compressed BLS public key
	BLSCompressedPubKeyLength = 96
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
	addrRaw := cosmossdktypes.AccAddress(goatcryp.Hash160Sum(privKey.PubKey().Bytes()))
	blsBytes, err := hex.DecodeString(config.AppConfig.RelayerBlsSk)
	if err != nil {
		log.Fatalf("Decode bls sk error: %v", err)
	}

	blsSk := new(goatcryp.PrivateKey).Deserialize(blsBytes)
	blsPk := new(goatcryp.PublicKey).From(blsSk).Compress()
	blsKeyHash := goatcryp.SHA256Sum(blsPk)
	reqMsg := relayertypes.NewOnBoardingVoterRequest(foundVoter.Epoch, addrRaw, blsKeyHash)
	sigMsg := relayertypes.VoteSignDoc(
		reqMsg.MethodName(), config.AppConfig.GoatChainID, epochVoter.Proposer, 0 /* sequence */, epochVoter.Epoch, reqMsg.SignDoc())
	// Use go-ethereum's signing function to generate a 65-byte signature that includes the recovery ID,
	// then remove the last byte (recovery ID) to convert it to a 64-byte signature,
	// to meet the requirements of crypto.VerifySignature for signature verification.
	ecdsaPriv, err := crypto.ToECDSA(privKeyBytes)
	if err != nil {
		log.Fatalf("RelayerSignProcessor BeginSig convert private key failed: %v", err)
	}
	sig65, err := crypto.Sign(sigMsg, ecdsaPriv)
	if err != nil {
		log.Fatalf("RelayerSignProcessor BeginSig sign tx key proof failed: %v", err)
	}
	txKeyProof := sig65[:64]
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

	// Validate signatures before submitting to P2P
	if !p.validateNewVoterSignatures(sigMsg, txKeyProof, blsKeyProof, privKey.PubKey().Bytes(), blsPk, addrRaw) {
		log.Errorf("RelayerSignProcessor BeginSig signature validation failed, aborting P2P submission")
		return
	}

	log.Infof("RelayerSignProcessor BeginSig all validations passed, submitting to P2P")
	p.state.EventBus.Publish(state.SigStart, msgSignNewVoter)
	p.sigStatus = true
}

// validateNewVoterSignatures validates all signatures before P2P submission
func (p *RelayerSignProcessor) validateNewVoterSignatures(sigMsg []byte, txKeyProof []byte, blsKeyProof []byte, voterTxKey []byte, voterBlsKey []byte, addrRaw cosmossdktypes.AccAddress) bool {
	log.Infof("RelayerSignProcessor BeginSig starting signature validation...")

	// Validate TX Key Proof
	txKeyValid := crypto.VerifySignature(voterTxKey, sigMsg, txKeyProof)
	if !txKeyValid {
		log.Errorf("RelayerSignProcessor BeginSig TX Key Proof validation failed")
		log.Errorf("Signature document: %s", hex.EncodeToString(sigMsg))
		log.Errorf("TX Key Proof: %s", hex.EncodeToString(txKeyProof))
		log.Errorf("Voter TX Key: %s", hex.EncodeToString(voterTxKey))
		return false
	}
	log.Infof("RelayerSignProcessor BeginSig TX Key Proof validation successful")

	// Validate BLS Key Proof
	blsKeyValid := goatcryp.Verify(voterBlsKey, sigMsg, blsKeyProof)
	if !blsKeyValid {
		log.Errorf("RelayerSignProcessor BeginSig BLS Key Proof validation failed")
		log.Errorf("Signature document: %s", hex.EncodeToString(sigMsg))
		log.Errorf("BLS Key Proof: %s", hex.EncodeToString(blsKeyProof))
		log.Errorf("Voter BLS Key: %s", hex.EncodeToString(voterBlsKey))
		return false
	}
	log.Infof("RelayerSignProcessor BeginSig BLS Key Proof validation successful")

	// Validate address calculation
	calculatedAddr := cosmossdktypes.AccAddress(goatcryp.Hash160Sum(voterTxKey))
	log.Infof("RelayerSignProcessor BeginSig calculated address: %s", calculatedAddr.String())
	log.Infof("RelayerSignProcessor BeginSig expected address: %s", addrRaw.String())

	if !strings.EqualFold(calculatedAddr.String(), addrRaw.String()) {
		log.Errorf("RelayerSignProcessor BeginSig address calculation error")
		return false
	}
	log.Infof("RelayerSignProcessor BeginSig address validation successful")

	// Validate parameter format
	if len(voterTxKey) != Secp256k1CompressedPubKeyLength {
		log.Errorf("RelayerSignProcessor BeginSig Voter TX Key length error: expected %d bytes, got %d bytes", Secp256k1CompressedPubKeyLength, len(voterTxKey))
		return false
	}
	if len(voterBlsKey) != BLSCompressedPubKeyLength {
		log.Errorf("RelayerSignProcessor BeginSig Voter BLS Key length error: expected %d bytes, got %d bytes", BLSCompressedPubKeyLength, len(voterBlsKey))
		return false
	}
	log.Infof("RelayerSignProcessor BeginSig parameter format validation successful")

	return true
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
