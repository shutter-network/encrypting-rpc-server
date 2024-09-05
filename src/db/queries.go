package db

import (
	"gorm.io/gorm"
)

func (db *PostgresDb) updateInclusion(txDetails TransactionDetails) error {
	if err := db.DB.Transaction(func(tx *gorm.DB) error {
		// Subquery to count rows with the same TxHash
		subQuery := tx.Model(&TransactionDetails{}).
			Select("COUNT(*) - 1").
			Where("tx_hash = ? AND encrypted_tx_hash IS NOT NULL AND encrypted_tx_hash <> ''", txDetails.TxHash)

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
		return err
	}
	return nil
}
