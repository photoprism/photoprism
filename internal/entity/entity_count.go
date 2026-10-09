package entity

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

// Count returns the number of records of the model that match all keys, given as Go field names
// in the order ModelValues returns their values.
func Count(m any, keys []string, values []any) int {
	if m == nil || len(keys) != len(values) {
		log.Debugf("entity: invalid parameters (count records)")
		return -1
	}

	db, count := UnscopedDb(), int64(0)

	stmt := db.Model(m)

	// Compose where condition.
	for k := range keys {
		stmt = stmt.Where(fmt.Sprintf("%s = ?", gorm.ToColumnName(keys[k])), values[k])
	}

	// Fetch count from database.
	if err := stmt.Count(&count).Error; err != nil {
		log.Debugf("entity: %s (count records)", err)
		return -1
	}

	return int(count)
}
