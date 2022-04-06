package entity

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/jinzhu/gorm"
)

// Save updates a record in the database, or inserts if it doesn't exist.
func Save(m interface{}, keyNames ...string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("index: save failed (%s)\nstack: %s", r, debug.Stack())
			log.Error(err)
		}
	}()

	// Try updating first.
	if err = Update(m, keyNames...); err == nil {
		return nil
	} else if err = UnscopedDb().Save(m).Error; err == nil {
		return nil
	} else if !strings.Contains(strings.ToLower(err.Error()), "lock") {
		return err
	} else if err = UnscopedDb().Save(m).Error; err != nil {
		return err
	}

	return nil
}

// Update updates an existing record in the database.
func Update(m interface{}, keyNames ...string) (err error) {
	// New entity?
	if Db().NewRecord(m) {
		return fmt.Errorf("new record")
	}

	values, keys, err := ModelValues(m, keyNames...)

	// Has keys and values?
	if err != nil {
		return err
	} else if len(keys) != len(keyNames) {
		return fmt.Errorf("record keys missing")
	}

	// Perform update.
	res := Db().Model(m).Updates(values)

	// Successful?
	if res.Error != nil {
		return err
	} else if res.RowsAffected > 1 {
		log.Debugf("entity: updated statement affected more than one record - possible bug")
		return nil
	} else if res.RowsAffected == 1 {
		return nil
	} else if Count(m, keyNames, keys) != 1 {
		return fmt.Errorf("record not found")
	}

	return err
}

// Count returns the number of records for a given a model and key values.
func Count(m interface{}, keys []string, values []interface{}) int {
	if m == nil || len(keys) != len(values) {
		log.Debugf("entity: invalid parameters (count records)")
		return -1
	}

	var count int

	stmt := Db().Model(m)

	for k := range keys {
		stmt.Where("? = ?", gorm.Expr(keys[k]), values[k])
	}

	if err := stmt.Count(&count).Error; err != nil {
		log.Debugf("entity: %s (count records)", err)
		return -1
	}

	return count
}
