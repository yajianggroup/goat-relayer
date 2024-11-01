package layer2

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

func (lis *Layer2Listener) IsDeposit(txid [32]byte, txout uint32) (bool, error) {
	return lis.contractBridge.IsDeposited(nil, txid, txout)
}

func (lis *Layer2Listener) GetWithdrawalSenderAddress(withdrawalID *big.Int) (common.Address, error) {
	withdrawal, err := lis.contractBridge.Withdrawals(&bind.CallOpts{}, withdrawalID)
	if err != nil {
		return common.Address{}, err
	}

	return withdrawal.Sender, nil
}
