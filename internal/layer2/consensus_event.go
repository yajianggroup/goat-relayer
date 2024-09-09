package layer2

import (
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	log "github.com/sirupsen/logrus"

	abcitypes "github.com/cometbft/cometbft/abci/types"
)

func (lis *Layer2Listener) processEvent(block uint64, event abcitypes.Event) {
	switch event.Type {
	case "new_epoch":
		lis.processNewEpochEvent(block, event.Attributes)
	case "finalized_proposal":
		lis.processFinalizedProposalEvent(block, event.Attributes)
	case "elected_proposer":
		lis.processElectedProposerEvent(block, event.Attributes)
	case "accepted_proposer":
		lis.processAcceptedProposerEvent(block, event.Attributes)
	case "voter_pending", "voter_on_boarding", "voter_boarded", "voter_off_boarding", "voter_activated", "voter_discharged":
		lis.processVoterEvent(block, event.Type, event.Attributes)

	case "new_block_hash":
		lis.processNewBtcBlockHash(block, event.Attributes)
	case "new_key":
		lis.processNewWalletKey(block, event.Attributes)
	case "new_deposit":
		lis.processNewDeposit(block, event.Attributes)
	case "new_withdrawal":
		lis.processNewWithdrawal(block, event.Attributes)
	case "approve_cancellation_withdrawal":
		lis.processCancelWithdrawal(block, event.Attributes)
	case "finalize_withdrawal":
		lis.processFinalizeWithdrawal(block, event.Attributes)

	default:
		// log.Debugf("Unrecognized event type: %s", event.Type)
	}
}

func (lis *Layer2Listener) processChainStatus(latestHeight uint64, catchingUp bool) {
	lis.state.UpdateL2ChainStatus(latestHeight, catchingUp)
}

func (lis *Layer2Listener) processEndBlock(block uint64) {
	log.Debugf("Abci end block %d", block)
	lis.state.UpdateL2InfoEndBlock(block)
}

func (lis *Layer2Listener) processFinalizeWithdrawal(block uint64, attributes []abcitypes.EventAttribute) {
	var txid string
	for _, attr := range attributes {
		key := attr.Key
		value := attr.Value

		if key == "txid" {
			txid = value
		}
	}
	log.Infof("Abci FinalizeWithdrawal, block: %d, txid: %s", block, txid)

	// TODO amount, address ?
}

func (lis *Layer2Listener) processCancelWithdrawal(block uint64, attributes []abcitypes.EventAttribute) {
	// TODO check uint64
	var id string
	for _, attr := range attributes {
		key := attr.Key
		value := attr.Value

		if key == "id" {
			id = value
		}
	}
	log.Infof("Abci ApproveCancelWithdrawal, block: %d, id: %s", block, id)

	// TODO
}

func (lis *Layer2Listener) processNewWithdrawal(block uint64, attributes []abcitypes.EventAttribute) {
	var txid string
	for _, attr := range attributes {
		key := attr.Key
		value := attr.Value

		if key == "txid" {
			txid = value
		}
	}
	log.Infof("Abci NewWithdrawal, block: %d, txid: %s", block, txid)

	// TODO amount, address ?
}

func (lis *Layer2Listener) processNewDeposit(block uint64, attributes []abcitypes.EventAttribute) {
	var txid string
	var txout uint64
	var address common.Address
	var amount uint64 // NOTE, db use float64
	for _, attr := range attributes {
		key := attr.Key
		value := attr.Value

		if key == "txid" {
			txid = value
		}
		if key == "txout" {
			txout, _ = strconv.ParseUint(value, 10, 64)
		}
		if key == "address" {
			address = common.HexToAddress(value)
		}
		if key == "amount" {
			amount, _ = strconv.ParseUint(value, 10, 64)
		}
	}
	log.Infof("Abci NewDeposit, block: %d, txid: %s, txout: %d, address: %v, amount: %d", block, txid, txout, address, amount)

	// TODO
}

func (lis *Layer2Listener) processNewWalletKey(block uint64, attributes []abcitypes.EventAttribute) {
	var walletType string
	var walletKey string
	for _, attr := range attributes {
		key := attr.Key
		value := attr.Value

		if key == "type" {
			walletType = value
		}
		if key == "key" {
			walletKey = value
		}
	}
	log.Infof("Abci NewKey: %s, typ: %s, block: %d", walletKey, walletType, block)
	// update
	if walletKey != "" && walletType != "" {
		err := lis.state.UpdateL2InfoWallet(block, walletType, walletKey)
		if err != nil {
			log.Errorf("Abci processNewWalletKey error: %v", err)
		}
	}
}

func (lis *Layer2Listener) processNewBtcBlockHash(block uint64, attributes []abcitypes.EventAttribute) {
	var height string
	var hash string
	for _, attr := range attributes {
		key := attr.Key
		value := attr.Value

		if key == "height" {
			height = value
		}
		if key == "hash" {
			hash = value
		}
	}
	log.Infof("Abci NewBlockHash: %s, block: %d, btcHeight: %s", hash, block, height)

	if height != "" && hash != "" {
		u64, _ := strconv.ParseUint(height, 10, 64)
		lis.state.UpdateL2InfoLatestBtc(block, u64)

		// manage BtcHeadState queue
		lis.state.UpdateProcessedBtcBlock(block, u64, hash)
	}
}

// Process new_epoch event
func (lis *Layer2Listener) processNewEpochEvent(block uint64, attributes []abcitypes.EventAttribute) {
	for _, attr := range attributes {
		key := attr.Key
		value := attr.Value

		if key == "epoch" {
			log.Infof("Abci NewEpoch: %s, block: %d", value, block)

			u64, _ := strconv.ParseUint(value, 10, 64)
			lis.state.UpdateL2InfoEpoch(block, uint(u64), "")
		}
	}
}

// Process finalized_proposal event
func (lis *Layer2Listener) processFinalizedProposalEvent(block uint64, attributes []abcitypes.EventAttribute) {
	for _, attr := range attributes {
		key := attr.Key
		value := attr.Value

		if key == "sequence" {
			log.Infof("Abci FinalizedProposal Sequence: %s, block: %d", value, block)

			// Not update
		}
	}
}

// Process elected_proposer event
func (lis *Layer2Listener) processElectedProposerEvent(block uint64, attributes []abcitypes.EventAttribute) {
	var epoch string
	var proposer string

	for _, attr := range attributes {
		key := attr.Key
		value := attr.Value

		switch key {
		case "epoch":
			epoch = value
		case "proposer":
			proposer = value
		}
	}

	log.Infof("Abci ElectedProposer: %s in Epoch: %s, block: %d", proposer, epoch, block)
	u64, _ := strconv.ParseUint(epoch, 10, 64)
	lis.state.UpdateL2InfoEpoch(block, uint(u64), proposer)
}

// Process accepted_proposer event
func (lis *Layer2Listener) processAcceptedProposerEvent(block uint64, attributes []abcitypes.EventAttribute) {
	for _, attr := range attributes {
		key := attr.Key
		value := attr.Value

		if key == "proposer" {
			log.Infof("Abci AcceptedProposer: %s, block: %d", value, block)
		}
	}
	// Not save this event yet
}

// Voter events
func (lis *Layer2Listener) processVoterEvent(block uint64, eventType string, attributes []abcitypes.EventAttribute) {
	for _, attr := range attributes {
		key := attr.Key
		value := attr.Value

		log.Infof("Abci %s - %s: %s, block: %d", eventType, key, value, block)

		// The value is addrStr, from goat.
		// addrRaw := sdktypes.AccAddress(goatcrypto.Hash160Sum(req.VoterTxKey))
		// addrStr, err := k.AddrCodec.BytesToString(addrRaw)
		// Relayer voter should calculate from secp256k1 private key,
		// then derive the public key, convert pk to address

		// means next epoch valid
		// if key == "voter_pending" {
		// }

		if key == "voter_on_boarding" {
			// TODO should pass event, notify voter to accept boarding
			// Call event bus, send block, voterAddr
			// Push state queue
		}

		if key == "voter_boarded" {
			// TODO should pass event, notify voter to mark boarded
			// Call event bus, send block, voterAddr
			// Update state queue status
		}

		if key == "voter_activated" {
			// TODO should pass event, notify voter to add to list
			// Call event bus, send block, voterAddr
			// Update state status (state -> db)
		}

		if key == "voter_discharged" {
			// TODO should pass event, notify voter to remove from list
			// Call event bus, send block, voterAddr
			// Update state status (state -> db)
		}
	}
}
