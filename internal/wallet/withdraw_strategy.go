package wallet

import (
	"fmt"
	"sort"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/goatnetwork/goat-relayer/internal/db"
	"github.com/goatnetwork/goat-relayer/internal/types"
)

// ConsolidateSmallUTXOs consolidate small utxos
//
// Parameters:
//
//	utxos - all unspent utxos
//	networkFee - network fee
//	threshold - threshold for small utxos
//	maxVin - maximum vin count
//	trigerNum - minimum small utxo
//
// Returns:
//
//	selectedUtxos - selected utxos for consolidation
//	totalAmount - total amount of selected utxos
//	finalAmount - final amount after consolidation, after deducting the network fee
//	error - error if any
func ConsolidateSmallUTXOs(utxos []*db.Utxo, networkFee, threshold int64, maxVin, trigerNum int) ([]*db.Utxo, int64, int64, error) {
	if networkFee > 200 {
		return nil, 0, 0, fmt.Errorf("network fee is too high, cannot consolidate")
	}
	if trigerNum < 10 {
		trigerNum = 500
	}
	if len(utxos) < trigerNum {
		return nil, 0, 0, fmt.Errorf("not enough utxos to consolidate")
	}

	var smallUTXOs []*db.Utxo
	var totalAmount int64 = 0

	// select lower than threshold (0.5 BTC)
	for _, utxo := range utxos {
		if utxo.ReceiverType != types.WALLET_TYPE_P2WPKH && utxo.ReceiverType != types.WALLET_TYPE_P2WSH && utxo.ReceiverType != types.WALLET_TYPE_P2PKH {
			continue
		}
		if utxo.Amount < threshold {
			smallUTXOs = append(smallUTXOs, utxo)
			totalAmount += utxo.Amount
		}
		// vin count limit
		if len(smallUTXOs) >= maxVin {
			break
		}
	}

	if len(smallUTXOs) < trigerNum {
		return nil, 0, 0, fmt.Errorf("not enough small utxos to consolidate")
	}

	// calculate transaction size and estimated fee
	utxoTypes := make([]string, len(smallUTXOs))
	for i, utxo := range smallUTXOs {
		utxoTypes[i] = utxo.ReceiverType
	}

	txSize := types.TransactionSizeEstimate(len(smallUTXOs), 1, utxoTypes) // 1 vout
	estimatedFee := txSize * networkFee

	// calculate the remaining amount after consolidation
	finalAmount := totalAmount - estimatedFee
	if finalAmount <= 0 {
		return nil, 0, 0, fmt.Errorf("consolidation fee is too high, cannot consolidate")
	}

	return smallUTXOs, totalAmount, finalAmount, nil
}

// SelectOptimalUTXOs select optimal utxos for withdrawal
//
// Parameters:
//
//	utxos - all unspent utxos
//	withdrawAmount - total amount to withdraw
//	networkFee - network fee
//	withdrawTotal - total number of withdrawals
//
// Returns:
//
//	selectedUtxos - selected utxos for withdrawal
//	totalSelectedAmount - total amount of selected utxos
//	withdrawAmount - total amount to withdraw (not minus tx fee yet)
//	changeAmount - change amount after withdrawal
//	estimatedFee - estimate fee
//	error - error if any
func SelectOptimalUTXOs(utxos []*db.Utxo, withdrawAmount, networkFee int64, withdrawTotal int) ([]*db.Utxo, int64, int64, int64, int64, error) {
	// sort utxos by amount from large to small
	sort.Slice(utxos, func(i, j int) bool {
		return utxos[i].Amount > utxos[j].Amount
	})

	var selectedUTXOs []*db.Utxo
	var totalSelectedAmount int64 = 0
	var maxVin int = 10
	if withdrawAmount > 50*1e8 {
		// max vin is 50 for large amount than 50 BTC
		maxVin = 50
	}

	// calculate the current transaction fee with no UTXO selected yet
	utxoTypes := make([]string, 0)
	txSize := types.TransactionSizeEstimate(len(utxoTypes), withdrawTotal+1, utxoTypes) // +1 for change output
	estimatedFee := txSize * networkFee

	// total target includes withdrawal amount and estimated fee
	totalTarget := withdrawAmount + estimatedFee

	// try to find UTXO that satisfies the total target including fee
	for _, utxo := range utxos {
		if totalSelectedAmount >= totalTarget {
			break
		}
		selectedUTXOs = append(selectedUTXOs, utxo)
		totalSelectedAmount += utxo.Amount

		// update transaction size and estimated fee with the current UTXO selection
		utxoTypes = append(utxoTypes, utxo.ReceiverType)
		txSize = types.TransactionSizeEstimate(len(selectedUTXOs), withdrawTotal+1, utxoTypes)
		estimatedFee = txSize * networkFee

		// recalculate the total target (withdraw amount + estimated fee)
		totalTarget = withdrawAmount // + estimatedFee

		// limit max vin
		if len(selectedUTXOs) >= maxVin {
			break
		}
	}

	// find the smallest utxo and append to selected utxos
	var smallestUTXO *db.Utxo
	for _, utxo := range utxos {
		// skip already selected utxos
		isAlreadySelected := false
		for _, selected := range selectedUTXOs {
			if selected.Txid == utxo.Txid && selected.OutIndex == utxo.OutIndex {
				isAlreadySelected = true
				break
			}
		}
		if isAlreadySelected {
			continue
		}

		// if find a smaller utxo, update the smallest utxo
		if smallestUTXO == nil || utxo.Amount < smallestUTXO.Amount {
			smallestUTXO = utxo
		}
	}

	// append the smallest utxo
	if smallestUTXO.Amount > 0 {
		selectedUTXOs = append(selectedUTXOs, smallestUTXO)
		totalSelectedAmount += smallestUTXO.Amount

		// update transaction size and estimated fee with the current UTXO selection
		utxoTypes = append(utxoTypes, smallestUTXO.ReceiverType)
		txSize = types.TransactionSizeEstimate(len(selectedUTXOs), withdrawTotal+1, utxoTypes)
		estimatedFee = txSize * networkFee

		// recalculate the total target (withdraw amount + estimated fee)
		totalTarget = withdrawAmount // + estimatedFee, fee should minus from withdraw txout value
	}

	// after selecting, check if we have enough UTXO to cover the total target
	if totalSelectedAmount < totalTarget {
		return nil, 0, 0, 0, 0, fmt.Errorf("not enough utxos to satisfy the withdrawal amount and network fee, withdraw amount: %d, selected amount: %d, estimated fee: %d", withdrawAmount, totalSelectedAmount, estimatedFee)
	}

	// calculate the change amount
	changeAmount := totalSelectedAmount - withdrawAmount // - estimatedFee, fee should minus from withdraw txout value

	return selectedUTXOs, totalSelectedAmount, withdrawAmount, changeAmount, estimatedFee, nil
}

// SelectWithdrawals select optimal withdrawals for withdrawal
//
// Parameters:
//
//	withdrawals - all withdrawals can start
//	networkFee - network fee
//	maxVout - maximum vout count
//
// Returns:
//
//	selectedWithdrawals - selected withdrawals
//	minTxFee - minimum transaction fee
//	error - error if any
func SelectWithdrawals(withdrawals []*db.Withdraw, networkFee int64, maxVout int) ([]*db.Withdraw, int64, int64, error) {
	// if network fee is too high, do not perform withdrawal
	if networkFee > 500 {
		return nil, 0, 0, fmt.Errorf("network fee too high, no withdrawals allowed")
	}

	// sort withdrawals by MaxTxFee in descending order
	sort.Slice(withdrawals, func(i, j int) bool {
		return withdrawals[i].TxPrice > withdrawals[j].TxPrice
	})

	// three groups
	var group1 []*db.Withdraw
	var group2 []*db.Withdraw
	var group3 []*db.Withdraw
	uNetworkFee := uint64(networkFee)

	// iterate withdrawals and group by MaxTxFee
	for _, withdrawal := range withdrawals {
		// group1: TxPrice > 150 and TxPrice >= networkFee.FeePerByte
		if withdrawal.TxPrice > 150 && withdrawal.TxPrice >= uNetworkFee {
			group1 = append(group1, withdrawal)
		} else if withdrawal.TxPrice > 50 && withdrawal.TxPrice <= 150 && withdrawal.TxPrice >= uNetworkFee {
			// group2: 50 < MaxTxFee <= 150 and MaxTxFee >= networkFee.FeePerByte
			group2 = append(group2, withdrawal)
		} else if withdrawal.TxPrice <= 50 && withdrawal.TxPrice >= uNetworkFee {
			// group3: MaxTxFee <= 50 and MaxTxFee >= networkFee
			group3 = append(group3, withdrawal)
		}
	}

	// group withdrawals and calculate the estimated fee
	applyGroup := func(group []*db.Withdraw, multiplier float64) ([]*db.Withdraw, int64, int64) {
		if len(group) == 0 {
			return nil, 0, 0
		}

		// limit the number of withdrawals to maxVout
		if len(group) > maxVout {
			group = group[:maxVout]
		}

		withdrawAmount := uint64(0)
		// calculate the minimum MaxTxFee in the group
		groupMinFee := group[0].TxPrice
		for _, withdrawal := range group {
			if withdrawal.TxPrice < groupMinFee {
				groupMinFee = withdrawal.TxPrice
			}
			withdrawAmount += withdrawal.Amount
		}

		// actual fee is the minimum of networkFee * multiplier and the minimum MaxTxFee in the group
		estimatedFee := int64(float64(networkFee) * multiplier)
		actualFee := int64(groupMinFee)
		if estimatedFee < int64(groupMinFee) {
			actualFee = estimatedFee
		}

		return group, int64(withdrawAmount), actualFee
	}

	// process group1 (MaxTxFee > 150, fee: networkFee * 1.25 or min MaxTxFee in the group)
	if len(group1) > 0 {
		selectedWithdrawals, withdrawAmount, minTxFee := applyGroup(group1, 1.25)
		return selectedWithdrawals, withdrawAmount, minTxFee, nil
	}

	// process group2 (50 < MaxTxFee <= 150, fee: networkFee * 1.1 or min MaxTxFee in the group)
	if len(group2) > 0 {
		selectedWithdrawals, withdrawAmount, minTxFee := applyGroup(group2, 1.1)
		return selectedWithdrawals, withdrawAmount, minTxFee, nil
	}

	// process group3 (MaxTxFee <= 50, fee: networkFee)
	if len(group3) > 0 {
		selectedWithdrawals, withdrawAmount, minTxFee := applyGroup(group3, 1.0)
		return selectedWithdrawals, withdrawAmount, minTxFee, nil
	}

	// no withdrawals found
	return nil, 0, 0, fmt.Errorf("no withdrawals found")
}

// CreateRawTransaction create raw transaction
//
// Parameters:
//
//	utxos - all unspent utxos
//	withdrawals - all withdrawals
//	changeAddress - change address
//	changeAmount - change amount
//	net - bitcoin network
//
// Returns:
//
//	tx - raw transaction
//	dustWithdraw - value lower than dust limit withdraw id
//	error - error if any
func CreateRawTransaction(utxos []*db.Utxo, withdrawals []*db.Withdraw, changeAddress string, changeAmount, estimatedFee int64, net *chaincfg.Params) (*wire.MsgTx, uint, error) {
	tx := wire.NewMsgTx(wire.TxVersion)

	// add utxos as transaction inputs
	for _, utxo := range utxos {
		hash, err := chainhash.NewHashFromStr(utxo.Txid)
		if err != nil {
			return nil, 0, err
		}
		outPoint := wire.NewOutPoint(hash, uint32(utxo.OutIndex))
		tx.AddTxIn(wire.NewTxIn(outPoint, nil, nil))
	}

	// actual fee for withdraw
	actualFee := int64(0)
	if len(withdrawals) > 0 {
		totalTxout := len(withdrawals)
		if changeAmount > 0 {
			totalTxout++
		}
		actualFee = estimatedFee / int64(totalTxout)
	}

	// add outputs (withdrawals)
	for _, withdrawal := range withdrawals {
		addr, err := btcutil.DecodeAddress(withdrawal.To, net)
		if err != nil {
			return nil, 0, err
		}
		pkScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, 0, err
		}
		val := int64(withdrawal.Amount) - actualFee
		if val <= types.DUST_LIMIT {
			return nil, withdrawal.ID, fmt.Errorf("withdrawal amount too small after fee deduction: %d", val)
		}
		// re-set TxFee field
		withdrawal.TxFee = withdrawal.Amount - uint64(val)
		tx.AddTxOut(wire.NewTxOut(val, pkScript))
	}

	// add change output
	if changeAmount-actualFee > 0 {
		changeAddr, err := btcutil.DecodeAddress(changeAddress, net)
		if err != nil {
			return nil, 0, err
		}
		changePkScript, err := txscript.PayToAddrScript(changeAddr)
		if err != nil {
			return nil, 0, err
		}
		tx.AddTxOut(wire.NewTxOut(changeAmount-actualFee, changePkScript))
	}

	return tx, 0, nil
}

// SignTransactionByPrivKey, use PrivKey to sign the transaction, and select the corresponding signature method according to the UTXO type
//
// Parameters:
//
//	privKey - private key
//	tx - transaction
//	utxos - all unspent utxos
//	net - bitcoin network
//
// Returns:
//
//	error - error if any
func SignTransactionByPrivKey(privKey *btcec.PrivateKey, tx *wire.MsgTx, utxos []*db.Utxo, net *chaincfg.Params) error {
	for i, utxo := range utxos {
		switch utxo.ReceiverType {
		// P2PKH
		case types.WALLET_TYPE_P2PKH:
			// P2PKH uncompressed address
			addr, err := btcutil.DecodeAddress(utxo.Sender, net)
			if err != nil {
				return err
			}
			pkScript, err := txscript.PayToAddrScript(addr)
			if err != nil {
				return err
			}

			// create P2PKH signature script
			sigScript, err := txscript.SignatureScript(tx, i, pkScript, txscript.SigHashAll, privKey, true)
			if err != nil {
				return err
			}

			// set signature script
			tx.TxIn[i].SignatureScript = sigScript

		// P2WPKH
		case types.WALLET_TYPE_P2WPKH:
			// P2WPKH needs witness data
			addr, err := btcutil.DecodeAddress(utxo.Sender, net)
			if err != nil {
				return err
			}
			pkScript, err := txscript.PayToAddrScript(addr)
			if err != nil {
				return err
			}

			// create witness signature
			witnessSig, err := txscript.RawTxInWitnessSignature(tx, txscript.NewTxSigHashes(tx, nil), i, utxo.Amount, pkScript, txscript.SigHashAll, privKey)
			if err != nil {
				return err
			}

			// set Witness data
			tx.TxIn[i].Witness = wire.TxWitness{
				witnessSig,                             // signature
				privKey.PubKey().SerializeCompressed(), // public key
			}

		// P2WSH
		case types.WALLET_TYPE_P2WSH:
			// P2WSH needs redeemScript
			// assume redeemScript is known

			// generate witness signature
			witnessSig, err := txscript.RawTxInWitnessSignature(tx, txscript.NewTxSigHashes(tx, nil), i, utxo.Amount, utxo.RedeemScript, txscript.SigHashAll, privKey)
			if err != nil {
				return err
			}

			// set Witness data
			tx.TxIn[i].Witness = wire.TxWitness{
				witnessSig,                             // signature
				privKey.PubKey().SerializeCompressed(), // public key
				utxo.RedeemScript,                      // redeem script
			}

		default:
			return fmt.Errorf("unknown UTXO type: %s", utxo.ReceiverType)
		}
	}

	return nil
}

// TODO sign with Fireblocks
