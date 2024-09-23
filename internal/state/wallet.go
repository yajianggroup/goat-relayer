package state

import (
	"fmt"

	"github.com/goatnetwork/goat-relayer/internal/db"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func (s *State) UpdateUtxoStatusProcessed(txid string, out uint) error {
	s.walletMu.Lock()
	defer s.walletMu.Unlock()

	utxo, err := s.getUtxo(txid, out)
	if err != nil {
		return err
	}

	if utxo.Status == "pending" || utxo.Status == "spent" {
		return fmt.Errorf("wallet state UpdateUtxoStatusProcessed failed, utxo %s:%d, status %s", txid, out, utxo.Status)
	}
	if utxo.Status == "processed" {
		return nil
	}
	utxo.Status = "processed"
	return s.saveUtxo(utxo)
}

func (s *State) AddUtxo(utxo *db.Utxo) error {
	s.walletMu.Lock()
	defer s.walletMu.Unlock()

	utxoExists, err := s.getUtxo(utxo.Txid, utxo.OutIndex)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	if err == nil {
		// update
		utxo.ID = utxoExists.ID
		utxo.Status = utxoExists.Status
	}

	if utxo.Status == "confirmed" {
		// TODO check deposit table (from layer2)
		// if found
		utxo.Source = "deposit"
		utxo.Status = "processed"
	}

	return s.saveUtxo(utxo)
}

func (s *State) getUtxo(txid string, out uint) (*db.Utxo, error) {
	var utxo db.Utxo
	result := s.dbm.GetWalletDB().Where("txid=? and out_index=?", txid, out).First(&utxo)
	if result.Error != nil {
		return nil, result.Error
	}
	return &utxo, nil
}

func (s *State) saveUtxo(utxo *db.Utxo) error {
	result := s.dbm.GetWalletDB().Save(utxo)
	if result.Error != nil {
		log.Errorf("State saveUTXO error: %v", result.Error)
		return result.Error
	}
	s.walletState.Utxo = utxo
	return nil
}
