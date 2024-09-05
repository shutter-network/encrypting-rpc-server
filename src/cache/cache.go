package cache

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shutter-network/encrypting-rpc-server/utils"
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

type ProcessTxEntryResp struct {
	SendStatus   bool
	UpdateStatus bool
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
	isCancellation := utils.IsCancellationTransaction(tx, fromAddress)
	return fmt.Sprintf("%s-%d-%t", fromAddress.Hex(), tx.Nonce(), isCancellation), nil
}

func (c *Cache) UpdateEntry(key string, tx *types.Transaction, cachedTime int64) {
	txInfo := TransactionInfo{Tx: tx, CachedTime: cachedTime}
	c.Data[key] = txInfo
	utils.Logger.Debug().Msgf("Cache entry at key [%s] updated to: Tx = [%s] and CachedTime = [%d]",
		key, c.Data[key].Tx.Hash().Hex(), c.Data[key].CachedTime)
}

func (c *Cache) ProcessTxEntry(newTx *types.Transaction, currentTime int64) (ProcessTxEntryResp, error) {
	key, err := c.Key(newTx)
	if err != nil {
		return ProcessTxEntryResp{
			SendStatus:   false,
			UpdateStatus: false,
		}, err
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
				utils.Logger.Debug().Msgf("A transaction already exists with a higher gas price. " +
					"Delaying transaction sending.")
				return ProcessTxEntryResp{
					SendStatus:   false,                                             // false -> tx won't be sent
					UpdateStatus: newTx.GasPrice().Cmp(existing.Tx.GasPrice()) != 0, // the db should be only updated when there is change in gas price
				}, nil
			}

			utils.Logger.Debug().Msgf("A transaction already exists with a lower gas price. " +
				"Updating transaction and delaying transaction sending.")
			c.UpdateEntry(key, newTx, existing.CachedTime)
			return ProcessTxEntryResp{
				SendStatus:   false,
				UpdateStatus: true, // new tx with higher gas -> update tx
			}, nil // false -> tx won't be sent
		}
	}

	// no tx sent in the last d seconds
	utils.Logger.Debug().Msgf("Adding transaction with hash [%s] and time [%v] to the cache at key [%s] \n", newTx.Hash(), currentTime, key)
	c.UpdateEntry(key, newTx, currentTime)
	return ProcessTxEntryResp{
		SendStatus:   true,
		UpdateStatus: true,
	}, nil // true -> send tx
}
