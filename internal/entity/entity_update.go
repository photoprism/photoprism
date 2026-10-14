package entity

import (
	"fmt"
	"reflect"

	"gorm.io/gorm"
)

// Update updates the values of an existing database record.
func Update(m any, keyNames ...string) (err error) {
	// Use an unscoped *gorm.DB connection, so that
	// soft-deleted database records can also be updated.
	db := UnscopedDb()

	// Return if the record has not been created yet.
	if isNewRecord(db, m) {
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
	result := db.Model(m).Updates(columnValues(db, m, values))

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

// isNewRecord reports whether the primary key of a model is zero, so it has not been created yet.
// Composite keys are checked by their first field.
func isNewRecord(db *gorm.DB, m any) bool {
	stmt := &gorm.Statement{DB: db}

	if err := stmt.Parse(m); err != nil || stmt.Schema == nil || len(stmt.Schema.PrimaryFields) == 0 {
		return true
	}

	field := stmt.Schema.PrioritizedPrimaryField

	if field == nil {
		field = stmt.Schema.PrimaryFields[0]
	}

	_, isZero := field.ValueOf(db.Statement.Context, reflect.Indirect(reflect.ValueOf(m)))

	return isZero
}

// columnValues returns the values whose field names map to table columns of the model,
// as the ORM would otherwise also update fields without a column.
func columnValues(db *gorm.DB, m any, values Values) Values {
	stmt := &gorm.Statement{DB: db}

	if err := stmt.Parse(m); err != nil || stmt.Schema == nil {
		return values
	}

	result := make(Values, len(values))

	for name, value := range values {
		if field := stmt.Schema.LookUpField(name); field != nil && field.DBName != "" {
			result[field.DBName] = value
		}
	}

	return result
}
