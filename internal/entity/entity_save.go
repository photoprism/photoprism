package entity

import (
	"fmt"
	"strings"
)

// Save updates a record in the database, or inserts
// if it does not exist.
func Save(m interface{}, keyNames ...string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("save: %s (panic)", r)
		}
	}()

	// Try a regular update first.
	if err = Update(m, keyNames...); err == nil {
		return nil
	}

	// Automatically insert/update record as needed.
	if err = UnscopedDb().Save(m).Error; err == nil {
		return nil
	}

	// Try again if database was locked, return otherwise.
	if !strings.Contains(strings.ToLower(err.Error()), "lock") {
		return err
	} else if err = UnscopedDb().Save(m).Error; err == nil {
		return nil
	}

	return err
}
