package cache

import (
	"fmt"
	"math/big"
	"testing"

	test_data "github.com/shutter-network/encrypting-rpc-server/test-data"
	"github.com/stretchr/testify/assert"
)

// todo update to cover all cases

func TestCache_Key(t *testing.T) {
	privateKey, fromAddress, err := test_data.GenerateKeyPair()
	assert.NoError(t, err, "Failed to generate key pair")

	chainID := big.NewInt(1)
	nonce := uint64(1)
	_, signedTx, err := test_data.Tx(privateKey, nonce, chainID)
	assert.NoError(t, err, "Failed to create signed transaction")

	c := NewCache(10)

	key, err := c.Key(signedTx)
	assert.NoError(t, err, "Failed to get key from transaction")

	expectedKey := fmt.Sprintf("%s-%d", fromAddress.Hex(), nonce)
	assert.Equal(t, expectedKey, key, "Expected key does not match")
}

func TestCache_UpdateEntry(t *testing.T) {
	c := NewCache(10)

	privateKey, fromAddress, err := test_data.GenerateKeyPair()
	assert.NoError(t, err, "Failed to generate key pair")

	_, signedTx, err := test_data.Tx(privateKey, 1, big.NewInt(1))
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
	assert.Equal(t, signedTx.Hash(), cachedTxInfo.Tx.Hash(), "Expected cached transaction hash does not match")
}
