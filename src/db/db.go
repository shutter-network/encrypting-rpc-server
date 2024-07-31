package db

import (
	"context"

	"github.com/lib/pq"
	"github.com/shutter-network/encrypting-rpc-server/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gorm_logger "gorm.io/gorm/logger"
)

type PostgresDb struct {
	ctx          context.Context
	cancel       context.CancelFunc
	DB           *gorm.DB
	addTxChannel chan TransactionDetails
}

type TransactionDetails struct {
	TxHash          string         `gorm:"primaryKey;uniqueIndex:p"`
	EncryptedTxHash pq.StringArray `gorm:"type:text[]"`
	InclusionTime   int64
	Retries         uint64
}

func InitialMigration(dbUrl string) (*PostgresDb, error) {

	gormConfig := &gorm.Config{Logger: gorm_logger.Default.LogMode(gorm_logger.Silent)}

	db, err := gorm.Open(postgres.Open(dbUrl), gormConfig)
	if err != nil {
		utils.Logger.Error().Err(err).Msg("failed to connect database")
	}

	// run migrations
	if err := db.AutoMigrate(TransactionDetails{}); err != nil {
		utils.Logger.Error().Err(err).Msg("failed to automigrate tables")
	}

	addTxChan := make(chan TransactionDetails)

	ctx, cancel := context.WithCancel(context.Background())
	return &PostgresDb{ctx: ctx, cancel: cancel, DB: db, addTxChannel: addTxChan}, nil
}

func (db *PostgresDb) InsertOrUpdateNewTx(txDetails TransactionDetails) {
	db.addTxChannel <- txDetails
}

func (db *PostgresDb) Start() error {
	for {
		select {
		case txDetails := <-db.addTxChannel:
			db.DB.Transaction(func(tx *gorm.DB) error {
				// Try to update existing record or insert a new one
				query := `
						INSERT INTO transaction_details (tx_hash, encrypted_tx_hash, inclusion_time, retries)
						VALUES (?, ?, ?, ?)
						ON CONFLICT (tx_hash)
						DO UPDATE SET
							encrypted_tx_hash = array(
								SELECT DISTINCT unnest(transaction_details.encrypted_tx_hash) || ?
								FROM transaction_details
								WHERE tx_hash = ?
							),
							inclusion_time = EXCLUDED.inclusion_time,
							retries = transaction_details.retries + 1
						WHERE transaction_details.tx_hash = EXCLUDED.tx_hash`

				result := tx.Exec(query, txDetails.TxHash, txDetails.EncryptedTxHash, txDetails.InclusionTime, txDetails.Retries, txDetails.EncryptedTxHash, txDetails.TxHash)
				if result.Error != nil {
					return result.Error
				}
				return nil
			})
		case <-db.ctx.Done():
			return nil
		}
	}
}

func (db *PostgresDb) Stop() {
	defer db.cancel()
	close(db.addTxChannel)
}
