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

func (c *Cache) UpdateEntry(key string, tx *types.Transaction, cachedTime int64) {
	txInfo := TransactionInfo{Tx: tx, CachedTime: cachedTime}
	c.Data[key] = txInfo
	utils.Logger.Debug().Msgf("Cache entry at key [%s] updated to: Tx = [%s] and CachedTime = [%d]",
		key, c.Data[key].Tx.Hash().Hex(), c.Data[key].CachedTime)
}

func (c *Cache) ProcessTxEntry(newTx *types.Transaction, currentTime int64) (bool, error) {
	key, err := c.Key(newTx)
	if err != nil {
		return false, err
	}

	c.Lock()
	defer c.Unlock()

	utils.Logger.Debug().Msgf("Attempting to update cache with key [%s] and transaction hash [%s]", key, newTx.Hash().Hex())
	if existing, found := c.Data[key]; found {
		if existing.Tx != nil {
			if existing.CachedTime+c.DelayFactor > currentTime { // we sent a transaction within less than d seconds
				// new tx with lower gas -> discard tx
				utils.Logger.Debug().Msgf("Found cache entry with key [%s], transaction data Tx [%s] and CachedTime [%d]",
					key, existing.Tx.Hash().Hex(), existing.CachedTime)
				if newTx.GasPrice().Cmp(existing.Tx.GasPrice()) <= 0 {
					utils.Logger.Debug().Msgf("A transaction already exists with a higher gas price. "+
						"Delaying transaction sending to [%d].", currentTime)
					return false, nil // false -> tx won't be sent
				}

				// new tx with higher gas -> update tx
				utils.Logger.Debug().Msgf("A transaction already exists with a lower gas price. "+
					"Updating transaction and delaying transaction sending to [%d].", currentTime)
				c.UpdateEntry(key, newTx, existing.CachedTime)
				return false, nil // false -> tx won't be sent
			}

			// we sent a transaction within more than d seconds
			utils.Logger.Debug().Msgf("Found cache entry with key [%s], transaction data Tx [%s] and CachedTime [%d]. Resending.",
				key, existing.Tx.Hash().Hex(), existing.CachedTime)

			if newTx.GasPrice().Cmp(existing.Tx.GasPrice()) <= 0 { // previous tx with higher gas price
				utils.Logger.Debug().Msgf("Keeping transaction with hash [%s] in the cache at key [%s]\n", newTx.Hash(), key)
				c.UpdateEntry(key, existing.Tx, currentTime)
				return true, nil // true -> tx will be sent
			} else {
				utils.Logger.Debug().Msgf("Adding transaction with hash [%s] to the cache at key [%s]\n", newTx.Hash(), key)
				c.UpdateEntry(key, newTx, currentTime)
				return true, nil // true -> tx will be sent
			}
		}
	}

	// first transaction
	utils.Logger.Debug().Msgf("Adding transaction with hash [%s] and time [%v] to the cache at key [%s] \n", newTx.Hash(), currentTime, key)
	c.UpdateEntry(key, newTx, currentTime)
	return true, nil // true -> send tx
}
