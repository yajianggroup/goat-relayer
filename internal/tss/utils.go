package tss

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strconv"

	"github.com/bnb-chain/tss-lib/v2/ecdsa/keygen"
	"github.com/bnb-chain/tss-lib/v2/tss"
	log "github.com/sirupsen/logrus"
)

func createPartyIDs(parties int) tss.SortedPartyIDs {
	partyIDs := make(tss.SortedPartyIDs, parties)
	for i := 0; i < parties; i++ {
		partyIDs[i] = tss.NewPartyID(strconv.Itoa(i), "", new(big.Int).SetInt64(int64(i)))
	}
	return partyIDs
}

func saveTSSData(data *keygen.LocalPartySaveData) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		log.Errorf("Unable to serialize TSS data: %v", err)
		return
	}

	dataDir := "tss_data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Errorf("Failed to create TSS data directory: %v", err)
		return
	}

	filePath := filepath.Join(dataDir, "tss_key_data.json")
	if err := os.WriteFile(filePath, dataBytes, 0644); err != nil {
		log.Errorf("Failed to save TSS data to file: %v", err)
		return
	}

	log.Infof("TSS data successfully saved to: %s", filePath)
}

func loadTSSData() (*keygen.LocalPartySaveData, error) {
	dataDir := "tss_data"
	filePath := filepath.Join(dataDir, "tss_key_data.json")

	dataBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to read TSS data file: %v", err)
	}

	var data keygen.LocalPartySaveData
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return nil, fmt.Errorf("unable to deserialize TSS data: %v", err)
	}

	log.Infof("Successfully loaded TSS data from %s", filePath)
	return &data, nil
}
