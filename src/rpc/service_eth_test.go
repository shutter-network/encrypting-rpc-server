package rpc_test

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shutter-network/encrypting-rpc-server/cache"
	"github.com/shutter-network/encrypting-rpc-server/rpc"
	"github.com/shutter-network/encrypting-rpc-server/test"
	"github.com/shutter-network/encrypting-rpc-server/testdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func initTest(t *testing.T) (*rpc.EthService, sqlmock.Sqlmock) {
	mockClient := new(MockEthereumClient)
	mockKeyperSetManager := new(MockKeyperSetManagerContract)
	mockKeyBroadcast := new(MockKeyBroadcastContract)
	mockSequencer := new(MockSequencerContract)

	privateKey, fromAddress, err := testdata.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	config := MockConfig()

	mockDb, db := test.NewPostgresTestDB(t)

	service := &rpc.EthService{
		Processor: rpc.Processor{
			Client:                   mockClient,
			SigningKey:               privateKey,
			SigningAddress:           &fromAddress,
			KeyBroadcastContract:     mockKeyBroadcast,
			SequencerContract:        mockSequencer,
			KeyperSetManagerContract: mockKeyperSetManager,
			Db:                       db,
		},
		Cache:              cache.NewCache(10),
		Config:             config,
		ProcessTransaction: mockProcessTransaction,
	}

	nonce := uint64(1)
	chainID := big.NewInt(1)
	blockNumber := uint64(1)
	accountBalance := big.NewInt(1000000000000000000)

	txReciept := types.Receipt{
		Status:    0,
		BlockHash: common.Hash{},
	}

	block := types.NewBlock(&types.Header{
		Time: uint64(time.Now().Unix()),
	}, nil, nil, nil)

	mockClient.On("PendingNonceAt", mock.Anything, fromAddress).Return(nonce, nil)
	mockClient.On("ChainID", mock.Anything).Return(chainID, nil)
	mockClient.On("BlockNumber", mock.Anything).Return(blockNumber, nil)
	mockClient.On("NonceAt", mock.Anything, fromAddress, (*big.Int)(nil)).Return(nonce, nil)
	mockClient.On("BalanceAt", mock.Anything, fromAddress, (*big.Int)(nil)).Return(accountBalance, nil)
	mockClient.On("TransactionReceipt", mock.Anything, mock.Anything).Return(&txReciept, nil)
	mockClient.On("BlockByHash", mock.Anything, mock.Anything).Return(block, nil)

	// reset counter at init
	mockProcessTransactionCallCount = 0

	return service, mockDb
}

func assertDynamicTxEquality(t *testing.T, cachedTx *types.Transaction, signedTx *types.Transaction) {
	assert.Equal(t, cachedTx.Nonce(), signedTx.Nonce(), "Expected transaction nonce does not match")
	assert.Equal(t, cachedTx.Hash(), signedTx.Hash(), "Expected transaction hash does not match")
	assert.Equal(t, cachedTx.To(), signedTx.To(), "Expected transaction to address does not match")
	assert.Equal(t, cachedTx.Value(), signedTx.Value(), "Expected transaction value does not match")
	assert.Equal(t, cachedTx.Gas(), signedTx.Gas(), "Expected transaction gas does not match")
	assert.Equal(t, cachedTx.GasFeeCap(), signedTx.GasFeeCap(), "Expected transaction max priority fee per gas does not match")
	assert.Equal(t, cachedTx.GasTipCap(), signedTx.GasTipCap(), "Expected transaction max fee per gas does not match")

	// Handle empty and nil slices equivalently
	if (cachedTx.Data() == nil && signedTx.Data() == nil) || (len(cachedTx.Data()) == 0 && len(signedTx.Data()) == 0) {
		assert.True(t, true)
	} else {
		assert.Equal(t, cachedTx.Data(), signedTx.Data(), "Expected transaction data does not match")
	}

	assert.Equal(t, cachedTx.ChainId(), signedTx.ChainId(), "Expected transaction chain ID does not match")

}

func TestGasPrice(t *testing.T) {
	service, _ := initTest(t)
	expectedGasPrice := big.NewInt(20000000000)
	expectedHexGasPrice := "0x9502f9000"

	service.Processor.Client.(*MockEthereumClient).On("SuggestGasPrice", mock.Anything).Return(expectedGasPrice, nil)

	actualGasPrice, err := service.GasPrice(context.Background())

	assert.NoError(t, err, "GasPrice should not return an error")
	assert.Equal(t, expectedHexGasPrice, actualGasPrice, "GasPrice should return the expected hex value")
	service.Processor.Client.(*MockEthereumClient).AssertCalled(t, "SuggestGasPrice", mock.Anything)
}

// First transaction gets sent and cache gets updated
func TestSendRawTransaction_Success(t *testing.T) {
	service, _ := initTest(t)
	rawTx1, signedTx, err := testdata.Tx(service.Processor.SigningKey, 1, big.NewInt(1))
	assert.NoError(t, err, "Failed to create signed transaction")

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
	assert.False(t, cachedTxInfo.Delayed, "The tx was tried only once")
	assert.Equal(t, time.Now().Unix(), cachedTxInfo.CachedTime, "Expected sending block does not match")
	assertDynamicTxEquality(t, signedTx, cachedTxInfo.Tx)
}

func TestSendRawTransaction_TransactionInvalidNonce_NotSent(t *testing.T) {
	service, _ := initTest(t)

	wrongNonce := uint64(0)
	chainID := big.NewInt(1)
	rawTx, _, err := testdata.Tx(service.Processor.SigningKey, wrongNonce, chainID)
	assert.NoError(t, err, "Failed to create signed transaction")

	_, err = service.SendRawTransaction(context.Background(), rawTx)
	assert.Error(t, err, "Expected the SendRawTransaction function to return an error")

	encodingErr, ok := err.(*rpc.EncodingError)
	assert.True(t, ok, "Expected error of type *EncodingError")
	assert.Equal(t, encodingErr.StatusCode, -32000, "Expected specific status code for invalid nonce error")
}

func TestSendRawTransaction_TransactionInvalid_GasCost_Higher(t *testing.T) {
	service, _ := initTest(t)

	highCost := new(big.Int).Mul(big.NewInt(10), big.NewInt(1e18))
	rawTx, _, err := testdata.TxWithGasPrice(service.Processor.SigningKey, 1, big.NewInt(1), highCost)
	assert.NoError(t, err, "Failed to create signed transaction")

	_, err = service.SendRawTransaction(context.Background(), rawTx)
	assert.Error(t, err, "Expected the SendRawTransaction function to return an error")

	encodingErr, ok := err.(*rpc.EncodingError)
	assert.True(t, ok, "Expected error of type *EncodingError")
	assert.Equal(t, encodingErr.StatusCode, -32000, "Expected specific status code for high gas cost error")
}

// First tx sent and resending delayed
func TestSendRawTransaction_SameNonce_SameGasPrice_Delayed(t *testing.T) {
	service, _ := initTest(t)
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
	assert.Equal(t, time.Now().Unix(), cachedTxInfo.CachedTime, "Expected sending block does not match")
	assertDynamicTxEquality(t, cachedTxInfo.Tx, signedTx1)
	assert.True(t, cachedTxInfo.Delayed, "Expected added in retry queue")

	// Only first transaction gets encrypted
	assert.Equal(t, mockProcessTransactionCallCount, 1, "Expected ProcessTransaction to be called once")
}

// First transaction gets sent, second tx with higher gas price gets delayed
func TestSendRawTransaction_SameNonce_HigherGasPrice_Delayed(t *testing.T) {
	service, _ := initTest(t)
	nonce := uint64(1)
	chainID := big.NewInt(1)

	// First transaction with error
	rawTx1, signedTx1, _ := testdata.Tx(service.Processor.SigningKey, nonce, chainID)
	_, err := service.SendRawTransaction(context.Background(), rawTx1)
	assert.NoError(t, err, "Expected transaction sending to succeed")

	// Send the second transaction
	twiceGasPrice := new(big.Int).Mul(signedTx1.GasPrice(), big.NewInt(2))
	rawTx2, signedTx2, _ := testdata.TxWithGasPrice(service.Processor.SigningKey, nonce, chainID, twiceGasPrice)

	_, err = service.SendRawTransaction(context.Background(), rawTx2)
	assert.NoError(t, err, "Expected transaction sending to succeed")

	// Check that the key is stored with the current block number
	key, err := service.Cache.Key(signedTx2)
	assert.NoError(t, err, "Expected cache to have key")

	// Transaction with higher gas price is stored and delayed
	cachedTxInfo, exists := service.Cache.Data[key]
	assert.True(t, exists, "Expected transaction information to be in the cache")
	assert.Equal(t, time.Now().Unix(), cachedTxInfo.CachedTime, "Expected sending block does not match")
	assertDynamicTxEquality(t, cachedTxInfo.Tx, signedTx2)
	assert.True(t, cachedTxInfo.Delayed, "Expected added in retry queue")

	// Only first transaction gets encrypted
	assert.Equal(t, mockProcessTransactionCallCount, 1, "Expected ProcessTransaction to be called once")
}

func TestNewTimeEvent_UpdateTxInfo(t *testing.T) {
	service, _ := initTest(t)
	currentTime := time.Now().Unix()
	chainID := big.NewInt(1)

	_, signedTx, _ := testdata.Tx(service.Processor.SigningKey, 1, chainID)

	key, err := service.Cache.Key(signedTx)
	if err != nil {
		t.Fatalf("Failed to create key: %v", err)
	}

	service.Cache.Data[key] = cache.TransactionInfo{Tx: signedTx, CachedTime: 1, Delayed: true}

	service.NewTimeEvent(context.Background(), currentTime)

	info, exists := service.Cache.Data[key]

	assert.True(t, exists, "Expected transaction information to be in the cache")
	assert.Equal(t, currentTime, info.CachedTime, "Expected cached time to be updated")
	assertDynamicTxEquality(t, signedTx, info.Tx)
	assert.False(t, info.Delayed, "Expected tx to be not tried again")
	assert.Equal(t, mockProcessTransactionCallCount, 1, "Expected ProcessTransaction to be called once")
}

func TestNewTimeEvent_KeepTxInfo(t *testing.T) {
	service, _ := initTest(t)
	currentTime := int64(13)
	chainID := big.NewInt(1)

	_, signedTx, _ := testdata.Tx(service.Processor.SigningKey, 1, chainID)

	key, err := service.Cache.Key(signedTx)
	if err != nil {
		t.Fatalf("Failed to create key: %v", err)
	}

	service.Cache.Data[key] = cache.TransactionInfo{Tx: signedTx, CachedTime: 4, Delayed: true}

	service.NewTimeEvent(context.Background(), currentTime)

	info, exists := service.Cache.Data[key]

	assert.True(t, exists, "Expected transaction information to be in the cache")
	assert.Equal(t, info.CachedTime, int64(4), "Expected cached time to remain unchanged")
	assertDynamicTxEquality(t, info.Tx, signedTx)
	assert.Equal(t, mockProcessTransactionCallCount, 0, "Expected ProcessTransaction to not be called")
}

func TestNewTimeEvent_DeleteTxInfo(t *testing.T) {
	service, _ := initTest(t)
	currentTime := int64(13)
	chainID := big.NewInt(1)

	_, signedTx, _ := testdata.Tx(service.Processor.SigningKey, 1, chainID)

	key, err := service.Cache.Key(signedTx)
	if err != nil {
		t.Fatalf("Failed to create key: %v", err)
	}

	service.Cache.Data[key] = cache.TransactionInfo{Tx: signedTx, CachedTime: 3, Delayed: false}

	service.NewTimeEvent(context.Background(), currentTime)

	_, exists := service.Cache.Data[key]

	assert.False(t, exists, "Expected transaction information to be deleted from the cache")
	assert.Equal(t, mockProcessTransactionCallCount, 0, "Expected ProcessTransaction to not be called")
}

func TestSendRawTransaction_GasLimitExceedsChainLimit_Error(t *testing.T) {
	service, _ := initTest(t)
	highGasLimit := service.Config.EncryptedGasLimit + 1
	nonce := uint64(1)
	chainID := big.NewInt(1)

	rawTx, _, err := testdata.TxWithGas(service.Processor.SigningKey, nonce, chainID, big.NewInt(2000000000), highGasLimit, big.NewInt(2000000000))
	assert.NoError(t, err, "Failed to create signed transaction")

	// Send the transaction
	_, err = service.SendRawTransaction(context.Background(), rawTx)
	// Expect an error here because gas limit exceeds the chain limit
	assert.Error(t, err, "Expected the SendRawTransaction function to return an error")

	encodingErr, ok := err.(*rpc.EncodingError)
	assert.True(t, ok, "Expected error of type *EncodingError")
	assert.Equal(t, encodingErr.StatusCode, -32000, "Expected specific status code for gas limit error")
	assert.Equal(t, encodingErr.Err.Error(), "gas limit exceeds encrypted gas limit (max gas limit allowed per shutterized block)")
}

func TestSendRawTransaction_IntrinsicGas_Error(t *testing.T) {
	service, _ := initTest(t)
	gasLimit := uint64(20000)
	nonce := uint64(1)
	chainID := big.NewInt(1)

	rawTx, _, err := testdata.TxWithGas(service.Processor.SigningKey, nonce, chainID, big.NewInt(2000000000), gasLimit, big.NewInt(2000000000))
	assert.NoError(t, err, "Failed to create signed transaction")

	// Send the transaction
	_, err = service.SendRawTransaction(context.Background(), rawTx)
	// Expect an error here because gas limit exceeds the chain limit
	assert.Error(t, err, "Expected the SendRawTransaction function to return an error")

	encodingErr, ok := err.(*rpc.EncodingError)
	assert.True(t, ok, "Expected error of type *EncodingError")
	assert.Equal(t, encodingErr.StatusCode, -32602, "Expected specific status code for gas limit error")
	assert.Equal(t, encodingErr.Err.Error(), "gas limit below the intrinsic gas limit 21000")
}

func TestSendRawTransaction_PriorityGas_Error(t *testing.T) {
	service, _ := initTest(t)
	gasLimit := uint64(30000)
	nonce := uint64(1)
	chainID := big.NewInt(1)

	rawTx, _, err := testdata.TxWithGas(service.Processor.SigningKey, nonce, chainID, big.NewInt(2000000000), gasLimit, big.NewInt(0))
	assert.NoError(t, err, "Failed to create signed transaction")

	txHash, err := service.SendRawTransaction(context.Background(), rawTx)
	assert.Error(t, err, "Failed to send raw transaction")
	assert.Nil(t, txHash)

	encodingErr, ok := err.(*rpc.EncodingError)
	assert.True(t, ok, "Expected error of type *EncodingError")
	assert.Equal(t, encodingErr.StatusCode, -32602, "Expected specific status code for invalid priority fee")
}