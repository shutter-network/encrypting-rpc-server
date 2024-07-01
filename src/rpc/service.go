package rpc

import (
	"context"
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/url"
	"math/big"
)

type Processor struct {
	URL                      string
	RPCUrl                   string
	SigningKey               *ecdsa.PrivateKey
	SigningAddress           *common.Address
	KeyperSetChangeLookAhead int
	Client                   EthereumClient
	KeyBroadcastContract     KeyBroadcastContract
	SequencerContract        SequencerContract
	KeyperSetManagerContract KeyperSetManagerContract
}

type Config struct {
	BackendURL        *url.URL
	HTTPListenAddress string
	DelayInSeconds    int
}

type RPCService interface {
	Name() string
	NewTimeEvent(ctx context.Context, newTime uint64)
	SendRawTransaction(ctx context.Context, s string) (*common.Hash, error)
	Init(processor Processor, config Config)
	SendTimeEvents(ctx context.Context, delayInSeconds int)
}

type EthereumClient interface {
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	ChainID(ctx context.Context) (*big.Int, error)
	BlockNumber(ctx context.Context) (uint64, error)
	SendTransaction(ctx context.Context, tx *types.Transaction) error
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error)
	NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error)
	BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
}

type KeyperSetManagerContract interface {
	GetKeyperSetIndexByBlock(opts *bind.CallOpts, blockNumber uint64) (uint64, error)
}

type KeyBroadcastContract interface {
	GetEonKey(opts *bind.CallOpts, eon uint64) ([]byte, error)
}

type SequencerContract interface {
	SubmitEncryptedTransaction(opts *bind.TransactOpts, eon uint64, identityPrefix [32]byte, encryptedTx []byte, gasLimit *big.Int) (*types.Transaction, error)
}

type EthClientWrapper struct {
	Client *ethclient.Client
}

func (w *EthClientWrapper) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	return w.Client.PendingNonceAt(ctx, account)
}

func (w *EthClientWrapper) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	return w.Client.SuggestGasPrice(ctx)
}

func (w *EthClientWrapper) ChainID(ctx context.Context) (*big.Int, error) {
	return w.Client.ChainID(ctx)
}

func (w *EthClientWrapper) BlockNumber(ctx context.Context) (uint64, error) {
	return w.Client.BlockNumber(ctx)
}

func (w *EthClientWrapper) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	return w.Client.SendTransaction(ctx, tx)
}

func (w *EthClientWrapper) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	return w.Client.TransactionReceipt(ctx, txHash)
}

func (w *EthClientWrapper) CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error) {
	return w.Client.CodeAt(ctx, account, blockNumber)
}

func (w *EthClientWrapper) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
	return w.Client.NonceAt(ctx, account, blockNumber)
}

func (w *EthClientWrapper) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	return w.Client.BalanceAt(ctx, account, blockNumber)
}
