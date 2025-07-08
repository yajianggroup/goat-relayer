package main

import (
	"encoding/csv"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/goatnetwork/goat-relayer/internal/types"
)

func main() {
	var (
		pubKeyHex   = flag.String("pubkey", "", "Public key in hex format")
		evmAddress  = flag.String("evm", "", "EVM address")
		networkType = flag.String("network", "mainnet", "Network type: mainnet, testnet3, regtest")
		csvFile     = flag.String("csv", "", "CSV file containing EVM addresses (one per line)")
		help        = flag.Bool("help", false, "Show help message")
	)
	flag.Parse()

	if *help {
		fmt.Println("Usage: p2wsh [options]")
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println("\nExamples:")
		fmt.Println("  p2wsh -pubkey 02... -evm 0x123... -network regtest")
		fmt.Println("  p2wsh -pubkey 02... -csv addresses.csv -network regtest")
		os.Exit(0)
	}

	if *pubKeyHex == "" {
		log.Fatal("Public key is required. Use -pubkey flag.")
	}

	// Decode public key
	pubKey, err := hex.DecodeString(*pubKeyHex)
	if err != nil {
		log.Fatalf("Invalid public key hex: %v", err)
	}

	// Get network parameters
	var net *chaincfg.Params
	switch *networkType {
	case "mainnet":
		net = &chaincfg.MainNetParams
	case "testnet3":
		net = &chaincfg.TestNet3Params
	case "regtest":
		net = &chaincfg.RegressionNetParams
	default:
		log.Fatalf("Invalid network type: %s", *networkType)
	}

	// Handle single EVM address
	if *evmAddress != "" {
		address, err := types.GenerateV0P2WSHAddress(pubKey, *evmAddress, net)
		if err != nil {
			log.Fatalf("Failed to generate P2WSH address: %v", err)
		}

		fmt.Printf("P2WSH Address: %s\n", address.String())
		fmt.Printf("Network: %s\n", *networkType)
		fmt.Printf("Public Key: %s\n", *pubKeyHex)
		fmt.Printf("EVM Address: %s\n", *evmAddress)
		return
	}

	// Handle CSV file
	if *csvFile != "" {
		evmAddresses, err := readEVMAddressesFromCSV(*csvFile)
		if err != nil {
			log.Fatalf("Failed to read CSV file: %v", err)
		}

		fmt.Printf("Generating P2WSH addresses for %d EVM addresses...\n", len(evmAddresses))
		fmt.Printf("Network: %s\n", *networkType)
		fmt.Printf("Public Key: %s\n", *pubKeyHex)
		fmt.Println()

		for i, evmAddr := range evmAddresses {
			address, err := types.GenerateV0P2WSHAddress(pubKey, evmAddr, net)
			if err != nil {
				log.Printf("Failed to generate P2WSH address for %s: %v", evmAddr, err)
				continue
			}
			fmt.Printf("%d. EVM: %s -> P2WSH: %s\n", i+1, evmAddr, address.String())
		}
		return
	}

	log.Fatal("Either -evm or -csv flag is required.")
}

func readEVMAddressesFromCSV(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var addresses []string
	for i, record := range records {
		if i == 0 && len(record) > 0 && strings.ToLower(strings.TrimSpace(record[0])) == "address" {
			// Skip header row
			continue
		}

		if len(record) > 0 {
			addr := strings.TrimSpace(record[0])
			if addr != "" {
				addresses = append(addresses, addr)
			}
		}
	}

	return addresses, nil
}
