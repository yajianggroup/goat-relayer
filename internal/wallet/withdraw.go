package wallet

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/goatnetwork/goat-relayer/internal/config"
	"github.com/goatnetwork/goat-relayer/internal/db"
	"github.com/goatnetwork/goat-relayer/internal/state"
	"github.com/goatnetwork/goat-relayer/internal/types"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func (w *WalletServer) withdrawLoop(ctx context.Context) {
	w.state.EventBus.Subscribe(state.SigFailed, w.withdrawSigFailChan)
	w.state.EventBus.Subscribe(state.SigFinish, w.withdrawSigFinishChan)
	w.state.EventBus.Subscribe(state.SigTimeout, w.withdrawSigTimeoutChan)

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
		case sigFail := <-w.withdrawSigFailChan:
			w.handleWithdrawSigFailed(sigFail, "failed")
		case sigTimeout := <-w.withdrawSigTimeoutChan:
			w.handleWithdrawSigFailed(sigTimeout, "timeout")
		case sigFinish := <-w.withdrawSigFinishChan:
			w.handleWithdrawSigFinish(sigFinish)
		case <-ticker.C:
			w.initWithdrawSig()
		}
	}
}

func (w *WalletServer) handleWithdrawSigFailed(sigFail interface{}, reason string) {
	log.Infof("WalletServer handleWithdrawSigFailed, reason: %s", reason)
}

func (w *WalletServer) handleWithdrawSigFinish(sigFinish interface{}) {
	log.Info("WalletServer handleWithdrawSigFinish")
}

func (w *WalletServer) initWithdrawSig() {
	log.Debug("WalletServer initWithdrawSig")

	// 1. check catching up, self is proposer
	if w.state.GetL2Info().Syncing {
		log.Infof("WalletServer initWithdrawSig ignore, layer2 is catching up")
		return
	}

	btcState := w.state.GetBtcHead()
	if btcState.Syncing {
		log.Infof("WalletServer initWithdrawSig ignore, btc is catching up")
		return
	}
	if btcState.NetworkFee > 500 {
		log.Infof("WalletServer initWithdrawSig ignore, btc network fee too high: %d", btcState.NetworkFee)
		return
	}

	w.sigMu.Lock()
	defer w.sigMu.Unlock()

	epochVoter := w.state.GetEpochVoter()
	if epochVoter.Proposer != config.AppConfig.RelayerAddress {
		if w.sigStatus {
			w.sigStatus = false
			// clean process, role changed, remove all status "create", "aggregating"
			w.cleanWithdrawProcess()
		}
		log.Debugf("WalletServer initWithdrawSig ignore, self is not proposer, epoch: %d, proposer: %s", epochVoter.Epoch, epochVoter.Proposer)
		return
	}

	// 2. check if there is a sig in progress
	if w.sigStatus {
		log.Debug("WalletServer initWithdrawSig ignore, there is a sig")
		return
	}
	// clean process, become proposer again, remove all status "create", "aggregating"
	w.cleanWithdrawProcess()

	// 3. query withraw list from db, status 'create'
	// if count > 150, built soon
	// else if count > 50, check oldest one, if than 2 hours (optional), built
	// else if check oldest one, if than 6 hours (optional), built
	// else go to 4
	// 4. do consolidation
	// 5. start bls sig

	// step 3
	withdraws, err := w.state.GetWithdrawsCanStart()
	if err != nil {
		log.Errorf("WalletServer initWithdrawSig getWithdrawsCanStart error: %v", err)
		return
	}
	if len(withdraws) == 0 {
		log.Infof("WalletServer initWithdrawSig no withdraw from db can start, count: %d", len(withdraws))
		return
	}
	selectedWithdraws, withdrawAmount, actualFee, err := SelectWithdrawals(withdraws, int64(btcState.NetworkFee), 150)
	if err != nil {
		log.Warnf("WalletServer initWithdrawSig SelectWithdrawals error: %v", err)
		return
	}
	if len(selectedWithdraws) == 0 {
		log.Infof("WalletServer initWithdrawSig no withdraw after filter can start, count: %d", len(selectedWithdraws))
		return
	}
	log.Infof("WalletServer initWithdrawSig SelectWithdrawals, withdrawAmount: %d, actualFee: %d, selectedWithdraws: %d", withdrawAmount, actualFee, len(selectedWithdraws))

	startBls := false
	wCount := len(selectedWithdraws)
	if wCount == 150 {
		startBls = true
	} else {
		sort.Slice(selectedWithdraws, func(i, j int) bool {
			return selectedWithdraws[i].CreatedAt.Unix() < selectedWithdraws[j].CreatedAt.Unix()
		})
		oldestWithdraw := selectedWithdraws[0]
		if wCount >= 50 && time.Since(oldestWithdraw.CreatedAt) > 2*time.Hour {
			startBls = true
		} else if time.Since(oldestWithdraw.CreatedAt) > 6*time.Hour {
			startBls = true
		}
	}
	// get pubkey
	pubkey, err := w.state.GetDepositKeyByBtcBlock(0)
	if err != nil {
		log.Fatalf("WalletServer get current change or consolidation key by btc height current err %v", err)
	}

	network := types.GetBTCNetwork(config.AppConfig.BTCNetworkType)

	// get vin
	pubkeyBytes, err := base64.StdEncoding.DecodeString(pubkey.PubKey)
	if err != nil {
		log.Fatalf("Base64 decode pubkey %s err %v", pubkey.PubKey, err)
	}
	p2wpkhAddress, err := types.GenerateP2WPKHAddress(pubkeyBytes, network)
	if err != nil {
		log.Fatalf("Gen P2WPKH address from pubkey %s err %v", pubkey.PubKey, err)
	}

	utxos, err := w.state.GetUtxoCanSpend()
	if err != nil {
		log.Errorf("WalletServer initWithdrawSig GetUtxoCanSpend error: %v", err)
		return
	}

	var msgSignSendOrder *types.MsgSignSendOrder

	if !startBls {
		log.Infof("WalletServer initWithdrawSig withdraw not start bls, count: %d, next to check consolidation", wCount)

		selectedUtxos, totalAmount, finalAmount, err := ConsolidateSmallUTXOs(utxos, int64(btcState.NetworkFee), 5*1e7, 50)
		if err != nil {
			log.Errorf("WalletServer initWithdrawSig ConsolidateSmallUTXOs error: %v", err)
			return
		}
		log.Infof("WalletServer initWithdrawSig ConsolidateSmallUTXOs,totalAmount: %d, finalAmount: %d, selectedUtxos: %d", totalAmount, finalAmount, len(selectedUtxos))

		startBls = true

		// create SendOrder for selectedUtxos consolidation
		tx, _, err := CreateRawTransaction(selectedUtxos, nil, p2wpkhAddress.EncodeAddress(), finalAmount, 0, network)
		if err != nil {
			log.Errorf("WalletServer initWithdrawSig CreateRawTransaction for consolidation error: %v", err)
			return
		}
		log.Infof("WalletServer initWithdrawSig CreateRawTransaction for consolidation, tx: %s", tx.TxHash().String())

		msgSignSendOrder, err = w.createSendOrder(tx, db.ORDER_TYPE_CONSOLIDATION, selectedUtxos, nil, totalAmount, 0, finalAmount, epochVoter, network)
		if err != nil {
			log.Errorf("WalletServer initWithdrawSig createSendOrder for consolidation error: %v", err)
			return
		}
	} else {
		// create SendOrder for selectedWithdraws
		selectOptimalUTXOs, totalSelectedAmount, _, changeAmount, estimateFee, err := SelectOptimalUTXOs(utxos, withdrawAmount, int64(btcState.NetworkFee), len(selectedWithdraws))
		if err != nil {
			log.Errorf("WalletServer initWithdrawSig SelectOptimalUTXOs error: %v", err)
			return
		}
		log.Infof("WalletServer initWithdrawSig SelectOptimalUTXOs, totalSelectedAmount: %d, withdrawAmount: %d, changeAmount: %d, selectedUtxos: %d", totalSelectedAmount, withdrawAmount, changeAmount, len(selectOptimalUTXOs))

		tx, dustWithdrawId, err := CreateRawTransaction(selectOptimalUTXOs, selectedWithdraws, p2wpkhAddress.EncodeAddress(), changeAmount, estimateFee, network)
		if err != nil {
			log.Errorf("WalletServer initWithdrawSig CreateRawTransaction for withdraw error: %v", err)
			if dustWithdrawId > 0 {
				// update this withdraw to closed
				err = w.state.CloseWithdraw(dustWithdrawId, "dust limit")
				if err != nil {
					log.Errorf("WalletServer initWithdrawSig CloseWithdraw for dust limit withdraw %d, error: %v", dustWithdrawId, err)
				} else {
					log.Infof("WalletServer initWithdrawSig CloseWithdraw for dust limit withdraw %d ok, ave tx fee %d", dustWithdrawId, estimateFee/int64(len(selectedWithdraws)))
				}
			}
			return
		}
		log.Infof("WalletServer initWithdrawSig CreateRawTransaction for withdraw, tx: %s", tx.TxHash().String())

		msgSignSendOrder, err = w.createSendOrder(tx, db.ORDER_TYPE_WITHDRAWAL, selectOptimalUTXOs, selectedWithdraws, totalSelectedAmount, withdrawAmount, changeAmount, epochVoter, network)
		if err != nil {
			log.Errorf("WalletServer initWithdrawSig createSendOrder for withdraw error: %v", err)
			return
		}
	}

	// w.sigStatus should update to false after layer2 InitalWithdraw callback
	if msgSignSendOrder != nil {
		w.sigStatus = true

		// send msg to bus
		w.state.EventBus.Publish(state.SigStart, msgSignSendOrder)
		log.Infof("WalletServer initWithdrawSig send MsgSignSendOrder to bus, requestId: %s", msgSignSendOrder.MsgSign.RequestId)
	}
}

// createSendOrder, create send order for selected utxos and withdraws (if orderType is consolidation, selectedWithdraws is nil)
func (w *WalletServer) createSendOrder(tx *wire.MsgTx, orderType string, selectedUtxos []*db.Utxo, selectedWithdraws []*db.Withdraw, utxoAmount, withdrawAmount, changeAmount int64, epochVoter db.EpochVoter, network *chaincfg.Params) (*types.MsgSignSendOrder, error) {
	// save order to db
	order := &db.SendOrder{
		OrderId:    uuid.New().String(),
		Proposer:   config.AppConfig.RelayerAddress,
		Amount:     uint64(utxoAmount),
		MaxTxFee:   uint64(utxoAmount - withdrawAmount - changeAmount),
		Status:     db.ORDER_STATUS_AGGREGATING,
		OrderType:  orderType,
		BtcBlock:   0,
		Txid:       tx.TxID(),
		SigScripts: nil, // after sign
		PkScripts:  nil, // after sign
		TxFee:      0,
		UpdatedAt:  time.Now(),
	}

	requestId := fmt.Sprintf("SENDORDER:%s:%s", config.AppConfig.RelayerAddress, order.OrderId)

	var vins []*db.Vin
	for i, txIn := range tx.TxIn {
		vin := &db.Vin{
			OrderId:      order.OrderId,
			BtcHeight:    uint64(txIn.PreviousOutPoint.Index),
			Txid:         txIn.PreviousOutPoint.Hash.String(),
			OutIndex:     int(txIn.PreviousOutPoint.Index),
			SigScript:    nil,
			RedeemScript: selectedUtxos[i].RedeemScript,
			Sender:       "",
			ReceiverType: selectedUtxos[i].ReceiverType,
			Source:       orderType,
			Status:       db.ORDER_STATUS_AGGREGATING,
			UpdatedAt:    time.Now(),
		}
		vins = append(vins, vin)
	}
	var vouts []*db.Vout
	for i, txOut := range tx.TxOut {
		withdrawId := ""
		if orderType == db.ORDER_TYPE_WITHDRAWAL && len(selectedWithdraws) > i {
			withdrawId = fmt.Sprintf("%d", selectedWithdraws[i].RequestId)
		}
		_, addresses, _, err := txscript.ExtractPkScriptAddrs(txOut.PkScript, network)
		if err != nil {
			return nil, err
		}
		vout := &db.Vout{
			OrderId:    order.OrderId,
			BtcHeight:  0,
			Txid:       tx.TxID(),
			OutIndex:   i,
			WithdrawId: withdrawId,
			Amount:     int64(txOut.Value),
			Receiver:   addresses[0].EncodeAddress(),
			Sender:     "",
			Source:     orderType,
			Status:     db.ORDER_STATUS_AGGREGATING,
			UpdatedAt:  time.Now(),
		}
		vouts = append(vouts, vout)
	}

	orderBytes, err := json.Marshal(order)
	if err != nil {
		return nil, err
	}
	utxoBytes, err := json.Marshal(selectedUtxos)
	if err != nil {
		return nil, err
	}

	vinBytes, err := json.Marshal(vins)
	if err != nil {
		return nil, err
	}

	voutBytes, err := json.Marshal(vouts)
	if err != nil {
		return nil, err
	}

	var withdrawBytes []byte
	if len(selectedWithdraws) > 0 {
		withdrawBytes, err = json.Marshal(selectedWithdraws)
		if err != nil {
			return nil, err
		}
	}

	msgSignSendOrder := &types.MsgSignSendOrder{
		MsgSign: types.MsgSign{
			RequestId:    requestId,
			Sequence:     epochVoter.Sequence,
			Epoch:        epochVoter.Epoch,
			IsProposer:   true,
			VoterAddress: epochVoter.Proposer,
			SigData:      nil,
			CreateTime:   time.Now().Unix(),
		},
		SendOrder: orderBytes,
		Utxos:     utxoBytes,
		Vins:      vinBytes,
		Vouts:     voutBytes,
		Withdraws: withdrawBytes,
	}

	// save
	err = w.state.CreateSendOrder(order, selectedUtxos, selectedWithdraws, vins, vouts)
	if err != nil {
		return nil, err
	}

	return msgSignSendOrder, nil
}

func (w *WalletServer) cleanWithdrawProcess() {
	// unset all status "create", "aggregating"
	err := w.state.CleanProcessingWithdraw()
	if err != nil {
		log.Fatal("WalletServer cleanWithdrawProcess unexpected error %v", err)
	}
}
