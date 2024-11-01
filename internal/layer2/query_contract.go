package layer2

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

func (lis *Layer2Listener) IsDeposit(txid [32]byte, txout uint32) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	opts := &bind.CallOpts{
		Context: ctx,
	}

	return lis.contractBridge.IsDeposited(opts, txid, txout)
}

func (lis *Layer2Listener) GetWithdrawalSenderAddress(withdrawalID *big.Int) (common.Address, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	opts := &bind.CallOpts{
		Context: ctx,
	}

	withdrawal, err := lis.contractBridge.Withdrawals(opts, withdrawalID)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to get withdrawal info: %w", err)
	}

	if (withdrawal.Sender == common.Address{}) {
		return common.Address{}, fmt.Errorf("invalid sender address returned")
	}

	return withdrawal.Sender, nil
}
