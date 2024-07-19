package cache

import (
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shutter-network/encrypting-rpc-server/utils"
	"sync"
)

type TransactionInfo struct {
	Tx         *types.Transaction
	CachedTime int64
}

type Cache struct {
	sync.RWMutex
	Data        map[string]TransactionInfo
	DelayFactor int64
}

func NewCache(delayFactor int64) *Cache {
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

func (c *Cache) UpdateEntry(newTx *types.Transaction, currentTime int64) (bool, error) {
	key, err := c.Key(newTx)
	if err != nil {
		return false, err
	}

	c.Lock()
	defer c.Unlock()

	utils.Logger.Debug().Msgf("Attempting to update cache with key [%s] and transaction hash [%s]", key, newTx.Hash().Hex())
	if existing, found := c.Data[key]; found {
		if existing.Tx != nil { // we sent a transaction in the last d seconds
			// new tx with lower gas -> discard tx
			utils.Logger.Debug().Msgf("Found cache entry with key [%s], transaction data Tx [%s] and CachedTime [%d]",
				key, existing.Tx.Hash().Hex(), existing.CachedTime)
			if newTx.GasPrice().Cmp(existing.Tx.GasPrice()) <= 0 {
				utils.Logger.Debug().Msgf("A transaction already exists with a higher gas price. "+
					"Delaying transaction sending to [%d].", currentTime)
				txInfo := TransactionInfo{existing.Tx, existing.CachedTime}
				c.Data[key] = txInfo
				return false, nil // false -> tx won't be sent
			}

			// new tx with higher gas -> update tx
			utils.Logger.Debug().Msgf("A transaction already exists with a lower gas price. "+
				"Updating transaction and delaying transaction sending to [%d].", currentTime)
			txInfo := TransactionInfo{newTx, existing.CachedTime}
			c.Data[key] = txInfo
			return false, nil // false -> tx won't be sent
		}
	}

	// no tx sent in the last d seconds
	utils.Logger.Debug().Msgf("Adding transaction with hash [%s] and time [%v] to the cache at key [%s] \n", newTx.Hash(), currentTime, key)
	c.Data[key] = TransactionInfo{Tx: newTx, CachedTime: currentTime}
	utils.Logger.Debug().Msgf("Cache entry updated to: Tx = nil and CachedTime = [%d]", c.Data[key].CachedTime)
	return true, nil // true -> send tx
}
