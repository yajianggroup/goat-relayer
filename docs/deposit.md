# Deposit Utxo Management

## Preverify deposit
Receive the new transactions by gRPC Server
Pre-verify the transaction (txHash + rawTx + evmAddress) without blockhash and blockheight since the server accept the unconfirmed transactions from frontend request or rpc request, which might not be confirmed by the bitcoin network. Received tx will be send to unconfirmed channel

## Wait for deposit confirmed & boardcast confirmed deposits
(a) Start a poller to query block height and block header by txhash every 10 min with 1 confirm
Query transactions at BtcCache(db.BtcCache) by txhash. Queried tx will be sent to confirmed channel,

(b) Start another poller to query the block status by block height every 10 min with 6 confirms(processed)
Query blocks at BtcLight(db.BtcBlock) by height of the block that the tx exist in. While the block has been consensused by GOAT network, which means the block has been confirmed 6 times by the bitcoin network, Update state to confirmed, then generate SPV proof, assembly the deposit and then broadcast to the relayer network.

## Verify SPV and send by proposer to mint GOAT
When another relayer receive the deposit:
(a) if the role is a regular voter, just store the data;
(b) if it's a proposer, verify deposits and then submit to GOAT network.
