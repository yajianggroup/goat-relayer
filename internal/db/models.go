package db

import (
	"time"

	log "github.com/sirupsen/logrus"

	"gorm.io/gorm"
)

type BTCTransaction struct {
	ID          uint      `gorm:"primaryKey"`
	TxID        string    `gorm:"uniqueIndex;not null"`
	RawTxData   string    `gorm:"type:text;not null"`
	ReceivedAt  time.Time `gorm:"not null"`
	Processed   bool      `gorm:"default:false"`
	ProcessedAt time.Time
}

type EVMSyncStatus struct {
	ID            uint      `gorm:"primaryKey"`
	LastSyncBlock uint64    `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`
}

type WithdrawalRecord struct {
	ID           uint      `gorm:"primaryKey"`
	WithdrawalID string    `gorm:"uniqueIndex;not null"`
	UserAddress  string    `gorm:"not null"`
	Amount       string    `gorm:"not null"`
	DetectedAt   time.Time `gorm:"not null"`
	OnChain      bool      `gorm:"default:false"`
	OnChainTxID  string
	Processed    bool `gorm:"default:false"`
	ProcessedAt  time.Time
}

// SubmitterRotation contains block number and current submitter
type SubmitterRotation struct {
	ID               uint64 `gorm:"primaryKey"`
	BlockNumber      uint64 `gorm:"not null"`
	CurrentSubmitter string `gorm:"not null"`
}

// Participant save all participants
type Participant struct {
	ID      uint64 `gorm:"primaryKey"`
	Address string `gorm:"uniqueIndex;not null"`
}

func MigrateDB(db *gorm.DB) {
	if err := db.AutoMigrate(&BTCTransaction{}, &EVMSyncStatus{}, &WithdrawalRecord{}, &SubmitterRotation{}, &Participant{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
}
