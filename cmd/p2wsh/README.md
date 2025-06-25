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

```bash
./bin/p2wsh -pubkey <public_key_hex> -evm <evm_address> [-network <network_type>]
```

### Parameters

- `-pubkey`: Public key in hex format (required)
- `-evm`: EVM address (required)
- `-network`: Network type, options: mainnet, testnet3, regtest (default: mainnet)
- `-help`: Show help message

### Examples

1. Generate mainnet P2WSH address:
```bash
./bin/p2wsh -pubkey 02a0434d9e47f3c86235477c7b1ae6ae5d3442d49b1943c2b752a68e2a47e247c7 -evm 0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6
```

2. Generate testnet P2WSH address:
```bash
./bin/p2wsh -pubkey 02a0434d9e47f3c86235477c7b1ae6ae5d3442d49b1943c2b752a68e2a47e247c7 -evm 0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6 -network testnet3
```

3. Generate regtest P2WSH address:
```bash
./bin/p2wsh -pubkey 02a0434d9e47f3c86235477c7b1ae6ae5d3442d49b1943c2b752a68e2a47e247c7 -evm 0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6 -network regtest
```

### Output Example

```
P2WSH Address: bc1q4hhretqpnkq0srzuhvmgjj8cnwushylqrjurkv9476dnlx8aa02s8r565l
Network: mainnet
Public Key: 02a0434d9e47f3c86235477c7b1ae6ae5d3442d49b1943c2b752a68e2a47e247c7
EVM Address: 0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6
```

## Notes

- Public key must be in valid hex format
- EVM address must be in valid Ethereum address format
- Generated P2WSH address contains timelock script for secure multi-signature transactions 
