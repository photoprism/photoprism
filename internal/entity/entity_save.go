package entity

import (
	"fmt"
)

// Save tries to update an existing record and falls back to insert semantics, retrying on lock errors.
func Save(m any, keyNames ...string) (err error) {
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
	if !isLockError(err) {
		return err
	} else if err = UnscopedDb().Save(m).Error; err == nil {
		return nil
	}

	return err
}
