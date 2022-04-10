package entity

import (
	"fmt"
	"runtime/debug"
)

// Save updates a record in the database, or inserts if it doesn't exist.
func Save(m interface{}, keyNames ...string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("entity: save failed (%s)\nstack: %s", r, debug.Stack())
			log.Error(err)
		}
	}()

	// Try updating first, then creating.
	if err = Update(m, keyNames...); err == nil {
		return nil
	} else if err = UnscopedDb().Create(m).Error; err == nil {
		return nil
	} else if err = UnscopedDb().Save(m).Error; err != nil {
		return err
	}

	return nil
}
