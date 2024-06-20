package rpc_test

import (
	"context"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shutter-network/encrypting-rpc-server/cache"
	"github.com/shutter-network/encrypting-rpc-server/rpc"
	test_data "github.com/shutter-network/encrypting-rpc-server/test-data"
	"github.com/shutter-network/encrypting-rpc-server/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"math/big"
	"testing"
)

// First transaction gets sent and cache gets updated
func TestSendRawTransaction_Success(t *testing.T) {
	mockClient := new(MockEthereumClient)
	mockKeyperSetManager := new(MockKeyperSetManagerContract)
	mockKeyBroadcast := new(MockKeyBroadcastContract)
	mockSequencer := new(MockSequencerContract)

	privateKey, fromAddress, err := test_data.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	config := test_data.MockConfig()

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
	}

	nonce := uint64(1)
	gasPrice := big.NewInt(2000000000)
	chainID := big.NewInt(1)
	blockNumber := uint64(1)
	accountBalance := big.NewInt(1000000000000000000)

	mockClient.On("PendingNonceAt", mock.Anything, fromAddress).Return(nonce, nil)
	mockClient.On("SuggestGasPrice", mock.Anything).Return(gasPrice, nil)
	mockClient.On("ChainID", mock.Anything).Return(chainID, nil)
	mockClient.On("BlockNumber", mock.Anything).Return(blockNumber, nil)
	mockClient.On("NonceAt", mock.Anything, fromAddress, (*big.Int)(nil)).Return(nonce, nil)
	mockClient.On("BalanceAt", mock.Anything, fromAddress, (*big.Int)(nil)).Return(accountBalance, nil)

	rawTx1, signedTx, err := test_data.Tx(privateKey, nonce, chainID)
	if err != nil {
		t.Fatalf("Failed to create signed transaction: %v", err)
	}

	receipt := &types.Receipt{
		Status: types.ReceiptStatusSuccessful,
		TxHash: signedTx.Hash(),
	}

	mockClient.On("WaitMined", mock.Anything, mock.Anything).Return(receipt, nil)
	mockClient.On("TransactionReceipt", mock.Anything, mock.Anything).Return(receipt, nil)

	// Send the transaction
	txHash, err := service.SendRawTransaction(context.Background(), rawTx1)
	utils.CheckErr(t, err, "Failed to send raw transaction 1")

	assert.NotNil(t, txHash)
	assert.Equal(t, signedTx.Hash().Hex(), txHash.Hex())

	// todo check cache
}
