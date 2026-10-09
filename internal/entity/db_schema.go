package entity

import (
	"fmt"

	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/pkg/dsn"
)

// DbHasTable reports whether the database has the table, and returns query errors where the ORM's
// MySQL check panics.
func DbHasTable(db *gorm.DB, table string) (bool, error) {
	if db == nil {
		return false, fmt.Errorf("db is nil")
	}

	var stmt string

	switch dsn.DialectFromDriver(db.Dialect().GetName()) {
	case dsn.DialectMySQL:
		stmt = "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?"
	case dsn.DialectSQLite:
		stmt = "SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?"
	default:
		return false, fmt.Errorf("unsupported dialect %s", db.Dialect().GetName())
	}

	var n int64

	if err := db.Raw(stmt, table).Row().Scan(&n); err != nil {
		return false, err
	}

	return n > 0, nil
}

// DbHasColumn reports whether the table has the column, and returns query errors where the ORM's
// MySQL check panics.
func DbHasColumn(db *gorm.DB, table, column string) (bool, error) {
	if db == nil {
		return false, fmt.Errorf("db is nil")
	}

	var stmt string

	switch dsn.DialectFromDriver(db.Dialect().GetName()) {
	case dsn.DialectMySQL:
		stmt = "SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = ? AND column_name = ?"
	case dsn.DialectSQLite:
		stmt = "SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?"
	default:
		return false, fmt.Errorf("unsupported dialect %s", db.Dialect().GetName())
	}

	var n int64

	if err := db.Raw(stmt, table, column).Row().Scan(&n); err != nil {
		return false, err
	}

	return n > 0, nil
}
