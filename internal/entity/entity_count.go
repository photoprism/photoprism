package entity

import (
	"gorm.io/gorm"
)

// Count returns the number of records for a given a model and key values.
func Count(m interface{}, keys []string, values []interface{}) int {
	if m == nil || len(keys) != len(values) {
		log.Debugf("entity: invalid parameters (count records)")
		return -1
	}

	db, count := UnscopedDb(), int64(0)

	stmt := db.Model(m)

	// Compose where condition.
	for k := range keys {
		stmt.Where("? = ?", gorm.Expr(keys[k]), values[k])
	}

	// Fetch count from database.
	if err := stmt.Count(&count).Error; err != nil {
		log.Debugf("entity: %s (count records)", err)
		return -1
	}

	return int(count)
}
