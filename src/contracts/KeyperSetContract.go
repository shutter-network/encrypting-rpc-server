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

// KeyperSetContractMetaData contains all meta data concerning the KeyperSetContract contract.
var KeyperSetContractMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"AlreadyFinalized\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"newMembers\",\"type\":\"address[]\"}],\"name\":\"addMembers\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"index\",\"type\":\"uint64\"}],\"name\":\"getMember\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getMembers\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getNumMembers\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getThreshold\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"a\",\"type\":\"address\"}],\"name\":\"isAllowedToBroadcastEonKey\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"isFinalized\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"setFinalized\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_broadcaster\",\"type\":\"address\"}],\"name\":\"setKeyBroadcaster\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"_threshold\",\"type\":\"uint64\"}],\"name\":\"setThreshold\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// KeyperSetContractABI is the input ABI used to generate the binding from.
// Deprecated: Use KeyperSetContractMetaData.ABI instead.
var KeyperSetContractABI = KeyperSetContractMetaData.ABI

// KeyperSetContract is an auto generated Go binding around an Ethereum contract.
type KeyperSetContract struct {
	KeyperSetContractCaller     // Read-only binding to the contract
	KeyperSetContractTransactor // Write-only binding to the contract
	KeyperSetContractFilterer   // Log filterer for contract events
}

// KeyperSetContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type KeyperSetContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// KeyperSetContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type KeyperSetContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// KeyperSetContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type KeyperSetContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// KeyperSetContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type KeyperSetContractSession struct {
	Contract     *KeyperSetContract // Generic contract binding to set the session for
	CallOpts     bind.CallOpts      // Call options to use throughout this session
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// KeyperSetContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type KeyperSetContractCallerSession struct {
	Contract *KeyperSetContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts            // Call options to use throughout this session
}

// KeyperSetContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type KeyperSetContractTransactorSession struct {
	Contract     *KeyperSetContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts            // Transaction auth options to use throughout this session
}

// KeyperSetContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type KeyperSetContractRaw struct {
	Contract *KeyperSetContract // Generic contract binding to access the raw methods on
}

// KeyperSetContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type KeyperSetContractCallerRaw struct {
	Contract *KeyperSetContractCaller // Generic read-only contract binding to access the raw methods on
}

// KeyperSetContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type KeyperSetContractTransactorRaw struct {
	Contract *KeyperSetContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewKeyperSetContract creates a new instance of KeyperSetContract, bound to a specific deployed contract.
func NewKeyperSetContract(address common.Address, backend bind.ContractBackend) (*KeyperSetContract, error) {
	contract, err := bindKeyperSetContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &KeyperSetContract{KeyperSetContractCaller: KeyperSetContractCaller{contract: contract}, KeyperSetContractTransactor: KeyperSetContractTransactor{contract: contract}, KeyperSetContractFilterer: KeyperSetContractFilterer{contract: contract}}, nil
}

// NewKeyperSetContractCaller creates a new read-only instance of KeyperSetContract, bound to a specific deployed contract.
func NewKeyperSetContractCaller(address common.Address, caller bind.ContractCaller) (*KeyperSetContractCaller, error) {
	contract, err := bindKeyperSetContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &KeyperSetContractCaller{contract: contract}, nil
}

// NewKeyperSetContractTransactor creates a new write-only instance of KeyperSetContract, bound to a specific deployed contract.
func NewKeyperSetContractTransactor(address common.Address, transactor bind.ContractTransactor) (*KeyperSetContractTransactor, error) {
	contract, err := bindKeyperSetContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &KeyperSetContractTransactor{contract: contract}, nil
}

// NewKeyperSetContractFilterer creates a new log filterer instance of KeyperSetContract, bound to a specific deployed contract.
func NewKeyperSetContractFilterer(address common.Address, filterer bind.ContractFilterer) (*KeyperSetContractFilterer, error) {
	contract, err := bindKeyperSetContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &KeyperSetContractFilterer{contract: contract}, nil
}

// bindKeyperSetContract binds a generic wrapper to an already deployed contract.
func bindKeyperSetContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := KeyperSetContractMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_KeyperSetContract *KeyperSetContractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _KeyperSetContract.Contract.KeyperSetContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_KeyperSetContract *KeyperSetContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _KeyperSetContract.Contract.KeyperSetContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_KeyperSetContract *KeyperSetContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _KeyperSetContract.Contract.KeyperSetContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_KeyperSetContract *KeyperSetContractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _KeyperSetContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_KeyperSetContract *KeyperSetContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _KeyperSetContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_KeyperSetContract *KeyperSetContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _KeyperSetContract.Contract.contract.Transact(opts, method, params...)
}

// GetMember is a free data retrieval call binding the contract method 0x2e8e6cad.
//
// Solidity: function getMember(uint64 index) view returns(address)
func (_KeyperSetContract *KeyperSetContractCaller) GetMember(opts *bind.CallOpts, index uint64) (common.Address, error) {
	var out []interface{}
	err := _KeyperSetContract.contract.Call(opts, &out, "getMember", index)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetMember is a free data retrieval call binding the contract method 0x2e8e6cad.
//
// Solidity: function getMember(uint64 index) view returns(address)
func (_KeyperSetContract *KeyperSetContractSession) GetMember(index uint64) (common.Address, error) {
	return _KeyperSetContract.Contract.GetMember(&_KeyperSetContract.CallOpts, index)
}

// GetMember is a free data retrieval call binding the contract method 0x2e8e6cad.
//
// Solidity: function getMember(uint64 index) view returns(address)
func (_KeyperSetContract *KeyperSetContractCallerSession) GetMember(index uint64) (common.Address, error) {
	return _KeyperSetContract.Contract.GetMember(&_KeyperSetContract.CallOpts, index)
}

// GetMembers is a free data retrieval call binding the contract method 0x9eab5253.
//
// Solidity: function getMembers() view returns(address[])
func (_KeyperSetContract *KeyperSetContractCaller) GetMembers(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _KeyperSetContract.contract.Call(opts, &out, "getMembers")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetMembers is a free data retrieval call binding the contract method 0x9eab5253.
//
// Solidity: function getMembers() view returns(address[])
func (_KeyperSetContract *KeyperSetContractSession) GetMembers() ([]common.Address, error) {
	return _KeyperSetContract.Contract.GetMembers(&_KeyperSetContract.CallOpts)
}

// GetMembers is a free data retrieval call binding the contract method 0x9eab5253.
//
// Solidity: function getMembers() view returns(address[])
func (_KeyperSetContract *KeyperSetContractCallerSession) GetMembers() ([]common.Address, error) {
	return _KeyperSetContract.Contract.GetMembers(&_KeyperSetContract.CallOpts)
}

// GetNumMembers is a free data retrieval call binding the contract method 0x17d5430a.
//
// Solidity: function getNumMembers() view returns(uint64)
func (_KeyperSetContract *KeyperSetContractCaller) GetNumMembers(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _KeyperSetContract.contract.Call(opts, &out, "getNumMembers")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// GetNumMembers is a free data retrieval call binding the contract method 0x17d5430a.
//
// Solidity: function getNumMembers() view returns(uint64)
func (_KeyperSetContract *KeyperSetContractSession) GetNumMembers() (uint64, error) {
	return _KeyperSetContract.Contract.GetNumMembers(&_KeyperSetContract.CallOpts)
}

// GetNumMembers is a free data retrieval call binding the contract method 0x17d5430a.
//
// Solidity: function getNumMembers() view returns(uint64)
func (_KeyperSetContract *KeyperSetContractCallerSession) GetNumMembers() (uint64, error) {
	return _KeyperSetContract.Contract.GetNumMembers(&_KeyperSetContract.CallOpts)
}

// GetThreshold is a free data retrieval call binding the contract method 0xe75235b8.
//
// Solidity: function getThreshold() view returns(uint64)
func (_KeyperSetContract *KeyperSetContractCaller) GetThreshold(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _KeyperSetContract.contract.Call(opts, &out, "getThreshold")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// GetThreshold is a free data retrieval call binding the contract method 0xe75235b8.
//
// Solidity: function getThreshold() view returns(uint64)
func (_KeyperSetContract *KeyperSetContractSession) GetThreshold() (uint64, error) {
	return _KeyperSetContract.Contract.GetThreshold(&_KeyperSetContract.CallOpts)
}

// GetThreshold is a free data retrieval call binding the contract method 0xe75235b8.
//
// Solidity: function getThreshold() view returns(uint64)
func (_KeyperSetContract *KeyperSetContractCallerSession) GetThreshold() (uint64, error) {
	return _KeyperSetContract.Contract.GetThreshold(&_KeyperSetContract.CallOpts)
}

// IsAllowedToBroadcastEonKey is a free data retrieval call binding the contract method 0xcde1532d.
//
// Solidity: function isAllowedToBroadcastEonKey(address a) view returns(bool)
func (_KeyperSetContract *KeyperSetContractCaller) IsAllowedToBroadcastEonKey(opts *bind.CallOpts, a common.Address) (bool, error) {
	var out []interface{}
	err := _KeyperSetContract.contract.Call(opts, &out, "isAllowedToBroadcastEonKey", a)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsAllowedToBroadcastEonKey is a free data retrieval call binding the contract method 0xcde1532d.
//
// Solidity: function isAllowedToBroadcastEonKey(address a) view returns(bool)
func (_KeyperSetContract *KeyperSetContractSession) IsAllowedToBroadcastEonKey(a common.Address) (bool, error) {
	return _KeyperSetContract.Contract.IsAllowedToBroadcastEonKey(&_KeyperSetContract.CallOpts, a)
}

// IsAllowedToBroadcastEonKey is a free data retrieval call binding the contract method 0xcde1532d.
//
// Solidity: function isAllowedToBroadcastEonKey(address a) view returns(bool)
func (_KeyperSetContract *KeyperSetContractCallerSession) IsAllowedToBroadcastEonKey(a common.Address) (bool, error) {
	return _KeyperSetContract.Contract.IsAllowedToBroadcastEonKey(&_KeyperSetContract.CallOpts, a)
}

// IsFinalized is a free data retrieval call binding the contract method 0x8d4e4083.
//
// Solidity: function isFinalized() view returns(bool)
func (_KeyperSetContract *KeyperSetContractCaller) IsFinalized(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _KeyperSetContract.contract.Call(opts, &out, "isFinalized")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsFinalized is a free data retrieval call binding the contract method 0x8d4e4083.
//
// Solidity: function isFinalized() view returns(bool)
func (_KeyperSetContract *KeyperSetContractSession) IsFinalized() (bool, error) {
	return _KeyperSetContract.Contract.IsFinalized(&_KeyperSetContract.CallOpts)
}

// IsFinalized is a free data retrieval call binding the contract method 0x8d4e4083.
//
// Solidity: function isFinalized() view returns(bool)
func (_KeyperSetContract *KeyperSetContractCallerSession) IsFinalized() (bool, error) {
	return _KeyperSetContract.Contract.IsFinalized(&_KeyperSetContract.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_KeyperSetContract *KeyperSetContractCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _KeyperSetContract.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_KeyperSetContract *KeyperSetContractSession) Owner() (common.Address, error) {
	return _KeyperSetContract.Contract.Owner(&_KeyperSetContract.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_KeyperSetContract *KeyperSetContractCallerSession) Owner() (common.Address, error) {
	return _KeyperSetContract.Contract.Owner(&_KeyperSetContract.CallOpts)
}

// AddMembers is a paid mutator transaction binding the contract method 0x6f4d469b.
//
// Solidity: function addMembers(address[] newMembers) returns()
func (_KeyperSetContract *KeyperSetContractTransactor) AddMembers(opts *bind.TransactOpts, newMembers []common.Address) (*types.Transaction, error) {
	return _KeyperSetContract.contract.Transact(opts, "addMembers", newMembers)
}

// AddMembers is a paid mutator transaction binding the contract method 0x6f4d469b.
//
// Solidity: function addMembers(address[] newMembers) returns()
func (_KeyperSetContract *KeyperSetContractSession) AddMembers(newMembers []common.Address) (*types.Transaction, error) {
	return _KeyperSetContract.Contract.AddMembers(&_KeyperSetContract.TransactOpts, newMembers)
}

// AddMembers is a paid mutator transaction binding the contract method 0x6f4d469b.
//
// Solidity: function addMembers(address[] newMembers) returns()
func (_KeyperSetContract *KeyperSetContractTransactorSession) AddMembers(newMembers []common.Address) (*types.Transaction, error) {
	return _KeyperSetContract.Contract.AddMembers(&_KeyperSetContract.TransactOpts, newMembers)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_KeyperSetContract *KeyperSetContractTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _KeyperSetContract.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_KeyperSetContract *KeyperSetContractSession) RenounceOwnership() (*types.Transaction, error) {
	return _KeyperSetContract.Contract.RenounceOwnership(&_KeyperSetContract.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_KeyperSetContract *KeyperSetContractTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _KeyperSetContract.Contract.RenounceOwnership(&_KeyperSetContract.TransactOpts)
}

// SetFinalized is a paid mutator transaction binding the contract method 0x1de77253.
//
// Solidity: function setFinalized() returns()
func (_KeyperSetContract *KeyperSetContractTransactor) SetFinalized(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _KeyperSetContract.contract.Transact(opts, "setFinalized")
}

// SetFinalized is a paid mutator transaction binding the contract method 0x1de77253.
//
// Solidity: function setFinalized() returns()
func (_KeyperSetContract *KeyperSetContractSession) SetFinalized() (*types.Transaction, error) {
	return _KeyperSetContract.Contract.SetFinalized(&_KeyperSetContract.TransactOpts)
}

// SetFinalized is a paid mutator transaction binding the contract method 0x1de77253.
//
// Solidity: function setFinalized() returns()
func (_KeyperSetContract *KeyperSetContractTransactorSession) SetFinalized() (*types.Transaction, error) {
	return _KeyperSetContract.Contract.SetFinalized(&_KeyperSetContract.TransactOpts)
}

// SetKeyBroadcaster is a paid mutator transaction binding the contract method 0x6a33e20e.
//
// Solidity: function setKeyBroadcaster(address _broadcaster) returns()
func (_KeyperSetContract *KeyperSetContractTransactor) SetKeyBroadcaster(opts *bind.TransactOpts, _broadcaster common.Address) (*types.Transaction, error) {
	return _KeyperSetContract.contract.Transact(opts, "setKeyBroadcaster", _broadcaster)
}

// SetKeyBroadcaster is a paid mutator transaction binding the contract method 0x6a33e20e.
//
// Solidity: function setKeyBroadcaster(address _broadcaster) returns()
func (_KeyperSetContract *KeyperSetContractSession) SetKeyBroadcaster(_broadcaster common.Address) (*types.Transaction, error) {
	return _KeyperSetContract.Contract.SetKeyBroadcaster(&_KeyperSetContract.TransactOpts, _broadcaster)
}

// SetKeyBroadcaster is a paid mutator transaction binding the contract method 0x6a33e20e.
//
// Solidity: function setKeyBroadcaster(address _broadcaster) returns()
func (_KeyperSetContract *KeyperSetContractTransactorSession) SetKeyBroadcaster(_broadcaster common.Address) (*types.Transaction, error) {
	return _KeyperSetContract.Contract.SetKeyBroadcaster(&_KeyperSetContract.TransactOpts, _broadcaster)
}

// SetThreshold is a paid mutator transaction binding the contract method 0x17c4de35.
//
// Solidity: function setThreshold(uint64 _threshold) returns()
func (_KeyperSetContract *KeyperSetContractTransactor) SetThreshold(opts *bind.TransactOpts, _threshold uint64) (*types.Transaction, error) {
	return _KeyperSetContract.contract.Transact(opts, "setThreshold", _threshold)
}

// SetThreshold is a paid mutator transaction binding the contract method 0x17c4de35.
//
// Solidity: function setThreshold(uint64 _threshold) returns()
func (_KeyperSetContract *KeyperSetContractSession) SetThreshold(_threshold uint64) (*types.Transaction, error) {
	return _KeyperSetContract.Contract.SetThreshold(&_KeyperSetContract.TransactOpts, _threshold)
}

// SetThreshold is a paid mutator transaction binding the contract method 0x17c4de35.
//
// Solidity: function setThreshold(uint64 _threshold) returns()
func (_KeyperSetContract *KeyperSetContractTransactorSession) SetThreshold(_threshold uint64) (*types.Transaction, error) {
	return _KeyperSetContract.Contract.SetThreshold(&_KeyperSetContract.TransactOpts, _threshold)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_KeyperSetContract *KeyperSetContractTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _KeyperSetContract.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_KeyperSetContract *KeyperSetContractSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _KeyperSetContract.Contract.TransferOwnership(&_KeyperSetContract.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_KeyperSetContract *KeyperSetContractTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _KeyperSetContract.Contract.TransferOwnership(&_KeyperSetContract.TransactOpts, newOwner)
}

// KeyperSetContractOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the KeyperSetContract contract.
type KeyperSetContractOwnershipTransferredIterator struct {
	Event *KeyperSetContractOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *KeyperSetContractOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(KeyperSetContractOwnershipTransferred)
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
		it.Event = new(KeyperSetContractOwnershipTransferred)
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
func (it *KeyperSetContractOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *KeyperSetContractOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// KeyperSetContractOwnershipTransferred represents a OwnershipTransferred event raised by the KeyperSetContract contract.
type KeyperSetContractOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_KeyperSetContract *KeyperSetContractFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*KeyperSetContractOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _KeyperSetContract.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &KeyperSetContractOwnershipTransferredIterator{contract: _KeyperSetContract.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_KeyperSetContract *KeyperSetContractFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *KeyperSetContractOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _KeyperSetContract.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(KeyperSetContractOwnershipTransferred)
				if err := _KeyperSetContract.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_KeyperSetContract *KeyperSetContractFilterer) ParseOwnershipTransferred(log types.Log) (*KeyperSetContractOwnershipTransferred, error) {
	event := new(KeyperSetContractOwnershipTransferred)
	if err := _KeyperSetContract.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
