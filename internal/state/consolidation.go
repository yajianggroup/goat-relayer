package state

import (
	"github.com/goatnetwork/goat-relayer/internal/db"
	"gorm.io/gorm"
)

type ConsolidationStateStore interface {
	IsConsolidationInProgress() bool
}

func (s *State) IsConsolidationInProgress() bool {
	s.walletMu.RLock()
	defer s.walletMu.RUnlock()

	var order db.SendOrder
	err := s.dbm.GetWalletDB().Where("status IN (?)", []string{db.ORDER_STATUS_AGGREGATING, db.ORDER_STATUS_INIT, db.ORDER_STATUS_PENDING}).Order("id desc").First(&order).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return false
	}
	return true
}
