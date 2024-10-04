package state

import (
	"fmt"
	"time"

	"github.com/goatnetwork/goat-relayer/internal/db"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// CreateWithdrawal, when a new withdrawal request is detected, save to unconfirmed
//
// Parameters:
//
//	address - to btc p2pkh, p2wpkh address
//	block - goat block
//	id - request id
//	txPrice - user set txPrice for withdraw
//	amount - user request withdraw amount of btc (unit satoshis)
func (s *State) CreateWithdrawal(address string, block, id, txPrice, amount uint64) error {
	s.walletMu.Lock()
	defer s.walletMu.Unlock()

	// check if exist, if not save to db
	_, err := s.getWithdrawByRequestId(id)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	if err == nil {
		// exist
		return nil
	}

	withdraw := &db.Withdraw{
		RequestId: id,
		GoatBlock: block,
		From:      "",
		To:        address,
		Amount:    amount,
		TxPrice:   txPrice,
		TxFee:     0,
		Status:    db.WITHDRAW_STATUS_CREATE,
		OrderId:   "",
		Reason:    "",
		UpdatedAt: time.Now(),
		CreatedAt: time.Now(),
	}

	return s.saveWithdraw(withdraw)
}

// CreateSendOrder, create a send order when start withdrawal or consolidation
func (s *State) CreateSendOrder(order *db.SendOrder, selectedUtxos []*db.Utxo, selectedWithdraws []*db.Withdraw, vins []*db.Vin, vouts []*db.Vout, isProposer bool) error {
	s.walletMu.Lock()
	defer s.walletMu.Unlock()

	err := s.dbm.GetWalletDB().Transaction(func(tx *gorm.DB) error {
		err := s.saveOrder(tx, order)
		if err != nil {
			return err
		}

		if err = tx.Create(&vins).Error; err != nil {
			return err
		}

		if err = tx.Create(&vouts).Error; err != nil {
			return err
		}

		// update utxo status
		for _, utxo := range selectedUtxos {
			var utxoInDb db.Utxo
			if err = tx.Where("txid = ? and out_index = ?", utxo.Txid, utxo.OutIndex).Order("id desc").First(&utxoInDb).Error; err != nil {
				return err
			}
			if isProposer {
				if utxoInDb.Status == db.UTXO_STATUS_UNCONFIRM || utxoInDb.Status == db.UTXO_STATUS_PENDING || utxoInDb.Status == db.UTXO_STATUS_SPENT {
					return fmt.Errorf("utxo status can not be make withdrawal or consolidation, utxo id: %d, status: %s", utxoInDb.ID, utxoInDb.Status)
				}
				err = tx.Model(&db.Utxo{}).Where("id = ?", utxoInDb.ID).Updates(&db.Utxo{Status: db.UTXO_STATUS_PENDING, UpdatedAt: time.Now()}).Error
				if err != nil {
					log.Errorf("State CreateSendOrder update utxo records error: %v", err)
					return err
				}
			} else {
				// voter perhaps receipt multiple orders with the same withdraw
				if utxoInDb.Status == db.UTXO_STATUS_UNCONFIRM || utxoInDb.Status == db.UTXO_STATUS_SPENT {
					return fmt.Errorf("utxo status can not be make withdrawal or consolidation, utxo id: %d, status: %s", utxoInDb.ID, utxoInDb.Status)
				}
				if utxoInDb.Status != db.UTXO_STATUS_PENDING {
					err = tx.Model(&db.Utxo{}).Where("id = ?", utxoInDb.ID).Updates(&db.Utxo{Status: db.UTXO_STATUS_PENDING, UpdatedAt: time.Now()}).Error
					if err != nil {
						log.Errorf("State CreateSendOrder update utxo records error: %v", err)
						return err
					}
				}
			}
		}

		// update withdraw status
		if order.OrderType == db.ORDER_TYPE_WITHDRAWAL {
			for _, withdraw := range selectedWithdraws {
				var withdrawInDb db.Withdraw
				if err = tx.Where("request_id = ?", withdraw.RequestId).First(&withdrawInDb).Error; err != nil {
					return err
				}
				if isProposer {
					if withdrawInDb.Status != db.WITHDRAW_STATUS_CREATE {
						return fmt.Errorf("withdraw status can not be make withdrawal, id: %d, status: %s", withdrawInDb.ID, withdrawInDb.Status)
					}
					err = tx.Model(&db.Withdraw{}).Where("id = ?", withdrawInDb.ID).Updates(&db.Utxo{Status: db.WITHDRAW_STATUS_AGGREGATING, UpdatedAt: time.Now()}).Error
					if err != nil {
						log.Errorf("State CreateSendOrder update withdraw records error: %v", err)
						return err
					}
				} else {
					// voter perhaps receipt multiple orders with the same withdraw
					if withdrawInDb.Status != db.WITHDRAW_STATUS_CREATE && withdrawInDb.Status != db.WITHDRAW_STATUS_AGGREGATING {
						return fmt.Errorf("withdraw status can not be make withdrawal, id: %d, status: %s", withdrawInDb.ID, withdrawInDb.Status)
					}
					if withdrawInDb.Status == db.WITHDRAW_STATUS_CREATE {
						err = tx.Model(&db.Withdraw{}).Where("id = ?", withdrawInDb.ID).Updates(&db.Utxo{Status: db.WITHDRAW_STATUS_AGGREGATING, UpdatedAt: time.Now()}).Error
						if err != nil {
							log.Errorf("State CreateSendOrder update withdraw records error: %v", err)
							return err
						}
					}
				}
			}
		}

		return nil
	})
	return err
}

func (s *State) UpdateWithdrawReplace(id, txPrice uint64) error {
	s.walletMu.Lock()
	defer s.walletMu.Unlock()

	withdraw, err := s.getWithdrawByRequestId(id)
	if err != nil {
		return err
	}
	if withdraw.Status != db.WITHDRAW_STATUS_CREATE && withdraw.Status != db.WITHDRAW_STATUS_AGGREGATING {
		// ignore if it is not create status
		// it is hard to update after start withdraw sig program
		return nil
	}

	withdraw.TxPrice = txPrice
	withdraw.UpdatedAt = time.Now()

	// TODO notify stop aggregating if it is aggregating status, set to closed

	return s.saveWithdraw(withdraw)
}

func (s *State) UpdateWithdrawCancel(id uint64) error {
	s.walletMu.Lock()
	defer s.walletMu.Unlock()

	withdraw, err := s.getWithdrawByRequestId(id)
	if err != nil {
		return err
	}
	if withdraw.Status != db.WITHDRAW_STATUS_CREATE && withdraw.Status != db.WITHDRAW_STATUS_AGGREGATING {
		// ignore if it is not create status
		// it is hard to update after start withdraw sig program
		return nil
	}

	withdraw.Status = db.WITHDRAW_STATUS_CLOSED
	withdraw.UpdatedAt = time.Now()

	// TODO notify stop aggregating if it is aggregating status, set to closed

	return s.saveWithdraw(withdraw)
}

// UpdateSendOrderConfirmed
// when a withdrawal or consolidation request is confirmed, save to confirmed
func (s *State) UpdateSendOrderConfirmed(txid string) error {
	s.walletMu.Lock()
	defer s.walletMu.Unlock()

	err := s.dbm.GetWalletDB().Transaction(func(tx *gorm.DB) error {
		order, err := s.getOrderByTxid(tx, txid)
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		// order found
		if order != nil && err == nil {
			if order.Status == db.ORDER_STATUS_CONFIRMED || order.Status == db.ORDER_STATUS_PROCESSED || order.Status == db.ORDER_STATUS_CLOSED {
				return nil
			}

			order.Status = db.ORDER_STATUS_CONFIRMED
			order.UpdatedAt = time.Now()

			err = s.saveOrder(tx, order)
			if err != nil {
				return err
			}
			err = s.updateOtherStatusByOrder(tx, order.OrderId, db.ORDER_STATUS_CONFIRMED)
			if err != nil {
				return err
			}
			return nil
		}

		// order not found, update by txid, it happens in recovery model
		// should check withdraw[0] when layer2 fast than BTC, if withdraw exists and status processed, update other status to processed
		otherStatus := db.WITHDRAW_STATUS_CONFIRMED
		found, err := s.hasWithdrawByTxidAndStatus(tx, txid, db.WITHDRAW_STATUS_PROCESSED)
		if err != nil {
			return err
		}
		if found {
			otherStatus = db.WITHDRAW_STATUS_PROCESSED
		}
		err = s.updateOtherStatusByTxid(tx, txid, otherStatus)
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func (s *State) UpdateWithdrawInitialized(txid string) error {
	s.walletMu.Lock()
	defer s.walletMu.Unlock()

	err := s.dbm.GetWalletDB().Transaction(func(tx *gorm.DB) error {
		order, err := s.getOrderByTxid(tx, txid)
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		// order found
		if order != nil && err == nil {
			if order.Status != db.ORDER_STATUS_AGGREGATING && order.Status != db.ORDER_STATUS_CLOSED {
				return nil
			}

			order.Status = db.ORDER_STATUS_INIT
			order.UpdatedAt = time.Now()

			err = s.saveOrder(tx, order)
			if err != nil {
				return err
			}
			err = s.updateOtherStatusByOrder(tx, order.OrderId, db.ORDER_STATUS_INIT)
			if err != nil {
				return err
			}
			return nil
		}
		// order not found, do nothing
		return nil
	})
	return err
}

func (s *State) UpdateWithdrawFinalized(txid string) error {
	s.walletMu.Lock()
	defer s.walletMu.Unlock()

	err := s.dbm.GetWalletDB().Transaction(func(tx *gorm.DB) error {
		order, err := s.getOrderByTxid(tx, txid)
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		// order found
		if order != nil && err == nil {
			if order.Status == db.ORDER_STATUS_PROCESSED {
				return nil
			}

			order.Status = db.ORDER_STATUS_PROCESSED
			order.UpdatedAt = time.Now()

			err = s.saveOrder(tx, order)
			if err != nil {
				return err
			}
			err = s.updateOtherStatusByOrder(tx, order.OrderId, db.ORDER_STATUS_PROCESSED)
			if err != nil {
				return err
			}
			return nil
		}

		// not found update by txid
		err = s.updateOtherStatusByTxid(tx, txid, db.ORDER_STATUS_PROCESSED)
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

// CleanProcessingWithdraw, clean all status "aggregating" orders and related withdraws
func (s *State) CleanProcessingWithdraw() error {
	s.walletMu.Lock()
	defer s.walletMu.Unlock()

	err := s.dbm.GetWalletDB().Transaction(func(tx *gorm.DB) error {
		orders, err := s.getOrderByStatuses(tx, db.ORDER_STATUS_AGGREGATING)
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		// order found
		if len(orders) > 0 && err == nil {
			for _, order := range orders {
				if order.Status != db.ORDER_STATUS_AGGREGATING {
					continue
				}
				order.Status = db.ORDER_STATUS_CLOSED
				order.UpdatedAt = time.Now()

				err = s.saveOrder(tx, order)
				if err != nil {
					return err
				}

				// update UTXO from pending to processed by vins
				vins, err := s.getVinsByOrderId(tx, order.OrderId)
				if err != nil {
					return err
				}
				for _, vin := range vins {
					var utxoInDb db.Utxo
					if err = tx.Where("txid = ? and out_index = ?", vin.Txid, vin.OutIndex).First(&utxoInDb).Error; err != nil {
						continue
					}
					if utxoInDb.Status != db.UTXO_STATUS_PENDING {
						continue
					}
					err = tx.Model(&db.Utxo{}).Where("id = ?", utxoInDb.ID).Updates(&db.Utxo{Status: db.UTXO_STATUS_PROCESSED, UpdatedAt: time.Now()}).Error
					if err != nil {
						log.Errorf("State CleanProcessingWithdraw update utxo txid %s - out %d error: %v", utxoInDb.Txid, utxoInDb.OutIndex, err)
						return err
					}
				}

				err = s.updateOtherStatusByOrder(tx, order.OrderId, db.ORDER_STATUS_CLOSED)
				if err != nil {
					return err
				}
				return nil
			}
		}

		// not found order, do not delete withdraw, just update status to create
		err = s.updateWithdrawStatusByStatuses(tx, db.WITHDRAW_STATUS_CREATE, db.WITHDRAW_STATUS_AGGREGATING)
		if err != nil {
			return err
		}

		// not found order, do not delete vin/vout, just update status to closed
		err = s.updateOtherStatusByStatuses(tx, db.WITHDRAW_STATUS_CLOSED, db.WITHDRAW_STATUS_CREATE, db.WITHDRAW_STATUS_AGGREGATING)
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func (s *State) GetWithdrawsCanStart() ([]*db.Withdraw, error) {
	s.walletMu.RLock()
	defer s.walletMu.RUnlock()

	withdraws, err := s.getWithdrawByStatuses(nil, "tx_price desc", db.WITHDRAW_STATUS_CREATE)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, nil
	}

	return withdraws, nil
}

func (s *State) CloseWithdraw(id uint, reason string) error {
	s.walletMu.RLock()
	defer s.walletMu.RUnlock()

	err := s.dbm.GetWalletDB().Model(&db.Withdraw{}).Where("id = ?", id).Updates(&db.Withdraw{Status: db.WITHDRAW_STATUS_CLOSED, Reason: reason, UpdatedAt: time.Now()}).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Errorf("State CloseWithdraw Withdraw by id: %d, reason: %s, error: %v", id, reason, err)
		return err
	}
	return nil
}

func (s *State) getWithdraw(evmTxId string) (*db.Withdraw, error) {
	var withdraw db.Withdraw
	result := s.dbm.GetWalletDB().Where("evm_tx_id=?", evmTxId).Order("id desc").First(&withdraw)
	if result.Error != nil {
		return nil, result.Error
	}
	return &withdraw, nil
}

func (s *State) getSendOrder(orderId string) (*db.SendOrder, error) {
	var order db.SendOrder
	result := s.dbm.GetWalletDB().Where("order_id=?", orderId).Order("id desc").First(&order)
	if result.Error != nil {
		return nil, result.Error
	}
	return &order, nil
}

// queryWithdrawByEvmTxId
func (s *State) getWithdrawByRequestId(requestId uint64) (*db.Withdraw, error) {
	var withdraw db.Withdraw
	result := s.dbm.GetWalletDB().Where("request_id=?", requestId).First(&withdraw)
	if result.Error != nil {
		return nil, result.Error
	}
	return &withdraw, nil
}

func (s *State) hasWithdrawByTxidAndStatus(tx *gorm.DB, txid string, status string) (bool, error) {
	if tx == nil {
		tx = s.dbm.GetWalletDB()
	}
	var withdraw db.Withdraw
	err := tx.Where("txid=? and status=?", txid, status).First(&withdraw).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return false, err
	}
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	return true, nil
}

func (s *State) getOrderByTxid(tx *gorm.DB, txid string) (*db.SendOrder, error) {
	if tx == nil {
		tx = s.dbm.GetWalletDB()
	}
	var order db.SendOrder
	result := tx.Where("txid=?", txid).Order("id DESC").First(&order)
	if result.Error != nil {
		return nil, result.Error
	}
	return &order, nil
}

func (s *State) getOrderByOrderId(tx *gorm.DB, orderId string) (*db.SendOrder, error) {
	if tx == nil {
		tx = s.dbm.GetWalletDB()
	}
	var order db.SendOrder
	result := tx.Where("order_id=?", orderId).First(&order)
	if result.Error != nil {
		return nil, result.Error
	}
	return &order, nil
}

func (s *State) getOrderByStatuses(tx *gorm.DB, statuses ...string) ([]*db.SendOrder, error) {
	if tx == nil {
		tx = s.dbm.GetWalletDB()
	}
	var orders []*db.SendOrder
	result := tx.Where("status in (?)", statuses).Find(&orders)
	if result.Error != nil {
		return nil, result.Error
	}
	return orders, nil
}

func (s *State) getWithdrawByStatuses(tx *gorm.DB, orderBy string, statuses ...string) ([]*db.Withdraw, error) {
	if tx == nil {
		tx = s.dbm.GetWalletDB()
	}
	if orderBy == "" {
		orderBy = "id asc"
	}
	var withdraws []*db.Withdraw
	result := tx.Where("status in (?)", statuses).Order(orderBy).Find(&withdraws)
	if result.Error != nil {
		return nil, result.Error
	}
	return withdraws, nil
}

func (s *State) saveWithdraw(withdraw *db.Withdraw) error {
	result := s.dbm.GetWalletDB().Save(withdraw)
	if result.Error != nil {
		log.Errorf("State saveWithdraw error: %v", result.Error)
		return result.Error
	}
	return nil
}

func (s *State) saveOrder(tx *gorm.DB, order *db.SendOrder) error {
	if tx == nil {
		tx = s.dbm.GetWalletDB()
	}
	result := tx.Save(order)
	if result.Error != nil {
		log.Errorf("State saveOrder error: %v", result.Error)
		return result.Error
	}
	return nil
}

func (s *State) updateOtherStatusByOrder(tx *gorm.DB, orderId string, status string) error {
	if tx == nil {
		tx = s.dbm.GetWalletDB()
	}
	err := tx.Model(&db.Withdraw{}).Where("order_id = ?", orderId).Updates(&db.Withdraw{Status: status, UpdatedAt: time.Now()}).Error
	if err != nil {
		log.Errorf("State updateOtherStatusByOrder Withdraw by order id: %s, status: %s, error: %v", orderId, status, err)
		return err
	}
	err = tx.Model(&db.Vin{}).Where("order_id = ?", orderId).Updates(&db.Vin{Status: status, UpdatedAt: time.Now()}).Error
	if err != nil {
		log.Errorf("State updateOtherStatusByOrder Vin by order id: %s, status: %s, error: %v", orderId, status, err)
		return err
	}
	err = tx.Model(&db.Vout{}).Where("order_id = ?", orderId).Updates(&db.Vout{Status: status, UpdatedAt: time.Now()}).Error
	if err != nil {
		log.Errorf("State updateOtherStatusByOrder Vout by order id: %s, status: %s, error: %v", orderId, status, err)
		return err
	}
	return nil
}

func (s *State) updateWithdrawStatusByStatuses(tx *gorm.DB, newStatus string, rawStatuses ...string) error {
	if tx == nil {
		tx = s.dbm.GetWalletDB()
	}
	err := tx.Model(&db.Withdraw{}).Where("status in (?)", rawStatuses).Updates(&db.Withdraw{Status: newStatus, UpdatedAt: time.Now()}).Error
	if err != nil {
		log.Errorf("State updateWithdrawStatusByStatuses Withdraw by raw status: %v, new status: %s, error: %v", rawStatuses, newStatus, err)
		return err
	}
	return nil
}

func (s *State) updateOtherStatusByStatuses(tx *gorm.DB, newStatus string, rawStatuses ...string) error {
	if tx == nil {
		tx = s.dbm.GetWalletDB()
	}
	err := tx.Model(&db.Vin{}).Where("status in (?)", rawStatuses).Updates(&db.Vin{Status: newStatus, UpdatedAt: time.Now()}).Error
	if err != nil {
		log.Errorf("State updateOtherStatusByStatuses Vin by raw statuses: %v, new status: %s, error: %v", rawStatuses, newStatus, err)
		return err
	}
	err = tx.Model(&db.Vout{}).Where("status in (?)", rawStatuses).Updates(&db.Vout{Status: newStatus, UpdatedAt: time.Now()}).Error
	if err != nil {
		log.Errorf("State updateOtherStatusByStatuses Vout by raw statuses: %v, new status: %s, error: %v", rawStatuses, newStatus, err)
		return err
	}
	return nil
}

// updateOtherStatusByTxid will update other status by txid, it is used in recovery model that can not restore send order
func (s *State) updateOtherStatusByTxid(tx *gorm.DB, txid string, status string) error {
	if tx == nil {
		tx = s.dbm.GetWalletDB()
	}
	err := tx.Model(&db.Withdraw{}).Where("txid = ?", txid).Updates(&db.Withdraw{Status: status, UpdatedAt: time.Now()}).Error
	if err != nil {
		log.Errorf("State updateOtherStatusByTxid Withdraw by order id: %s, status: %s, error: %v", txid, status, err)
		return err
	}
	err = tx.Model(&db.Vin{}).Where("txid = ?", txid).Updates(&db.Vin{Status: status, UpdatedAt: time.Now()}).Error
	if err != nil {
		log.Errorf("State updateOtherStatusByTxid Vin by order id: %s, status: %s, error: %v", txid, status, err)
		return err
	}
	err = tx.Model(&db.Vout{}).Where("txid = ?", txid).Updates(&db.Vout{Status: status, UpdatedAt: time.Now()}).Error
	if err != nil {
		log.Errorf("State updateOtherStatusByTxid Vout by order id: %s, status: %s, error: %v", txid, status, err)
		return err
	}
	return nil
}
