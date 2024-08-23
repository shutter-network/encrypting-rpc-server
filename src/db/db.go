package db

import (
	"context"
	"fmt"

	"github.com/shutter-network/encrypting-rpc-server/utils"
)

func (db *PostgresDb) InsertNewTx(txDetails TransactionDetails) {
	db.addTxCh <- txDetails
}

// txhash and inclusion time are mandatory fields to update the finalised tx
func (db *PostgresDb) FinaliseTx(receipt TransactionDetails) {
	db.inclusionCh <- receipt
}

func (db *PostgresDb) Start(ctx context.Context) {
	sqlDb, err := db.DB.DB()
	if err != nil {
		utils.Logger.Info().Msgf("cannot initiate sqlDb | err: %v", err)
		panic(fmt.Sprintf("cannot initiate sqlDb | err: %v", err))
	}
	defer sqlDb.Close()
	for {

		select {
		case txDetails := <-db.addTxCh:
			if err := db.DB.Create(txDetails).Error; err != nil {
				utils.Logger.Info().Msgf("Error recording tx | txHash: %s | err: %v", txDetails.TxHash, err)
				continue
			}
		case txDetails := <-db.inclusionCh:
			if err := db.updateInclusion(txDetails); err != nil {
				utils.Logger.Info().Msgf("Error updating inclusion time | txHash: %s | err: %v", txDetails.TxHash, err)
				continue
			}

		case <-ctx.Done():
			close(db.addTxCh)
			close(db.inclusionCh)
			return
		}
	}
}
