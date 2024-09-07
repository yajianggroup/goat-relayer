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

// BitcoinContractMetaData contains all meta data concerning the BitcoinContract contract.
var BitcoinContractMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_height\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"_hash\",\"type\":\"bytes32\"},{\"internalType\":\"string\",\"name\":\"_network\",\"type\":\"string\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"AccessDenied\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"Forbidden\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"height\",\"type\":\"uint256\"}],\"name\":\"NewBlockHash\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"height\",\"type\":\"uint256\"}],\"name\":\"blockHash\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"latestHeight\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"networkName\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_hash\",\"type\":\"bytes32\"}],\"name\":\"newBlockHash\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"startHeight\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// BitcoinContractABI is the input ABI used to generate the binding from.
// Deprecated: Use BitcoinContractMetaData.ABI instead.
var BitcoinContractABI = BitcoinContractMetaData.ABI

// BitcoinContract is an auto generated Go binding around an Ethereum contract.
type BitcoinContract struct {
	BitcoinContractCaller     // Read-only binding to the contract
	BitcoinContractTransactor // Write-only binding to the contract
	BitcoinContractFilterer   // Log filterer for contract events
}

// BitcoinContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type BitcoinContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BitcoinContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BitcoinContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BitcoinContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BitcoinContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BitcoinContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BitcoinContractSession struct {
	Contract     *BitcoinContract  // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BitcoinContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BitcoinContractCallerSession struct {
	Contract *BitcoinContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts          // Call options to use throughout this session
}

// BitcoinContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BitcoinContractTransactorSession struct {
	Contract     *BitcoinContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// BitcoinContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type BitcoinContractRaw struct {
	Contract *BitcoinContract // Generic contract binding to access the raw methods on
}

// BitcoinContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BitcoinContractCallerRaw struct {
	Contract *BitcoinContractCaller // Generic read-only contract binding to access the raw methods on
}

// BitcoinContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BitcoinContractTransactorRaw struct {
	Contract *BitcoinContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBitcoinContract creates a new instance of BitcoinContract, bound to a specific deployed contract.
func NewBitcoinContract(address common.Address, backend bind.ContractBackend) (*BitcoinContract, error) {
	contract, err := bindBitcoinContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &BitcoinContract{BitcoinContractCaller: BitcoinContractCaller{contract: contract}, BitcoinContractTransactor: BitcoinContractTransactor{contract: contract}, BitcoinContractFilterer: BitcoinContractFilterer{contract: contract}}, nil
}

// NewBitcoinContractCaller creates a new read-only instance of BitcoinContract, bound to a specific deployed contract.
func NewBitcoinContractCaller(address common.Address, caller bind.ContractCaller) (*BitcoinContractCaller, error) {
	contract, err := bindBitcoinContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BitcoinContractCaller{contract: contract}, nil
}

// NewBitcoinContractTransactor creates a new write-only instance of BitcoinContract, bound to a specific deployed contract.
func NewBitcoinContractTransactor(address common.Address, transactor bind.ContractTransactor) (*BitcoinContractTransactor, error) {
	contract, err := bindBitcoinContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BitcoinContractTransactor{contract: contract}, nil
}

// NewBitcoinContractFilterer creates a new log filterer instance of BitcoinContract, bound to a specific deployed contract.
func NewBitcoinContractFilterer(address common.Address, filterer bind.ContractFilterer) (*BitcoinContractFilterer, error) {
	contract, err := bindBitcoinContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BitcoinContractFilterer{contract: contract}, nil
}

// bindBitcoinContract binds a generic wrapper to an already deployed contract.
func bindBitcoinContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := BitcoinContractMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BitcoinContract *BitcoinContractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BitcoinContract.Contract.BitcoinContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BitcoinContract *BitcoinContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BitcoinContract.Contract.BitcoinContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BitcoinContract *BitcoinContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BitcoinContract.Contract.BitcoinContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BitcoinContract *BitcoinContractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BitcoinContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BitcoinContract *BitcoinContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BitcoinContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BitcoinContract *BitcoinContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BitcoinContract.Contract.contract.Transact(opts, method, params...)
}

// BlockHash is a free data retrieval call binding the contract method 0x85df51fd.
//
// Solidity: function blockHash(uint256 height) view returns(bytes32)
func (_BitcoinContract *BitcoinContractCaller) BlockHash(opts *bind.CallOpts, height *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _BitcoinContract.contract.Call(opts, &out, "blockHash", height)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// BlockHash is a free data retrieval call binding the contract method 0x85df51fd.
//
// Solidity: function blockHash(uint256 height) view returns(bytes32)
func (_BitcoinContract *BitcoinContractSession) BlockHash(height *big.Int) ([32]byte, error) {
	return _BitcoinContract.Contract.BlockHash(&_BitcoinContract.CallOpts, height)
}

// BlockHash is a free data retrieval call binding the contract method 0x85df51fd.
//
// Solidity: function blockHash(uint256 height) view returns(bytes32)
func (_BitcoinContract *BitcoinContractCallerSession) BlockHash(height *big.Int) ([32]byte, error) {
	return _BitcoinContract.Contract.BlockHash(&_BitcoinContract.CallOpts, height)
}

// LatestHeight is a free data retrieval call binding the contract method 0xe405bbc3.
//
// Solidity: function latestHeight() view returns(uint256)
func (_BitcoinContract *BitcoinContractCaller) LatestHeight(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BitcoinContract.contract.Call(opts, &out, "latestHeight")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// LatestHeight is a free data retrieval call binding the contract method 0xe405bbc3.
//
// Solidity: function latestHeight() view returns(uint256)
func (_BitcoinContract *BitcoinContractSession) LatestHeight() (*big.Int, error) {
	return _BitcoinContract.Contract.LatestHeight(&_BitcoinContract.CallOpts)
}

// LatestHeight is a free data retrieval call binding the contract method 0xe405bbc3.
//
// Solidity: function latestHeight() view returns(uint256)
func (_BitcoinContract *BitcoinContractCallerSession) LatestHeight() (*big.Int, error) {
	return _BitcoinContract.Contract.LatestHeight(&_BitcoinContract.CallOpts)
}

// NetworkName is a free data retrieval call binding the contract method 0x107bf28c.
//
// Solidity: function networkName() view returns(string)
func (_BitcoinContract *BitcoinContractCaller) NetworkName(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _BitcoinContract.contract.Call(opts, &out, "networkName")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// NetworkName is a free data retrieval call binding the contract method 0x107bf28c.
//
// Solidity: function networkName() view returns(string)
func (_BitcoinContract *BitcoinContractSession) NetworkName() (string, error) {
	return _BitcoinContract.Contract.NetworkName(&_BitcoinContract.CallOpts)
}

// NetworkName is a free data retrieval call binding the contract method 0x107bf28c.
//
// Solidity: function networkName() view returns(string)
func (_BitcoinContract *BitcoinContractCallerSession) NetworkName() (string, error) {
	return _BitcoinContract.Contract.NetworkName(&_BitcoinContract.CallOpts)
}

// StartHeight is a free data retrieval call binding the contract method 0x26a6557a.
//
// Solidity: function startHeight() view returns(uint256)
func (_BitcoinContract *BitcoinContractCaller) StartHeight(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BitcoinContract.contract.Call(opts, &out, "startHeight")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// StartHeight is a free data retrieval call binding the contract method 0x26a6557a.
//
// Solidity: function startHeight() view returns(uint256)
func (_BitcoinContract *BitcoinContractSession) StartHeight() (*big.Int, error) {
	return _BitcoinContract.Contract.StartHeight(&_BitcoinContract.CallOpts)
}

// StartHeight is a free data retrieval call binding the contract method 0x26a6557a.
//
// Solidity: function startHeight() view returns(uint256)
func (_BitcoinContract *BitcoinContractCallerSession) StartHeight() (*big.Int, error) {
	return _BitcoinContract.Contract.StartHeight(&_BitcoinContract.CallOpts)
}

// NewBlockHash is a paid mutator transaction binding the contract method 0x94f490bd.
//
// Solidity: function newBlockHash(bytes32 _hash) returns()
func (_BitcoinContract *BitcoinContractTransactor) NewBlockHash(opts *bind.TransactOpts, _hash [32]byte) (*types.Transaction, error) {
	return _BitcoinContract.contract.Transact(opts, "newBlockHash", _hash)
}

// NewBlockHash is a paid mutator transaction binding the contract method 0x94f490bd.
//
// Solidity: function newBlockHash(bytes32 _hash) returns()
func (_BitcoinContract *BitcoinContractSession) NewBlockHash(_hash [32]byte) (*types.Transaction, error) {
	return _BitcoinContract.Contract.NewBlockHash(&_BitcoinContract.TransactOpts, _hash)
}

// NewBlockHash is a paid mutator transaction binding the contract method 0x94f490bd.
//
// Solidity: function newBlockHash(bytes32 _hash) returns()
func (_BitcoinContract *BitcoinContractTransactorSession) NewBlockHash(_hash [32]byte) (*types.Transaction, error) {
	return _BitcoinContract.Contract.NewBlockHash(&_BitcoinContract.TransactOpts, _hash)
}

// BitcoinContractNewBlockHashIterator is returned from FilterNewBlockHash and is used to iterate over the raw logs and unpacked data for NewBlockHash events raised by the BitcoinContract contract.
type BitcoinContractNewBlockHashIterator struct {
	Event *BitcoinContractNewBlockHash // Event containing the contract specifics and raw log

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
func (it *BitcoinContractNewBlockHashIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BitcoinContractNewBlockHash)
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
		it.Event = new(BitcoinContractNewBlockHash)
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
func (it *BitcoinContractNewBlockHashIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BitcoinContractNewBlockHashIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BitcoinContractNewBlockHash represents a NewBlockHash event raised by the BitcoinContract contract.
type BitcoinContractNewBlockHash struct {
	Height *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterNewBlockHash is a free log retrieval operation binding the contract event 0xdd5483f1119d050d70b0fe3ed9db0b5f41b3ec55838346cbb624efe0565b0133.
//
// Solidity: event NewBlockHash(uint256 height)
func (_BitcoinContract *BitcoinContractFilterer) FilterNewBlockHash(opts *bind.FilterOpts) (*BitcoinContractNewBlockHashIterator, error) {

	logs, sub, err := _BitcoinContract.contract.FilterLogs(opts, "NewBlockHash")
	if err != nil {
		return nil, err
	}
	return &BitcoinContractNewBlockHashIterator{contract: _BitcoinContract.contract, event: "NewBlockHash", logs: logs, sub: sub}, nil
}

// WatchNewBlockHash is a free log subscription operation binding the contract event 0xdd5483f1119d050d70b0fe3ed9db0b5f41b3ec55838346cbb624efe0565b0133.
//
// Solidity: event NewBlockHash(uint256 height)
func (_BitcoinContract *BitcoinContractFilterer) WatchNewBlockHash(opts *bind.WatchOpts, sink chan<- *BitcoinContractNewBlockHash) (event.Subscription, error) {

	logs, sub, err := _BitcoinContract.contract.WatchLogs(opts, "NewBlockHash")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BitcoinContractNewBlockHash)
				if err := _BitcoinContract.contract.UnpackLog(event, "NewBlockHash", log); err != nil {
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

// ParseNewBlockHash is a log parse operation binding the contract event 0xdd5483f1119d050d70b0fe3ed9db0b5f41b3ec55838346cbb624efe0565b0133.
//
// Solidity: event NewBlockHash(uint256 height)
func (_BitcoinContract *BitcoinContractFilterer) ParseNewBlockHash(log types.Log) (*BitcoinContractNewBlockHash, error) {
	event := new(BitcoinContractNewBlockHash)
	if err := _BitcoinContract.contract.UnpackLog(event, "NewBlockHash", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
