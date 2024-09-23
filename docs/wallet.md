# Goat Relayer Wallet
This document mainly describes the wallet management of Relayer, including spendable utxo, withdrawal application discovery, withdrawal processing, p2wsh consolidation, p2pkh consolidation, etc.

## GORM schema
```
Utxo: source (deposit|other), receiver_type (P2PKH P2SH P2WSH P2WPKH P2TR), status
|- source -> deposit, it contains P2PKH, P2WSH receiver_types
Withdraw: evm_tx_id, from (evm address), to (btc address), amount, status
|
SendOrder: an order to select utxo to process, withdrawal or consolidation
｜- operate -> init, sign, send to btc (pending, rbf-request), finalize (feedback to goat)
Vin: SendOrder in items, from utxo P2PKH type and status of confirmed or processed
|
Vout: SendOrder out items, contain change to self address
```

## State update
The status update mainly refers to the two-way status update of utxo and withdraw|deposit, in order to be compatible with the disorder of layer2 and btc listener of the newly added voter node

### * Btc Poller 
Step 1,
> check the utxo vouts from send order vin in Vin table

Step 2,
> save to Uxto with status of confirmed or spent

Step 3,
> check the utxo vins from send order vin in Vin table

Step 4
> save to Vin, Vout (sender is MPC) with status processed if not exists, and order id set to 'restore';  update Vin status to processed if exists, update SendOrder status to confirmed if not processed


### * Layer2 Listener
Step 1,
> detect user withdraw tx, save to Withdraw table

Step 2,
> update Withdraw table by event

Step 3,
> restore order id?, Vin?, Vout?

## Data flow
User Deposit 1 
> Deposit -> Utxo (P2PKH)

User Deposit 0 
> Deposit -> Utxo (P2WSH)

Other Transfer *
> Transfer -> Utxo (P2PKH)

Consolidation * child collection of Other Transfer
> Utxo (P2WSH) -> Utxo (P2PKH)

Withdrawal
> Utxo (P2PKH) ->  Withdraw + Utxo (P2PKH)

## Consolidation


## Withdrawal

