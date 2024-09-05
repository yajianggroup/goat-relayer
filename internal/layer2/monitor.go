package layer2

import (
	"context"
	"math/big"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/goatnetwork/goat-relayer/internal/config"
	"github.com/goatnetwork/goat-relayer/internal/db"
	"gorm.io/gorm"
)

var (
	votingManagerABI = `[{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"newParticipant","type":"address"}],"name":"ParticipantAdded","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"participant","type":"address"}],"name":"ParticipantRemoved","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"requester","type":"address"},{"indexed":true,"internalType":"address","name":"currentSubmitter","type":"address"}],"name":"SubmitterRotationRequested","type":"event"},{"inputs":[{"internalType":"address","name":"targetAddress","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"bytes","name":"txInfo","type":"bytes"},{"internalType":"uint256","name":"chainId","type":"uint256"},{"internalType":"bytes","name":"extraInfo","type":"bytes"},{"internalType":"bytes","name":"signature","type":"bytes"}],"name":"submitDepositInfo","outputs":[],"stateMutability":"nonpayable","type":"function"}]`
)

type Layer2Monitor interface {
	StartLayer2Monitor()
}

type Layer2MonitorImpl struct {
	db.DatabaseManager
}

func (l2m *Layer2MonitorImpl) StartLayer2Monitor() {
	client, parsedVotingManagerABI, votingManagerAddress := getClientAndAbi()

	// Get latest sync height
	var syncStatus db.EVMSyncStatus
	result := l2m.GetDB().First(&syncStatus)
	if result.Error != nil && result.Error == gorm.ErrRecordNotFound {
		syncStatus.LastSyncBlock = uint64(config.AppConfig.L2StartHeight)
		syncStatus.UpdatedAt = time.Now()
		l2m.GetDB().Create(&syncStatus)
	}

	for {
		latestBlock, err := client.BlockNumber(context.Background())
		if err != nil {
			log.Fatalf("Error getting latest block number: %v", err)
		}

		targetBlock := latestBlock - uint64(config.AppConfig.L2Confirmations)

		if syncStatus.LastSyncBlock < targetBlock {
			fromBlock := syncStatus.LastSyncBlock + 1
			toBlock := min(fromBlock+uint64(config.AppConfig.L2MaxBlockRange)-1, targetBlock)

			log.WithFields(log.Fields{
				"fromBlock": fromBlock,
				"toBlock":   toBlock,
			}).Info("Syncing L2 goat events")

			filterQuery := ethereum.FilterQuery{
				FromBlock: big.NewInt(int64(fromBlock)),
				ToBlock:   big.NewInt(int64(toBlock)),
				Addresses: []common.Address{votingManagerAddress},
			}

			logs, err := client.FilterLogs(context.Background(), filterQuery)
			if err != nil {
				log.Errorf("Failed to filter logs: %v", err)
				// Next loop
				time.Sleep(config.AppConfig.L2RequestInterval)
				continue
			}

			for _, vLog := range logs {
				processLog(vLog, parsedVotingManagerABI, l2m.GetDB())
				if syncStatus.LastSyncBlock < vLog.BlockNumber {
					syncStatus.LastSyncBlock = vLog.BlockNumber
					syncStatus.UpdatedAt = time.Now()
					l2m.GetDB().Save(&syncStatus)
				}
			}

			if syncStatus.LastSyncBlock < toBlock {
				// Save sync status
				syncStatus.LastSyncBlock = toBlock
				syncStatus.UpdatedAt = time.Now()
				l2m.GetDB().Save(&syncStatus)
			}
		}

		// Next loop
		time.Sleep(config.AppConfig.L2RequestInterval)
	}
}

func getClientAndAbi() (*ethclient.Client, abi.ABI, common.Address) {
	client, err := ethclient.Dial(config.AppConfig.L2RPC)
	if err != nil {
		log.Fatalf("Error creating Layer2 EVM RPC client: %v", err)
	}

	// Decode ABI
	parsedVotingManagerABI, err := abi.JSON(strings.NewReader(votingManagerABI))
	if err != nil {
		log.Fatalf("Error parsing VotingManager ABI: %v", err)
	}

	votingManagerAddress := common.HexToAddress(config.AppConfig.VotingContract)
	return client, parsedVotingManagerABI, votingManagerAddress
}

func processLog(vLog types.Log, parsedABI abi.ABI, database *gorm.DB) {
	// TODO  DepositInfoSubmitted event parse and update deposit tx in db
	switch vLog.Topics[0].Hex() {
	case parsedABI.Events["SubmitterRotationRequested"].ID.Hex():
		event := struct {
			Requester        common.Address
			CurrentSubmitter common.Address
		}{}

		err := parsedABI.UnpackIntoInterface(&event, "SubmitterRotationRequested", vLog.Data)
		if err != nil {
			log.Errorf("Error unpacking SubmitterRotationRequested event: %v", err)
		}

		// save current rotation
		var submitterRotation db.SubmitterRotation
		result := database.First(&submitterRotation)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				submitterRotation.BlockNumber = vLog.BlockNumber
				submitterRotation.CurrentSubmitter = event.CurrentSubmitter.Hex()
				err = database.Create(&submitterRotation).Error
			} else {
				log.Fatalf("Error query db: %v", result.Error)
			}
		} else {
			submitterRotation.BlockNumber = vLog.BlockNumber
			submitterRotation.CurrentSubmitter = event.CurrentSubmitter.Hex()
			err = database.Save(&submitterRotation).Error
		}

		if err != nil {
			log.Fatalf("Error saving SubmitterRotation: %v", err)
		}

		selfAddress := crypto.PubkeyToAddress(config.AppConfig.L2PrivateKey.PublicKey)
		if event.CurrentSubmitter == selfAddress {
			// TODO start submitDeposit call
		} else {
			// TODO stop submitDeposit
		}

	case parsedABI.Events["ParticipantAdded"].ID.Hex():
		event := struct {
			NewParticipant common.Address
		}{}

		err := parsedABI.UnpackIntoInterface(&event, "ParticipantAdded", vLog.Data)
		if err != nil {
			log.Errorf("Error unpacking ParticipantAdded event: %v", err)
			return
		}

		// save locked relayer member from db
		participant := db.Participant{
			Address: event.NewParticipant.Hex(),
		}
		err = database.FirstOrCreate(&participant, "address = ?", participant.Address).Error
		if err != nil {
			log.Fatalf("Error adding Participant: %v", err)
		} else {
			log.Infof("Participant added: %s", event.NewParticipant.Hex())
		}

	case parsedABI.Events["ParticipantRemoved"].ID.Hex():
		event := struct {
			Participant common.Address
		}{}

		err := parsedABI.UnpackIntoInterface(&event, "ParticipantRemoved", vLog.Data)
		if err != nil {
			log.Errorf("Error unpacking ParticipantRemoved event: %v", err)
			return
		}

		// remove relayer member from db
		err = database.Where("address = ?", event.Participant.Hex()).Delete(&db.Participant{}).Error
		if err != nil {
			log.Fatalf("Error removing Participant: %v", err)
		} else {
			log.Infof("Participant removed: %s", event.Participant.Hex())
		}
	}
}

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}
