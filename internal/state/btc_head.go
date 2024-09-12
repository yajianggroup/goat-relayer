package state

import (
	"time"

	"github.com/goatnetwork/goat-relayer/internal/db"
	"gorm.io/gorm"
)

/*
AddUnconfirmBtcBlock
when btc scanner detected a new block, save to unconfirmed,
proposer will start an off-chain signature
*/
func (s *State) AddUnconfirmBtcBlock(height uint64, hash string) error {
	s.btcHeadMu.Lock()
	defer s.btcHeadMu.Unlock()

	// check if exist, if not save to db
	_, err := s.queryBtcBlockByHeigh(height)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	if err == nil {
		// exist height
		return nil
	}

	status := "unconfirm"
	if height <= s.btcHeadState.Latest.Height {
		status = "processed"
	}

	btcBlock := &db.BtcBlock{
		Status:    status,
		UpdatedAt: time.Now(),
		Height:    height,
		Hash:      hash,
	}
	result := s.dbm.GetBtcLightDB().Save(btcBlock)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s *State) SaveConfirmBtcBlock(height uint64, hash string) error {
	s.btcHeadMu.Lock()
	defer s.btcHeadMu.Unlock()

	btcBlock, err := s.queryBtcBlockByHeigh(height)
	status := "confirmed"
	if height <= s.btcHeadState.Latest.Height {
		status = "processed"
	}

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			btcBlock = &db.BtcBlock{
				Status:    status,
				UpdatedAt: time.Now(),
				Height:    height,
				Hash:      hash,
			}
		} else {
			// DB error
			return err
		}
	} else {
		if btcBlock.Status != "processed" {
			btcBlock.Status = status
			btcBlock.Hash = hash
		}
		btcBlock.UpdatedAt = time.Now()
	}
	result := s.dbm.GetBtcLightDB().Save(btcBlock)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s *State) UpdateProcessedBtcBlock(block uint64, height uint64, hash string) error {
	s.btcHeadMu.Lock()
	defer s.btcHeadMu.Unlock()

	btcBlock, err := s.queryBtcBlockByHeigh(height)
	if err != nil {
		// query db failed, update cache
		// this will happen when L2 scan fast than btc
		s.btcHeadState.Latest = db.BtcBlock{
			Height:    height,
			Hash:      hash,
			UpdatedAt: time.Now(),
		}
		return err
	}

	btcBlock.Status = "processed"

	// TODO update height <= height
	result := s.dbm.GetBtcLightDB().Save(btcBlock)
	if result.Error != nil {
		return result.Error
	}

	s.btcHeadState.Latest = *btcBlock
	for i, bb := range s.btcHeadState.UnconfirmQueue {
		if bb.Height == height {
			if i == len(s.btcHeadState.UnconfirmQueue)-1 {
				s.btcHeadState.UnconfirmQueue = s.btcHeadState.UnconfirmQueue[:i]
			} else {
				s.btcHeadState.UnconfirmQueue = append(s.btcHeadState.UnconfirmQueue[:i], s.btcHeadState.UnconfirmQueue[i+1:]...)
			}
			break
		}
	}
	for i, bb := range s.btcHeadState.SigQueue {
		if bb.Height == height {
			if i == len(s.btcHeadState.SigQueue)-1 {
				s.btcHeadState.SigQueue = s.btcHeadState.SigQueue[:i]
			} else {
				s.btcHeadState.SigQueue = append(s.btcHeadState.SigQueue[:i], s.btcHeadState.SigQueue[i+1:]...)
			}
			break
		}
	}

	return nil
}

// GetBtcBlockForSign
func (s *State) GetBtcBlockForSign(size int) ([]*db.BtcBlock, error) {
	s.btcHeadMu.RLock()
	defer s.btcHeadMu.RUnlock()

	from := s.btcHeadState.Latest.Height
	var btcBlocks []*db.BtcBlock
	result := s.dbm.GetBtcLightDB().Where("height > ?", from).Order("height asc").Limit(size).Find(&btcBlocks)
	if result.Error != nil {
		return nil, result.Error
	}
	return btcBlocks, nil
}

// GetCurrentBtcBlock
func (s *State) GetCurrentBtcBlock() (db.BtcBlock, error) {
	s.btcHeadMu.RLock()
	defer s.btcHeadMu.RUnlock()

	return s.btcHeadState.Latest, nil
}

func (s *State) queryBtcBlockByHeigh(height uint64) (*db.BtcBlock, error) {
	var btcBlock db.BtcBlock
	result := s.dbm.GetBtcLightDB().Where("height = ?", height).First(&btcBlock)
	if result.Error != nil {
		return nil, result.Error
	}
	return &btcBlock, nil
}
