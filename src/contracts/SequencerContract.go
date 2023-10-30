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

// SequencerContractMetaData contains all meta data concerning the SequencerContract contract.
var SequencerContractMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"InsufficientFee\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"message\",\"type\":\"bytes\"}],\"name\":\"DecryptionProgressSubmitted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"eon\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"identityPrefix\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"encryptedTransaction\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"gasLimit\",\"type\":\"uint256\"}],\"name\":\"TransactionSubmitted\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"message\",\"type\":\"bytes\"}],\"name\":\"submitDecryptionProgress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"eon\",\"type\":\"uint64\"},{\"internalType\":\"bytes32\",\"name\":\"identityPrefix\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"encryptedTransaction\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"gasLimit\",\"type\":\"uint256\"}],\"name\":\"submitEncryptedTransaction\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"}]",
}

// SequencerContractABI is the input ABI used to generate the binding from.
// Deprecated: Use SequencerContractMetaData.ABI instead.
var SequencerContractABI = SequencerContractMetaData.ABI

// SequencerContract is an auto generated Go binding around an Ethereum contract.
type SequencerContract struct {
	SequencerContractCaller     // Read-only binding to the contract
	SequencerContractTransactor // Write-only binding to the contract
	SequencerContractFilterer   // Log filterer for contract events
}

// SequencerContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type SequencerContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SequencerContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SequencerContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SequencerContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SequencerContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SequencerContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SequencerContractSession struct {
	Contract     *SequencerContract // Generic contract binding to set the session for
	CallOpts     bind.CallOpts      // Call options to use throughout this session
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// SequencerContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SequencerContractCallerSession struct {
	Contract *SequencerContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts            // Call options to use throughout this session
}

// SequencerContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SequencerContractTransactorSession struct {
	Contract     *SequencerContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts            // Transaction auth options to use throughout this session
}

// SequencerContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type SequencerContractRaw struct {
	Contract *SequencerContract // Generic contract binding to access the raw methods on
}

// SequencerContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SequencerContractCallerRaw struct {
	Contract *SequencerContractCaller // Generic read-only contract binding to access the raw methods on
}

// SequencerContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SequencerContractTransactorRaw struct {
	Contract *SequencerContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSequencerContract creates a new instance of SequencerContract, bound to a specific deployed contract.
func NewSequencerContract(address common.Address, backend bind.ContractBackend) (*SequencerContract, error) {
	contract, err := bindSequencerContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SequencerContract{SequencerContractCaller: SequencerContractCaller{contract: contract}, SequencerContractTransactor: SequencerContractTransactor{contract: contract}, SequencerContractFilterer: SequencerContractFilterer{contract: contract}}, nil
}

// NewSequencerContractCaller creates a new read-only instance of SequencerContract, bound to a specific deployed contract.
func NewSequencerContractCaller(address common.Address, caller bind.ContractCaller) (*SequencerContractCaller, error) {
	contract, err := bindSequencerContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SequencerContractCaller{contract: contract}, nil
}

// NewSequencerContractTransactor creates a new write-only instance of SequencerContract, bound to a specific deployed contract.
func NewSequencerContractTransactor(address common.Address, transactor bind.ContractTransactor) (*SequencerContractTransactor, error) {
	contract, err := bindSequencerContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SequencerContractTransactor{contract: contract}, nil
}

// NewSequencerContractFilterer creates a new log filterer instance of SequencerContract, bound to a specific deployed contract.
func NewSequencerContractFilterer(address common.Address, filterer bind.ContractFilterer) (*SequencerContractFilterer, error) {
	contract, err := bindSequencerContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SequencerContractFilterer{contract: contract}, nil
}

// bindSequencerContract binds a generic wrapper to an already deployed contract.
func bindSequencerContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := SequencerContractMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SequencerContract *SequencerContractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SequencerContract.Contract.SequencerContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SequencerContract *SequencerContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SequencerContract.Contract.SequencerContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SequencerContract *SequencerContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SequencerContract.Contract.SequencerContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SequencerContract *SequencerContractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SequencerContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SequencerContract *SequencerContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SequencerContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SequencerContract *SequencerContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SequencerContract.Contract.contract.Transact(opts, method, params...)
}

// SubmitDecryptionProgress is a paid mutator transaction binding the contract method 0x2d32522e.
//
// Solidity: function submitDecryptionProgress(bytes message) returns()
func (_SequencerContract *SequencerContractTransactor) SubmitDecryptionProgress(opts *bind.TransactOpts, message []byte) (*types.Transaction, error) {
	return _SequencerContract.contract.Transact(opts, "submitDecryptionProgress", message)
}

// SubmitDecryptionProgress is a paid mutator transaction binding the contract method 0x2d32522e.
//
// Solidity: function submitDecryptionProgress(bytes message) returns()
func (_SequencerContract *SequencerContractSession) SubmitDecryptionProgress(message []byte) (*types.Transaction, error) {
	return _SequencerContract.Contract.SubmitDecryptionProgress(&_SequencerContract.TransactOpts, message)
}

// SubmitDecryptionProgress is a paid mutator transaction binding the contract method 0x2d32522e.
//
// Solidity: function submitDecryptionProgress(bytes message) returns()
func (_SequencerContract *SequencerContractTransactorSession) SubmitDecryptionProgress(message []byte) (*types.Transaction, error) {
	return _SequencerContract.Contract.SubmitDecryptionProgress(&_SequencerContract.TransactOpts, message)
}

// SubmitEncryptedTransaction is a paid mutator transaction binding the contract method 0x6a69d2e1.
//
// Solidity: function submitEncryptedTransaction(uint64 eon, bytes32 identityPrefix, bytes encryptedTransaction, uint256 gasLimit) payable returns()
func (_SequencerContract *SequencerContractTransactor) SubmitEncryptedTransaction(opts *bind.TransactOpts, eon uint64, identityPrefix [32]byte, encryptedTransaction []byte, gasLimit *big.Int) (*types.Transaction, error) {
	return _SequencerContract.contract.Transact(opts, "submitEncryptedTransaction", eon, identityPrefix, encryptedTransaction, gasLimit)
}

// SubmitEncryptedTransaction is a paid mutator transaction binding the contract method 0x6a69d2e1.
//
// Solidity: function submitEncryptedTransaction(uint64 eon, bytes32 identityPrefix, bytes encryptedTransaction, uint256 gasLimit) payable returns()
func (_SequencerContract *SequencerContractSession) SubmitEncryptedTransaction(eon uint64, identityPrefix [32]byte, encryptedTransaction []byte, gasLimit *big.Int) (*types.Transaction, error) {
	return _SequencerContract.Contract.SubmitEncryptedTransaction(&_SequencerContract.TransactOpts, eon, identityPrefix, encryptedTransaction, gasLimit)
}

// SubmitEncryptedTransaction is a paid mutator transaction binding the contract method 0x6a69d2e1.
//
// Solidity: function submitEncryptedTransaction(uint64 eon, bytes32 identityPrefix, bytes encryptedTransaction, uint256 gasLimit) payable returns()
func (_SequencerContract *SequencerContractTransactorSession) SubmitEncryptedTransaction(eon uint64, identityPrefix [32]byte, encryptedTransaction []byte, gasLimit *big.Int) (*types.Transaction, error) {
	return _SequencerContract.Contract.SubmitEncryptedTransaction(&_SequencerContract.TransactOpts, eon, identityPrefix, encryptedTransaction, gasLimit)
}

// SequencerContractDecryptionProgressSubmittedIterator is returned from FilterDecryptionProgressSubmitted and is used to iterate over the raw logs and unpacked data for DecryptionProgressSubmitted events raised by the SequencerContract contract.
type SequencerContractDecryptionProgressSubmittedIterator struct {
	Event *SequencerContractDecryptionProgressSubmitted // Event containing the contract specifics and raw log

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
func (it *SequencerContractDecryptionProgressSubmittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SequencerContractDecryptionProgressSubmitted)
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
		it.Event = new(SequencerContractDecryptionProgressSubmitted)
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
func (it *SequencerContractDecryptionProgressSubmittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SequencerContractDecryptionProgressSubmittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SequencerContractDecryptionProgressSubmitted represents a DecryptionProgressSubmitted event raised by the SequencerContract contract.
type SequencerContractDecryptionProgressSubmitted struct {
	Message []byte
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterDecryptionProgressSubmitted is a free log retrieval operation binding the contract event 0xa9a0645b33a70f18b8d490681d637cb46a859ec51707787e6f46b942f90e8f59.
//
// Solidity: event DecryptionProgressSubmitted(bytes message)
func (_SequencerContract *SequencerContractFilterer) FilterDecryptionProgressSubmitted(opts *bind.FilterOpts) (*SequencerContractDecryptionProgressSubmittedIterator, error) {

	logs, sub, err := _SequencerContract.contract.FilterLogs(opts, "DecryptionProgressSubmitted")
	if err != nil {
		return nil, err
	}
	return &SequencerContractDecryptionProgressSubmittedIterator{contract: _SequencerContract.contract, event: "DecryptionProgressSubmitted", logs: logs, sub: sub}, nil
}

// WatchDecryptionProgressSubmitted is a free log subscription operation binding the contract event 0xa9a0645b33a70f18b8d490681d637cb46a859ec51707787e6f46b942f90e8f59.
//
// Solidity: event DecryptionProgressSubmitted(bytes message)
func (_SequencerContract *SequencerContractFilterer) WatchDecryptionProgressSubmitted(opts *bind.WatchOpts, sink chan<- *SequencerContractDecryptionProgressSubmitted) (event.Subscription, error) {

	logs, sub, err := _SequencerContract.contract.WatchLogs(opts, "DecryptionProgressSubmitted")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SequencerContractDecryptionProgressSubmitted)
				if err := _SequencerContract.contract.UnpackLog(event, "DecryptionProgressSubmitted", log); err != nil {
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

// ParseDecryptionProgressSubmitted is a log parse operation binding the contract event 0xa9a0645b33a70f18b8d490681d637cb46a859ec51707787e6f46b942f90e8f59.
//
// Solidity: event DecryptionProgressSubmitted(bytes message)
func (_SequencerContract *SequencerContractFilterer) ParseDecryptionProgressSubmitted(log types.Log) (*SequencerContractDecryptionProgressSubmitted, error) {
	event := new(SequencerContractDecryptionProgressSubmitted)
	if err := _SequencerContract.contract.UnpackLog(event, "DecryptionProgressSubmitted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SequencerContractTransactionSubmittedIterator is returned from FilterTransactionSubmitted and is used to iterate over the raw logs and unpacked data for TransactionSubmitted events raised by the SequencerContract contract.
type SequencerContractTransactionSubmittedIterator struct {
	Event *SequencerContractTransactionSubmitted // Event containing the contract specifics and raw log

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
func (it *SequencerContractTransactionSubmittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SequencerContractTransactionSubmitted)
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
		it.Event = new(SequencerContractTransactionSubmitted)
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
func (it *SequencerContractTransactionSubmittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SequencerContractTransactionSubmittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SequencerContractTransactionSubmitted represents a TransactionSubmitted event raised by the SequencerContract contract.
type SequencerContractTransactionSubmitted struct {
	Eon                  uint64
	IdentityPrefix       [32]byte
	Sender               common.Address
	EncryptedTransaction []byte
	GasLimit             *big.Int
	Raw                  types.Log // Blockchain specific contextual infos
}

// FilterTransactionSubmitted is a free log retrieval operation binding the contract event 0x6515f8e10d22a184f86cfbaeb024db7afde82add43a1b1c065e8d202e43ef1a0.
//
// Solidity: event TransactionSubmitted(uint64 eon, bytes32 identityPrefix, address sender, bytes encryptedTransaction, uint256 gasLimit)
func (_SequencerContract *SequencerContractFilterer) FilterTransactionSubmitted(opts *bind.FilterOpts) (*SequencerContractTransactionSubmittedIterator, error) {

	logs, sub, err := _SequencerContract.contract.FilterLogs(opts, "TransactionSubmitted")
	if err != nil {
		return nil, err
	}
	return &SequencerContractTransactionSubmittedIterator{contract: _SequencerContract.contract, event: "TransactionSubmitted", logs: logs, sub: sub}, nil
}

// WatchTransactionSubmitted is a free log subscription operation binding the contract event 0x6515f8e10d22a184f86cfbaeb024db7afde82add43a1b1c065e8d202e43ef1a0.
//
// Solidity: event TransactionSubmitted(uint64 eon, bytes32 identityPrefix, address sender, bytes encryptedTransaction, uint256 gasLimit)
func (_SequencerContract *SequencerContractFilterer) WatchTransactionSubmitted(opts *bind.WatchOpts, sink chan<- *SequencerContractTransactionSubmitted) (event.Subscription, error) {

	logs, sub, err := _SequencerContract.contract.WatchLogs(opts, "TransactionSubmitted")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SequencerContractTransactionSubmitted)
				if err := _SequencerContract.contract.UnpackLog(event, "TransactionSubmitted", log); err != nil {
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

// ParseTransactionSubmitted is a log parse operation binding the contract event 0x6515f8e10d22a184f86cfbaeb024db7afde82add43a1b1c065e8d202e43ef1a0.
//
// Solidity: event TransactionSubmitted(uint64 eon, bytes32 identityPrefix, address sender, bytes encryptedTransaction, uint256 gasLimit)
func (_SequencerContract *SequencerContractFilterer) ParseTransactionSubmitted(log types.Log) (*SequencerContractTransactionSubmitted, error) {
	event := new(SequencerContractTransactionSubmitted)
	if err := _SequencerContract.contract.UnpackLog(event, "TransactionSubmitted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
