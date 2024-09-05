package db

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/goatnetwork/goat-relayer/internal/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DatabaseManager interface {
	InitDB()
	GetDB() *gorm.DB
}

type DatabaseManagerImpl struct {
	DB *gorm.DB
}

func (dm *DatabaseManagerImpl) InitDB() {
	dbDir := config.AppConfig.DbDir
	if err := os.MkdirAll(dbDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create database directory: %v", err)
	}

	dbPath := filepath.Join(dbDir, "relayer_data.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	dm.DB = db
	log.Debugf("Database connected successfully, path: %s", dbPath)

	MigrateDB(dm.DB)
	log.Debugf("Database migration completed successfully")
}

func (dm *DatabaseManagerImpl) GetDB() *gorm.DB {
	return dm.DB
}
