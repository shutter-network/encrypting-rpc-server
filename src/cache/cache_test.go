package cache

import (
	"fmt"
	"github.com/shutter-network/encrypting-rpc-server/testdata"
	"math/big"
	"sync"
	"testing"

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

	c.UpdateEntry(key, signedTx, cachedTime)

	entry, exists := c.Data[key]
	assert.True(t, exists, "Key should exist in cache after UpdateEntry")
	assert.Equal(t, signedTx, entry.Tx, "Transaction should match the updated one")
	assert.Equal(t, cachedTime, entry.CachedTime, "CachedTime should match the updated one")
}

func TestCache_ProcessTxEntry(t *testing.T) {
	c := NewCache(10)

	privateKey, fromAddress, err := testdata.GenerateKeyPair()
	assert.NoError(t, err, "Failed to generate key pair")

	_, signedTx, err := testdata.Tx(privateKey, 1, big.NewInt(1))
	assert.NoError(t, err, "Failed to create signed transaction")

	sendStatus, err := c.ProcessTxEntry(signedTx, 100)
	assert.NoError(t, err, "Failed to update entry")
	assert.True(t, sendStatus, "Expected transaction to be added to cache")

	key, err := c.Key(signedTx)
	assert.NoError(t, err, "Failed to get key from cache")

	expectedKey := fmt.Sprintf("%s-%d", fromAddress.Hex(), signedTx.Nonce())
	assert.Equal(t, expectedKey, key, "Expected key does not match")

	// Verify that the transaction in the cache matches the signed transaction
	cachedTxInfo, exists := c.Data[key]
	assert.True(t, exists, "Expected transaction to be in the cache")
	assert.Equal(t, signedTx, cachedTxInfo.Tx, "Expected cached transaction to be nil")
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
		assert.True(t, executed)
	}()

	go func() {
		defer wg.Done()
		_, newTx, err := testdata.Tx(privateKey, nonce+1, chainID)
		assert.Nil(t, err, "Error while creating transaction")
		currentBlock := int64(2)
		executed, err := c.ProcessTxEntry(newTx, currentBlock)
		assert.Nil(t, err)
		assert.True(t, executed)
	}()
	wg.Wait()

	assert.Equal(t, 2, len(c.Data), "Expected cache to contain 2 entries")
}
