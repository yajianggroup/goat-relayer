package migrations

import (
	"gorm.io/gorm"
)

// AddSafeboxTaskOrderId adds the OrderId field to the SafeboxTask table
func AddSafeboxTaskOrderId(tx *gorm.DB) error {
	// First, add the field but do not set it as non-null
	if err := tx.Exec("ALTER TABLE safebox_tasks ADD COLUMN order_id VARCHAR(255)").Error; err != nil {
		return err
	}

	// Update existing records, set default values
	if err := tx.Exec("UPDATE safebox_tasks SET order_id = '' WHERE order_id IS NULL").Error; err != nil {
		return err
	}

	// Add an index
	if err := tx.Exec("CREATE INDEX safeboxtask_orderid_index ON safebox_tasks (order_id)").Error; err != nil {
		return err
	}

	// Finally, set the field as non-null
	if err := tx.Exec("ALTER TABLE safebox_tasks MODIFY COLUMN order_id VARCHAR(255) NOT NULL").Error; err != nil {
		return err
	}

	return nil
}
