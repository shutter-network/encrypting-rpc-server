// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contracts

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

// KeyperSetManagerContractMetaData contains all meta data concerning the KeyperSetManagerContract contract.
var KeyperSetManagerContractMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"AlreadyHaveKeyperSet\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"KeyperSetNotFinalized\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NoActiveKeyperSet\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"activationSlot\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"keyperSetContract\",\"type\":\"address\"}],\"name\":\"KeyperSetAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"activationSlot\",\"type\":\"uint64\"},{\"internalType\":\"address\",\"name\":\"keyperSetContract\",\"type\":\"address\"}],\"name\":\"addKeyperSet\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"index\",\"type\":\"uint64\"}],\"name\":\"getKeyperSetActivationSlot\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"index\",\"type\":\"uint64\"}],\"name\":\"getKeyperSetAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"slot\",\"type\":\"uint64\"}],\"name\":\"getKeyperSetIndexBySlot\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getNumKeyperSets\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// KeyperSetManagerContractABI is the input ABI used to generate the binding from.
// Deprecated: Use KeyperSetManagerContractMetaData.ABI instead.
var KeyperSetManagerContractABI = KeyperSetManagerContractMetaData.ABI

// KeyperSetManagerContract is an auto generated Go binding around an Ethereum contract.
type KeyperSetManagerContract struct {
	KeyperSetManagerContractCaller     // Read-only binding to the contract
	KeyperSetManagerContractTransactor // Write-only binding to the contract
	KeyperSetManagerContractFilterer   // Log filterer for contract events
}

// KeyperSetManagerContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type KeyperSetManagerContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// KeyperSetManagerContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type KeyperSetManagerContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// KeyperSetManagerContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type KeyperSetManagerContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// KeyperSetManagerContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type KeyperSetManagerContractSession struct {
	Contract     *KeyperSetManagerContract // Generic contract binding to set the session for
	CallOpts     bind.CallOpts             // Call options to use throughout this session
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// KeyperSetManagerContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type KeyperSetManagerContractCallerSession struct {
	Contract *KeyperSetManagerContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                   // Call options to use throughout this session
}

// KeyperSetManagerContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type KeyperSetManagerContractTransactorSession struct {
	Contract     *KeyperSetManagerContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                   // Transaction auth options to use throughout this session
}

// KeyperSetManagerContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type KeyperSetManagerContractRaw struct {
	Contract *KeyperSetManagerContract // Generic contract binding to access the raw methods on
}

// KeyperSetManagerContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type KeyperSetManagerContractCallerRaw struct {
	Contract *KeyperSetManagerContractCaller // Generic read-only contract binding to access the raw methods on
}

// KeyperSetManagerContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type KeyperSetManagerContractTransactorRaw struct {
	Contract *KeyperSetManagerContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewKeyperSetManagerContract creates a new instance of KeyperSetManagerContract, bound to a specific deployed contract.
func NewKeyperSetManagerContract(address common.Address, backend bind.ContractBackend) (*KeyperSetManagerContract, error) {
	contract, err := bindKeyperSetManagerContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &KeyperSetManagerContract{KeyperSetManagerContractCaller: KeyperSetManagerContractCaller{contract: contract}, KeyperSetManagerContractTransactor: KeyperSetManagerContractTransactor{contract: contract}, KeyperSetManagerContractFilterer: KeyperSetManagerContractFilterer{contract: contract}}, nil
}

// NewKeyperSetManagerContractCaller creates a new read-only instance of KeyperSetManagerContract, bound to a specific deployed contract.
func NewKeyperSetManagerContractCaller(address common.Address, caller bind.ContractCaller) (*KeyperSetManagerContractCaller, error) {
	contract, err := bindKeyperSetManagerContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &KeyperSetManagerContractCaller{contract: contract}, nil
}

// NewKeyperSetManagerContractTransactor creates a new write-only instance of KeyperSetManagerContract, bound to a specific deployed contract.
func NewKeyperSetManagerContractTransactor(address common.Address, transactor bind.ContractTransactor) (*KeyperSetManagerContractTransactor, error) {
	contract, err := bindKeyperSetManagerContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &KeyperSetManagerContractTransactor{contract: contract}, nil
}

// NewKeyperSetManagerContractFilterer creates a new log filterer instance of KeyperSetManagerContract, bound to a specific deployed contract.
func NewKeyperSetManagerContractFilterer(address common.Address, filterer bind.ContractFilterer) (*KeyperSetManagerContractFilterer, error) {
	contract, err := bindKeyperSetManagerContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &KeyperSetManagerContractFilterer{contract: contract}, nil
}

// bindKeyperSetManagerContract binds a generic wrapper to an already deployed contract.
func bindKeyperSetManagerContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := KeyperSetManagerContractMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_KeyperSetManagerContract *KeyperSetManagerContractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _KeyperSetManagerContract.Contract.KeyperSetManagerContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_KeyperSetManagerContract *KeyperSetManagerContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _KeyperSetManagerContract.Contract.KeyperSetManagerContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_KeyperSetManagerContract *KeyperSetManagerContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _KeyperSetManagerContract.Contract.KeyperSetManagerContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_KeyperSetManagerContract *KeyperSetManagerContractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _KeyperSetManagerContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_KeyperSetManagerContract *KeyperSetManagerContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _KeyperSetManagerContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_KeyperSetManagerContract *KeyperSetManagerContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _KeyperSetManagerContract.Contract.contract.Transact(opts, method, params...)
}

// GetKeyperSetActivationSlot is a free data retrieval call binding the contract method 0x6254bd0b.
//
// Solidity: function getKeyperSetActivationSlot(uint64 index) view returns(uint64)
func (_KeyperSetManagerContract *KeyperSetManagerContractCaller) GetKeyperSetActivationSlot(opts *bind.CallOpts, index uint64) (uint64, error) {
	var out []interface{}
	err := _KeyperSetManagerContract.contract.Call(opts, &out, "getKeyperSetActivationSlot", index)

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// GetKeyperSetActivationSlot is a free data retrieval call binding the contract method 0x6254bd0b.
//
// Solidity: function getKeyperSetActivationSlot(uint64 index) view returns(uint64)
func (_KeyperSetManagerContract *KeyperSetManagerContractSession) GetKeyperSetActivationSlot(index uint64) (uint64, error) {
	return _KeyperSetManagerContract.Contract.GetKeyperSetActivationSlot(&_KeyperSetManagerContract.CallOpts, index)
}

// GetKeyperSetActivationSlot is a free data retrieval call binding the contract method 0x6254bd0b.
//
// Solidity: function getKeyperSetActivationSlot(uint64 index) view returns(uint64)
func (_KeyperSetManagerContract *KeyperSetManagerContractCallerSession) GetKeyperSetActivationSlot(index uint64) (uint64, error) {
	return _KeyperSetManagerContract.Contract.GetKeyperSetActivationSlot(&_KeyperSetManagerContract.CallOpts, index)
}

// GetKeyperSetAddress is a free data retrieval call binding the contract method 0xf90f3bed.
//
// Solidity: function getKeyperSetAddress(uint64 index) view returns(address)
func (_KeyperSetManagerContract *KeyperSetManagerContractCaller) GetKeyperSetAddress(opts *bind.CallOpts, index uint64) (common.Address, error) {
	var out []interface{}
	err := _KeyperSetManagerContract.contract.Call(opts, &out, "getKeyperSetAddress", index)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetKeyperSetAddress is a free data retrieval call binding the contract method 0xf90f3bed.
//
// Solidity: function getKeyperSetAddress(uint64 index) view returns(address)
func (_KeyperSetManagerContract *KeyperSetManagerContractSession) GetKeyperSetAddress(index uint64) (common.Address, error) {
	return _KeyperSetManagerContract.Contract.GetKeyperSetAddress(&_KeyperSetManagerContract.CallOpts, index)
}

// GetKeyperSetAddress is a free data retrieval call binding the contract method 0xf90f3bed.
//
// Solidity: function getKeyperSetAddress(uint64 index) view returns(address)
func (_KeyperSetManagerContract *KeyperSetManagerContractCallerSession) GetKeyperSetAddress(index uint64) (common.Address, error) {
	return _KeyperSetManagerContract.Contract.GetKeyperSetAddress(&_KeyperSetManagerContract.CallOpts, index)
}

// GetKeyperSetIndexBySlot is a free data retrieval call binding the contract method 0xcf751d5f.
//
// Solidity: function getKeyperSetIndexBySlot(uint64 slot) view returns(uint64)
func (_KeyperSetManagerContract *KeyperSetManagerContractCaller) GetKeyperSetIndexBySlot(opts *bind.CallOpts, slot uint64) (uint64, error) {
	var out []interface{}
	err := _KeyperSetManagerContract.contract.Call(opts, &out, "getKeyperSetIndexBySlot", slot)

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// GetKeyperSetIndexBySlot is a free data retrieval call binding the contract method 0xcf751d5f.
//
// Solidity: function getKeyperSetIndexBySlot(uint64 slot) view returns(uint64)
func (_KeyperSetManagerContract *KeyperSetManagerContractSession) GetKeyperSetIndexBySlot(slot uint64) (uint64, error) {
	return _KeyperSetManagerContract.Contract.GetKeyperSetIndexBySlot(&_KeyperSetManagerContract.CallOpts, slot)
}

// GetKeyperSetIndexBySlot is a free data retrieval call binding the contract method 0xcf751d5f.
//
// Solidity: function getKeyperSetIndexBySlot(uint64 slot) view returns(uint64)
func (_KeyperSetManagerContract *KeyperSetManagerContractCallerSession) GetKeyperSetIndexBySlot(slot uint64) (uint64, error) {
	return _KeyperSetManagerContract.Contract.GetKeyperSetIndexBySlot(&_KeyperSetManagerContract.CallOpts, slot)
}

// GetNumKeyperSets is a free data retrieval call binding the contract method 0xf2e6100a.
//
// Solidity: function getNumKeyperSets() view returns(uint64)
func (_KeyperSetManagerContract *KeyperSetManagerContractCaller) GetNumKeyperSets(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _KeyperSetManagerContract.contract.Call(opts, &out, "getNumKeyperSets")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// GetNumKeyperSets is a free data retrieval call binding the contract method 0xf2e6100a.
//
// Solidity: function getNumKeyperSets() view returns(uint64)
func (_KeyperSetManagerContract *KeyperSetManagerContractSession) GetNumKeyperSets() (uint64, error) {
	return _KeyperSetManagerContract.Contract.GetNumKeyperSets(&_KeyperSetManagerContract.CallOpts)
}

// GetNumKeyperSets is a free data retrieval call binding the contract method 0xf2e6100a.
//
// Solidity: function getNumKeyperSets() view returns(uint64)
func (_KeyperSetManagerContract *KeyperSetManagerContractCallerSession) GetNumKeyperSets() (uint64, error) {
	return _KeyperSetManagerContract.Contract.GetNumKeyperSets(&_KeyperSetManagerContract.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_KeyperSetManagerContract *KeyperSetManagerContractCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _KeyperSetManagerContract.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_KeyperSetManagerContract *KeyperSetManagerContractSession) Owner() (common.Address, error) {
	return _KeyperSetManagerContract.Contract.Owner(&_KeyperSetManagerContract.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_KeyperSetManagerContract *KeyperSetManagerContractCallerSession) Owner() (common.Address, error) {
	return _KeyperSetManagerContract.Contract.Owner(&_KeyperSetManagerContract.CallOpts)
}

// AddKeyperSet is a paid mutator transaction binding the contract method 0xd3877c43.
//
// Solidity: function addKeyperSet(uint64 activationSlot, address keyperSetContract) returns()
func (_KeyperSetManagerContract *KeyperSetManagerContractTransactor) AddKeyperSet(opts *bind.TransactOpts, activationSlot uint64, keyperSetContract common.Address) (*types.Transaction, error) {
	return _KeyperSetManagerContract.contract.Transact(opts, "addKeyperSet", activationSlot, keyperSetContract)
}

// AddKeyperSet is a paid mutator transaction binding the contract method 0xd3877c43.
//
// Solidity: function addKeyperSet(uint64 activationSlot, address keyperSetContract) returns()
func (_KeyperSetManagerContract *KeyperSetManagerContractSession) AddKeyperSet(activationSlot uint64, keyperSetContract common.Address) (*types.Transaction, error) {
	return _KeyperSetManagerContract.Contract.AddKeyperSet(&_KeyperSetManagerContract.TransactOpts, activationSlot, keyperSetContract)
}

// AddKeyperSet is a paid mutator transaction binding the contract method 0xd3877c43.
//
// Solidity: function addKeyperSet(uint64 activationSlot, address keyperSetContract) returns()
func (_KeyperSetManagerContract *KeyperSetManagerContractTransactorSession) AddKeyperSet(activationSlot uint64, keyperSetContract common.Address) (*types.Transaction, error) {
	return _KeyperSetManagerContract.Contract.AddKeyperSet(&_KeyperSetManagerContract.TransactOpts, activationSlot, keyperSetContract)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_KeyperSetManagerContract *KeyperSetManagerContractTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _KeyperSetManagerContract.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_KeyperSetManagerContract *KeyperSetManagerContractSession) RenounceOwnership() (*types.Transaction, error) {
	return _KeyperSetManagerContract.Contract.RenounceOwnership(&_KeyperSetManagerContract.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_KeyperSetManagerContract *KeyperSetManagerContractTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _KeyperSetManagerContract.Contract.RenounceOwnership(&_KeyperSetManagerContract.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_KeyperSetManagerContract *KeyperSetManagerContractTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _KeyperSetManagerContract.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_KeyperSetManagerContract *KeyperSetManagerContractSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _KeyperSetManagerContract.Contract.TransferOwnership(&_KeyperSetManagerContract.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_KeyperSetManagerContract *KeyperSetManagerContractTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _KeyperSetManagerContract.Contract.TransferOwnership(&_KeyperSetManagerContract.TransactOpts, newOwner)
}

// KeyperSetManagerContractKeyperSetAddedIterator is returned from FilterKeyperSetAdded and is used to iterate over the raw logs and unpacked data for KeyperSetAdded events raised by the KeyperSetManagerContract contract.
type KeyperSetManagerContractKeyperSetAddedIterator struct {
	Event *KeyperSetManagerContractKeyperSetAdded // Event containing the contract specifics and raw log

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
func (it *KeyperSetManagerContractKeyperSetAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(KeyperSetManagerContractKeyperSetAdded)
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
		it.Event = new(KeyperSetManagerContractKeyperSetAdded)
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
func (it *KeyperSetManagerContractKeyperSetAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *KeyperSetManagerContractKeyperSetAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// KeyperSetManagerContractKeyperSetAdded represents a KeyperSetAdded event raised by the KeyperSetManagerContract contract.
type KeyperSetManagerContractKeyperSetAdded struct {
	ActivationSlot    uint64
	KeyperSetContract common.Address
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterKeyperSetAdded is a free log retrieval operation binding the contract event 0x6605cb866297050f9f49ae7e0b38e0e4c8178d4b176e24332bc01672818d707b.
//
// Solidity: event KeyperSetAdded(uint64 activationSlot, address keyperSetContract)
func (_KeyperSetManagerContract *KeyperSetManagerContractFilterer) FilterKeyperSetAdded(opts *bind.FilterOpts) (*KeyperSetManagerContractKeyperSetAddedIterator, error) {

	logs, sub, err := _KeyperSetManagerContract.contract.FilterLogs(opts, "KeyperSetAdded")
	if err != nil {
		return nil, err
	}
	return &KeyperSetManagerContractKeyperSetAddedIterator{contract: _KeyperSetManagerContract.contract, event: "KeyperSetAdded", logs: logs, sub: sub}, nil
}

// WatchKeyperSetAdded is a free log subscription operation binding the contract event 0x6605cb866297050f9f49ae7e0b38e0e4c8178d4b176e24332bc01672818d707b.
//
// Solidity: event KeyperSetAdded(uint64 activationSlot, address keyperSetContract)
func (_KeyperSetManagerContract *KeyperSetManagerContractFilterer) WatchKeyperSetAdded(opts *bind.WatchOpts, sink chan<- *KeyperSetManagerContractKeyperSetAdded) (event.Subscription, error) {

	logs, sub, err := _KeyperSetManagerContract.contract.WatchLogs(opts, "KeyperSetAdded")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(KeyperSetManagerContractKeyperSetAdded)
				if err := _KeyperSetManagerContract.contract.UnpackLog(event, "KeyperSetAdded", log); err != nil {
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

// ParseKeyperSetAdded is a log parse operation binding the contract event 0x6605cb866297050f9f49ae7e0b38e0e4c8178d4b176e24332bc01672818d707b.
//
// Solidity: event KeyperSetAdded(uint64 activationSlot, address keyperSetContract)
func (_KeyperSetManagerContract *KeyperSetManagerContractFilterer) ParseKeyperSetAdded(log types.Log) (*KeyperSetManagerContractKeyperSetAdded, error) {
	event := new(KeyperSetManagerContractKeyperSetAdded)
	if err := _KeyperSetManagerContract.contract.UnpackLog(event, "KeyperSetAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// KeyperSetManagerContractOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the KeyperSetManagerContract contract.
type KeyperSetManagerContractOwnershipTransferredIterator struct {
	Event *KeyperSetManagerContractOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *KeyperSetManagerContractOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(KeyperSetManagerContractOwnershipTransferred)
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
		it.Event = new(KeyperSetManagerContractOwnershipTransferred)
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
func (it *KeyperSetManagerContractOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *KeyperSetManagerContractOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// KeyperSetManagerContractOwnershipTransferred represents a OwnershipTransferred event raised by the KeyperSetManagerContract contract.
type KeyperSetManagerContractOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_KeyperSetManagerContract *KeyperSetManagerContractFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*KeyperSetManagerContractOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _KeyperSetManagerContract.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &KeyperSetManagerContractOwnershipTransferredIterator{contract: _KeyperSetManagerContract.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_KeyperSetManagerContract *KeyperSetManagerContractFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *KeyperSetManagerContractOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _KeyperSetManagerContract.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(KeyperSetManagerContractOwnershipTransferred)
				if err := _KeyperSetManagerContract.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_KeyperSetManagerContract *KeyperSetManagerContractFilterer) ParseOwnershipTransferred(log types.Log) (*KeyperSetManagerContractOwnershipTransferred, error) {
	event := new(KeyperSetManagerContractOwnershipTransferred)
	if err := _KeyperSetManagerContract.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
