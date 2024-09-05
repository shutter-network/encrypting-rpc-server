package db

import (
	"gorm.io/gorm"
)

func (db *PostgresDb) updateInclusion(txDetails TransactionDetails) error {
	if err := db.DB.Transaction(func(tx *gorm.DB) error {

		// Update all rows with new inclusion_time
		if err := tx.Model(&TransactionDetails{}).
			Where("tx_hash = ?", txDetails.TxHash).
			Updates(map[string]interface{}{
				"inclusion_time": txDetails.InclusionTime,
			}).Error; err != nil {
			// Return any error will rollback the transaction
			return err
		}
		// Return nil to commit the transaction
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (db *PostgresDb) GetRetries(txHash string) (retries int64, err error) {
	if err := db.DB.Model(TransactionDetails{}).Where("tx_hash = ?", txHash).Count(&retries).Error; err != nil {
		return 0, err
	}
	return retries, nil
}
