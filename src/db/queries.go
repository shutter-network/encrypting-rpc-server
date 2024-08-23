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
