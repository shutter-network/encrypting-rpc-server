package db

import (
	"context"
	"fmt"

	"github.com/shutter-network/encrypting-rpc-server/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gorm_logger "gorm.io/gorm/logger"
)

const BufferSize = 10

type PostgresDb struct {
	DB          *gorm.DB
	addTxCh     chan TransactionDetails
	inclusionCh chan TransactionDetails
}

type TransactionDetails struct {
	Address         string `gorm:"primaryKey;index:idx_address_nonce"`
	Nonce           uint64 `gorm:"primaryKey;index:idx_address_nonce"`
	TxHash          string `gorm:"primaryKey;index:idx_tx_hash"`
	EncryptedTxHash string `gorm:"primaryKey"`
	InclusionTime   uint64
	Retries         uint64
}

func InitialMigration(dbUrl string) (*PostgresDb, error) {

	gormConfig := &gorm.Config{Logger: gorm_logger.Default.LogMode(gorm_logger.Silent)}

	db, err := gorm.Open(postgres.Open(dbUrl), gormConfig)
	if err != nil {
		utils.Logger.Error().Err(err).Msg("failed to connect database")
	}

	err = createRoleIfNotExists(db, "postgres", "postgres")
	if err != nil {
		return nil, fmt.Errorf("failed to create role: %v", err)
	}
	// run migrations
	if err := db.AutoMigrate(TransactionDetails{}); err != nil {
		utils.Logger.Error().Err(err).Msg("failed to automigrate tables")
		return nil, fmt.Errorf("failed to automigrate tables")
	}

	inclusionCh := make(chan TransactionDetails, BufferSize)
	addTxCh := make(chan TransactionDetails, BufferSize)

	return &PostgresDb{DB: db, addTxCh: addTxCh, inclusionCh: inclusionCh}, nil
}

func (db *PostgresDb) InsertNewTx(txDetails TransactionDetails) {
	db.addTxCh <- txDetails
}

func (db *PostgresDb) FinaliseTx(receipt TransactionDetails) {
	db.inclusionCh <- receipt
}

func (db *PostgresDb) Start(ctx context.Context) error {
	for {
		select {
		case txDetails := <-db.addTxCh:
			if err := db.DB.Create(txDetails).Error; err != nil {
				utils.Logger.Info().Msgf("Error recording tx | txHash: %s | err: %v", txDetails.TxHash, err)
				continue
			}
		case txDetails := <-db.inclusionCh:
			println(txDetails.TxHash, "got hereerererere")
			if err := db.DB.Transaction(func(tx *gorm.DB) error {
				// Subquery to count rows with the same TxHash
				subQuery := tx.Model(&TransactionDetails{}).
					Select("COUNT(*) - 1").
					Where("tx_hash = ?", txDetails.TxHash)
					// Group("tx_hash")

				// Update all rows with new inclusion_time and retries count
				if err := tx.Model(&TransactionDetails{}).
					Where("tx_hash = ?", txDetails.TxHash).
					Updates(map[string]interface{}{
						"inclusion_time": txDetails.InclusionTime,
						"retries":        gorm.Expr("(?)", subQuery),
					}).Error; err != nil {
					// Return any error will rollback the transaction
					return err
				}
				// Return nil to commit the transaction
				return nil
			}); err != nil {
				utils.Logger.Info().Msgf("Error updating inclusion time | txHash: %s | err: %v", txDetails.TxHash, err)
				continue
			}

		case <-ctx.Done():
			close(db.addTxCh)
			close(db.inclusionCh)
			return nil
		}
	}
}

func createRoleIfNotExists(db *gorm.DB, roleName, password string) error {
	// Define the SQL statement with CREATEDB permission
	sql := fmt.Sprintf(`
    DO $$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = '%s') THEN
            EXECUTE 'CREATE ROLE %s WITH LOGIN CREATEDB PASSWORD ''%s''';
        END IF;
    END $$;
    `, roleName, roleName, password)

	// Execute the SQL statement
	if err := db.Exec(sql).Error; err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}
	return nil
}
