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
	TriedAgain bool
}

type Cache struct {
	sync.RWMutex
	Data                   map[string]TransactionInfo
	WaitingForReceiptCache map[string]bool //the key here should be tx hash
	DelayFactor            int64
}

type ProcessTxEntryResp struct {
	SendStatus   bool
	UpdateStatus bool
}

func NewCache(delayFactor int64) *Cache {
	return &Cache{
		Data:                   make(map[string]TransactionInfo),
		DelayFactor:            delayFactor,
		WaitingForReceiptCache: make(map[string]bool),
	}
}

func (c *Cache) Key(tx *types.Transaction) (string, error) {
	fromAddress, err := utils.SenderAddress(tx)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s-%d", fromAddress.Hex(), tx.Nonce()), nil
}

func (c *Cache) UpdateEntry(key string, tx *types.Transaction, cachedTime int64, triedAgain bool) {
	txInfo := TransactionInfo{Tx: tx, CachedTime: cachedTime, TriedAgain: triedAgain}
	c.Data[key] = txInfo
	utils.Logger.Debug().Msgf("Cache entry at key [%s] updated to: CachedTime = [%d]",
		key, c.Data[key].CachedTime)
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
		if existing.TriedAgain { // we sent a transaction in the last d seconds
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
			c.UpdateEntry(key, newTx, existing.CachedTime, true)
			return ProcessTxEntryResp{
				SendStatus:   false,
				UpdateStatus: true, // new tx with higher gas -> update tx
			}, nil // false -> tx won't be sent
		}
		utils.Logger.Debug().Msgf("Found cache entry with nil value.")
		utils.Logger.Debug().Msgf("Adding transaction with hash [%s] to the cache at key [%s]\n", newTx.Hash(), key)
		c.UpdateEntry(key, newTx, existing.CachedTime, true)

		return ProcessTxEntryResp{
			SendStatus:   false,
			UpdateStatus: newTx.GasPrice().Cmp(existing.Tx.GasPrice()) != 0, // we should record it if there is any change in gas price
		}, nil // false -> tx won't be sent
	}

	// no tx sent in the last d seconds
	utils.Logger.Debug().Msgf("Adding transaction with hash [%s] and time [%v] to the cache at key [%s] \n", newTx.Hash(), currentTime, key)
	c.UpdateEntry(key, newTx, currentTime, false)
	return ProcessTxEntryResp{
		SendStatus:   true,
		UpdateStatus: true,
	}, nil // true -> send tx
}
