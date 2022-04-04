package entity

import (
	"fmt"
	"reflect"
	"runtime/debug"
	"strings"
)

// Save updates an entity in the database, or inserts if it doesn't exist.
func Save(m interface{}, primaryKeys ...string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("index: save failed (%s)\nstack: %s", r, debug.Stack())
			log.Error(err)
		}
	}()

	if err := Update(m, primaryKeys...); err == nil {
		return nil
	} else if err := UnscopedDb().Save(m).Error; err == nil {
		return nil
	} else if !strings.Contains(strings.ToLower(err.Error()), "lock") {
		return err
	} else if err := UnscopedDb().Save(m).Error; err != nil {
		return err
	}

	return nil
}

// Update updates an existing entity in the database.
func Update(m interface{}, primaryKeys ...string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("index: update failed (%s)\nstack: %s", r, debug.Stack())
			log.Error(err)
		}
	}()

	// Return with error if a primary key is empty.
	v := reflect.ValueOf(m).Elem()
	for _, k := range primaryKeys {
		if field := v.FieldByName(k); !field.CanSet() || field.IsZero() {
			return fmt.Errorf("empty primary key '%s'", k)
		}
	}

	err = UnscopedDb().FirstOrCreate(m, GetValues(m)).Error

	return err
}
