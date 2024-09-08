# State Management

## Relayer State

```
State
|
|- L2 Info (from cosmos & evm)
|  |- Single Record: Height, Syncing, Threshold
|
|- Voter
|  |- list: pubkey, blskey
|  |  |- epoch, voters, proposer
|  |- queue
|     |- add request
|     |- remove request
|
|- BTC head
|  |- Confirmed (from cosmos)
|     |- state: start_height, latest_height
|     |- block list: <block> height, hash,
|  |- Unconfirm queue (BTC)
|     |- block: height, hash, status(init, pending, confirmed)
|  |- Off-chain consensus (Libp2p)
|     |- block: height, hash,
|
|- Wallet
|     Every utxo scaned from BTC scanner and detected status from L2 withdraw (if new voter startup)
|     BTC scanner: find Received-UTXO, Sent-UTXO
|     Offchain: maintain the withdraw queue, multi-sig for withdraw UTXOs
|  |- received-utxo: uid, txid, out_index, amount, receiver, sender, status
|     |- is-deposit
|     |- unknown
|  |- withdraw (from comsos || evm)
|     |- evm_tx: txid, block, amount, max_tx_fee, from, to, status
|  |- queue: txid (orderid when unconfirm), tx_fee_real, tx_fee_suggest, status
|     |- this queue is for off-chain consensus via Libp2p
|     |- txid refers to send-vin, send-vout
|  |- sent-vin: uid, txid, vin_txid, vin_vout, amount, sender, status
|     |- is-withdraw
|     |- unknown
|  |- sent-vout: uid, txid, out_index, withdraw_id, amount, receiver, sender, status
|     |- is-withdraw
|     |- unknown
|
```