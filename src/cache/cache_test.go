package cache

import (
	"fmt"
	"math/big"
	"sync"
	"testing"

	testdata "github.com/shutter-network/encrypting-rpc-server/test-data"
	"github.com/stretchr/testify/assert"
)

// todo update to cover all cases

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
	c := NewCache(10)

	privateKey, fromAddress, err := testdata.GenerateKeyPair()
	assert.NoError(t, err, "Failed to generate key pair")

	_, signedTx, err := testdata.Tx(privateKey, 1, big.NewInt(1))
	assert.NoError(t, err, "Failed to create signed transaction")

	updated, err := c.UpdateEntry(signedTx, 100)
	assert.NoError(t, err, "Failed to update entry")
	assert.True(t, updated, "Expected transaction to be added to cache")

	key, err := c.Key(signedTx)
	assert.NoError(t, err, "Failed to get key from cache")

	expectedKey := fmt.Sprintf("%s-%d", fromAddress.Hex(), signedTx.Nonce())
	assert.Equal(t, expectedKey, key, "Expected key does not match")

	// Verify that the transaction in the cache matches the signed transaction
	cachedTxInfo, exists := c.Data[key]
	assert.True(t, exists, "Expected transaction to be in the cache")
	assert.Nil(t, cachedTxInfo.Tx, "Expected cached transaction to be nil")
}

func TestCacheConcurrentUpdateEntry(t *testing.T) {
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
		currentBlock := uint64(1)
		executed, err := c.UpdateEntry(newTx, currentBlock)
		assert.Nil(t, err)
		assert.True(t, executed)
	}()

	go func() {
		defer wg.Done()
		_, newTx, err := testdata.Tx(privateKey, nonce+1, chainID)
		assert.Nil(t, err, "Error while creating transaction")
		currentBlock := uint64(2)
		executed, err := c.UpdateEntry(newTx, currentBlock)
		assert.Nil(t, err)
		assert.True(t, executed)
	}()
	wg.Wait()

	assert.Equal(t, 2, len(c.Data), "Expected cache to contain 2 entries")
}
