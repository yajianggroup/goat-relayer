package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/goatnetwork/goat-relayer/internal/types"
)

func main() {
	var (
		pubKeyHex   = flag.String("pubkey", "", "Public key in hex format")
		evmAddress  = flag.String("evm", "", "EVM address")
		networkType = flag.String("network", "mainnet", "Network type: mainnet, testnet3, regtest")
		help        = flag.Bool("help", false, "Show help message")
	)
	flag.Parse()

	if *help {
		fmt.Println("Usage: p2wsh [options]")
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *pubKeyHex == "" {
		log.Fatal("Public key is required. Use -pubkey flag.")
	}

	if *evmAddress == "" {
		log.Fatal("EVM address is required. Use -evm flag.")
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

	// Generate P2WSH address
	address, err := types.GenerateV0P2WSHAddress(pubKey, *evmAddress, net)
	if err != nil {
		log.Fatalf("Failed to generate P2WSH address: %v", err)
	}

	fmt.Printf("P2WSH Address: %s\n", address.String())
	fmt.Printf("Network: %s\n", *networkType)
	fmt.Printf("Public Key: %s\n", *pubKeyHex)
	fmt.Printf("EVM Address: %s\n", *evmAddress)
}
