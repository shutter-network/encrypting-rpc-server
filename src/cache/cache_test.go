package cache

import (
	"fmt"
	"math/big"
	"sync"
	"testing"

	"github.com/shutter-network/encrypting-rpc-server/testdata"

	"github.com/stretchr/testify/assert"
)

func TestCache_Key(t *testing.T) {
	privateKey, fromAddress, err := testdata.GenerateKeyPair()
	assert.NoError(t, err, "Failed to generate key pair")

	chainID := big.NewInt(1)
	nonce := uint64(1)
	_, signedTx, err := testdata.Tx(privateKey, nonce, chainID)
	assert.NoError(t, err, "Failed to create signed transaction")

	c := NewCache(10)

	key, err := c.Key(signedTx)
	assert.NoError(t, err, "Failed to get key from transaction")

	expectedKey := fmt.Sprintf("%s-%d", fromAddress.Hex(), nonce)
	assert.Equal(t, expectedKey, key, "Expected key does not match")
}

func TestCache_UpdateEntry(t *testing.T) {
	privateKey, _, err := testdata.GenerateKeyPair()
	assert.NoError(t, err, "Failed to generate key pair")

	chainID := big.NewInt(1)
	nonce := uint64(1)
	_, signedTx, err := testdata.Tx(privateKey, nonce, chainID)
	assert.NoError(t, err, "Failed to create signed transaction")

	c := NewCache(10)

	key := "testKey"
	cachedTime := int64(1234)

	c.UpdateEntry(key, signedTx, cachedTime, false)

	entry, exists := c.Data[key]
	assert.True(t, exists, "Key should exist in cache after UpdateEntry")
	assert.Equal(t, signedTx, entry.Tx, "Transaction should match the updated one")
	assert.Equal(t, cachedTime, entry.CachedTime, "CachedTime should match the updated one")
	assert.Equal(t, false, entry.Delayed, "CachedTime should match the updated one")
}

func TestCache_ProcessTxEntry(t *testing.T) {
	c := NewCache(10)

	privateKey, fromAddress, err := testdata.GenerateKeyPair()
	assert.NoError(t, err, "Failed to generate key pair")

	_, signedTx, err := testdata.Tx(privateKey, 1, big.NewInt(1))
	assert.NoError(t, err, "Failed to create signed transaction")

	sendStatus, err := c.ProcessTxEntry(signedTx, 100)
	assert.NoError(t, err, "Failed to update entry")
	assert.True(t, sendStatus.SendStatus, "Expected transaction send status to be true")

	key, err := c.Key(signedTx)
	assert.NoError(t, err, "Failed to get key from cache")

	expectedKey := fmt.Sprintf("%s-%d", fromAddress.Hex(), signedTx.Nonce())
	assert.Equal(t, expectedKey, key, "Expected key does not match")

	// Verify that the transaction in the cache matches the signed transaction
	cachedTxInfo, exists := c.Data[key]
	assert.True(t, exists, "Expected transaction to be in the cache")
	assert.Equal(t, false, cachedTxInfo.Delayed, "The tx was tried just once, so Delayed should be falso")
	assert.Equal(t, signedTx, cachedTxInfo.Tx, "Cached transaction should match the updated one")
}

func TestCache_ProcessTxEntryTwice_SameTx(t *testing.T) {
	c := NewCache(10)

	privateKey, fromAddress, err := testdata.GenerateKeyPair()
	assert.NoError(t, err, "Failed to generate key pair")

	_, signedTx, err := testdata.Tx(privateKey, 1, big.NewInt(1))
	assert.NoError(t, err, "Failed to create signed transaction")

	sendStatus, err := c.ProcessTxEntry(signedTx, 100)
	assert.NoError(t, err, "Failed to update entry")
	assert.True(t, sendStatus.SendStatus, "Expected transaction send status to be true")

	key, err := c.Key(signedTx)
	assert.NoError(t, err, "Failed to get key from cache")

	expectedKey := fmt.Sprintf("%s-%d", fromAddress.Hex(), signedTx.Nonce())
	assert.Equal(t, expectedKey, key, "Expected key does not match")

	// Verify that the transaction in the cache matches the signed transaction
	cachedTxInfo, exists := c.Data[key]
	assert.True(t, exists, "Expected transaction to be in the cache")
	assert.False(t, cachedTxInfo.Delayed, "The tx was tried just once, so Delayed should be falso")
	assert.Equal(t, signedTx, cachedTxInfo.Tx, "Cached transaction should match the updated one")

	sendStatus, err = c.ProcessTxEntry(signedTx, 100)
	assert.NoError(t, err, "Failed to update entry")
	assert.False(t, sendStatus.SendStatus, "Expected transaction send status to be false")
	assert.False(t, sendStatus.UpdateStatus, "Expected transaction update status to be false, because of no change in tx")

	cachedTxInfo, exists = c.Data[key]
	assert.True(t, exists, "Expected transaction to be in the cache")
	assert.True(t, cachedTxInfo.Delayed, "The tx was tried more than once, so Delayed should be true")
}

func TestCache_ProcessTxEntryTwice_GasPriceChange(t *testing.T) {
	c := NewCache(10)

	privateKey, fromAddress, err := testdata.GenerateKeyPair()
	assert.NoError(t, err, "Failed to generate key pair")

	_, signedTx, err := testdata.Tx(privateKey, 1, big.NewInt(1))
	assert.NoError(t, err, "Failed to create signed transaction")

	sendStatus, err := c.ProcessTxEntry(signedTx, 100)
	assert.NoError(t, err, "Failed to update entry")
	assert.True(t, sendStatus.SendStatus, "Expected transaction send status to be true")

	key, err := c.Key(signedTx)
	assert.NoError(t, err, "Failed to get key from cache")

	expectedKey := fmt.Sprintf("%s-%d", fromAddress.Hex(), signedTx.Nonce())
	assert.Equal(t, expectedKey, key, "Expected key does not match")

	// Verify that the transaction in the cache matches the signed transaction
	cachedTxInfo, exists := c.Data[key]
	assert.True(t, exists, "Expected transaction to be in the cache")
	assert.False(t, cachedTxInfo.Delayed, "The tx was tried just once, so Delayed should be falso")
	assert.Equal(t, signedTx, cachedTxInfo.Tx, "Cached transaction should match the updated one")

	_, newTx, err := testdata.TxWithGasPrice(privateKey, signedTx.Nonce(), signedTx.ChainId(), big.NewInt(3000000000))
	assert.NoError(t, err, "Failed to generate new tx")

	sendStatus, err = c.ProcessTxEntry(newTx, 100)
	assert.NoError(t, err, "Failed to update entry")
	assert.False(t, sendStatus.SendStatus, "Expected transaction send status to be false")
	assert.True(t, sendStatus.UpdateStatus, "Expected transaction update status to be true, because of change in gas price")

	cachedTxInfo, exists = c.Data[key]
	assert.True(t, exists, "Expected transaction to be in the cache")
	assert.True(t, cachedTxInfo.Delayed, "The tx was tried more than once, so Delayed should be true")
}

func TestCache_ConcurrentUpdateEntry(t *testing.T) {
	c := NewCache(10)
	var wg sync.WaitGroup
	wg.Add(2)

	chainID := big.NewInt(3)
	nonce := uint64(1)
	privateKey, _, err := testdata.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	go func() {
		defer wg.Done()
		_, newTx, err := testdata.Tx(privateKey, nonce, chainID)
		assert.Nil(t, err, "Error while creating transaction")
		currentBlock := int64(1)
		executed, err := c.ProcessTxEntry(newTx, currentBlock)
		assert.Nil(t, err)
		assert.True(t, executed.SendStatus)
	}()

	go func() {
		defer wg.Done()
		_, newTx, err := testdata.Tx(privateKey, nonce+1, chainID)
		assert.Nil(t, err, "Error while creating transaction")
		currentBlock := int64(2)
		executed, err := c.ProcessTxEntry(newTx, currentBlock)
		assert.Nil(t, err)
		assert.True(t, executed.SendStatus)
	}()
	wg.Wait()

	assert.Equal(t, 2, len(c.Data), "Expected cache to contain 2 entries")
}
