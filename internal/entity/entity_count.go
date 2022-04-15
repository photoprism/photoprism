package entity

import (
	"github.com/jinzhu/gorm"
)

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
