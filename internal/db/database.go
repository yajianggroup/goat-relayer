package db

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"

	"github.com/goatnetwork/goat-relayer/internal/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DatabaseManager struct {
	db    *gorm.DB
	cache *leveldb.DB
}

func NewDatabaseManager() *DatabaseManager {
	dm := &DatabaseManager{}
	dm.initDB()
	dm.initCache()
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

func (dm *DatabaseManager) initCache() {
	dbPath := filepath.Join(config.AppConfig.DbDir, "btc_cache.db")
	db, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		log.Fatalf("Failed to open cache database: %v", err)
	}
	dm.cache = db
	log.Debugf("Cache database connected successfully, path: %s", dbPath)
}

func (dm *DatabaseManager) GetDB() *gorm.DB {
	return dm.db
}

func (dm *DatabaseManager) GetCacheDB() *leveldb.DB {
	return dm.cache
}

func (dm *DatabaseManager) Close() {
	if dm.cache != nil {
		dm.cache.Close()
	}
}
