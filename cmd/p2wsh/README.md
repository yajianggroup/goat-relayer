# P2WSH Address Generator

This tool is used to generate P2WSH (Pay-to-Witness-Script-Hash) addresses.

## Build

```bash
make p2wsh
```

Or build directly:

```bash
go build -o bin/p2wsh cmd/p2wsh/main.go
```

## Usage

### Single Address Generation

```bash
./bin/p2wsh -pubkey <public_key_hex> -evm <evm_address> [-network <network_type>]
```

### Batch Address Generation from CSV

```bash
./bin/p2wsh -pubkey <public_key_hex> -csv <csv_file> [-network <network_type>]
```

### Parameters

- `-pubkey`: Public key in hex format (required)
- `-evm`: EVM address (required for single address generation)
- `-csv`: CSV file containing EVM addresses (required for batch generation)
- `-network`: Network type, options: mainnet, testnet3, regtest (default: mainnet)
- `-help`: Show help message

### CSV File Format

The CSV file should contain EVM addresses, one per line. The first line can be a header (e.g., "address") and will be automatically skipped.

Example CSV format:
```csv
address
0xab0c51effd6c6a71a900f8ee8daada61f08b0e59
0x6c0f425a6f7c63341e04b076e92ee1ce6ff98299
0x7598c930f071102058510b5ebc45f3b72f7a5644
```

### Examples

1. Generate single mainnet P2WSH address:
```bash
./bin/p2wsh -pubkey 02a0434d9e47f3c86235477c7b1ae6ae5d3442d49b1943c2b752a68e2a47e247c7 -evm 0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6
```

2. Generate single testnet P2WSH address:
```bash
./bin/p2wsh -pubkey 02a0434d9e47f3c86235477c7b1ae6ae5d3442d49b1943c2b752a68e2a47e247c7 -evm 0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6 -network testnet3
```

3. Generate single regtest P2WSH address:
```bash
./bin/p2wsh -pubkey 02a0434d9e47f3c86235477c7b1ae6ae5d3442d49b1943c2b752a68e2a47e247c7 -evm 0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6 -network regtest
```

4. Generate batch P2WSH addresses from CSV file:
```bash
./bin/p2wsh -pubkey 02a0434d9e47f3c86235477c7b1ae6ae5d3442d49b1943c2b752a68e2a47e247c7 -csv addresses.csv -network regtest
```

### Output Examples

#### Single Address Output
```
P2WSH Address: bc1q4hhretqpnkq0srzuhvmgjj8cnwushylqrjurkv9476dnlx8aa02s8r565l
Network: mainnet
Public Key: 02a0434d9e47f3c86235477c7b1ae6ae5d3442d49b1943c2b752a68e2a47e247c7
EVM Address: 0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6
```

#### Batch Address Output
```
Generating P2WSH addresses for 300 EVM addresses...
Network: regtest
Public Key: 0383560def84048edefe637d0119a4428dd12a42765a118b2bf77984057633c50e

1. EVM: 0xab0c51effd6c6a71a900f8ee8daada61f08b0e59 -> P2WSH: bcrt1qrkym99zgam8z4qewnrzw6s8ch9slrgvlzs050z4pzs30gytkqwkqp3h8xx
2. EVM: 0x6c0f425a6f7c63341e04b076e92ee1ce6ff98299 -> P2WSH: bcrt1qqk9hmq2emnjesthl3a0k4sfpgu6c5an8afrhg75gtrq57mgxf8hs0dep3x
3. EVM: 0x7598c930f071102058510b5ebc45f3b72f7a5644 -> P2WSH: bcrt1qkjwlu5nyj2qy8l2046msuvk9dx79nup0dwpwccnlnqa7xljq9evsfextnc
...
```

## Notes

- Public key must be in valid hex format (compressed format recommended)
- EVM address must be in valid Ethereum address format
- Generated P2WSH address contains timelock script for secure multi-signature transactions
- For batch processing, the CSV file should contain one EVM address per line
- The first line of the CSV file can be a header and will be automatically skipped
- Either `-evm` or `-csv` flag must be provided (but not both)
