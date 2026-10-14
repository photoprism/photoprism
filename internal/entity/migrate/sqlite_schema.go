package migrate

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/dsn"
)

// sqliteLegacyTypes matches column types that older versions declared in SQLite tables and that differ
// from the current ones; varchar columns are declared as before.
var sqliteLegacyTypes = regexp.MustCompile(`(?i)\b(varbinary|mediumblob|longtext|bool)\b`)

// sqliteColumnType returns the current SQLite type for a legacy column type.
func sqliteColumnType(s string) string {
	switch t := strings.ToLower(s); {
	case strings.HasPrefix(t, "varbinary"):
		return "blob(0)"
	case t == "mediumblob":
		return "blob"
	case t == "longtext":
		return "text"
	case t == "bool":
		return "numeric"
	default:
		return s
	}
}

// sqliteIndex is an index as stored in sqlite_master.
type sqliteIndex struct {
	Name string
	SQL  string `gorm:"column:sql"`
}

// sqliteColumn is a column as returned by PRAGMA table_info.
type sqliteColumn struct {
	Name string
	Type string
}

// ConvertSQLiteSchema recreates the SQLite tables of the models whose definitions still use legacy
// column types, so that they match the current schema and AutoMigrate leaves them unchanged.
// Each table is rebuilt once in its own transaction; numbers stored in text columns become text.
func ConvertSQLiteSchema(db *gorm.DB, models map[string]any) error {
	if db == nil || db.Dialector == nil {
		return fmt.Errorf("migrate: no database connection")
	} else if dsn.DialectFromDriver(db.Name()) != dsn.DialectSQLite {
		return nil
	}

	for name, model := range models {
		var ddl string

		if err := db.Raw("SELECT sql FROM sqlite_master WHERE type = 'table' AND name = ?", name).Row().Scan(&ddl); errors.Is(err, sql.ErrNoRows) {
			// The table does not exist yet.
			continue
		} else if err != nil {
			return fmt.Errorf("migrate: %s in %s (convert schema)", err, clean.Log(name))
		} else if !sqliteLegacyTypes.MatchString(ddl) {
			continue
		}

		start := time.Now()

		if err := convertSQLiteTable(db, name, model); err != nil {
			return fmt.Errorf("migrate: %s in %s (convert schema)", err, clean.Log(name))
		}

		log.Infof("migrate: converted %s table schema [%s]", clean.Log(name), time.Since(start))
	}

	return nil
}

// convertSQLiteTable renames a table, creates it again from the model, copies the rows, and drops the
// renamed table. Columns and indexes that the model does not define are kept.
func convertSQLiteTable(db *gorm.DB, name string, model any) error {
	old := name + "__v1"

	return db.Transaction(func(tx *gorm.DB) error {
		// Index statements are read before the rename, which changes the table name they refer to.
		var indexes []sqliteIndex

		if err := tx.Raw("SELECT name, sql FROM sqlite_master WHERE type = 'index' AND tbl_name = ? AND sql IS NOT NULL", name).Scan(&indexes).Error; err != nil {
			return err
		} else if err = tx.Exec(fmt.Sprintf(`ALTER TABLE %s RENAME TO %s`, sqliteIdent(name), sqliteIdent(old))).Error; err != nil {
			return err
		}

		// Index names are unique per database, so the indexes of the renamed table go first.
		for _, index := range indexes {
			if err := tx.Exec(fmt.Sprintf(`DROP INDEX %s`, sqliteIdent(index.Name))).Error; err != nil {
				return err
			}
		}

		if err := tx.Migrator().CreateTable(model); err != nil {
			return err
		}

		var oldColumns, newColumns []sqliteColumn

		if err := tx.Raw(fmt.Sprintf("PRAGMA table_info(%s)", sqliteIdent(old))).Scan(&oldColumns).Error; err != nil {
			return err
		} else if err = tx.Raw(fmt.Sprintf("PRAGMA table_info(%s)", sqliteIdent(name))).Scan(&newColumns).Error; err != nil {
			return err
		}

		types := make(map[string]string, len(newColumns))

		// Column names are compared case-insensitively, as in SQLite.
		for _, col := range newColumns {
			types[strings.ToLower(col.Name)] = strings.ToLower(col.Type)
		}

		cols := make([]string, 0, len(oldColumns))
		values := make([]string, 0, len(oldColumns))

		for _, col := range oldColumns {
			newType, found := types[strings.ToLower(col.Name)]

			if !found {
				newType = sqliteColumnType(col.Type)

				if err := tx.Exec(fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s %s`, sqliteIdent(name), sqliteIdent(col.Name), newType)).Error; err != nil {
					return err
				}
			}

			cols = append(cols, sqliteIdent(col.Name))

			// Values that SQLite stored as numbers because of the numeric affinity of the legacy type.
			if newType == "text" || strings.HasPrefix(newType, "varchar") || strings.HasPrefix(newType, "blob") {
				values = append(values, fmt.Sprintf(`CASE WHEN typeof(%[1]s) IN ('integer', 'real') THEN CAST(%[1]s AS TEXT) ELSE %[1]s END`, sqliteIdent(col.Name)))
			} else {
				values = append(values, sqliteIdent(col.Name))
			}
		}

		if err := tx.Exec(fmt.Sprintf(`INSERT INTO %s (%s) SELECT %s FROM %s`,
			sqliteIdent(name), strings.Join(cols, ", "), strings.Join(values, ", "), sqliteIdent(old))).Error; err != nil {
			return err
		}

		// Restores the strings of path and name columns that the legacy type stored as numbers.
		if repair := sqliteRepairs[name]; repair != nil {
			if err := repair(tx, old); err != nil {
				log.Warnf("migrate: %s while restoring numeric values in %s (convert schema)", err, clean.Log(name))
			}
		}

		// Keeps the counter of an AUTOINCREMENT key, so IDs of deleted rows are not used again.
		if err := tx.Exec(`DELETE FROM sqlite_sequence WHERE name = ? AND EXISTS (SELECT 1 FROM sqlite_sequence WHERE name = ?)`, name, old).Error; err != nil {
			return err
		} else if err = tx.Exec(`UPDATE sqlite_sequence SET name = ? WHERE name = ?`, name, old).Error; err != nil {
			return err
		}

		if err := tx.Exec(fmt.Sprintf(`DROP TABLE %s`, sqliteIdent(old))).Error; err != nil {
			return err
		}

		// Creates the indexes that migrations added, after the rows, so a unique index sees all of them.
		for _, index := range indexes {
			var found int64

			if err := tx.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type = 'index' AND name = ?", index.Name).Row().Scan(&found); err != nil {
				return err
			} else if found > 0 {
				continue
			} else if err = tx.Exec(index.SQL).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// sqliteIdent returns a quoted SQLite identifier.
func sqliteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}
