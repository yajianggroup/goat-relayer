package db

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/goatnetwork/goat-relayer/internal/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DatabaseManager struct {
	db *gorm.DB
}

func NewDatabaseManager() *DatabaseManager {
	dm := &DatabaseManager{}

	dm.initDB()

	return dm
}

func (dm *DatabaseManager) initDB() {
	dbDir := config.AppConfig.DbDir
	if err := os.MkdirAll(dbDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create database directory: %v", err)
	}

	dbPath := filepath.Join(dbDir, "relayer_data.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	dm.db = db
	log.Debugf("Database connected successfully, path: %s", dbPath)

	dm.migrateDB()
	log.Debugf("Database migration completed successfully")
}

func (dm *DatabaseManager) GetDB() *gorm.DB {
	return dm.db
}
