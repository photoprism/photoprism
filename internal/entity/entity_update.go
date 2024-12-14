package entity

import (
	"fmt"
	"reflect"

	"github.com/ulule/deepcopier"
)

// Checks if the primary key is populated
func NewRecord(m interface{}) (result bool, err error) {
	tx := UnscopedDb()
	tx.Statement.Dest = m

	reflectValue := reflect.Indirect(reflect.ValueOf(m))
	for reflectValue.Kind() == reflect.Ptr || reflectValue.Kind() == reflect.Interface {
		reflectValue = reflect.Indirect(reflectValue)
	}

	switch reflectValue.Kind() {
	case reflect.Struct:
		if err := tx.Statement.Parse(m); err == nil && tx.Statement.Schema != nil {
			for _, pf := range tx.Statement.Schema.PrimaryFields {
				if _, isZero := pf.ValueOf(tx.Statement.Context, reflectValue); isZero {
					return true, nil
				}
			}
		}
		return false, nil
	default:
		return true, fmt.Errorf("interface %s not recognised", reflectValue.Kind().String())
	}
}

// Update updates the values of an existing database record.
func Update(m interface{}, keyNames ...string) (err error) {
	// Use an unscoped *gorm.DB connection, so that
	// soft-deleted database records can also be updated.
	db := UnscopedDb()

	// Return if the record has not been created yet.
	if newrec, err := NewRecord(m); newrec == true || err != nil {
		if err != nil {
			return err
		} else {
			return fmt.Errorf("new record")
		}
	}

	// Create a backup to use as otherwise we are going to remove a bunch of data from m
	// which may break other things.
	var backup interface{} = reflect.New(reflect.ValueOf(m).Elem().Type()).Interface()
	//log.Debugf("backup = %v", backup)
	if err = deepcopier.Copy(m).To(backup); err != nil {
		log.Debugf("entity_update deepcopier failed with error %v", err)
		return err
	}

	// Extract interface slice with all values including zero.
	values, keys, err := ModelValuesStructOption(backup, false, keyNames...)

	// Check if the number of keys matches the number of values.
	if err != nil {
		return err
	} else if len(keys) != len(keyNames) {
		return fmt.Errorf("record keys missing")
	}

	// Get the counter before attempting an update as calls after a constraint failure don't work.
	counter := Count(m, keyNames, keys)

	// Execute update statement.
	result := db.Model(backup).Updates(values)

	// Push the UpdatedAt if found back into the original m
	if _, found := reflect.ValueOf(backup).Elem().Type().FieldByName("UpdatedAt"); found {
		updatedAt := reflect.ValueOf(backup).Elem().FieldByName("UpdatedAt")
		reflect.ValueOf(m).Elem().FieldByName("UpdatedAt").Set(updatedAt)
	}

	// Return an error if the update has failed.
	if err = result.Error; err != nil {
		if counter != 1 {
			return fmt.Errorf("record not found and %v", err)
		} else {
			return err
		}
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
