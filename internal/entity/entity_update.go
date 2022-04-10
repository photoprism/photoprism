package entity

import (
	"fmt"
)

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
