package cache

import (
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shutter-network/encrypting-rpc-server/utils"
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

func (c *Cache) Key(tx *types.Transaction) (string, error) {
	fromAddress, err := utils.SenderAddress(tx)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s-%d", fromAddress.Hex(), tx.Nonce()), nil
}

func (c *Cache) ResetEntry(tx *types.Transaction, currentBlock uint64) (bool, error) {
	c.Lock()
	defer c.Unlock()
	key, err := c.Key(tx)
	if err != nil {
		return false, err
	}

	c.Data[key] = TransactionInfo{Tx: nil, SendingBlock: currentBlock}
	return true, nil
}

func (c *Cache) UpdateEntry(newTx *types.Transaction, currentBlock uint64) (bool, error) {
	key, err := c.Key(newTx)
	if err != nil {
		return false, err
	}

	c.Lock()
	defer c.Unlock()

	utils.Logger.Debug().Msgf("Attempting to update cache with key [%s] and transaction hash [%s]", key, newTx.Hash().Hex())

	if existing, found := c.Data[key]; found {
		if newTx.GasPrice().Cmp(existing.Tx.GasPrice()) <= 0 {
			utils.Logger.Debug().Msgf("A transaction already exists with a higher gas price. "+
				"Delaying transaction sending to [%d].", currentBlock)
			existing.SendingBlock = currentBlock
			c.Data[key] = existing
			return false, nil
		}
	}

	utils.Logger.Info().Msgf("Adding transaction with hash [%s] to the cache at key [%s]\n", newTx.Hash(), key)
	c.Data[key] = TransactionInfo{Tx: newTx, SendingBlock: currentBlock + c.DelayFactor}

	return true, nil
}

func (c *Cache) DeleteEntry(tx *types.Transaction) (bool, error) {
	key, err := c.Key(tx)
	if err != nil {
		return false, err
	}

	c.Lock()
	defer c.Unlock()

	utils.Logger.Info().Msgf("Deleting entry at key %s", key)
	delete(c.Data, key)
	return true, nil
}
