package entity

import (
	"fmt"
	"reflect"
	"strings"
)

// Values returns entity values as string map.
func Values(m interface{}, omit ...string) (result map[string]interface{}) {
	skip := func(name string) bool {
		for _, s := range omit {
			if name == s {
				return true
			}
		}

		return false
	}

	result = make(map[string]interface{})

	elem := reflect.ValueOf(m).Elem()
	relType := elem.Type()

	for i := 0; i < relType.NumField(); i++ {
		name := relType.Field(i).Name

		if skip(name) {
			continue
		}

		result[name] = elem.Field(i).Interface()
	}

	return result
}

// Save updates an entity in the database, or inserts if it doesn't exist.
func Save(m interface{}, primaryKeys ...string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("save: %s (panic)", r)
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
			err = fmt.Errorf("update: %s (panic)", r)
			log.Error(err)
		}
	}()

	v := reflect.ValueOf(m).Elem()

	// Abort if a primary key is zero.
	for _, k := range primaryKeys {
		if field := v.FieldByName(k); field.IsZero() {
			return fmt.Errorf("key '%s' not found", k)
		}
	}

	// Update all values except primary keys.
	if res := UnscopedDb().Model(m).Updates(Values(m, primaryKeys...)); res.Error != nil {
		return res.Error
	} else if res.RowsAffected > 1 {
		log.Warnf("update: more than one row affected")
	} else if res.RowsAffected == 0 {
		// MariaDB may report zero rows in case no data was actually changed, even though the row exists.
		log.Tracef("update: no rows affected")
	}

	return nil
}
