package layer2

import (
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/goatnetwork/goat-relayer/internal/layer2/abis"
	log "github.com/sirupsen/logrus"
)

func (lis *Layer2Listener) processGoatLogs(vLog types.Log) {
	if len(vLog.Topics) == 0 {
		log.Debug("No topics found in the log")
		return
	}
	var err error
	switch vLog.Address {
	case abis.BitcoinAddress:
		err = lis.processBitcoinEvent(vLog)
	case abis.BridgeAddress:
		err = lis.processBridgeEvent(vLog)
	case abis.RelayerAddress:
		err = lis.processRelayerEvent(vLog)
	}
	if err != nil {
		log.Errorf("Error processing log: %v", err)
	}
}

func (lis *Layer2Listener) processBitcoinEvent(vLog types.Log) error {
	event, err := lis.contractBitcoin.BitcoinContractFilterer.ParseNewBlockHash(vLog)
	if err != nil {
		log.Warnf("Unpacking NewBlockHash event from ContractBitcoin: %v", err)
		return nil
	}
	blockHash, err := lis.contractBitcoin.BitcoinContractCaller.BlockHash(nil, event.Height)
	if err != nil {
		return fmt.Errorf("query btc block hash from layer2: %v", err)
	}
	log.Infof("Bitcoin chain height updated: %s, hash: %s, goat block: %d", event.Height.String(), hex.EncodeToString(blockHash[:]), event.Raw.BlockNumber)

	// TODO update self l2 state

	return nil
}

func (lis *Layer2Listener) processBridgeEvent(vLog types.Log) error {
	eventDeposit, err := lis.contractBridge.BridgeContractFilterer.ParseDeposit(vLog)
	if err == nil {
		log.Infof("Bridge deposit updated, target %s, amount: %s, txid: %s, txout: %d", eventDeposit.Target.Hex(), eventDeposit.Amount.String(), hex.EncodeToString(eventDeposit.Txid[:]), eventDeposit.Txout)
		// TODO update deposit queue status from rpc
	}

	eventWithdraw, err := lis.contractBridge.BridgeContractFilterer.ParseWithdraw(vLog)
	if err == nil {
		log.Infof("Bridge withdraw created, id %s, sender %s, receiver %s, amount: %s, tax: %s, maxTxPrice: %s", eventWithdraw.Id.String(), eventWithdraw.From.Hex(), eventWithdraw.Receiver, eventWithdraw.Amount.String(), eventWithdraw.Tax.String(), eventWithdraw.MaxTxPrice.String())
		// TODO push withdraw queue
	}

	eventCanceling, err := lis.contractBridge.BridgeContractFilterer.ParseCanceling(vLog)
	if err == nil {
		log.Infof("Bridge withdraw canceling, id %s", eventCanceling.Id.String())
		// TODO update withdraw queue status to canceling
	}

	eventCanceled, err := lis.contractBridge.BridgeContractFilterer.ParseCanceled(vLog)
	if err == nil {
		log.Infof("Bridge withdraw canceled, id %s", eventCanceled.Id.String())
		// TODO update withdraw queue status to canceled
	}

	eventRefund, err := lis.contractBridge.BridgeContractFilterer.ParseRefund(vLog)
	if err == nil {
		log.Infof("Bridge withdraw refund, id %s", eventRefund.Id.String())
		// TODO update withdraw queue status to refund
	}

	eventPaid, err := lis.contractBridge.BridgeContractFilterer.ParsePaid(vLog)
	if err == nil {
		log.Infof("Bridge withdraw paied, id %s, received %s, txid: %s, txout: %d", eventPaid.Id.String(), eventPaid.Value.String(), hex.EncodeToString(eventPaid.Txid[:]), eventPaid.Txout)
		// TODO update withdraw queue status to paid, or remove from queue
	}

	return nil
}

func (lis *Layer2Listener) processRelayerEvent(vLog types.Log) error {
	eventAdd, err := lis.contractRelayer.RelayerContractFilterer.ParseAddedVoter(vLog)
	if err == nil {
		log.Infof("Relayer add voter, address %s, key hash: %s", hex.EncodeToString(eventAdd.Voter[:]), hex.EncodeToString(eventAdd.KeyHash[:]))
		// TODO update self voter status and add to list
	}

	eventRemove, err := lis.contractRelayer.RelayerContractFilterer.ParseRemovedVoter(vLog)
	if err == nil {
		log.Infof("Relayer remove voter, address %s", hex.EncodeToString(eventRemove.Voter[:]))
		// TODO update self voter status and remove from list
	}

	// TODO rotate?

	return nil
}
