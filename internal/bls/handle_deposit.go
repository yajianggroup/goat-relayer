// handle_deposit.go handle deposit bls sig
package bls

import (
	"context"
	"encoding/json"

	"github.com/goatnetwork/goat-relayer/internal/config"
	"github.com/goatnetwork/goat-relayer/internal/state"
	"github.com/goatnetwork/goat-relayer/internal/types"
	bitcointypes "github.com/goatnetwork/goat/x/bitcoin/types"
	relayertypes "github.com/goatnetwork/goat/x/relayer/types"
	log "github.com/sirupsen/logrus"
)

func (s *Signer) handleSigStartNewDeposit(ctx context.Context, e types.MsgSignDeposit) error {
	// do not send p2p here, it doesn't need to aggregate sign here
	isProposer := s.IsProposer()
	if isProposer {
		log.Info("Proposer submit NewDeposits to consensus")

		headers := make(map[uint64][]byte)
		headers[e.BlockNumber] = e.BlockHeader
		headersBytes, err := json.Marshal(headers)
		if err != nil {
			log.Errorf("Failed to marshal headers: %v", err)
			return err
		}

		pubKey := relayertypes.DecodePublicKey(e.RelayerPubkey)
		deposits := make([]*bitcointypes.Deposit, 1)
		deposits[0] = &bitcointypes.Deposit{
			Version:           e.Version,
			BlockNumber:       e.BlockNumber,
			TxIndex:           e.TxIndex,
			NoWitnessTx:       e.NoWitnessTx,
			OutputIndex:       e.OutputIndex,
			IntermediateProof: e.IntermediateProof,
			EvmAddress:        e.EvmAddress,
			RelayerPubkey:     pubKey,
		}

		msgDeposits := &bitcointypes.MsgNewDeposits{
			Proposer:     e.Proposer,
			BlockHeaders: headersBytes,
			Deposits:     deposits,
		}

		err = s.RetrySubmit(ctx, e.RequestId, msgDeposits, config.AppConfig.L2SubmitRetry)
		if err != nil {
			log.Errorf("Proposer submit NewDeposit to consensus error, request id: %s, err: %v", e.RequestId, err)
			// feedback SigFailed, deposit should module subscribe it to save UTXO or mark confirm
			s.state.EventBus.Publish(state.SigFailed, e)
			return err
		}

		// feedback SigFinish, deposit should module subscribe it to save UTXO or mark confirm
		s.state.EventBus.Publish(state.SigFinish, e)
		log.Infof("Proposer submit MsgNewDeposit to consensus ok, request id: %s", e.RequestId)
	}
	return nil
}
