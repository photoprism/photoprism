package entity

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

const (
	// sqliteDefaultStringSize is the size declared for text columns without a size tag.
	sqliteDefaultStringSize = 255
	// sqliteMaxSizedType is the largest size declared in a column type; larger columns are text or blob.
	sqliteMaxSizedType = 65535
)

// sqliteDialector is the SQLite dialector with column types that keep the query plans of existing schemas.
type sqliteDialector struct {
	*sqlite.Dialector
}

// Migrator returns the schema migrator of the dialector.
func (d sqliteDialector) Migrator(db *gorm.DB) gorm.Migrator {
	return sqlite.Migrator{Migrator: migrator.Migrator{Config: migrator.Config{
		DB:                          db,
		Dialector:                   d,
		CreateIndexAfterCreateTable: true,
	}}}
}

// DataTypeOf returns the column type of a field. SQLite estimates row widths from declared types when
// planning joins, so strings are varchar(n) as before, and binary columns are blob(0), which has blob
// affinity and the width estimate of the varbinary columns they replace.
func (d sqliteDialector) DataTypeOf(field *schema.Field) string {
	t := d.Dialector.DataTypeOf(field)

	switch {
	case t == "text" && field.Size <= 0:
		return fmt.Sprintf("varchar(%d)", sqliteDefaultStringSize)
	case t == "text" && field.Size <= sqliteMaxSizedType:
		return fmt.Sprintf("varchar(%d)", field.Size)
	case t == "blob" && field.Size > 0 && field.Size <= sqliteMaxSizedType:
		return "blob(0)"
	default:
		return t
	}
}
