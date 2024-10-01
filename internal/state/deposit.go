package state

import (
	"encoding/json"
	"time"

	"github.com/goatnetwork/goat-relayer/internal/db"
	"github.com/goatnetwork/goat-relayer/internal/types"
	"gorm.io/gorm"
)

/*
AddUnconfirmDeposit
when utxo scanner detected a new transaction in 1 confirm, save to unconfirmed,
*/
func (s *State) AddUnconfirmDeposit(txHash string, rawTx string, evmAddr string, signVersion uint32) error {
	s.depositMu.Lock()
	defer s.depositMu.Unlock()

	// check if exist, if not save to db
	_, err := s.queryDepositByTxHash(txHash)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	if err == nil {
		// exist height
		return nil
	}

	status := db.DEPOSIT_STATUS_UNCONFIRM

	deposit := &db.Deposit{
		Status:      status,
		UpdatedAt:   time.Now(),
		TxHash:      txHash,
		RawTx:       rawTx,
		EvmAddr:     evmAddr,
		SignVersion: signVersion,
	}
	result := s.dbm.GetBtcCacheDB().Save(deposit)
	if result.Error != nil {
		return result.Error
	}

	s.depositState.UnconfirmQueue = append(s.depositState.UnconfirmQueue, deposit)

	return nil
}

/*
SaveConfirmDeposit
when utxo scanner detected a new transaction in 6 confirm, save to confirmed
*/
func (s *State) SaveConfirmDeposit(txHash string, rawTx string, evmAddr string) error {
	s.depositMu.Lock()
	defer s.depositMu.Unlock()

	deposit, err := s.queryDepositByTxHash(txHash)
	status := db.DEPOSIT_STATUS_CONFIRMED

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			deposit = &db.Deposit{
				Status:    status,
				UpdatedAt: time.Now(),
			}
		} else {
			// DB error
			return err
		}
	} else {
		if deposit.Status != db.DEPOSIT_STATUS_PROCESSED {
			deposit.Status = status
		}
		deposit.UpdatedAt = time.Now()
	}
	result := s.dbm.GetBtcCacheDB().Save(deposit)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

/*
UpdateProcessedDeposit
when utxo committer process a deposit to consensus, save to processed
*/
func (s *State) UpdateProcessedDeposit(txHash string) error {
	s.depositMu.Lock()
	defer s.depositMu.Unlock()

	// TODO should rewrite this
	deposit, err := s.queryDepositByTxHash(txHash)
	if err != nil {
		// query db failed, update cache
		s.depositState.Latest = db.Deposit{
			TxHash:    deposit.TxHash,
			RawTx:     deposit.RawTx,
			EvmAddr:   deposit.EvmAddr,
			UpdatedAt: time.Now(),
		}
		return err
	}
	deposit.Status = db.DEPOSIT_STATUS_PROCESSED

	result := s.dbm.GetBtcCacheDB().Save(deposit)
	if result.Error != nil {
		return result.Error
	}

	s.depositState.Latest = *deposit
	// remove from unconfirm queue
	for i, ds := range s.depositState.UnconfirmQueue {
		if ds.TxHash == txHash {
			if i == len(s.depositState.UnconfirmQueue)-1 {
				s.depositState.UnconfirmQueue = s.depositState.UnconfirmQueue[:i]
			} else {
				s.depositState.UnconfirmQueue = append(s.depositState.UnconfirmQueue[:i], s.depositState.UnconfirmQueue[i+1:]...)
			}
			break
		}
	}
	// remove from sig queue
	for i, ds := range s.depositState.SigQueue {
		if ds.TxHash == txHash {
			if i == len(s.depositState.SigQueue)-1 {
				s.depositState.SigQueue = s.depositState.SigQueue[:i]
			} else {
				s.depositState.SigQueue = append(s.depositState.SigQueue[:i], s.depositState.SigQueue[i+1:]...)
			}
			break
		}
	}

	return nil
}

// GetDepositForSign get deposits for sign
func (s *State) GetDepositForSign(size int) ([]*db.Deposit, error) {
	s.depositMu.RLock()
	defer s.depositMu.RUnlock()

	from := s.depositState.Latest.ID
	var deposits []*db.Deposit
	result := s.dbm.GetBtcCacheDB().Where("id > ?", from).Order("id asc").Limit(size).Find(&deposits)
	if result.Error != nil {
		return nil, result.Error
	}
	return deposits, nil
}

// GetCurrentDeposit get current deposit
func (s *State) GetCurrentDeposit() (db.Deposit, error) {
	s.depositMu.RLock()
	defer s.depositMu.RUnlock()

	return s.depositState.Latest, nil
}

func (s *State) queryDepositByTxHash(txHash string) (*db.Deposit, error) {
	var deposit db.Deposit
	result := s.dbm.GetBtcCacheDB().Where("tx_hash = ?", txHash).First(&deposit)
	if result.Error != nil {
		return nil, result.Error
	}
	return &deposit, nil
}

func (s *State) QueryUnConfirmDeposit() ([]db.Deposit, error) {
	var deposit []db.Deposit
	result := s.dbm.GetBtcCacheDB().Where("status = ?", db.DEPOSIT_STATUS_UNCONFIRM).Find(&deposit)
	if result.Error != nil {
		return nil, result.Error
	}
	return deposit, nil
}

// QueryBtcBlockDataByHeight query btc block data by height
func (s *State) QueryBtcBlockDataByBlockHashes(blockHashes []string) ([]db.BtcBlockData, error) {
	var btcBlockData []db.BtcBlockData
	result := s.dbm.GetBtcCacheDB().Where("block_hash IN (?)", blockHashes).Find(&btcBlockData)
	if result.Error != nil {
		return nil, result.Error
	}
	return btcBlockData, nil
}

func (s *State) UpdateConfirmedDepositsByBtcHeight(blockHeight uint64, blockHash string) (err error) {
	s.depositMu.Lock()
	defer s.depositMu.Unlock()

	var btcBlockData db.BtcBlockData
	result := s.dbm.GetBtcCacheDB().Where("height = ?", blockHeight).Find(&btcBlockData)
	if result.Error != nil {
		return result.Error
	}

	// Save to confirmed while the block has the deposit tx
	if err != nil {
		return err
	}
	for _, deposit := range s.depositState.UnconfirmQueue {
		txHashes := make([]string, 0)
		err := json.Unmarshal([]byte(btcBlockData.TxHashes), &txHashes)
		if err != nil {
			return err
		}
		merkleRoot, proofBytes, txIndex, err := types.GenerateSPVProof(deposit.TxHash, txHashes)
		if deposit.Status == "unconfirm" && txIndex != -1 {
			if err != nil {
				return err
			}
			deposit.Status = "confirmed"
			deposit.BlockHash = blockHash
			deposit.BlockHeight = blockHeight
			deposit.TxIndex = uint64(txIndex)
			deposit.MerkleRoot = string(merkleRoot)
			deposit.Proof = string(proofBytes)
			deposit.UpdatedAt = time.Now()
			result := s.dbm.GetBtcCacheDB().Save(deposit)
			if result.Error != nil {
				return result.Error
			}
		}
	}

	return nil
}
