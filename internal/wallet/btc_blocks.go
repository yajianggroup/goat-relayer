package wallet

import (
	"context"
	"encoding/base64"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/goatnetwork/goat-relayer/internal/config"
	"github.com/goatnetwork/goat-relayer/internal/db"
	"github.com/goatnetwork/goat-relayer/internal/types"
	log "github.com/sirupsen/logrus"
)

// extract Withdrawal and Consolidation
// vin sender is tss, log vin and vout
func (w *WalletServer) blockScanLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Info("WalletService blockScanLoop stopping...")
			return
		case block := <-w.blockCh:
			btcBlock := block.(types.BtcBlockExt)
			log.Debugf("Btc block wallet module filter, height: %d, hash: %s", btcBlock.BlockNumber, btcBlock.BlockHash().String())

			// get pubkey
			pubkey, err := w.state.GetDepositKeyByBtcBlock(btcBlock.BlockNumber)
			if err != nil {
				log.Fatalf("Get current deposit key by btc height %d err %v", btcBlock.BlockNumber, err)
				continue
			}

			var network *chaincfg.Params
			switch config.AppConfig.BTCNetworkType {
			case "":
				network = &chaincfg.MainNetParams
			case "mainnet":
				network = &chaincfg.MainNetParams
			case "regtest":
				network = &chaincfg.RegressionNetParams
			case "testnet3":
				network = &chaincfg.TestNet3Params
			}

			// TODO verify SPV again

			// get vin
			pubkeyBytes, err := base64.StdEncoding.DecodeString(pubkey.PubKey)
			if err != nil {
				log.Fatalf("Base64 decode pubkey %s err %v", pubkey.PubKey, err)
				continue
			}
			p2pkhAddress, err := types.GenerateP2PKHAddress(pubkeyBytes, network)
			if err != nil {
				log.Fatalf("Gen P2PKH address from pubkey %s err %v", pubkey.PubKey, err)
				continue
			}
			p2wpkhAddress, err := types.GenerateP2WPKHAddress(pubkeyBytes, network)
			if err != nil {
				log.Fatalf("Gen P2WPKH address from pubkey %s err %v", pubkey.PubKey, err)
				continue
			}

			var utxos []*db.Utxo
			var vins []*db.Vin
			var vouts []*db.Vout

			for _, tx := range btcBlock.Transactions {
				isUtxo, isVin := false, false
				for _, vin := range tx.TxIn {
					_, addresses, requireSigs, err := txscript.ExtractPkScriptAddrs(vin.SignatureScript, network)
					if err != nil {
						log.Errorf("Error extracting input address, %v", err)
						continue
					}
					if addresses == nil {
						log.Errorf("Error extracting input address nil")
						continue
					}
					if requireSigs > 1 {
						// ignore multi sigs
						continue
					}

					if addresses[0].EncodeAddress() == p2pkhAddress.EncodeAddress() || addresses[0].EncodeAddress() == p2wpkhAddress.EncodeAddress() {
						// should save vin db logic
						isVin = true
						vins = append(vins, &db.Vin{})
					}
				}

				for _, vout := range tx.TxOut {
					_, addresses, requireSigs, err := txscript.ExtractPkScriptAddrs(vout.PkScript, network)
					if err != nil {
						log.Errorf("Error extracting output address, %v", err)
						continue
					}
					if addresses == nil {
						log.Errorf("Error extracting output address nil")
						continue
					}
					if requireSigs > 1 {
						// ignore multi sigs
						continue
					}

					if addresses[0].EncodeAddress() == p2pkhAddress.EncodeAddress() || addresses[0].EncodeAddress() == p2wpkhAddress.EncodeAddress() {
						// should save vout db logic
						isUtxo = true
						utxos = append(utxos, &db.Utxo{})
					}
					vouts = append(vouts, &db.Vout{})
				}

				if isUtxo {
					// TODO save utxo db, check if it is deposit from layer2
					// txout -> utxo

					// TODO if err, fatalf

				}
				if isVin {
					// TODO save vin, vout db, check if it is withdraw from layer2
				}

				// Note, if isUtxo && isVin, it is withdrawal|consolidation with change out to self

				// Note, vout is P2WSH can not find here, it should query from BTC client in recovery model
			}
		}
	}
}
