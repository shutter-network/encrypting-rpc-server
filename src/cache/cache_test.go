package cache

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"testing"
)

// todo update to cover all cases

func TestCache(t *testing.T) {
	c := NewCache(10)
	toAddress := common.HexToAddress("0x123")
	txData := &types.LegacyTx{
		Nonce:    0,
		To:       &toAddress,
		Value:    big.NewInt(0),
		Gas:      21000,
		GasPrice: big.NewInt(1),
	}
	tx := types.NewTx(txData)

	updated := c.UpdateEntry(tx, 100)
	if !updated {
		t.Error("Expected transaction to be added to cache")
	}

	c.ResetEntry(tx.Nonce(), 110)
	entry := c.Data[c.Key(common.Address{}, tx.Nonce())]
	if entry.Tx != nil || entry.SendingBlock != 110 {
		t.Error("Expected entry to be reset with nil transaction and new sending block")
	}

	c.DeleteEntry(toAddress, tx.Nonce())
	if _, exists := c.Data[c.Key(toAddress, tx.Nonce())]; exists {
		t.Error("Expected entry to be deleted from cache")
	}
}
