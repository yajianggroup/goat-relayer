package wallet

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/goatnetwork/goat-relayer/internal/config"
	"github.com/goatnetwork/goat-relayer/internal/db"
	"github.com/goatnetwork/goat-relayer/internal/http"
	"github.com/goatnetwork/goat-relayer/internal/state"
	"github.com/goatnetwork/goat-relayer/internal/types"
	log "github.com/sirupsen/logrus"
)

type OrderBroadcaster interface {
	Start(ctx context.Context)
	Stop()
	broadcastOrders()
	broadcastPendingCheck()
}

type RemoteClient interface {
	SendRawTransaction(tx *wire.MsgTx, utxos []*db.Utxo, orderType string) (txHash string, exist bool, err error)
	CheckPending(txid string, externalTxId string) (revert bool, confirmations uint64, err error)
}

type BtcClient struct {
	client *rpcclient.Client
}

type FireblocksClient struct {
	client *http.FireblocksProposal
}

type BaseOrderBroadcaster struct {
	remoteClient RemoteClient
	state        *state.State

	txBroadcastMu              sync.Mutex
	txBroadcastStatus          bool
	txBroadcastFinishBtcHeight uint64
}

var (
	_ RemoteClient = (*BtcClient)(nil)
	_ RemoteClient = (*FireblocksClient)(nil)

	_ OrderBroadcaster = (*BaseOrderBroadcaster)(nil)
)

func (c *BtcClient) SendRawTransaction(tx *wire.MsgTx, utxos []*db.Utxo, orderType string) (txHash string, exist bool, err error) {
	txid := tx.TxHash().String()
	privKeyBytes, err := hex.DecodeString(config.AppConfig.FireblocksPrivKey)
	if err != nil {
		return txid, false, fmt.Errorf("decode privKey error: %v", err)
	}
	privKey, _ := btcec.PrivKeyFromBytes(privKeyBytes)
	// sign the transaction
	err = SignTransactionByPrivKey(privKey, tx, utxos, types.GetBTCNetwork(config.AppConfig.BTCNetworkType))
	if err != nil {
		return txid, false, fmt.Errorf("sign tx %s error: %v", txid, err)
	}
	_, err = c.client.SendRawTransaction(tx, false)
	if err != nil {
		if rpcErr, ok := err.(*btcjson.RPCError); ok {
			switch rpcErr.Code {
			case btcjson.ErrRPCTxAlreadyInChain:
				return txid, true, err
			}
		}
		return txid, false, err
	}
	return txid, false, nil
}

func (c *BtcClient) CheckPending(txid string, externalTxId string) (revert bool, confirmations uint64, err error) {
	txHash, err := chainhash.NewHashFromStr(txid)
	if err != nil {
		return false, 0, fmt.Errorf("new hash from str error: %v, raw txid: %s", err, txid)
	}
	txRawResult, err := c.client.GetRawTransactionVerbose(txHash)
	if err != nil {
		return false, 0, fmt.Errorf("get raw transaction verbose error: %v, txid: %s", err, txid)
	}

	// TODO:
	// 1. if tx is no found in 10 minutes, revert to submit again
	// 2. if tx is no found after 72 hours (perhaps removed from mempool), revert to submit again

	return false, txRawResult.Confirmations, nil
}

func (c *FireblocksClient) SendRawTransaction(tx *wire.MsgTx, utxos []*db.Utxo, orderType string) (txHash string, exist bool, err error) {
	txid := tx.TxHash().String()
	rawMessage, err := GenerateRawMeessageToFireblocks(tx, utxos, types.GetBTCNetwork(config.AppConfig.BTCNetworkType))
	if err != nil {
		return txid, false, fmt.Errorf("generate raw message to fireblocks error: %v", err)
	}

	resp, err := c.client.PostRawSigningRequest(rawMessage, fmt.Sprintf("%s:%s", orderType, txid))
	if err != nil {
		return txid, false, fmt.Errorf("post raw signing request error: %v", err)
	}
	log.Debugf("PostRawSigningRequest resp: %+v", resp)

	return resp.ID, false, nil
}

func (c *FireblocksClient) CheckPending(txid string, externalTxId string) (revert bool, confirmations uint64, err error) {
	txDetails, err := c.client.QueryTransaction(externalTxId)
	if err != nil {
		return false, 0, fmt.Errorf("get tx details from fireblocks error: %v, txid: %s", err, txid)
	}

	failedStatus := []string{"CANCELLING", "CANCELLED", "BLOCKED", "REJECTED", "FAILED"}
	// Check if txDetails.Status is in failedStatus
	for _, status := range failedStatus {
		if txDetails.Status == status {
			return true, 0, nil
		}
	}

	return false, uint64(txDetails.NumOfConfirmations), nil
}

func NewOrderBroadcaster(btcClient *rpcclient.Client, state *state.State) OrderBroadcaster {
	orderBroadcaster := &BaseOrderBroadcaster{
		state: state,
	}
	if config.AppConfig.BTCNetworkType == "regtest" {
		orderBroadcaster.remoteClient = &BtcClient{
			client: btcClient,
		}
	} else {
		orderBroadcaster.remoteClient = &FireblocksClient{
			client: http.NewFireblocksProposal(),
		}
	}
	return orderBroadcaster
}

// txBroadcastLoop is a loop that broadcasts withdrawal and consolidation orders to the network
// check orders pending status, if it is failed, broadcast it again
func (b *BaseOrderBroadcaster) Start(ctx context.Context) {
	log.Debug("BaseOrderBroadcaster start")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			b.Stop()
			return
		case <-ticker.C:
			b.broadcastOrders()
			b.broadcastPendingCheck()
		}
	}
}

func (b *BaseOrderBroadcaster) Stop() {
}

// broadcastOrders is a function that broadcasts withdrawal and consolidation orders to the network
func (b *BaseOrderBroadcaster) broadcastOrders() {
	l2Info := b.state.GetL2Info()
	if l2Info.Syncing {
		log.Infof("OrderBroadcaster broadcastOrders ignore, layer2 is catching up")
		return
	}

	btcState := b.state.GetBtcHead()
	if btcState.Syncing {
		log.Infof("OrderBroadcaster broadcastOrders ignore, btc is catching up")
		return
	}
	if btcState.NetworkFee > 500 {
		log.Infof("OrderBroadcaster broadcastOrders ignore, btc network fee too high: %d", btcState.NetworkFee)
		return
	}

	b.txBroadcastMu.Lock()
	defer b.txBroadcastMu.Unlock()

	epochVoter := b.state.GetEpochVoter()
	if epochVoter.Proposer != config.AppConfig.RelayerAddress {
		log.Debugf("OrderBroadcaster broadcastOrders ignore, self is not proposer, epoch: %d, proposer: %s", epochVoter.Epoch, epochVoter.Proposer)
		return
	}

	// 2. check if there is a sig in progress
	if b.txBroadcastStatus {
		log.Debug("WalletServer broadcastOrders ignore, there is a broadcast in progress")
		return
	}
	if l2Info.LatestBtcHeight <= b.txBroadcastFinishBtcHeight+1 {
		log.Debugf("WalletServer broadcastOrders ignore, last finish broadcast in this block: %d", b.txBroadcastFinishBtcHeight)
		return
	}

	sendOrders, err := b.state.GetSendOrderInitlized()
	if err != nil {
		log.Errorf("OrderBroadcaster broadcastOrders error: %v", err)
		return
	}
	if len(sendOrders) == 0 {
		log.Debug("OrderBroadcaster broadcastOrders ignore, no withdraw for broadcast")
		return
	}

	for _, sendOrder := range sendOrders {
		tx, err := types.DeserializeTransaction(sendOrder.NoWitnessTx)
		if err != nil {
			log.Errorf("OrderBroadcaster broadcastOrders deserialize tx error: %v, txid: %s", err, sendOrder.Txid)
			continue
		}

		utxos, err := b.state.GetUtxoByOrderId(sendOrder.OrderId)
		if err != nil {
			log.Errorf("WalletServer broadcastOrders get utxos error: %v, txid: %s", err, sendOrder.Txid)
			continue
		}

		// broadcast the transaction and update sendOrder status
		txHash, exist, err := b.remoteClient.SendRawTransaction(tx, utxos, sendOrder.OrderType)
		if err != nil {
			if exist {
				log.Warnf("WalletServer broadcastOrders tx already in chain, txid: %s, err: %v", sendOrder.Txid, err)
				if sendOrder.Status == db.ORDER_STATUS_CONFIRMED {
					continue
				}
				err = b.state.UpdateSendOrderConfirmed(sendOrder.Txid)
				if err != nil {
					log.Errorf("WalletServer broadcastOrders update sendOrder %s status error: %v", sendOrder.Txid, err)
				}
				continue
			}
			continue
		}
		log.Infof("WalletServer broadcastOrders tx broadcast success, txid: %s", txHash)

		// update sendOrder status to pending
		err = b.state.UpdateSendOrderPending(sendOrder.Txid, txHash)
		if err != nil {
			log.Errorf("WalletServer broadcastOrders update sendOrder status error: %v, txid: %s", err, sendOrder.Txid)
			continue
		}
	}
}

// broadcastPendingCheck is a function that checks the pending status of the orders
// if it is failed, broadcast it again
func (b *BaseOrderBroadcaster) broadcastPendingCheck() {
	// TODO: query pending orders (limit 50 per time)

	// TODO: call remoteClient.CheckPending to check the pending status of the orders
}
