package cache

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"sync"
)

type TransactionInfo struct {
	Tx           *types.Transaction
	SendingBlock uint64
}

type Cache struct {
	sync.RWMutex
	Data        map[string]TransactionInfo
	DelayFactor uint64
}

func NewCache(delayFactor uint64) *Cache {
	return &Cache{
		Data:        make(map[string]TransactionInfo),
		DelayFactor: delayFactor,
	}
}

func (c *Cache) Key(address common.Address, nonce uint64) string {
	return fmt.Sprintf("%s-%d", address.Hex(), nonce)
}

func (c *Cache) ResetEntry(nonce uint64, currentBlock uint64) {
	c.Lock()
	defer c.Unlock()
	key := c.Key(common.Address{}, nonce)
	c.Data[key] = TransactionInfo{Tx: nil, SendingBlock: currentBlock}
}

func (c *Cache) UpdateEntry(newTx *types.Transaction, currentBlock uint64) bool {
	key := c.Key(*newTx.To(), newTx.Nonce())
	c.Lock()
	defer c.Unlock()

	if existing, found := c.Data[key]; found {
		if newTx.GasPrice().Cmp(existing.Tx.GasPrice()) <= 0 {
			fmt.Printf("A transaction already exists with a higher gas price. Delaying transaction sending.")
			existing.SendingBlock = currentBlock
			c.Data[key] = existing
			return false
		}
	}

	fmt.Printf("Adding transaction to the cache.")
	c.Data[key] = TransactionInfo{Tx: newTx, SendingBlock: currentBlock + c.DelayFactor}

	return true
}

func (c *Cache) DeleteEntry(address common.Address, nonce uint64) {
	key := c.Key(address, nonce)
	c.Lock()
	defer c.Unlock()
	delete(c.Data, key)
	fmt.Printf("Cancelling transaction %s-%d\n", address.Hex(), nonce)
}
