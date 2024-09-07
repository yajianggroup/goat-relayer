// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package abis

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// BridgeContractMetaData contains all meta data concerning the BridgeContract contract.
var BridgeContractMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"AccessDenied\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"AddressInsufficientBalance\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"FailedInnerCall\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"Forbidden\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidAddress\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"MalformedTax\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnableInvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"RateLimitExceeded\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"TaxTooHigh\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"Canceled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"Canceling\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"txid\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"txout\",\"type\":\"uint32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"tax\",\"type\":\"uint256\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint16\",\"name\":\"rate\",\"type\":\"uint16\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"max\",\"type\":\"uint64\"}],\"name\":\"DepositTaxUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"txid\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"txout\",\"type\":\"uint32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Paid\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"maxTxPrice\",\"type\":\"uint256\"}],\"name\":\"RBF\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"name\":\"RateLimitUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"Refund\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"tax\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"maxTxPrice\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"receiver\",\"type\":\"string\"}],\"name\":\"Withdraw\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint16\",\"name\":\"rate\",\"type\":\"uint16\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"max\",\"type\":\"uint64\"}],\"name\":\"WithdrawalTaxUpdated\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_wid\",\"type\":\"uint256\"}],\"name\":\"cancel1\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_wid\",\"type\":\"uint256\"}],\"name\":\"cancel2\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_txid\",\"type\":\"bytes32\"},{\"internalType\":\"uint32\",\"name\":\"_txout\",\"type\":\"uint32\"},{\"internalType\":\"address\",\"name\":\"_target\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"deposit\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"tax\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_txid\",\"type\":\"bytes32\"},{\"internalType\":\"uint32\",\"name\":\"_txout\",\"type\":\"uint32\"}],\"name\":\"isDeposited\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_wid\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"_txid\",\"type\":\"bytes32\"},{\"internalType\":\"uint32\",\"name\":\"_txout\",\"type\":\"uint32\"},{\"internalType\":\"uint256\",\"name\":\"_received\",\"type\":\"uint256\"}],\"name\":\"paid\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"param\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"rateLimit\",\"type\":\"uint16\"},{\"internalType\":\"uint16\",\"name\":\"depositTaxBP\",\"type\":\"uint16\"},{\"internalType\":\"uint64\",\"name\":\"maxDepositTax\",\"type\":\"uint64\"},{\"internalType\":\"uint16\",\"name\":\"withdrawalTaxBP\",\"type\":\"uint16\"},{\"internalType\":\"uint64\",\"name\":\"maxWithdrawalTax\",\"type\":\"uint64\"},{\"internalType\":\"uint16\",\"name\":\"_res1\",\"type\":\"uint16\"},{\"internalType\":\"uint64\",\"name\":\"_res2\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"receipts\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"txid\",\"type\":\"bytes32\"},{\"internalType\":\"uint32\",\"name\":\"txout\",\"type\":\"uint32\"},{\"internalType\":\"uint256\",\"name\":\"received\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_wid\",\"type\":\"uint256\"}],\"name\":\"refund\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_wid\",\"type\":\"uint256\"},{\"internalType\":\"uint16\",\"name\":\"_maxTxPrice\",\"type\":\"uint16\"}],\"name\":\"replaceByFee\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"_bp\",\"type\":\"uint16\"},{\"internalType\":\"uint64\",\"name\":\"_max\",\"type\":\"uint64\"}],\"name\":\"setDepositTax\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"_sec\",\"type\":\"uint16\"}],\"name\":\"setRateLimit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"_bp\",\"type\":\"uint16\"},{\"internalType\":\"uint64\",\"name\":\"_max\",\"type\":\"uint64\"}],\"name\":\"setWithdrawalTax\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"id\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_receiver\",\"type\":\"string\"},{\"internalType\":\"uint16\",\"name\":\"_maxTxPrice\",\"type\":\"uint16\"}],\"name\":\"withdraw\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"withdrawals\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"tax\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxTxPrice\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"updatedAt\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"receiver\",\"type\":\"string\"},{\"internalType\":\"enumIBridge.WithdrawalStatus\",\"name\":\"status\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// BridgeContractABI is the input ABI used to generate the binding from.
// Deprecated: Use BridgeContractMetaData.ABI instead.
var BridgeContractABI = BridgeContractMetaData.ABI

// BridgeContract is an auto generated Go binding around an Ethereum contract.
type BridgeContract struct {
	BridgeContractCaller     // Read-only binding to the contract
	BridgeContractTransactor // Write-only binding to the contract
	BridgeContractFilterer   // Log filterer for contract events
}

// BridgeContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type BridgeContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BridgeContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BridgeContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BridgeContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BridgeContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BridgeContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BridgeContractSession struct {
	Contract     *BridgeContract   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BridgeContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BridgeContractCallerSession struct {
	Contract *BridgeContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// BridgeContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BridgeContractTransactorSession struct {
	Contract     *BridgeContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// BridgeContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type BridgeContractRaw struct {
	Contract *BridgeContract // Generic contract binding to access the raw methods on
}

// BridgeContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BridgeContractCallerRaw struct {
	Contract *BridgeContractCaller // Generic read-only contract binding to access the raw methods on
}

// BridgeContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BridgeContractTransactorRaw struct {
	Contract *BridgeContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBridgeContract creates a new instance of BridgeContract, bound to a specific deployed contract.
func NewBridgeContract(address common.Address, backend bind.ContractBackend) (*BridgeContract, error) {
	contract, err := bindBridgeContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &BridgeContract{BridgeContractCaller: BridgeContractCaller{contract: contract}, BridgeContractTransactor: BridgeContractTransactor{contract: contract}, BridgeContractFilterer: BridgeContractFilterer{contract: contract}}, nil
}

// NewBridgeContractCaller creates a new read-only instance of BridgeContract, bound to a specific deployed contract.
func NewBridgeContractCaller(address common.Address, caller bind.ContractCaller) (*BridgeContractCaller, error) {
	contract, err := bindBridgeContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BridgeContractCaller{contract: contract}, nil
}

// NewBridgeContractTransactor creates a new write-only instance of BridgeContract, bound to a specific deployed contract.
func NewBridgeContractTransactor(address common.Address, transactor bind.ContractTransactor) (*BridgeContractTransactor, error) {
	contract, err := bindBridgeContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BridgeContractTransactor{contract: contract}, nil
}

// NewBridgeContractFilterer creates a new log filterer instance of BridgeContract, bound to a specific deployed contract.
func NewBridgeContractFilterer(address common.Address, filterer bind.ContractFilterer) (*BridgeContractFilterer, error) {
	contract, err := bindBridgeContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BridgeContractFilterer{contract: contract}, nil
}

// bindBridgeContract binds a generic wrapper to an already deployed contract.
func bindBridgeContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := BridgeContractMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BridgeContract *BridgeContractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BridgeContract.Contract.BridgeContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BridgeContract *BridgeContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BridgeContract.Contract.BridgeContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BridgeContract *BridgeContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BridgeContract.Contract.BridgeContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BridgeContract *BridgeContractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BridgeContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BridgeContract *BridgeContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BridgeContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BridgeContract *BridgeContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BridgeContract.Contract.contract.Transact(opts, method, params...)
}

// IsDeposited is a free data retrieval call binding the contract method 0x1ccc92c7.
//
// Solidity: function isDeposited(bytes32 _txid, uint32 _txout) view returns(bool)
func (_BridgeContract *BridgeContractCaller) IsDeposited(opts *bind.CallOpts, _txid [32]byte, _txout uint32) (bool, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "isDeposited", _txid, _txout)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsDeposited is a free data retrieval call binding the contract method 0x1ccc92c7.
//
// Solidity: function isDeposited(bytes32 _txid, uint32 _txout) view returns(bool)
func (_BridgeContract *BridgeContractSession) IsDeposited(_txid [32]byte, _txout uint32) (bool, error) {
	return _BridgeContract.Contract.IsDeposited(&_BridgeContract.CallOpts, _txid, _txout)
}

// IsDeposited is a free data retrieval call binding the contract method 0x1ccc92c7.
//
// Solidity: function isDeposited(bytes32 _txid, uint32 _txout) view returns(bool)
func (_BridgeContract *BridgeContractCallerSession) IsDeposited(_txid [32]byte, _txout uint32) (bool, error) {
	return _BridgeContract.Contract.IsDeposited(&_BridgeContract.CallOpts, _txid, _txout)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BridgeContract *BridgeContractCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BridgeContract *BridgeContractSession) Owner() (common.Address, error) {
	return _BridgeContract.Contract.Owner(&_BridgeContract.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BridgeContract *BridgeContractCallerSession) Owner() (common.Address, error) {
	return _BridgeContract.Contract.Owner(&_BridgeContract.CallOpts)
}

// Param is a free data retrieval call binding the contract method 0x883d87b1.
//
// Solidity: function param() view returns(uint16 rateLimit, uint16 depositTaxBP, uint64 maxDepositTax, uint16 withdrawalTaxBP, uint64 maxWithdrawalTax, uint16 _res1, uint64 _res2)
func (_BridgeContract *BridgeContractCaller) Param(opts *bind.CallOpts) (struct {
	RateLimit        uint16
	DepositTaxBP     uint16
	MaxDepositTax    uint64
	WithdrawalTaxBP  uint16
	MaxWithdrawalTax uint64
	Res1             uint16
	Res2             uint64
}, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "param")

	outstruct := new(struct {
		RateLimit        uint16
		DepositTaxBP     uint16
		MaxDepositTax    uint64
		WithdrawalTaxBP  uint16
		MaxWithdrawalTax uint64
		Res1             uint16
		Res2             uint64
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.RateLimit = *abi.ConvertType(out[0], new(uint16)).(*uint16)
	outstruct.DepositTaxBP = *abi.ConvertType(out[1], new(uint16)).(*uint16)
	outstruct.MaxDepositTax = *abi.ConvertType(out[2], new(uint64)).(*uint64)
	outstruct.WithdrawalTaxBP = *abi.ConvertType(out[3], new(uint16)).(*uint16)
	outstruct.MaxWithdrawalTax = *abi.ConvertType(out[4], new(uint64)).(*uint64)
	outstruct.Res1 = *abi.ConvertType(out[5], new(uint16)).(*uint16)
	outstruct.Res2 = *abi.ConvertType(out[6], new(uint64)).(*uint64)

	return *outstruct, err

}

// Param is a free data retrieval call binding the contract method 0x883d87b1.
//
// Solidity: function param() view returns(uint16 rateLimit, uint16 depositTaxBP, uint64 maxDepositTax, uint16 withdrawalTaxBP, uint64 maxWithdrawalTax, uint16 _res1, uint64 _res2)
func (_BridgeContract *BridgeContractSession) Param() (struct {
	RateLimit        uint16
	DepositTaxBP     uint16
	MaxDepositTax    uint64
	WithdrawalTaxBP  uint16
	MaxWithdrawalTax uint64
	Res1             uint16
	Res2             uint64
}, error) {
	return _BridgeContract.Contract.Param(&_BridgeContract.CallOpts)
}

// Param is a free data retrieval call binding the contract method 0x883d87b1.
//
// Solidity: function param() view returns(uint16 rateLimit, uint16 depositTaxBP, uint64 maxDepositTax, uint16 withdrawalTaxBP, uint64 maxWithdrawalTax, uint16 _res1, uint64 _res2)
func (_BridgeContract *BridgeContractCallerSession) Param() (struct {
	RateLimit        uint16
	DepositTaxBP     uint16
	MaxDepositTax    uint64
	WithdrawalTaxBP  uint16
	MaxWithdrawalTax uint64
	Res1             uint16
	Res2             uint64
}, error) {
	return _BridgeContract.Contract.Param(&_BridgeContract.CallOpts)
}

// Receipts is a free data retrieval call binding the contract method 0x0f7ee1ec.
//
// Solidity: function receipts(uint256 id) view returns(bytes32 txid, uint32 txout, uint256 received)
func (_BridgeContract *BridgeContractCaller) Receipts(opts *bind.CallOpts, id *big.Int) (struct {
	Txid     [32]byte
	Txout    uint32
	Received *big.Int
}, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "receipts", id)

	outstruct := new(struct {
		Txid     [32]byte
		Txout    uint32
		Received *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Txid = *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)
	outstruct.Txout = *abi.ConvertType(out[1], new(uint32)).(*uint32)
	outstruct.Received = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// Receipts is a free data retrieval call binding the contract method 0x0f7ee1ec.
//
// Solidity: function receipts(uint256 id) view returns(bytes32 txid, uint32 txout, uint256 received)
func (_BridgeContract *BridgeContractSession) Receipts(id *big.Int) (struct {
	Txid     [32]byte
	Txout    uint32
	Received *big.Int
}, error) {
	return _BridgeContract.Contract.Receipts(&_BridgeContract.CallOpts, id)
}

// Receipts is a free data retrieval call binding the contract method 0x0f7ee1ec.
//
// Solidity: function receipts(uint256 id) view returns(bytes32 txid, uint32 txout, uint256 received)
func (_BridgeContract *BridgeContractCallerSession) Receipts(id *big.Int) (struct {
	Txid     [32]byte
	Txout    uint32
	Received *big.Int
}, error) {
	return _BridgeContract.Contract.Receipts(&_BridgeContract.CallOpts, id)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 id) view returns(bool)
func (_BridgeContract *BridgeContractCaller) SupportsInterface(opts *bind.CallOpts, id [4]byte) (bool, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "supportsInterface", id)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 id) view returns(bool)
func (_BridgeContract *BridgeContractSession) SupportsInterface(id [4]byte) (bool, error) {
	return _BridgeContract.Contract.SupportsInterface(&_BridgeContract.CallOpts, id)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 id) view returns(bool)
func (_BridgeContract *BridgeContractCallerSession) SupportsInterface(id [4]byte) (bool, error) {
	return _BridgeContract.Contract.SupportsInterface(&_BridgeContract.CallOpts, id)
}

// Withdrawals is a free data retrieval call binding the contract method 0x5cc07076.
//
// Solidity: function withdrawals(uint256 ) view returns(address sender, uint256 amount, uint256 tax, uint256 maxTxPrice, uint256 updatedAt, string receiver, uint8 status)
func (_BridgeContract *BridgeContractCaller) Withdrawals(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Sender     common.Address
	Amount     *big.Int
	Tax        *big.Int
	MaxTxPrice *big.Int
	UpdatedAt  *big.Int
	Receiver   string
	Status     uint8
}, error) {
	var out []interface{}
	err := _BridgeContract.contract.Call(opts, &out, "withdrawals", arg0)

	outstruct := new(struct {
		Sender     common.Address
		Amount     *big.Int
		Tax        *big.Int
		MaxTxPrice *big.Int
		UpdatedAt  *big.Int
		Receiver   string
		Status     uint8
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Sender = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.Amount = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.Tax = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.MaxTxPrice = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.UpdatedAt = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.Receiver = *abi.ConvertType(out[5], new(string)).(*string)
	outstruct.Status = *abi.ConvertType(out[6], new(uint8)).(*uint8)

	return *outstruct, err

}

// Withdrawals is a free data retrieval call binding the contract method 0x5cc07076.
//
// Solidity: function withdrawals(uint256 ) view returns(address sender, uint256 amount, uint256 tax, uint256 maxTxPrice, uint256 updatedAt, string receiver, uint8 status)
func (_BridgeContract *BridgeContractSession) Withdrawals(arg0 *big.Int) (struct {
	Sender     common.Address
	Amount     *big.Int
	Tax        *big.Int
	MaxTxPrice *big.Int
	UpdatedAt  *big.Int
	Receiver   string
	Status     uint8
}, error) {
	return _BridgeContract.Contract.Withdrawals(&_BridgeContract.CallOpts, arg0)
}

// Withdrawals is a free data retrieval call binding the contract method 0x5cc07076.
//
// Solidity: function withdrawals(uint256 ) view returns(address sender, uint256 amount, uint256 tax, uint256 maxTxPrice, uint256 updatedAt, string receiver, uint8 status)
func (_BridgeContract *BridgeContractCallerSession) Withdrawals(arg0 *big.Int) (struct {
	Sender     common.Address
	Amount     *big.Int
	Tax        *big.Int
	MaxTxPrice *big.Int
	UpdatedAt  *big.Int
	Receiver   string
	Status     uint8
}, error) {
	return _BridgeContract.Contract.Withdrawals(&_BridgeContract.CallOpts, arg0)
}

// Cancel1 is a paid mutator transaction binding the contract method 0x84a64c12.
//
// Solidity: function cancel1(uint256 _wid) returns()
func (_BridgeContract *BridgeContractTransactor) Cancel1(opts *bind.TransactOpts, _wid *big.Int) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "cancel1", _wid)
}

// Cancel1 is a paid mutator transaction binding the contract method 0x84a64c12.
//
// Solidity: function cancel1(uint256 _wid) returns()
func (_BridgeContract *BridgeContractSession) Cancel1(_wid *big.Int) (*types.Transaction, error) {
	return _BridgeContract.Contract.Cancel1(&_BridgeContract.TransactOpts, _wid)
}

// Cancel1 is a paid mutator transaction binding the contract method 0x84a64c12.
//
// Solidity: function cancel1(uint256 _wid) returns()
func (_BridgeContract *BridgeContractTransactorSession) Cancel1(_wid *big.Int) (*types.Transaction, error) {
	return _BridgeContract.Contract.Cancel1(&_BridgeContract.TransactOpts, _wid)
}

// Cancel2 is a paid mutator transaction binding the contract method 0xc19dd320.
//
// Solidity: function cancel2(uint256 _wid) returns()
func (_BridgeContract *BridgeContractTransactor) Cancel2(opts *bind.TransactOpts, _wid *big.Int) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "cancel2", _wid)
}

// Cancel2 is a paid mutator transaction binding the contract method 0xc19dd320.
//
// Solidity: function cancel2(uint256 _wid) returns()
func (_BridgeContract *BridgeContractSession) Cancel2(_wid *big.Int) (*types.Transaction, error) {
	return _BridgeContract.Contract.Cancel2(&_BridgeContract.TransactOpts, _wid)
}

// Cancel2 is a paid mutator transaction binding the contract method 0xc19dd320.
//
// Solidity: function cancel2(uint256 _wid) returns()
func (_BridgeContract *BridgeContractTransactorSession) Cancel2(_wid *big.Int) (*types.Transaction, error) {
	return _BridgeContract.Contract.Cancel2(&_BridgeContract.TransactOpts, _wid)
}

// Deposit is a paid mutator transaction binding the contract method 0xb55ada39.
//
// Solidity: function deposit(bytes32 _txid, uint32 _txout, address _target, uint256 _amount) returns(uint256 tax)
func (_BridgeContract *BridgeContractTransactor) Deposit(opts *bind.TransactOpts, _txid [32]byte, _txout uint32, _target common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "deposit", _txid, _txout, _target, _amount)
}

// Deposit is a paid mutator transaction binding the contract method 0xb55ada39.
//
// Solidity: function deposit(bytes32 _txid, uint32 _txout, address _target, uint256 _amount) returns(uint256 tax)
func (_BridgeContract *BridgeContractSession) Deposit(_txid [32]byte, _txout uint32, _target common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _BridgeContract.Contract.Deposit(&_BridgeContract.TransactOpts, _txid, _txout, _target, _amount)
}

// Deposit is a paid mutator transaction binding the contract method 0xb55ada39.
//
// Solidity: function deposit(bytes32 _txid, uint32 _txout, address _target, uint256 _amount) returns(uint256 tax)
func (_BridgeContract *BridgeContractTransactorSession) Deposit(_txid [32]byte, _txout uint32, _target common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _BridgeContract.Contract.Deposit(&_BridgeContract.TransactOpts, _txid, _txout, _target, _amount)
}

// Paid is a paid mutator transaction binding the contract method 0xb670ab5e.
//
// Solidity: function paid(uint256 _wid, bytes32 _txid, uint32 _txout, uint256 _received) returns()
func (_BridgeContract *BridgeContractTransactor) Paid(opts *bind.TransactOpts, _wid *big.Int, _txid [32]byte, _txout uint32, _received *big.Int) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "paid", _wid, _txid, _txout, _received)
}

// Paid is a paid mutator transaction binding the contract method 0xb670ab5e.
//
// Solidity: function paid(uint256 _wid, bytes32 _txid, uint32 _txout, uint256 _received) returns()
func (_BridgeContract *BridgeContractSession) Paid(_wid *big.Int, _txid [32]byte, _txout uint32, _received *big.Int) (*types.Transaction, error) {
	return _BridgeContract.Contract.Paid(&_BridgeContract.TransactOpts, _wid, _txid, _txout, _received)
}

// Paid is a paid mutator transaction binding the contract method 0xb670ab5e.
//
// Solidity: function paid(uint256 _wid, bytes32 _txid, uint32 _txout, uint256 _received) returns()
func (_BridgeContract *BridgeContractTransactorSession) Paid(_wid *big.Int, _txid [32]byte, _txout uint32, _received *big.Int) (*types.Transaction, error) {
	return _BridgeContract.Contract.Paid(&_BridgeContract.TransactOpts, _wid, _txid, _txout, _received)
}

// Refund is a paid mutator transaction binding the contract method 0x278ecde1.
//
// Solidity: function refund(uint256 _wid) returns()
func (_BridgeContract *BridgeContractTransactor) Refund(opts *bind.TransactOpts, _wid *big.Int) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "refund", _wid)
}

// Refund is a paid mutator transaction binding the contract method 0x278ecde1.
//
// Solidity: function refund(uint256 _wid) returns()
func (_BridgeContract *BridgeContractSession) Refund(_wid *big.Int) (*types.Transaction, error) {
	return _BridgeContract.Contract.Refund(&_BridgeContract.TransactOpts, _wid)
}

// Refund is a paid mutator transaction binding the contract method 0x278ecde1.
//
// Solidity: function refund(uint256 _wid) returns()
func (_BridgeContract *BridgeContractTransactorSession) Refund(_wid *big.Int) (*types.Transaction, error) {
	return _BridgeContract.Contract.Refund(&_BridgeContract.TransactOpts, _wid)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BridgeContract *BridgeContractTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BridgeContract *BridgeContractSession) RenounceOwnership() (*types.Transaction, error) {
	return _BridgeContract.Contract.RenounceOwnership(&_BridgeContract.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BridgeContract *BridgeContractTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _BridgeContract.Contract.RenounceOwnership(&_BridgeContract.TransactOpts)
}

// ReplaceByFee is a paid mutator transaction binding the contract method 0xb3dd64dd.
//
// Solidity: function replaceByFee(uint256 _wid, uint16 _maxTxPrice) returns()
func (_BridgeContract *BridgeContractTransactor) ReplaceByFee(opts *bind.TransactOpts, _wid *big.Int, _maxTxPrice uint16) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "replaceByFee", _wid, _maxTxPrice)
}

// ReplaceByFee is a paid mutator transaction binding the contract method 0xb3dd64dd.
//
// Solidity: function replaceByFee(uint256 _wid, uint16 _maxTxPrice) returns()
func (_BridgeContract *BridgeContractSession) ReplaceByFee(_wid *big.Int, _maxTxPrice uint16) (*types.Transaction, error) {
	return _BridgeContract.Contract.ReplaceByFee(&_BridgeContract.TransactOpts, _wid, _maxTxPrice)
}

// ReplaceByFee is a paid mutator transaction binding the contract method 0xb3dd64dd.
//
// Solidity: function replaceByFee(uint256 _wid, uint16 _maxTxPrice) returns()
func (_BridgeContract *BridgeContractTransactorSession) ReplaceByFee(_wid *big.Int, _maxTxPrice uint16) (*types.Transaction, error) {
	return _BridgeContract.Contract.ReplaceByFee(&_BridgeContract.TransactOpts, _wid, _maxTxPrice)
}

// SetDepositTax is a paid mutator transaction binding the contract method 0xb3f33eda.
//
// Solidity: function setDepositTax(uint16 _bp, uint64 _max) returns()
func (_BridgeContract *BridgeContractTransactor) SetDepositTax(opts *bind.TransactOpts, _bp uint16, _max uint64) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "setDepositTax", _bp, _max)
}

// SetDepositTax is a paid mutator transaction binding the contract method 0xb3f33eda.
//
// Solidity: function setDepositTax(uint16 _bp, uint64 _max) returns()
func (_BridgeContract *BridgeContractSession) SetDepositTax(_bp uint16, _max uint64) (*types.Transaction, error) {
	return _BridgeContract.Contract.SetDepositTax(&_BridgeContract.TransactOpts, _bp, _max)
}

// SetDepositTax is a paid mutator transaction binding the contract method 0xb3f33eda.
//
// Solidity: function setDepositTax(uint16 _bp, uint64 _max) returns()
func (_BridgeContract *BridgeContractTransactorSession) SetDepositTax(_bp uint16, _max uint64) (*types.Transaction, error) {
	return _BridgeContract.Contract.SetDepositTax(&_BridgeContract.TransactOpts, _bp, _max)
}

// SetRateLimit is a paid mutator transaction binding the contract method 0xd653d820.
//
// Solidity: function setRateLimit(uint16 _sec) returns()
func (_BridgeContract *BridgeContractTransactor) SetRateLimit(opts *bind.TransactOpts, _sec uint16) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "setRateLimit", _sec)
}

// SetRateLimit is a paid mutator transaction binding the contract method 0xd653d820.
//
// Solidity: function setRateLimit(uint16 _sec) returns()
func (_BridgeContract *BridgeContractSession) SetRateLimit(_sec uint16) (*types.Transaction, error) {
	return _BridgeContract.Contract.SetRateLimit(&_BridgeContract.TransactOpts, _sec)
}

// SetRateLimit is a paid mutator transaction binding the contract method 0xd653d820.
//
// Solidity: function setRateLimit(uint16 _sec) returns()
func (_BridgeContract *BridgeContractTransactorSession) SetRateLimit(_sec uint16) (*types.Transaction, error) {
	return _BridgeContract.Contract.SetRateLimit(&_BridgeContract.TransactOpts, _sec)
}

// SetWithdrawalTax is a paid mutator transaction binding the contract method 0x8aa4af89.
//
// Solidity: function setWithdrawalTax(uint16 _bp, uint64 _max) returns()
func (_BridgeContract *BridgeContractTransactor) SetWithdrawalTax(opts *bind.TransactOpts, _bp uint16, _max uint64) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "setWithdrawalTax", _bp, _max)
}

// SetWithdrawalTax is a paid mutator transaction binding the contract method 0x8aa4af89.
//
// Solidity: function setWithdrawalTax(uint16 _bp, uint64 _max) returns()
func (_BridgeContract *BridgeContractSession) SetWithdrawalTax(_bp uint16, _max uint64) (*types.Transaction, error) {
	return _BridgeContract.Contract.SetWithdrawalTax(&_BridgeContract.TransactOpts, _bp, _max)
}

// SetWithdrawalTax is a paid mutator transaction binding the contract method 0x8aa4af89.
//
// Solidity: function setWithdrawalTax(uint16 _bp, uint64 _max) returns()
func (_BridgeContract *BridgeContractTransactorSession) SetWithdrawalTax(_bp uint16, _max uint64) (*types.Transaction, error) {
	return _BridgeContract.Contract.SetWithdrawalTax(&_BridgeContract.TransactOpts, _bp, _max)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BridgeContract *BridgeContractTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BridgeContract *BridgeContractSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _BridgeContract.Contract.TransferOwnership(&_BridgeContract.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BridgeContract *BridgeContractTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _BridgeContract.Contract.TransferOwnership(&_BridgeContract.TransactOpts, newOwner)
}

// Withdraw is a paid mutator transaction binding the contract method 0xa81de869.
//
// Solidity: function withdraw(string _receiver, uint16 _maxTxPrice) payable returns()
func (_BridgeContract *BridgeContractTransactor) Withdraw(opts *bind.TransactOpts, _receiver string, _maxTxPrice uint16) (*types.Transaction, error) {
	return _BridgeContract.contract.Transact(opts, "withdraw", _receiver, _maxTxPrice)
}

// Withdraw is a paid mutator transaction binding the contract method 0xa81de869.
//
// Solidity: function withdraw(string _receiver, uint16 _maxTxPrice) payable returns()
func (_BridgeContract *BridgeContractSession) Withdraw(_receiver string, _maxTxPrice uint16) (*types.Transaction, error) {
	return _BridgeContract.Contract.Withdraw(&_BridgeContract.TransactOpts, _receiver, _maxTxPrice)
}

// Withdraw is a paid mutator transaction binding the contract method 0xa81de869.
//
// Solidity: function withdraw(string _receiver, uint16 _maxTxPrice) payable returns()
func (_BridgeContract *BridgeContractTransactorSession) Withdraw(_receiver string, _maxTxPrice uint16) (*types.Transaction, error) {
	return _BridgeContract.Contract.Withdraw(&_BridgeContract.TransactOpts, _receiver, _maxTxPrice)
}

// BridgeContractCanceledIterator is returned from FilterCanceled and is used to iterate over the raw logs and unpacked data for Canceled events raised by the BridgeContract contract.
type BridgeContractCanceledIterator struct {
	Event *BridgeContractCanceled // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BridgeContractCanceledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeContractCanceled)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BridgeContractCanceled)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BridgeContractCanceledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeContractCanceledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeContractCanceled represents a Canceled event raised by the BridgeContract contract.
type BridgeContractCanceled struct {
	Id  *big.Int
	Raw types.Log // Blockchain specific contextual infos
}

// FilterCanceled is a free log retrieval operation binding the contract event 0x829a8683c544ad289ce92d3ce06e9ebad69b18a6916e60ec766c2c217461d8e9.
//
// Solidity: event Canceled(uint256 indexed id)
func (_BridgeContract *BridgeContractFilterer) FilterCanceled(opts *bind.FilterOpts, id []*big.Int) (*BridgeContractCanceledIterator, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}

	logs, sub, err := _BridgeContract.contract.FilterLogs(opts, "Canceled", idRule)
	if err != nil {
		return nil, err
	}
	return &BridgeContractCanceledIterator{contract: _BridgeContract.contract, event: "Canceled", logs: logs, sub: sub}, nil
}

// WatchCanceled is a free log subscription operation binding the contract event 0x829a8683c544ad289ce92d3ce06e9ebad69b18a6916e60ec766c2c217461d8e9.
//
// Solidity: event Canceled(uint256 indexed id)
func (_BridgeContract *BridgeContractFilterer) WatchCanceled(opts *bind.WatchOpts, sink chan<- *BridgeContractCanceled, id []*big.Int) (event.Subscription, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}

	logs, sub, err := _BridgeContract.contract.WatchLogs(opts, "Canceled", idRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeContractCanceled)
				if err := _BridgeContract.contract.UnpackLog(event, "Canceled", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseCanceled is a log parse operation binding the contract event 0x829a8683c544ad289ce92d3ce06e9ebad69b18a6916e60ec766c2c217461d8e9.
//
// Solidity: event Canceled(uint256 indexed id)
func (_BridgeContract *BridgeContractFilterer) ParseCanceled(log types.Log) (*BridgeContractCanceled, error) {
	event := new(BridgeContractCanceled)
	if err := _BridgeContract.contract.UnpackLog(event, "Canceled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeContractCancelingIterator is returned from FilterCanceling and is used to iterate over the raw logs and unpacked data for Canceling events raised by the BridgeContract contract.
type BridgeContractCancelingIterator struct {
	Event *BridgeContractCanceling // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BridgeContractCancelingIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeContractCanceling)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BridgeContractCanceling)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BridgeContractCancelingIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeContractCancelingIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeContractCanceling represents a Canceling event raised by the BridgeContract contract.
type BridgeContractCanceling struct {
	Id  *big.Int
	Raw types.Log // Blockchain specific contextual infos
}

// FilterCanceling is a free log retrieval operation binding the contract event 0x0106f4416537efff55311ef5e2f9c2a48204fcf84731f2b9d5091d23fc52160c.
//
// Solidity: event Canceling(uint256 indexed id)
func (_BridgeContract *BridgeContractFilterer) FilterCanceling(opts *bind.FilterOpts, id []*big.Int) (*BridgeContractCancelingIterator, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}

	logs, sub, err := _BridgeContract.contract.FilterLogs(opts, "Canceling", idRule)
	if err != nil {
		return nil, err
	}
	return &BridgeContractCancelingIterator{contract: _BridgeContract.contract, event: "Canceling", logs: logs, sub: sub}, nil
}

// WatchCanceling is a free log subscription operation binding the contract event 0x0106f4416537efff55311ef5e2f9c2a48204fcf84731f2b9d5091d23fc52160c.
//
// Solidity: event Canceling(uint256 indexed id)
func (_BridgeContract *BridgeContractFilterer) WatchCanceling(opts *bind.WatchOpts, sink chan<- *BridgeContractCanceling, id []*big.Int) (event.Subscription, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}

	logs, sub, err := _BridgeContract.contract.WatchLogs(opts, "Canceling", idRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeContractCanceling)
				if err := _BridgeContract.contract.UnpackLog(event, "Canceling", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseCanceling is a log parse operation binding the contract event 0x0106f4416537efff55311ef5e2f9c2a48204fcf84731f2b9d5091d23fc52160c.
//
// Solidity: event Canceling(uint256 indexed id)
func (_BridgeContract *BridgeContractFilterer) ParseCanceling(log types.Log) (*BridgeContractCanceling, error) {
	event := new(BridgeContractCanceling)
	if err := _BridgeContract.contract.UnpackLog(event, "Canceling", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeContractDepositIterator is returned from FilterDeposit and is used to iterate over the raw logs and unpacked data for Deposit events raised by the BridgeContract contract.
type BridgeContractDepositIterator struct {
	Event *BridgeContractDeposit // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BridgeContractDepositIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeContractDeposit)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BridgeContractDeposit)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BridgeContractDepositIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeContractDepositIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeContractDeposit represents a Deposit event raised by the BridgeContract contract.
type BridgeContractDeposit struct {
	Target common.Address
	Amount *big.Int
	Txid   [32]byte
	Txout  uint32
	Tax    *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDeposit is a free log retrieval operation binding the contract event 0xbc0e2d4f64f63e9c6b07a1665a26f689b20e42e836968119499db41c2d315efa.
//
// Solidity: event Deposit(address indexed target, uint256 indexed amount, bytes32 txid, uint32 txout, uint256 tax)
func (_BridgeContract *BridgeContractFilterer) FilterDeposit(opts *bind.FilterOpts, target []common.Address, amount []*big.Int) (*BridgeContractDepositIterator, error) {

	var targetRule []interface{}
	for _, targetItem := range target {
		targetRule = append(targetRule, targetItem)
	}
	var amountRule []interface{}
	for _, amountItem := range amount {
		amountRule = append(amountRule, amountItem)
	}

	logs, sub, err := _BridgeContract.contract.FilterLogs(opts, "Deposit", targetRule, amountRule)
	if err != nil {
		return nil, err
	}
	return &BridgeContractDepositIterator{contract: _BridgeContract.contract, event: "Deposit", logs: logs, sub: sub}, nil
}

// WatchDeposit is a free log subscription operation binding the contract event 0xbc0e2d4f64f63e9c6b07a1665a26f689b20e42e836968119499db41c2d315efa.
//
// Solidity: event Deposit(address indexed target, uint256 indexed amount, bytes32 txid, uint32 txout, uint256 tax)
func (_BridgeContract *BridgeContractFilterer) WatchDeposit(opts *bind.WatchOpts, sink chan<- *BridgeContractDeposit, target []common.Address, amount []*big.Int) (event.Subscription, error) {

	var targetRule []interface{}
	for _, targetItem := range target {
		targetRule = append(targetRule, targetItem)
	}
	var amountRule []interface{}
	for _, amountItem := range amount {
		amountRule = append(amountRule, amountItem)
	}

	logs, sub, err := _BridgeContract.contract.WatchLogs(opts, "Deposit", targetRule, amountRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeContractDeposit)
				if err := _BridgeContract.contract.UnpackLog(event, "Deposit", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDeposit is a log parse operation binding the contract event 0xbc0e2d4f64f63e9c6b07a1665a26f689b20e42e836968119499db41c2d315efa.
//
// Solidity: event Deposit(address indexed target, uint256 indexed amount, bytes32 txid, uint32 txout, uint256 tax)
func (_BridgeContract *BridgeContractFilterer) ParseDeposit(log types.Log) (*BridgeContractDeposit, error) {
	event := new(BridgeContractDeposit)
	if err := _BridgeContract.contract.UnpackLog(event, "Deposit", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeContractDepositTaxUpdatedIterator is returned from FilterDepositTaxUpdated and is used to iterate over the raw logs and unpacked data for DepositTaxUpdated events raised by the BridgeContract contract.
type BridgeContractDepositTaxUpdatedIterator struct {
	Event *BridgeContractDepositTaxUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BridgeContractDepositTaxUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeContractDepositTaxUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BridgeContractDepositTaxUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BridgeContractDepositTaxUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeContractDepositTaxUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeContractDepositTaxUpdated represents a DepositTaxUpdated event raised by the BridgeContract contract.
type BridgeContractDepositTaxUpdated struct {
	Rate uint16
	Max  uint64
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterDepositTaxUpdated is a free log retrieval operation binding the contract event 0x1007ff7aec53e9626ce51f25d4e093f290f60da8019c8cf489f0ae2f21ebf76a.
//
// Solidity: event DepositTaxUpdated(uint16 rate, uint64 max)
func (_BridgeContract *BridgeContractFilterer) FilterDepositTaxUpdated(opts *bind.FilterOpts) (*BridgeContractDepositTaxUpdatedIterator, error) {

	logs, sub, err := _BridgeContract.contract.FilterLogs(opts, "DepositTaxUpdated")
	if err != nil {
		return nil, err
	}
	return &BridgeContractDepositTaxUpdatedIterator{contract: _BridgeContract.contract, event: "DepositTaxUpdated", logs: logs, sub: sub}, nil
}

// WatchDepositTaxUpdated is a free log subscription operation binding the contract event 0x1007ff7aec53e9626ce51f25d4e093f290f60da8019c8cf489f0ae2f21ebf76a.
//
// Solidity: event DepositTaxUpdated(uint16 rate, uint64 max)
func (_BridgeContract *BridgeContractFilterer) WatchDepositTaxUpdated(opts *bind.WatchOpts, sink chan<- *BridgeContractDepositTaxUpdated) (event.Subscription, error) {

	logs, sub, err := _BridgeContract.contract.WatchLogs(opts, "DepositTaxUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeContractDepositTaxUpdated)
				if err := _BridgeContract.contract.UnpackLog(event, "DepositTaxUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDepositTaxUpdated is a log parse operation binding the contract event 0x1007ff7aec53e9626ce51f25d4e093f290f60da8019c8cf489f0ae2f21ebf76a.
//
// Solidity: event DepositTaxUpdated(uint16 rate, uint64 max)
func (_BridgeContract *BridgeContractFilterer) ParseDepositTaxUpdated(log types.Log) (*BridgeContractDepositTaxUpdated, error) {
	event := new(BridgeContractDepositTaxUpdated)
	if err := _BridgeContract.contract.UnpackLog(event, "DepositTaxUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeContractOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the BridgeContract contract.
type BridgeContractOwnershipTransferredIterator struct {
	Event *BridgeContractOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BridgeContractOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeContractOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BridgeContractOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BridgeContractOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeContractOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeContractOwnershipTransferred represents a OwnershipTransferred event raised by the BridgeContract contract.
type BridgeContractOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BridgeContract *BridgeContractFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*BridgeContractOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BridgeContract.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BridgeContractOwnershipTransferredIterator{contract: _BridgeContract.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BridgeContract *BridgeContractFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *BridgeContractOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BridgeContract.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeContractOwnershipTransferred)
				if err := _BridgeContract.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BridgeContract *BridgeContractFilterer) ParseOwnershipTransferred(log types.Log) (*BridgeContractOwnershipTransferred, error) {
	event := new(BridgeContractOwnershipTransferred)
	if err := _BridgeContract.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeContractPaidIterator is returned from FilterPaid and is used to iterate over the raw logs and unpacked data for Paid events raised by the BridgeContract contract.
type BridgeContractPaidIterator struct {
	Event *BridgeContractPaid // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BridgeContractPaidIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeContractPaid)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BridgeContractPaid)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BridgeContractPaidIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeContractPaidIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeContractPaid represents a Paid event raised by the BridgeContract contract.
type BridgeContractPaid struct {
	Id    *big.Int
	Txid  [32]byte
	Txout uint32
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterPaid is a free log retrieval operation binding the contract event 0xb74f5dbf34aabe02f20ff775b898acf1a9f70e4fbd48ad50548acae86e1ccd78.
//
// Solidity: event Paid(uint256 indexed id, bytes32 txid, uint32 txout, uint256 value)
func (_BridgeContract *BridgeContractFilterer) FilterPaid(opts *bind.FilterOpts, id []*big.Int) (*BridgeContractPaidIterator, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}

	logs, sub, err := _BridgeContract.contract.FilterLogs(opts, "Paid", idRule)
	if err != nil {
		return nil, err
	}
	return &BridgeContractPaidIterator{contract: _BridgeContract.contract, event: "Paid", logs: logs, sub: sub}, nil
}

// WatchPaid is a free log subscription operation binding the contract event 0xb74f5dbf34aabe02f20ff775b898acf1a9f70e4fbd48ad50548acae86e1ccd78.
//
// Solidity: event Paid(uint256 indexed id, bytes32 txid, uint32 txout, uint256 value)
func (_BridgeContract *BridgeContractFilterer) WatchPaid(opts *bind.WatchOpts, sink chan<- *BridgeContractPaid, id []*big.Int) (event.Subscription, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}

	logs, sub, err := _BridgeContract.contract.WatchLogs(opts, "Paid", idRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeContractPaid)
				if err := _BridgeContract.contract.UnpackLog(event, "Paid", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParsePaid is a log parse operation binding the contract event 0xb74f5dbf34aabe02f20ff775b898acf1a9f70e4fbd48ad50548acae86e1ccd78.
//
// Solidity: event Paid(uint256 indexed id, bytes32 txid, uint32 txout, uint256 value)
func (_BridgeContract *BridgeContractFilterer) ParsePaid(log types.Log) (*BridgeContractPaid, error) {
	event := new(BridgeContractPaid)
	if err := _BridgeContract.contract.UnpackLog(event, "Paid", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeContractRBFIterator is returned from FilterRBF and is used to iterate over the raw logs and unpacked data for RBF events raised by the BridgeContract contract.
type BridgeContractRBFIterator struct {
	Event *BridgeContractRBF // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BridgeContractRBFIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeContractRBF)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BridgeContractRBF)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BridgeContractRBFIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeContractRBFIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeContractRBF represents a RBF event raised by the BridgeContract contract.
type BridgeContractRBF struct {
	Id         *big.Int
	MaxTxPrice *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterRBF is a free log retrieval operation binding the contract event 0xe6190c43c634720bcd5b72bb0b920b88f8d0a15ae7e5dcbb0582c1bb2f98df4b.
//
// Solidity: event RBF(uint256 indexed id, uint256 maxTxPrice)
func (_BridgeContract *BridgeContractFilterer) FilterRBF(opts *bind.FilterOpts, id []*big.Int) (*BridgeContractRBFIterator, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}

	logs, sub, err := _BridgeContract.contract.FilterLogs(opts, "RBF", idRule)
	if err != nil {
		return nil, err
	}
	return &BridgeContractRBFIterator{contract: _BridgeContract.contract, event: "RBF", logs: logs, sub: sub}, nil
}

// WatchRBF is a free log subscription operation binding the contract event 0xe6190c43c634720bcd5b72bb0b920b88f8d0a15ae7e5dcbb0582c1bb2f98df4b.
//
// Solidity: event RBF(uint256 indexed id, uint256 maxTxPrice)
func (_BridgeContract *BridgeContractFilterer) WatchRBF(opts *bind.WatchOpts, sink chan<- *BridgeContractRBF, id []*big.Int) (event.Subscription, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}

	logs, sub, err := _BridgeContract.contract.WatchLogs(opts, "RBF", idRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeContractRBF)
				if err := _BridgeContract.contract.UnpackLog(event, "RBF", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRBF is a log parse operation binding the contract event 0xe6190c43c634720bcd5b72bb0b920b88f8d0a15ae7e5dcbb0582c1bb2f98df4b.
//
// Solidity: event RBF(uint256 indexed id, uint256 maxTxPrice)
func (_BridgeContract *BridgeContractFilterer) ParseRBF(log types.Log) (*BridgeContractRBF, error) {
	event := new(BridgeContractRBF)
	if err := _BridgeContract.contract.UnpackLog(event, "RBF", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeContractRateLimitUpdatedIterator is returned from FilterRateLimitUpdated and is used to iterate over the raw logs and unpacked data for RateLimitUpdated events raised by the BridgeContract contract.
type BridgeContractRateLimitUpdatedIterator struct {
	Event *BridgeContractRateLimitUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BridgeContractRateLimitUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeContractRateLimitUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BridgeContractRateLimitUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BridgeContractRateLimitUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeContractRateLimitUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeContractRateLimitUpdated represents a RateLimitUpdated event raised by the BridgeContract contract.
type BridgeContractRateLimitUpdated struct {
	Arg0 uint16
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterRateLimitUpdated is a free log retrieval operation binding the contract event 0xe536f709e7276119ff965216f1bbd671ef9ea99059743501129a0c9bec5d37ed.
//
// Solidity: event RateLimitUpdated(uint16 arg0)
func (_BridgeContract *BridgeContractFilterer) FilterRateLimitUpdated(opts *bind.FilterOpts) (*BridgeContractRateLimitUpdatedIterator, error) {

	logs, sub, err := _BridgeContract.contract.FilterLogs(opts, "RateLimitUpdated")
	if err != nil {
		return nil, err
	}
	return &BridgeContractRateLimitUpdatedIterator{contract: _BridgeContract.contract, event: "RateLimitUpdated", logs: logs, sub: sub}, nil
}

// WatchRateLimitUpdated is a free log subscription operation binding the contract event 0xe536f709e7276119ff965216f1bbd671ef9ea99059743501129a0c9bec5d37ed.
//
// Solidity: event RateLimitUpdated(uint16 arg0)
func (_BridgeContract *BridgeContractFilterer) WatchRateLimitUpdated(opts *bind.WatchOpts, sink chan<- *BridgeContractRateLimitUpdated) (event.Subscription, error) {

	logs, sub, err := _BridgeContract.contract.WatchLogs(opts, "RateLimitUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeContractRateLimitUpdated)
				if err := _BridgeContract.contract.UnpackLog(event, "RateLimitUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRateLimitUpdated is a log parse operation binding the contract event 0xe536f709e7276119ff965216f1bbd671ef9ea99059743501129a0c9bec5d37ed.
//
// Solidity: event RateLimitUpdated(uint16 arg0)
func (_BridgeContract *BridgeContractFilterer) ParseRateLimitUpdated(log types.Log) (*BridgeContractRateLimitUpdated, error) {
	event := new(BridgeContractRateLimitUpdated)
	if err := _BridgeContract.contract.UnpackLog(event, "RateLimitUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeContractRefundIterator is returned from FilterRefund and is used to iterate over the raw logs and unpacked data for Refund events raised by the BridgeContract contract.
type BridgeContractRefundIterator struct {
	Event *BridgeContractRefund // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BridgeContractRefundIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeContractRefund)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BridgeContractRefund)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BridgeContractRefundIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeContractRefundIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeContractRefund represents a Refund event raised by the BridgeContract contract.
type BridgeContractRefund struct {
	Id  *big.Int
	Raw types.Log // Blockchain specific contextual infos
}

// FilterRefund is a free log retrieval operation binding the contract event 0x2e1897b0591d764356194f7a795238a87c1987c7a877568e50d829d547c92b97.
//
// Solidity: event Refund(uint256 indexed id)
func (_BridgeContract *BridgeContractFilterer) FilterRefund(opts *bind.FilterOpts, id []*big.Int) (*BridgeContractRefundIterator, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}

	logs, sub, err := _BridgeContract.contract.FilterLogs(opts, "Refund", idRule)
	if err != nil {
		return nil, err
	}
	return &BridgeContractRefundIterator{contract: _BridgeContract.contract, event: "Refund", logs: logs, sub: sub}, nil
}

// WatchRefund is a free log subscription operation binding the contract event 0x2e1897b0591d764356194f7a795238a87c1987c7a877568e50d829d547c92b97.
//
// Solidity: event Refund(uint256 indexed id)
func (_BridgeContract *BridgeContractFilterer) WatchRefund(opts *bind.WatchOpts, sink chan<- *BridgeContractRefund, id []*big.Int) (event.Subscription, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}

	logs, sub, err := _BridgeContract.contract.WatchLogs(opts, "Refund", idRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeContractRefund)
				if err := _BridgeContract.contract.UnpackLog(event, "Refund", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRefund is a log parse operation binding the contract event 0x2e1897b0591d764356194f7a795238a87c1987c7a877568e50d829d547c92b97.
//
// Solidity: event Refund(uint256 indexed id)
func (_BridgeContract *BridgeContractFilterer) ParseRefund(log types.Log) (*BridgeContractRefund, error) {
	event := new(BridgeContractRefund)
	if err := _BridgeContract.contract.UnpackLog(event, "Refund", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeContractWithdrawIterator is returned from FilterWithdraw and is used to iterate over the raw logs and unpacked data for Withdraw events raised by the BridgeContract contract.
type BridgeContractWithdrawIterator struct {
	Event *BridgeContractWithdraw // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BridgeContractWithdrawIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeContractWithdraw)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BridgeContractWithdraw)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BridgeContractWithdrawIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeContractWithdrawIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeContractWithdraw represents a Withdraw event raised by the BridgeContract contract.
type BridgeContractWithdraw struct {
	Id         *big.Int
	From       common.Address
	Amount     *big.Int
	Tax        *big.Int
	MaxTxPrice *big.Int
	Receiver   string
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterWithdraw is a free log retrieval operation binding the contract event 0xa7aa0bf5f2a5106eb711bca96c8ea6ea0d99ec2c496d4d2c5dc0af8217d43c7c.
//
// Solidity: event Withdraw(uint256 indexed id, address indexed from, uint256 amount, uint256 tax, uint256 maxTxPrice, string receiver)
func (_BridgeContract *BridgeContractFilterer) FilterWithdraw(opts *bind.FilterOpts, id []*big.Int, from []common.Address) (*BridgeContractWithdrawIterator, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}

	logs, sub, err := _BridgeContract.contract.FilterLogs(opts, "Withdraw", idRule, fromRule)
	if err != nil {
		return nil, err
	}
	return &BridgeContractWithdrawIterator{contract: _BridgeContract.contract, event: "Withdraw", logs: logs, sub: sub}, nil
}

// WatchWithdraw is a free log subscription operation binding the contract event 0xa7aa0bf5f2a5106eb711bca96c8ea6ea0d99ec2c496d4d2c5dc0af8217d43c7c.
//
// Solidity: event Withdraw(uint256 indexed id, address indexed from, uint256 amount, uint256 tax, uint256 maxTxPrice, string receiver)
func (_BridgeContract *BridgeContractFilterer) WatchWithdraw(opts *bind.WatchOpts, sink chan<- *BridgeContractWithdraw, id []*big.Int, from []common.Address) (event.Subscription, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}

	logs, sub, err := _BridgeContract.contract.WatchLogs(opts, "Withdraw", idRule, fromRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeContractWithdraw)
				if err := _BridgeContract.contract.UnpackLog(event, "Withdraw", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseWithdraw is a log parse operation binding the contract event 0xa7aa0bf5f2a5106eb711bca96c8ea6ea0d99ec2c496d4d2c5dc0af8217d43c7c.
//
// Solidity: event Withdraw(uint256 indexed id, address indexed from, uint256 amount, uint256 tax, uint256 maxTxPrice, string receiver)
func (_BridgeContract *BridgeContractFilterer) ParseWithdraw(log types.Log) (*BridgeContractWithdraw, error) {
	event := new(BridgeContractWithdraw)
	if err := _BridgeContract.contract.UnpackLog(event, "Withdraw", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BridgeContractWithdrawalTaxUpdatedIterator is returned from FilterWithdrawalTaxUpdated and is used to iterate over the raw logs and unpacked data for WithdrawalTaxUpdated events raised by the BridgeContract contract.
type BridgeContractWithdrawalTaxUpdatedIterator struct {
	Event *BridgeContractWithdrawalTaxUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BridgeContractWithdrawalTaxUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BridgeContractWithdrawalTaxUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BridgeContractWithdrawalTaxUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BridgeContractWithdrawalTaxUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BridgeContractWithdrawalTaxUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BridgeContractWithdrawalTaxUpdated represents a WithdrawalTaxUpdated event raised by the BridgeContract contract.
type BridgeContractWithdrawalTaxUpdated struct {
	Rate uint16
	Max  uint64
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterWithdrawalTaxUpdated is a free log retrieval operation binding the contract event 0x59b73ca79bcb3dcb02c4d2b81e1a2da4c9fd9857ed81cfb16c5431b502f8c71b.
//
// Solidity: event WithdrawalTaxUpdated(uint16 rate, uint64 max)
func (_BridgeContract *BridgeContractFilterer) FilterWithdrawalTaxUpdated(opts *bind.FilterOpts) (*BridgeContractWithdrawalTaxUpdatedIterator, error) {

	logs, sub, err := _BridgeContract.contract.FilterLogs(opts, "WithdrawalTaxUpdated")
	if err != nil {
		return nil, err
	}
	return &BridgeContractWithdrawalTaxUpdatedIterator{contract: _BridgeContract.contract, event: "WithdrawalTaxUpdated", logs: logs, sub: sub}, nil
}

// WatchWithdrawalTaxUpdated is a free log subscription operation binding the contract event 0x59b73ca79bcb3dcb02c4d2b81e1a2da4c9fd9857ed81cfb16c5431b502f8c71b.
//
// Solidity: event WithdrawalTaxUpdated(uint16 rate, uint64 max)
func (_BridgeContract *BridgeContractFilterer) WatchWithdrawalTaxUpdated(opts *bind.WatchOpts, sink chan<- *BridgeContractWithdrawalTaxUpdated) (event.Subscription, error) {

	logs, sub, err := _BridgeContract.contract.WatchLogs(opts, "WithdrawalTaxUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BridgeContractWithdrawalTaxUpdated)
				if err := _BridgeContract.contract.UnpackLog(event, "WithdrawalTaxUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseWithdrawalTaxUpdated is a log parse operation binding the contract event 0x59b73ca79bcb3dcb02c4d2b81e1a2da4c9fd9857ed81cfb16c5431b502f8c71b.
//
// Solidity: event WithdrawalTaxUpdated(uint16 rate, uint64 max)
func (_BridgeContract *BridgeContractFilterer) ParseWithdrawalTaxUpdated(log types.Log) (*BridgeContractWithdrawalTaxUpdated, error) {
	event := new(BridgeContractWithdrawalTaxUpdated)
	if err := _BridgeContract.contract.UnpackLog(event, "WithdrawalTaxUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
