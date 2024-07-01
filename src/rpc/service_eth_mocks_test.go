package rpc_test

import (
	"context"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shutter-network/encrypting-rpc-server/rpc"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/url"
	"github.com/stretchr/testify/mock"
	"math/big"
)

type MockEthereumClient struct {
	mock.Mock
}

type MockKeyperSetManagerContract struct {
	mock.Mock
}

type MockKeyBroadcastContract struct {
	mock.Mock
}

type MockSequencerContract struct {
	mock.Mock
}

func MockConfig() rpc.Config {
	return rpc.Config{
		BackendURL:        &url.URL{},
		HTTPListenAddress: ":8546",
		DelayInSeconds:    10,
	}
}

var mockProcessTransactionCallCount int

func mockProcessTransaction(tx *types.Transaction, ctx context.Context, service *rpc.EthService, blockNumber uint64, b []byte) (*types.Transaction, error) {
	mockProcessTransactionCallCount++
	return tx, nil
}

func mockWaitMined(ctx context.Context, client bind.DeployBackend, tx *types.Transaction) (*types.Receipt, error) {
	receipt := &types.Receipt{
		Status: types.ReceiptStatusSuccessful,
		TxHash: tx.Hash(),
	}
	return receipt, nil
}

func (m *MockEthereumClient) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	args := m.Called(ctx, account)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockEthereumClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockEthereumClient) ChainID(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockEthereumClient) BlockNumber(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockEthereumClient) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	args := m.Called(ctx, tx)
	return args.Error(0)
}

func (m *MockEthereumClient) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	args := m.Called(ctx, txHash)
	return args.Get(0).(*types.Receipt), args.Error(1)
}

func (m *MockEthereumClient) CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error) {
	args := m.Called(ctx, account, blockNumber)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockEthereumClient) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
	args := m.Called(ctx, account, blockNumber)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockEthereumClient) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	args := m.Called(ctx, account, blockNumber)
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockKeyperSetManagerContract) GetKeyperSetIndexByBlock(opts *bind.CallOpts, blockNumber uint64) (uint64, error) {
	args := m.Called(opts, blockNumber)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockKeyBroadcastContract) GetEonKey(opts *bind.CallOpts, eon uint64) ([]byte, error) {
	args := m.Called(opts, eon)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockSequencerContract) SubmitEncryptedTransaction(opts *bind.TransactOpts, eon uint64, identityPrefix [32]byte, encryptedTx []byte, gasLimit *big.Int) (*types.Transaction, error) {
	args := m.Called(opts, eon, identityPrefix, encryptedTx, gasLimit)
	return args.Get(0).(*types.Transaction), args.Error(1)
}
