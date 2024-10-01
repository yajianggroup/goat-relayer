// handle_wallet.go handle wallet send order bls sig
// contains withdrawal and consolidation
package bls

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/goatnetwork/goat-relayer/internal/config"
	"github.com/goatnetwork/goat-relayer/internal/db"
	"github.com/goatnetwork/goat-relayer/internal/p2p"
	"github.com/goatnetwork/goat-relayer/internal/state"
	"github.com/goatnetwork/goat-relayer/internal/types"
	goatcryp "github.com/goatnetwork/goat/pkg/crypto"
	bitcointypes "github.com/goatnetwork/goat/x/bitcoin/types"
	relayertypes "github.com/goatnetwork/goat/x/relayer/types"
	"github.com/kelindar/bitmap"
	log "github.com/sirupsen/logrus"
)

// handleSigStartSendOrder handle start send order sig event
func (s *Signer) handleSigStartSendOrder(ctx context.Context, e types.MsgSignSendOrder) error {
	canSign := s.CanSign()
	isProposer := s.IsProposer()
	if !canSign || !isProposer {
		log.Debugf("Ignore SigStart SendOrder request id %s, canSign: %v, isProposer: %v", e.RequestId, canSign, isProposer)
		log.Debugf("Current l2 context, catching up: %v, self address: %s, proposer: %s", s.state.GetL2Info().Syncing, s.address, s.state.GetEpochVoter().Proposer)
		return fmt.Errorf("cannot start sig %s in current l2 context, catching up: %v, is proposer: %v", e.RequestId, !canSign, isProposer)
	}

	// request id format: SENDORDER:VoterAddr:OrderId
	// check map
	_, ok := s.sigExists(e.RequestId)
	if ok {
		return fmt.Errorf("sig send order exists: %s", e.RequestId)
	}

	var order db.SendOrder
	err := json.Unmarshal(e.SendOrder, &order)
	if err != nil {
		log.Debug("Cannot unmarshal send order from msg")
		return err
	}

	// build sign
	newSign := &types.MsgSignSendOrder{
		MsgSign: types.MsgSign{
			RequestId:    e.RequestId,
			Sequence:     e.Sequence,
			Epoch:        e.Epoch,
			IsProposer:   true,
			VoterAddress: s.address, // proposer address
			SigData:      s.makeSigSendOrder(order.OrderType, e.WithdrawIds, order.NoWitnessTx, order.TxFee),
			CreateTime:   time.Now().Unix(),
		},
		SendOrder: e.SendOrder,
		Utxos:     e.Utxos,
		Vins:      e.Vins,
		Vouts:     e.Vouts,
		Withdraws: e.Withdraws,

		WithdrawIds: e.WithdrawIds,
	}

	// p2p broadcast
	p2pMsg := p2p.Message{
		MessageType: p2p.MessageTypeSigReq,
		RequestId:   e.RequestId,
		DataType:    "MsgSignSendOrder",
		Data:        *newSign,
	}
	if err := p2p.PublishMessage(ctx, p2pMsg); err != nil {
		log.Errorf("SigStart public MsgSignSendOrder to p2p error, request id: %s, err: %v", e.RequestId, err)
		return err
	}

	s.sigMu.Lock()
	s.sigMap[e.RequestId] = make(map[string]interface{})
	s.sigMap[e.RequestId][s.address] = *newSign
	timeoutDuration := config.AppConfig.BlsSigTimeout
	s.sigTimeoutMap[e.RequestId] = time.Now().Add(timeoutDuration)
	s.sigMu.Unlock()
	log.Infof("SigStart broadcast MsgSignSendOrder ok, request id: %s", e.RequestId)

	// If voters count is 1, should submit soon
	rpcMsg, err := s.aggSigSendOrder(e.RequestId)
	if err != nil {
		log.Warnf("SigStart proposer process MsgSignSendOrder aggregate sig, request id: %s, err: %v", e.RequestId, err)
		return nil
	}

	if order.OrderType == db.ORDER_TYPE_WITHDRAWAL {
		err = s.RetrySubmit(ctx, e.RequestId, rpcMsg, config.AppConfig.L2SubmitRetry)
		if err != nil {
			log.Errorf("SigStart proposer submit MsgSignSendOrder to RPC error, request id: %s, err: %v", e.RequestId, err)
			s.removeSigMap(e.RequestId, false)
			return err
		}
		log.Infof("SigStart proposer submit MsgSignSendOrder to RPC ok, request id: %s", e.RequestId)
	}

	s.removeSigMap(e.RequestId, false)

	// feedback SigFinish
	s.state.EventBus.Publish(state.SigFinish, e)
	return nil
}

func (s *Signer) makeSigSendOrder(orderType string, withdrawIds []uint64, noWitnessTx []byte, txFee uint64) []byte {
	voters := make(bitmap.Bitmap, 5)
	votes := &relayertypes.Votes{
		Sequence:  0,
		Epoch:     0,
		Voters:    voters.ToBytes(),
		Signature: nil,
	}
	epochVoter := s.state.GetEpochVoter()
	if orderType == db.ORDER_TYPE_WITHDRAWAL {
		msg := bitcointypes.MsgInitializeWithdrawal{
			Proposer: "",
			Vote:     votes,
			Proposal: &bitcointypes.WithdrawalProposal{
				Id:          withdrawIds,
				NoWitnessTx: noWitnessTx,
				TxFee:       txFee,
			},
		}
		sigDoc := relayertypes.VoteSignDoc(msg.MethodName(), config.AppConfig.GoatChainID, epochVoter.Proposer, epochVoter.Sequence, uint64(epochVoter.Epoch), msg.VoteSigDoc())
		return goatcryp.Sign(s.sk, sigDoc)
	} else {
		msg := bitcointypes.MsgNewConsolidation{
			Proposer:    "",
			Vote:        votes,
			NoWitnessTx: noWitnessTx,
		}
		sigDoc := relayertypes.VoteSignDoc(msg.MethodName(), config.AppConfig.GoatChainID, epochVoter.Proposer, epochVoter.Sequence, uint64(epochVoter.Epoch), msg.VoteSigDoc())
		return goatcryp.Sign(s.sk, sigDoc)
	}
}

func (s *Signer) aggSigSendOrder(requestId string) (interface{}, error) {
	epochVoter := s.state.GetEpochVoter()

	voteMap, ok := s.sigExists(requestId)
	if !ok {
		return nil, fmt.Errorf("no sig found of send order, request id: %s", requestId)
	}
	voterAll := strings.Split(epochVoter.VoteAddrList, ",")
	proposer := ""
	orderType := db.ORDER_TYPE_WITHDRAWAL
	var txFee, epoch, sequence uint64
	var noWitnessTx []byte
	var withdrawIds []uint64
	var bmp bitmap.Bitmap
	var proposerSig []byte
	voteSig := make([][]byte, 0)

	for address, msg := range voteMap {
		msgSendOrder := msg.(types.MsgSignSendOrder)
		var order db.SendOrder
		err := json.Unmarshal(msgSendOrder.SendOrder, &order)
		if err != nil {
			log.Debug("Cannot unmarshal send order from vote msg")
			return nil, err
		}
		if msgSendOrder.IsProposer {
			proposer = address // proposer address
			sequence = msgSendOrder.Sequence
			epoch = msgSendOrder.Epoch
			withdrawIds = msgSendOrder.WithdrawIds
			proposerSig = msgSendOrder.SigData

			txFee = order.TxFee
			noWitnessTx = order.NoWitnessTx
			orderType = order.OrderType
		} else {
			pos := indexOfSlice(voterAll, address) // voter address
			log.Debugf("Bitmap check, pos: %d, address: %s, all: %s", pos, address, epochVoter.VoteAddrList)
			if pos >= 0 {
				bmp.Set(uint32(pos))
				voteSig = append(voteSig, msgSendOrder.SigData)
			}
		}
	}

	if proposer == "" {
		return nil, fmt.Errorf("missing proposer sig msg of send order, request id: %s", requestId)
	}

	if epoch != epochVoter.Epoch {
		return nil, fmt.Errorf("incorrect epoch of send order, request id: %s, msg epoch: %d, current epoch: %d", requestId, epoch, epochVoter.Epoch)
	}
	if sequence != epochVoter.Sequence {
		return nil, fmt.Errorf("incorrect sequence of send order, request id: %s, msg sequence: %d, current sequence: %d", requestId, sequence, epochVoter.Sequence)
	}

	voteSig = append([][]byte{proposerSig}, voteSig...)

	// check threshold
	threshold := Threshold(len(voterAll))
	if len(voteSig) < threshold {
		return nil, fmt.Errorf("threshold not reach of send order, request id: %s, has sig: %d, threshold: %d", requestId, len(voteSig), threshold)
	}

	// aggregate
	aggSig, err := goatcryp.AggregateSignatures(voteSig)
	if err != nil {
		return nil, err
	}

	votes := &relayertypes.Votes{
		Sequence:  sequence,
		Epoch:     epoch,
		Voters:    bmp.ToBytes(),
		Signature: aggSig,
	}

	if orderType == db.ORDER_TYPE_WITHDRAWAL {
		msgWithdrawal := bitcointypes.MsgInitializeWithdrawal{
			Proposer: proposer,
			Vote:     votes,
			Proposal: &bitcointypes.WithdrawalProposal{
				Id:          withdrawIds,
				NoWitnessTx: noWitnessTx,
				TxFee:       txFee,
			},
		}
		return &msgWithdrawal, nil
	} else {
		msgConsolidation := bitcointypes.MsgNewConsolidation{
			Proposer:    proposer,
			Vote:        votes,
			NoWitnessTx: noWitnessTx,
		}
		return &msgConsolidation, nil
	}
}
