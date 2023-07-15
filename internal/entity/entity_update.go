package entity

import (
	"fmt"
)

// Update updates an existing record in the database.
func Update(m interface{}, keyNames ...string) (err error) {
	// Unscoped so soft-deleted records can still be updated.
	db := UnscopedDb()

	// New entity?
	if db.NewRecord(m) {
		return fmt.Errorf("new record")
	}

	// Extract interface slice with all values including zero.
	values, keys, err := ModelValues(m, keyNames...)

	// Has keys and values?
	if err != nil {
		return err
	} else if len(keys) != len(keyNames) {
		return fmt.Errorf("record keys missing")
	}

	// Update values.
	result := db.Model(m).Updates(values)

	// Successful?
	if result.Error != nil {
		return err
	} else if result.RowsAffected > 1 {
		log.Debugf("entity: updated statement affected more than one record - you may have found a bug")
		return nil
	} else if result.RowsAffected == 1 {
		return nil
	} else if Count(m, keyNames, keys) != 1 {
		return fmt.Errorf("record not found")
	}

	return err
}
