package entity

import (
	"fmt"
)

// Update updates the values of an existing database record.
func Update(m interface{}, keyNames ...string) (err error) {
	// Use an unscoped *gorm.DB connection, so that
	// soft-deleted database records can also be updated.
	db := UnscopedDb()

	// Return if the record has not been created yet.
	if db.NewRecord(m) {
		return fmt.Errorf("new record")
	}

	// Extract interface slice with all values including zero.
	values, keys, err := ModelValues(m, keyNames...)

	// Check if the number of keys matches the number of values.
	if err != nil {
		return err
	} else if len(keys) != len(keyNames) {
		return fmt.Errorf("record keys missing")
	}

	// Execute update statement.
	result := db.Model(m).Updates(values)

	// Return an error if the update has failed.
	if err = result.Error; err != nil {
		return err
	}

	// Verify number of updated rows.
	if result.RowsAffected > 1 {
		log.Debugf("entity: updated statement affected more than one record - you may have found a bug")
		return nil
	} else if result.RowsAffected == 1 {
		return nil
	} else if Count(m, keyNames, keys) != 1 {
		return fmt.Errorf("record not found")
	}

	return nil
}
