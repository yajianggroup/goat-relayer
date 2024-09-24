package state

import (
	"time"

	"github.com/goatnetwork/goat-relayer/internal/db"
	"gorm.io/gorm"
)

// CreateWithdrawal
// when a new withdrawal request is detected, save to unconfirmed
func (s *State) CreateWithdrawal(evmTxId string, from string, to string, amount float64, maxTxFee uint) error {
	s.walletMu.Lock()
	defer s.walletMu.Unlock()

	// check if exist, if not save to db
	_, err := s.queryWithdrawByEvmTxId(evmTxId)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	if err == nil {
		// exist
		return nil
	}

	withdraw := &db.Withdraw{
		EvmTxId:   evmTxId,
		From:      from,
		To:        to,
		Amount:    amount,
		MaxTxFee:  maxTxFee,
		Status:    "unconfirm",
		UpdatedAt: time.Now(),
	}
	result := s.dbm.GetWalletDB().Save(withdraw)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// SaveConfirmWithdraw
// when a withdrawal request is confirmed, save to confirmed
func (s *State) SaveConfirmWithdraw(evmTxId string) error {
	s.walletMu.Lock()
	defer s.walletMu.Unlock()

	withdraw, err := s.queryWithdrawByEvmTxId(evmTxId)
	if err != nil {
		return err
	}

	withdraw.Status = "confirmed"
	withdraw.UpdatedAt = time.Now()

	result := s.dbm.GetWalletDB().Save(withdraw)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// UpdateProcessedWithdraw
// when a withdrawal request is processed, save to processed
func (s *State) UpdateProcessedWithdraw(txId string) error {
	s.walletMu.Lock()
	defer s.walletMu.Unlock()

	withdraw, err := s.queryWithdrawByEvmTxId(txId)
	if err != nil {
		return err
	}

	withdraw.Status = "processed"
	withdraw.UpdatedAt = time.Now()

	result := s.dbm.GetWalletDB().Save(withdraw)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// queryWithdrawByEvmTxId
func (s *State) queryWithdrawByEvmTxId(evmTxId string) (*db.Withdraw, error) {
	var withdraw db.Withdraw
	result := s.dbm.GetWalletDB().Where("evm_tx_id=?", evmTxId).First(&withdraw)
	if result.Error != nil {
		return nil, result.Error
	}
	return &withdraw, nil
}
