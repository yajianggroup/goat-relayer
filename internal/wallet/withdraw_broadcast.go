package wallet

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/goatnetwork/goat-relayer/internal/config"
	"github.com/goatnetwork/goat-relayer/internal/db"
	"github.com/goatnetwork/goat-relayer/internal/types"
	log "github.com/sirupsen/logrus"
)

func (w *WalletServer) txBroadcastLoop(ctx context.Context) {
	log.Debug("WalletServer withdrawProcessLoop")
	// init status process, if restart && layer2 status is up to date, remove all status "create", "aggregating"
	if !w.state.GetBtcHead().Syncing {
		w.cleanWithdrawProcess()
	}

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.handleTxBroadcast()
		}
	}
}

func (w *WalletServer) handleTxBroadcast() {
	l2Info := w.state.GetL2Info()
	if l2Info.Syncing {
		log.Infof("WalletServer handleTxBroadcast ignore, layer2 is catching up")
		return
	}

	btcState := w.state.GetBtcHead()
	if btcState.Syncing {
		log.Infof("WalletServer handleTxBroadcast ignore, btc is catching up")
		return
	}
	if btcState.NetworkFee > 500 {
		log.Infof("WalletServer handleTxBroadcast ignore, btc network fee too high: %d", btcState.NetworkFee)
		return
	}

	w.txBroadcastMu.Lock()
	defer w.txBroadcastMu.Unlock()

	epochVoter := w.state.GetEpochVoter()
	if epochVoter.Proposer != config.AppConfig.RelayerAddress {
		// do not clean immediately
		if w.txBroadcastStatus && l2Info.Height > epochVoter.Height+1 {
			w.txBroadcastStatus = false
			// clean process, role changed, remove all status "create", "aggregating"
			w.cleanWithdrawProcess()
		}
		log.Debugf("WalletServer handleTxBroadcast ignore, self is not proposer, epoch: %d, proposer: %s", epochVoter.Epoch, epochVoter.Proposer)
		return
	}

	// 2. check if there is a sig in progress
	if w.txBroadcastStatus {
		log.Debug("WalletServer handleTxBroadcast ignore, there is a broadcast in progress")
		return
	}
	if l2Info.LatestBtcHeight <= w.txBroadcastFinishBtcHeight+1 {
		log.Debugf("WalletServer handleTxBroadcast ignore, last finish broadcast in this block: %d", w.txBroadcastFinishBtcHeight)
		return
	}

	sendOrders, err := w.state.GetSendOrderInitlized()
	if err != nil {
		log.Errorf("WalletServer handleTxBroadcast error: %v", err)
		return
	}
	if len(sendOrders) == 0 {
		log.Debug("WalletServer handleTxBroadcast ignore, no withdraw for broadcast")
		return
	}

	privKeyBytes, err := hex.DecodeString(config.AppConfig.FireblocksPrivKey)
	if err != nil {
		log.Errorf("WalletServer handleTxBroadcast decode privKey error: %v", err)
		return
	}
	privKey, _ := btcec.PrivKeyFromBytes(privKeyBytes)

	for _, sendOrder := range sendOrders {
		tx, err := types.DeserializeTransaction(sendOrder.NoWitnessTx)
		if err != nil {
			log.Errorf("WalletServer handleTxBroadcast deserialize tx error: %v, txid: %s", err, sendOrder.Txid)
			continue
		}

		utxos, err := w.state.GetUtxoByOrderId(sendOrder.OrderId)
		if err != nil {
			log.Errorf("WalletServer handleTxBroadcast get utxos error: %v, txid: %s", err, sendOrder.Txid)
			continue
		}

		// sign the transaction
		err = SignTransactionByPrivKey(privKey, tx, utxos, types.GetBTCNetwork(config.AppConfig.BTCNetworkType))
		if err != nil {
			log.Errorf("WalletServer handleTxBroadcast sign tx error: %v, txid: %s", err, sendOrder.Txid)
			continue
		}

		// broadcast the transaction and update sendOrder status
		txHash, err := w.btcClient.SendRawTransaction(tx, false)
		if err != nil {
			if rpcErr, ok := err.(*btcjson.RPCError); ok {
				switch rpcErr.Code {
				case btcjson.ErrRPCTxAlreadyInChain:
					log.Infof("WalletServer handleTxBroadcast tx already in chain, txid: %s", sendOrder.Txid)
					if sendOrder.Status == db.ORDER_STATUS_CONFIRMED {
						continue
					}
					err = w.state.UpdateSendOrderConfirmed(sendOrder.Txid)
					if err != nil {
						log.Errorf("WalletServer handleTxBroadcast update sendOrder status error: %v, txid: %s", err, sendOrder.Txid)
					}
					continue
				default:
					log.Errorf("WalletServer handleTxBroadcast broadcast tx error: %v, txid: %s", rpcErr, sendOrder.Txid)
				}
			}
			continue
		}
		log.Infof("WalletServer handleTxBroadcast tx broadcast success, txid: %s", txHash.String())

		// update sendOrder status to pending
		err = w.state.UpdateSendOrderPending(txHash.String())
		if err != nil {
			log.Errorf("WalletServer handleTxBroadcast update sendOrder status error: %v, txid: %s", err, sendOrder.Txid)
			continue
		}
	}

}
