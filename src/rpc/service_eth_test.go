package rpc_test

import (
	"context"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shutter-network/encrypting-rpc-server/cache"
	"github.com/shutter-network/encrypting-rpc-server/rpc"
	testdata "github.com/shutter-network/encrypting-rpc-server/test-data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"math/big"
	"testing"
)

func setupTest(t *testing.T) (*rpc.EthService, *MockEthereumClient) {
	mockClient := new(MockEthereumClient)
	mockKeyperSetManager := new(MockKeyperSetManagerContract)
	mockKeyBroadcast := new(MockKeyBroadcastContract)
	mockSequencer := new(MockSequencerContract)

	privateKey, fromAddress, err := testdata.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	config := MockConfig()

	service := &rpc.EthService{
		Processor: rpc.Processor{
			Client:                   mockClient,
			SigningKey:               privateKey,
			SigningAddress:           &fromAddress,
			KeyBroadcastContract:     mockKeyBroadcast,
			SequencerContract:        mockSequencer,
			KeyperSetManagerContract: mockKeyperSetManager,
		},
		Cache:              cache.NewCache(10),
		Config:             config,
		ProcessTransaction: mockProcessTransaction,
		WaitMinedFunc:      mockWaitMined,
	}

	nonce := uint64(1)
	chainID := big.NewInt(1)
	blockNumber := uint64(1)
	accountBalance := big.NewInt(1000000000000000000)

	mockClient.On("PendingNonceAt", mock.Anything, fromAddress).Return(nonce, nil)
	mockClient.On("ChainID", mock.Anything).Return(chainID, nil)
	mockClient.On("BlockNumber", mock.Anything).Return(blockNumber, nil)
	mockClient.On("NonceAt", mock.Anything, fromAddress, (*big.Int)(nil)).Return(nonce, nil)
	mockClient.On("BalanceAt", mock.Anything, fromAddress, (*big.Int)(nil)).Return(accountBalance, nil)

	return service, mockClient
}

func assertDynamicTxEquality(t *testing.T, cachedTxInfo cache.TransactionInfo, signedTx *types.Transaction) {
	assert.Equal(t, cachedTxInfo.Tx.Nonce(), signedTx.Nonce(), "Expected transaction nonce does not match")
	assert.Equal(t, cachedTxInfo.Tx.Hash(), signedTx.Hash(), "Expected transaction hash does not match")
	assert.Equal(t, cachedTxInfo.Tx.To(), signedTx.To(), "Expected transaction to address does not match")
	assert.Equal(t, cachedTxInfo.Tx.Value(), signedTx.Value(), "Expected transaction value does not match")
	assert.Equal(t, cachedTxInfo.Tx.Gas(), signedTx.Gas(), "Expected transaction gas does not match")
	assert.Equal(t, cachedTxInfo.Tx.GasFeeCap(), signedTx.GasFeeCap(), "Expected transaction max priority fee per gas does not match")
	assert.Equal(t, cachedTxInfo.Tx.GasTipCap(), signedTx.GasTipCap(), "Expected transaction max fee per gas does not match")

	// Handle empty and nil slices equivalently
	if (cachedTxInfo.Tx.Data() == nil && signedTx.Data() == nil) || (len(cachedTxInfo.Tx.Data()) == 0 && len(signedTx.Data()) == 0) {
		assert.True(t, true)
	} else {
		assert.Equal(t, cachedTxInfo.Tx.Data(), signedTx.Data(), "Expected transaction data does not match")
	}

	assert.Equal(t, cachedTxInfo.Tx.ChainId(), signedTx.ChainId(), "Expected transaction chain ID does not match")

}

// First transaction gets sent and cache gets updated
func TestSendRawTransaction_Success(t *testing.T) {
	service, mockClient := setupTest(t)
	rawTx1, signedTx, err := testdata.Tx(service.Processor.SigningKey, 1, big.NewInt(1))
	assert.NoError(t, err, "Failed to create signed transaction")

	receipt := &types.Receipt{
		Status: types.ReceiptStatusSuccessful,
		TxHash: signedTx.Hash(),
	}

	mockClient.On("TransactionReceipt", mock.Anything, mock.Anything).Return(receipt, nil)

	// Send the transaction
	txHash, err := service.SendRawTransaction(context.Background(), rawTx1)
	assert.NoError(t, err, "Failed to send raw transaction")
	assert.NotNil(t, txHash)
	assert.Equal(t, signedTx.Hash().Hex(), txHash.Hex())

	// Cache key exists
	key, err := service.Cache.Key(signedTx)
	assert.NoError(t, err, "Failed to get key from cache")

	// Cache entry updated correctly
	cachedTxInfo, exists := service.Cache.Data[key]
	assert.True(t, exists, "Expected transaction information to be in the cache")
	assert.Equal(t, cachedTxInfo.SendingBlock, uint64(1), "Expected sending block does not match")
	assertDynamicTxEquality(t, cachedTxInfo, signedTx)
}

// First tx sent and resending delayed
func TestSendRawTransaction_SameNonce_SameGasPrice_Delayed(t *testing.T) {
	service, _ := setupTest(t)
	nonce := uint64(1)
	chainID := big.NewInt(1)

	// First transaction with error
	rawTx1, signedTx1, _ := testdata.Tx(service.Processor.SigningKey, nonce, chainID)
	_, err := service.SendRawTransaction(context.Background(), rawTx1)
	assert.NoError(t, err, "Expected transaction sending to succeed")

	// Send the second transaction
	rawTx2, signedTx2, _ := testdata.Tx(service.Processor.SigningKey, nonce, chainID)
	_, err = service.SendRawTransaction(context.Background(), rawTx2)
	assert.NoError(t, err, "Expected transaction sending to succeed")

	// Check that the key is stored with the current block number
	key, err := service.Cache.Key(signedTx2)
	assert.NoError(t, err, "Expected cache to have key")

	// Transaction resending is delayed
	cachedTxInfo, exists := service.Cache.Data[key]
	assert.True(t, exists, "Expected transaction information to be in the cache")
	assert.Equal(t, cachedTxInfo.SendingBlock, uint64(1)+10, "Expected sending block does not match")
	assertDynamicTxEquality(t, cachedTxInfo, signedTx1)

	// Only first transaction gets encrypted
	assert.Equal(t, mockProcessTransactionCallCount, 1, "Expected ProcessTransaction to be called once")
}

// First transaction gets sent, second tx with higher gas price gets delayed
func TestSendRawTransaction_SameNonce_HigherGasPrice_Delayed(t *testing.T) {
	service, _ := setupTest(t)
	nonce := uint64(1)
	chainID := big.NewInt(1)

	// First transaction with error
	rawTx1, signedTx1, _ := testdata.Tx(service.Processor.SigningKey, nonce, chainID)
	_, err := service.SendRawTransaction(context.Background(), rawTx1)
	assert.NoError(t, err, "Expected transaction sending to succeed")

	// Send the second transaction
	twiceGasPrice := new(big.Int).Mul(signedTx1.GasPrice(), big.NewInt(2))
	rawTx2, signedTx2, _ := testdata.TxWithGas(service.Processor.SigningKey, nonce, chainID, twiceGasPrice)

	_, err = service.SendRawTransaction(context.Background(), rawTx2)
	assert.NoError(t, err, "Expected transaction sending to succeed")

	// Check that the key is stored with the current block number
	key, err := service.Cache.Key(signedTx2)
	assert.NoError(t, err, "Expected cache to have key")

	// Transaction with higher gas price is stored and delayed
	cachedTxInfo, exists := service.Cache.Data[key]
	assert.True(t, exists, "Expected transaction information to be in the cache")
	assert.Equal(t, cachedTxInfo.SendingBlock, uint64(1)+10, "Expected sending block does not match")
	assertDynamicTxEquality(t, cachedTxInfo, signedTx2)

	// Only first transaction gets encrypted
	assert.Equal(t, mockProcessTransactionCallCount, 1, "Expected ProcessTransaction to be called once")
}
