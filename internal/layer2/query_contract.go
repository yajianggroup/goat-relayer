package layer2

func (lis *Layer2Listener) IsDeposit(txid [32]byte, txout uint32) (bool, error) {
	return lis.contractBridge.IsDeposited(nil, txid, txout)
}
