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

// RelayerContractMetaData contains all meta data concerning the RelayerContract contract.
var RelayerContractMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnableInvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes20\",\"name\":\"voter\",\"type\":\"bytes20\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"keyHash\",\"type\":\"bytes32\"}],\"name\":\"AddedVoter\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes20\",\"name\":\"voter\",\"type\":\"bytes20\"}],\"name\":\"RemovedVoter\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"MAX_VOTER_COUNT\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes20\",\"name\":\"voter\",\"type\":\"bytes20\"},{\"internalType\":\"bytes32\",\"name\":\"vtkey\",\"type\":\"bytes32\"}],\"name\":\"addVoter\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"vtkh\",\"type\":\"bytes32\"}],\"name\":\"pubkeys\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"exists\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes20\",\"name\":\"voter\",\"type\":\"bytes20\"}],\"name\":\"removeVoter\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"total\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes20\",\"name\":\"voter\",\"type\":\"bytes20\"}],\"name\":\"voters\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"exists\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// RelayerContractABI is the input ABI used to generate the binding from.
// Deprecated: Use RelayerContractMetaData.ABI instead.
var RelayerContractABI = RelayerContractMetaData.ABI

// RelayerContract is an auto generated Go binding around an Ethereum contract.
type RelayerContract struct {
	RelayerContractCaller     // Read-only binding to the contract
	RelayerContractTransactor // Write-only binding to the contract
	RelayerContractFilterer   // Log filterer for contract events
}

// RelayerContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type RelayerContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RelayerContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type RelayerContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RelayerContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type RelayerContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RelayerContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type RelayerContractSession struct {
	Contract     *RelayerContract  // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// RelayerContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type RelayerContractCallerSession struct {
	Contract *RelayerContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts          // Call options to use throughout this session
}

// RelayerContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type RelayerContractTransactorSession struct {
	Contract     *RelayerContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// RelayerContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type RelayerContractRaw struct {
	Contract *RelayerContract // Generic contract binding to access the raw methods on
}

// RelayerContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type RelayerContractCallerRaw struct {
	Contract *RelayerContractCaller // Generic read-only contract binding to access the raw methods on
}

// RelayerContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type RelayerContractTransactorRaw struct {
	Contract *RelayerContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewRelayerContract creates a new instance of RelayerContract, bound to a specific deployed contract.
func NewRelayerContract(address common.Address, backend bind.ContractBackend) (*RelayerContract, error) {
	contract, err := bindRelayerContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &RelayerContract{RelayerContractCaller: RelayerContractCaller{contract: contract}, RelayerContractTransactor: RelayerContractTransactor{contract: contract}, RelayerContractFilterer: RelayerContractFilterer{contract: contract}}, nil
}

// NewRelayerContractCaller creates a new read-only instance of RelayerContract, bound to a specific deployed contract.
func NewRelayerContractCaller(address common.Address, caller bind.ContractCaller) (*RelayerContractCaller, error) {
	contract, err := bindRelayerContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &RelayerContractCaller{contract: contract}, nil
}

// NewRelayerContractTransactor creates a new write-only instance of RelayerContract, bound to a specific deployed contract.
func NewRelayerContractTransactor(address common.Address, transactor bind.ContractTransactor) (*RelayerContractTransactor, error) {
	contract, err := bindRelayerContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &RelayerContractTransactor{contract: contract}, nil
}

// NewRelayerContractFilterer creates a new log filterer instance of RelayerContract, bound to a specific deployed contract.
func NewRelayerContractFilterer(address common.Address, filterer bind.ContractFilterer) (*RelayerContractFilterer, error) {
	contract, err := bindRelayerContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &RelayerContractFilterer{contract: contract}, nil
}

// bindRelayerContract binds a generic wrapper to an already deployed contract.
func bindRelayerContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := RelayerContractMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_RelayerContract *RelayerContractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _RelayerContract.Contract.RelayerContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_RelayerContract *RelayerContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _RelayerContract.Contract.RelayerContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_RelayerContract *RelayerContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _RelayerContract.Contract.RelayerContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_RelayerContract *RelayerContractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _RelayerContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_RelayerContract *RelayerContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _RelayerContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_RelayerContract *RelayerContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _RelayerContract.Contract.contract.Transact(opts, method, params...)
}

// MAXVOTERCOUNT is a free data retrieval call binding the contract method 0x827fb811.
//
// Solidity: function MAX_VOTER_COUNT() view returns(uint16)
func (_RelayerContract *RelayerContractCaller) MAXVOTERCOUNT(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _RelayerContract.contract.Call(opts, &out, "MAX_VOTER_COUNT")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// MAXVOTERCOUNT is a free data retrieval call binding the contract method 0x827fb811.
//
// Solidity: function MAX_VOTER_COUNT() view returns(uint16)
func (_RelayerContract *RelayerContractSession) MAXVOTERCOUNT() (uint16, error) {
	return _RelayerContract.Contract.MAXVOTERCOUNT(&_RelayerContract.CallOpts)
}

// MAXVOTERCOUNT is a free data retrieval call binding the contract method 0x827fb811.
//
// Solidity: function MAX_VOTER_COUNT() view returns(uint16)
func (_RelayerContract *RelayerContractCallerSession) MAXVOTERCOUNT() (uint16, error) {
	return _RelayerContract.Contract.MAXVOTERCOUNT(&_RelayerContract.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_RelayerContract *RelayerContractCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _RelayerContract.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_RelayerContract *RelayerContractSession) Owner() (common.Address, error) {
	return _RelayerContract.Contract.Owner(&_RelayerContract.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_RelayerContract *RelayerContractCallerSession) Owner() (common.Address, error) {
	return _RelayerContract.Contract.Owner(&_RelayerContract.CallOpts)
}

// Pubkeys is a free data retrieval call binding the contract method 0x98611f12.
//
// Solidity: function pubkeys(bytes32 vtkh) view returns(bool exists)
func (_RelayerContract *RelayerContractCaller) Pubkeys(opts *bind.CallOpts, vtkh [32]byte) (bool, error) {
	var out []interface{}
	err := _RelayerContract.contract.Call(opts, &out, "pubkeys", vtkh)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Pubkeys is a free data retrieval call binding the contract method 0x98611f12.
//
// Solidity: function pubkeys(bytes32 vtkh) view returns(bool exists)
func (_RelayerContract *RelayerContractSession) Pubkeys(vtkh [32]byte) (bool, error) {
	return _RelayerContract.Contract.Pubkeys(&_RelayerContract.CallOpts, vtkh)
}

// Pubkeys is a free data retrieval call binding the contract method 0x98611f12.
//
// Solidity: function pubkeys(bytes32 vtkh) view returns(bool exists)
func (_RelayerContract *RelayerContractCallerSession) Pubkeys(vtkh [32]byte) (bool, error) {
	return _RelayerContract.Contract.Pubkeys(&_RelayerContract.CallOpts, vtkh)
}

// Total is a free data retrieval call binding the contract method 0x2ddbd13a.
//
// Solidity: function total() view returns(uint16)
func (_RelayerContract *RelayerContractCaller) Total(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _RelayerContract.contract.Call(opts, &out, "total")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// Total is a free data retrieval call binding the contract method 0x2ddbd13a.
//
// Solidity: function total() view returns(uint16)
func (_RelayerContract *RelayerContractSession) Total() (uint16, error) {
	return _RelayerContract.Contract.Total(&_RelayerContract.CallOpts)
}

// Total is a free data retrieval call binding the contract method 0x2ddbd13a.
//
// Solidity: function total() view returns(uint16)
func (_RelayerContract *RelayerContractCallerSession) Total() (uint16, error) {
	return _RelayerContract.Contract.Total(&_RelayerContract.CallOpts)
}

// Voters is a free data retrieval call binding the contract method 0x1a41e890.
//
// Solidity: function voters(bytes20 voter) view returns(bool exists)
func (_RelayerContract *RelayerContractCaller) Voters(opts *bind.CallOpts, voter [20]byte) (bool, error) {
	var out []interface{}
	err := _RelayerContract.contract.Call(opts, &out, "voters", voter)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Voters is a free data retrieval call binding the contract method 0x1a41e890.
//
// Solidity: function voters(bytes20 voter) view returns(bool exists)
func (_RelayerContract *RelayerContractSession) Voters(voter [20]byte) (bool, error) {
	return _RelayerContract.Contract.Voters(&_RelayerContract.CallOpts, voter)
}

// Voters is a free data retrieval call binding the contract method 0x1a41e890.
//
// Solidity: function voters(bytes20 voter) view returns(bool exists)
func (_RelayerContract *RelayerContractCallerSession) Voters(voter [20]byte) (bool, error) {
	return _RelayerContract.Contract.Voters(&_RelayerContract.CallOpts, voter)
}

// AddVoter is a paid mutator transaction binding the contract method 0xd782d4f7.
//
// Solidity: function addVoter(bytes20 voter, bytes32 vtkey) returns()
func (_RelayerContract *RelayerContractTransactor) AddVoter(opts *bind.TransactOpts, voter [20]byte, vtkey [32]byte) (*types.Transaction, error) {
	return _RelayerContract.contract.Transact(opts, "addVoter", voter, vtkey)
}

// AddVoter is a paid mutator transaction binding the contract method 0xd782d4f7.
//
// Solidity: function addVoter(bytes20 voter, bytes32 vtkey) returns()
func (_RelayerContract *RelayerContractSession) AddVoter(voter [20]byte, vtkey [32]byte) (*types.Transaction, error) {
	return _RelayerContract.Contract.AddVoter(&_RelayerContract.TransactOpts, voter, vtkey)
}

// AddVoter is a paid mutator transaction binding the contract method 0xd782d4f7.
//
// Solidity: function addVoter(bytes20 voter, bytes32 vtkey) returns()
func (_RelayerContract *RelayerContractTransactorSession) AddVoter(voter [20]byte, vtkey [32]byte) (*types.Transaction, error) {
	return _RelayerContract.Contract.AddVoter(&_RelayerContract.TransactOpts, voter, vtkey)
}

// RemoveVoter is a paid mutator transaction binding the contract method 0x6b75a2b3.
//
// Solidity: function removeVoter(bytes20 voter) returns()
func (_RelayerContract *RelayerContractTransactor) RemoveVoter(opts *bind.TransactOpts, voter [20]byte) (*types.Transaction, error) {
	return _RelayerContract.contract.Transact(opts, "removeVoter", voter)
}

// RemoveVoter is a paid mutator transaction binding the contract method 0x6b75a2b3.
//
// Solidity: function removeVoter(bytes20 voter) returns()
func (_RelayerContract *RelayerContractSession) RemoveVoter(voter [20]byte) (*types.Transaction, error) {
	return _RelayerContract.Contract.RemoveVoter(&_RelayerContract.TransactOpts, voter)
}

// RemoveVoter is a paid mutator transaction binding the contract method 0x6b75a2b3.
//
// Solidity: function removeVoter(bytes20 voter) returns()
func (_RelayerContract *RelayerContractTransactorSession) RemoveVoter(voter [20]byte) (*types.Transaction, error) {
	return _RelayerContract.Contract.RemoveVoter(&_RelayerContract.TransactOpts, voter)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_RelayerContract *RelayerContractTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _RelayerContract.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_RelayerContract *RelayerContractSession) RenounceOwnership() (*types.Transaction, error) {
	return _RelayerContract.Contract.RenounceOwnership(&_RelayerContract.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_RelayerContract *RelayerContractTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _RelayerContract.Contract.RenounceOwnership(&_RelayerContract.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_RelayerContract *RelayerContractTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _RelayerContract.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_RelayerContract *RelayerContractSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _RelayerContract.Contract.TransferOwnership(&_RelayerContract.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_RelayerContract *RelayerContractTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _RelayerContract.Contract.TransferOwnership(&_RelayerContract.TransactOpts, newOwner)
}

// RelayerContractAddedVoterIterator is returned from FilterAddedVoter and is used to iterate over the raw logs and unpacked data for AddedVoter events raised by the RelayerContract contract.
type RelayerContractAddedVoterIterator struct {
	Event *RelayerContractAddedVoter // Event containing the contract specifics and raw log

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
func (it *RelayerContractAddedVoterIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RelayerContractAddedVoter)
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
		it.Event = new(RelayerContractAddedVoter)
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
func (it *RelayerContractAddedVoterIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RelayerContractAddedVoterIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RelayerContractAddedVoter represents a AddedVoter event raised by the RelayerContract contract.
type RelayerContractAddedVoter struct {
	Voter   [20]byte
	KeyHash [32]byte
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterAddedVoter is a free log retrieval operation binding the contract event 0x2940969868a5545331aa91a95aaa97def154a66f4ed34985622a4136be3b1b04.
//
// Solidity: event AddedVoter(bytes20 indexed voter, bytes32 keyHash)
func (_RelayerContract *RelayerContractFilterer) FilterAddedVoter(opts *bind.FilterOpts, voter [][20]byte) (*RelayerContractAddedVoterIterator, error) {

	var voterRule []interface{}
	for _, voterItem := range voter {
		voterRule = append(voterRule, voterItem)
	}

	logs, sub, err := _RelayerContract.contract.FilterLogs(opts, "AddedVoter", voterRule)
	if err != nil {
		return nil, err
	}
	return &RelayerContractAddedVoterIterator{contract: _RelayerContract.contract, event: "AddedVoter", logs: logs, sub: sub}, nil
}

// WatchAddedVoter is a free log subscription operation binding the contract event 0x2940969868a5545331aa91a95aaa97def154a66f4ed34985622a4136be3b1b04.
//
// Solidity: event AddedVoter(bytes20 indexed voter, bytes32 keyHash)
func (_RelayerContract *RelayerContractFilterer) WatchAddedVoter(opts *bind.WatchOpts, sink chan<- *RelayerContractAddedVoter, voter [][20]byte) (event.Subscription, error) {

	var voterRule []interface{}
	for _, voterItem := range voter {
		voterRule = append(voterRule, voterItem)
	}

	logs, sub, err := _RelayerContract.contract.WatchLogs(opts, "AddedVoter", voterRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RelayerContractAddedVoter)
				if err := _RelayerContract.contract.UnpackLog(event, "AddedVoter", log); err != nil {
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

// ParseAddedVoter is a log parse operation binding the contract event 0x2940969868a5545331aa91a95aaa97def154a66f4ed34985622a4136be3b1b04.
//
// Solidity: event AddedVoter(bytes20 indexed voter, bytes32 keyHash)
func (_RelayerContract *RelayerContractFilterer) ParseAddedVoter(log types.Log) (*RelayerContractAddedVoter, error) {
	event := new(RelayerContractAddedVoter)
	if err := _RelayerContract.contract.UnpackLog(event, "AddedVoter", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RelayerContractOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the RelayerContract contract.
type RelayerContractOwnershipTransferredIterator struct {
	Event *RelayerContractOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *RelayerContractOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RelayerContractOwnershipTransferred)
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
		it.Event = new(RelayerContractOwnershipTransferred)
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
func (it *RelayerContractOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RelayerContractOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RelayerContractOwnershipTransferred represents a OwnershipTransferred event raised by the RelayerContract contract.
type RelayerContractOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_RelayerContract *RelayerContractFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*RelayerContractOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _RelayerContract.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &RelayerContractOwnershipTransferredIterator{contract: _RelayerContract.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_RelayerContract *RelayerContractFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *RelayerContractOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _RelayerContract.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RelayerContractOwnershipTransferred)
				if err := _RelayerContract.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_RelayerContract *RelayerContractFilterer) ParseOwnershipTransferred(log types.Log) (*RelayerContractOwnershipTransferred, error) {
	event := new(RelayerContractOwnershipTransferred)
	if err := _RelayerContract.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RelayerContractRemovedVoterIterator is returned from FilterRemovedVoter and is used to iterate over the raw logs and unpacked data for RemovedVoter events raised by the RelayerContract contract.
type RelayerContractRemovedVoterIterator struct {
	Event *RelayerContractRemovedVoter // Event containing the contract specifics and raw log

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
func (it *RelayerContractRemovedVoterIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RelayerContractRemovedVoter)
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
		it.Event = new(RelayerContractRemovedVoter)
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
func (it *RelayerContractRemovedVoterIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RelayerContractRemovedVoterIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RelayerContractRemovedVoter represents a RemovedVoter event raised by the RelayerContract contract.
type RelayerContractRemovedVoter struct {
	Voter [20]byte
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterRemovedVoter is a free log retrieval operation binding the contract event 0xa4d0070f0847ad2444266e520aecd08cc31d75b1f1facc657f007083f9857778.
//
// Solidity: event RemovedVoter(bytes20 indexed voter)
func (_RelayerContract *RelayerContractFilterer) FilterRemovedVoter(opts *bind.FilterOpts, voter [][20]byte) (*RelayerContractRemovedVoterIterator, error) {

	var voterRule []interface{}
	for _, voterItem := range voter {
		voterRule = append(voterRule, voterItem)
	}

	logs, sub, err := _RelayerContract.contract.FilterLogs(opts, "RemovedVoter", voterRule)
	if err != nil {
		return nil, err
	}
	return &RelayerContractRemovedVoterIterator{contract: _RelayerContract.contract, event: "RemovedVoter", logs: logs, sub: sub}, nil
}

// WatchRemovedVoter is a free log subscription operation binding the contract event 0xa4d0070f0847ad2444266e520aecd08cc31d75b1f1facc657f007083f9857778.
//
// Solidity: event RemovedVoter(bytes20 indexed voter)
func (_RelayerContract *RelayerContractFilterer) WatchRemovedVoter(opts *bind.WatchOpts, sink chan<- *RelayerContractRemovedVoter, voter [][20]byte) (event.Subscription, error) {

	var voterRule []interface{}
	for _, voterItem := range voter {
		voterRule = append(voterRule, voterItem)
	}

	logs, sub, err := _RelayerContract.contract.WatchLogs(opts, "RemovedVoter", voterRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RelayerContractRemovedVoter)
				if err := _RelayerContract.contract.UnpackLog(event, "RemovedVoter", log); err != nil {
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

// ParseRemovedVoter is a log parse operation binding the contract event 0xa4d0070f0847ad2444266e520aecd08cc31d75b1f1facc657f007083f9857778.
//
// Solidity: event RemovedVoter(bytes20 indexed voter)
func (_RelayerContract *RelayerContractFilterer) ParseRemovedVoter(log types.Log) (*RelayerContractRemovedVoter, error) {
	event := new(RelayerContractRemovedVoter)
	if err := _RelayerContract.contract.UnpackLog(event, "RemovedVoter", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
