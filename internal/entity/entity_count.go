package entity

import (
	"sync"

	"github.com/photoprism/photoprism/internal/functions"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// Count returns the number of records for a given a model and key values.
func Count(m interface{}, keys []string, values []interface{}) int {
	if m == nil || len(keys) != len(values) {
		log.Debugf("entity: invalid parameters (count records)")
		return -1
	}

	db, count := UnscopedDb(), int64(0)

	stmt := db.Model(m)

	// Assume that the caller has passed in Schema named columns and we need to translate to db name columns.
	// Even if they are db name columns, this shouldn't break things.
	mSchema, _ := schema.Parse(m, &sync.Map{}, db.NamingStrategy)
	mTableName := mSchema.Table

	// Compose where condition.
	for k := range keys {
		stmt.Where("? = ?", gorm.Expr(db.NamingStrategy.ColumnName(mTableName, keys[k])), values[k])
	}

	// Fetch count from database.
	if err := stmt.Count(&count).Error; err != nil {
		log.Debugf("entity: %s (count records)", err)
		return -1
	}

	return functions.SafeInt64toint(count)
}
