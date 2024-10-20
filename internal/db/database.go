package db

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goatnetwork/goat-relayer/internal/config"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type DatabaseManager struct {
	l2SyncDb   *gorm.DB
	l2InfoDb   *gorm.DB
	btcLightDb *gorm.DB
	walletDb   *gorm.DB
	btcCacheDb *gorm.DB
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

	databaseConfigs := []struct {
		dbPath string
		dbRef  **gorm.DB
		dbName string
	}{
		{filepath.Join(dbDir, "l2_sync.db"), &dm.l2SyncDb, "Database 1"},
		{filepath.Join(dbDir, "l2_info.db"), &dm.l2InfoDb, "Database 2"},
		{filepath.Join(dbDir, "btc_light.db"), &dm.btcLightDb, "Database 3"},
		{filepath.Join(dbDir, "wallet_order.db"), &dm.walletDb, "Database 4"},
		{filepath.Join(dbDir, "btc_cache.db"), &dm.btcCacheDb, "Database 5"},
	}

	for _, dbConfig := range databaseConfigs {
		if err := dm.connectDatabase(dbConfig.dbPath, dbConfig.dbRef, dbConfig.dbName); err != nil {
			log.Fatalf("Failed to connect to %s: %v", dbConfig.dbName, err)
		}
	}

	dm.autoMigrate()
	log.Debugf("Database migration completed successfully")
}

func (dm *DatabaseManager) connectDatabase(dbPath string, dbRef **gorm.DB, dbName string) error {
	// open database and set WAL mode
	db, err := gorm.Open(sqlite.Open(dbPath+"?_journal_mode=WAL"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", dbName, err)
	}

	*dbRef = db
	log.Debugf("%s connected successfully in WAL mode, path: %s", dbName, dbPath)
	return nil
}

func (dm *DatabaseManager) GetL2SyncDB() *gorm.DB {
	return dm.l2SyncDb
}

func (dm *DatabaseManager) GetL2InfoDB() *gorm.DB {
	return dm.l2InfoDb
}

func (dm *DatabaseManager) GetBtcLightDB() *gorm.DB {
	return dm.btcLightDb
}

func (dm *DatabaseManager) GetWalletDB() *gorm.DB {
	return dm.walletDb
}

func (dm *DatabaseManager) GetBtcCacheDB() *gorm.DB {
	return dm.btcCacheDb
}
