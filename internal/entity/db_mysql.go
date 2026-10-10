package entity

import (
	"encoding/hex"
	"regexp"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// mysqlHexDefault matches the hexadecimal literal MariaDB reports as default of a binary column.
var mysqlHexDefault = regexp.MustCompile(`^(?i)x'([0-9a-f]*)'$`)

// mysqlDialector is the MySQL/MariaDB dialector with a migrator that reads binary defaults as text.
type mysqlDialector struct {
	*mysql.Dialector
}

// Migrator returns the schema migrator of the dialector.
func (d mysqlDialector) Migrator(db *gorm.DB) gorm.Migrator {
	return mysqlMigrator{Migrator: d.Dialector.Migrator(db).(mysql.Migrator)}
}

// mysqlMigrator compares column defaults of VARBINARY columns with their text value, as MariaDB
// reports them as hexadecimal literals and AutoMigrate would otherwise alter them on every start.
type mysqlMigrator struct {
	mysql.Migrator
}

// ColumnTypes returns the column types of a table, with binary defaults decoded.
func (m mysqlMigrator) ColumnTypes(value any) ([]gorm.ColumnType, error) {
	columns, err := m.Migrator.ColumnTypes(value)

	if err != nil {
		return columns, err
	}

	for i, c := range columns {
		if dv, ok := c.DefaultValue(); ok {
			if match := mysqlHexDefault.FindStringSubmatch(dv); match != nil {
				if b, decodeErr := hex.DecodeString(match[1]); decodeErr == nil {
					columns[i] = textDefaultColumn{columnType: c, defaultValue: string(b)}
				}
			}
		}
	}

	return columns, nil
}

// columnType is the interface embedded by textDefaultColumn, named so that its ColumnType method is promoted.
type columnType = gorm.ColumnType

// textDefaultColumn is a column type whose default value has been decoded.
type textDefaultColumn struct {
	columnType
	defaultValue string
}

// DefaultValue returns the decoded default value.
func (c textDefaultColumn) DefaultValue() (string, bool) {
	return c.defaultValue, true
}
