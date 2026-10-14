package entity

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// mysqlHexDefault matches the hexadecimal literal MariaDB reports as default of a binary column.
var mysqlHexDefault = regexp.MustCompile(`^(?i)x'([0-9a-f]*)'$`)

// mysqlDataType matches the base name and length of a column type, e.g. varchar(160).
var mysqlDataType = regexp.MustCompile(`^([a-z]+)(?:\((\d+)[^)]*\))?`)

// mysqlDialector is the MySQL/MariaDB dialector with a migrator that keeps existing columns compatible.
type mysqlDialector struct {
	*mysql.Dialector
}

// Migrator returns the schema migrator of the dialector.
func (d mysqlDialector) Migrator(db *gorm.DB) gorm.Migrator {
	return mysqlMigrator{Migrator: d.Dialector.Migrator(db).(mysql.Migrator)}
}

// mysqlMigrator reads binary defaults as text, as MariaDB reports them as hexadecimal literals, and
// alters existing columns only in ways that cannot truncate or convert their values.
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

// mysqlBaseType returns the base name of a column type, with boolean as tinyint.
func mysqlBaseType(name string) string {
	if name = strings.ToLower(name); name == "bool" || name == "boolean" {
		return "tinyint"
	}

	return name
}

// AlterColumn changes a column to match its field, but keeps a column that is longer or has a different base
// type, as databases created by older releases may hold values that would be truncated or converted.
// A differing default is set without rebuilding the column.
func (m mysqlMigrator) AlterColumn(value any, name string) error {
	return m.RunWithValue(value, func(stmt *gorm.Statement) error {
		if stmt.Schema == nil {
			return m.Migrator.AlterColumn(value, name)
		}

		field := stmt.Schema.LookUpField(name)

		if field == nil {
			return m.Migrator.AlterColumn(value, name)
		}

		columns, err := m.ColumnTypes(value)

		if err != nil {
			return err
		}

		var current gorm.ColumnType

		for _, c := range columns {
			if strings.EqualFold(c.Name(), field.DBName) {
				current = c
				break
			}
		}

		if current == nil {
			return m.Migrator.AlterColumn(value, name)
		}

		want := strings.ToLower(m.Dialector.DataTypeOf(field))
		match := mysqlDataType.FindStringSubmatch(want)
		haveLen, hasLen := current.Length()

		if match == nil {
			return m.Migrator.AlterColumn(value, name)
		}

		wantLen, _ := strconv.ParseInt(match[2], 10, 64)

		switch {
		case hasLen && wantLen > haveLen && mysqlBaseType(match[1]) == mysqlBaseType(current.DatabaseTypeName()):
			return m.growColumn(stmt, field)
		case mysqlBaseType(match[1]) != mysqlBaseType(current.DatabaseTypeName()):
			log.Infof("migrate: kept type %s of %s.%s, which differs from %s", strings.ToLower(current.DatabaseTypeName()),
				stmt.Table, field.DBName, want)
		case hasLen && wantLen > 0 && wantLen < haveLen:
			log.Infof("migrate: kept length %d of %s.%s, which exceeds %s", haveLen, stmt.Table, field.DBName, want)
		}

		// Sets a differing default in place, as changing it with MODIFY COLUMN may rebuild the table.
		// A default the column cannot hold, e.g. -1 in an unsigned column, is logged and left for a migration.
		if !mysqlDefaultDiffers(current, field) {
			// Default matches.
		} else if mysqlNegativeUnsigned(current, field) {
			log.Infof("migrate: kept unsigned %s.%s, which cannot have the default %s", stmt.Table, field.DBName, field.DefaultValue)
		} else if err = m.setDefault(stmt, field); err != nil {
			log.Warnf("migrate: %s while setting the default of %s.%s", err, stmt.Table, field.DBName)
		}

		// Kept columns are not changed otherwise.
		if mysqlBaseType(match[1]) != mysqlBaseType(current.DatabaseTypeName()) || hasLen && wantLen > 0 && wantLen < haveLen {
			return nil
		}

		// Columns that do not accept NULL although the field does are still changed.
		if nullable, ok := current.Nullable(); ok && !nullable && !field.NotNull && !field.PrimaryKey {
			return m.Migrator.AlterColumn(value, name)
		}

		return nil
	})
}

// mysqlDefaultDiffers reports whether the default of a column differs from the default of its field.
func mysqlDefaultDiffers(current gorm.ColumnType, field *schema.Field) bool {
	want := field.HasDefaultValue && (field.DefaultValueInterface != nil ||
		field.DefaultValue != "" && !strings.EqualFold(field.DefaultValue, "NULL"))
	have, ok := current.DefaultValue()

	if ok && strings.EqualFold(have, "NULL") {
		ok = false
	}

	switch {
	case !want:
		return ok
	case !ok:
		return true
	case field.DefaultValueInterface != nil:
		if b, isBool := field.DefaultValueInterface.(bool); isBool {
			v, _ := strconv.ParseBool(have)
			return v != b
		}

		return have != fmt.Sprint(field.DefaultValueInterface)
	default:
		return have != field.DefaultValue && have != strings.Trim(field.DefaultValue, "'\"")
	}
}

// mysqlNegativeUnsigned reports whether the default of a field is negative while its column is unsigned.
func mysqlNegativeUnsigned(current gorm.ColumnType, field *schema.Field) bool {
	columnType, ok := current.ColumnType()

	return ok && strings.Contains(strings.ToLower(columnType), "unsigned") &&
		strings.HasPrefix(strings.Trim(field.DefaultValue, "'\" "), "-")
}

// mysqlCharsetName matches the character set and collation names that a column definition may repeat.
var mysqlCharsetName = regexp.MustCompile(`^[a-z0-9_]+$`)

// growColumn lengthens a column to the size of its field and keeps its character set and collation,
// which MODIFY COLUMN would otherwise reset to the table default.
func (m mysqlMigrator) growColumn(stmt *gorm.Statement, field *schema.Field) error {
	var charset struct {
		Charset   string
		Collation string
	}

	if err := m.DB.Raw("SELECT COALESCE(character_set_name, '') AS charset, COALESCE(collation_name, '') AS collation "+
		"FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = ? AND column_name = ?",
		stmt.Table, field.DBName).Scan(&charset).Error; err != nil {
		return err
	}

	def := m.FullDataTypeOf(field)

	if mysqlCharsetName.MatchString(charset.Charset) && mysqlCharsetName.MatchString(charset.Collation) {
		if end := strings.Index(def.SQL, ")"); end > 0 {
			def.SQL = def.SQL[:end+1] + " CHARACTER SET " + charset.Charset + " COLLATE " + charset.Collation + def.SQL[end+1:]
		}
	}

	return m.DB.Exec("ALTER TABLE ? MODIFY COLUMN ? ?", m.CurrentTable(stmt), clause.Column{Name: field.DBName}, def).Error
}

// setDefault sets or drops the default value of a column as defined by its field.
func (m mysqlMigrator) setDefault(stmt *gorm.Statement, field *schema.Field) error {
	table, column := m.CurrentTable(stmt), clause.Column{Name: field.DBName}

	switch {
	case !field.HasDefaultValue || field.DefaultValueInterface == nil && (field.DefaultValue == "" || strings.EqualFold(field.DefaultValue, "NULL")):
		return m.DB.Exec("ALTER TABLE ? ALTER COLUMN ? DROP DEFAULT", table, column).Error
	case field.DefaultValueInterface != nil:
		return m.DB.Exec(fmt.Sprintf("ALTER TABLE ? ALTER COLUMN ? SET DEFAULT %s",
			m.Dialector.Explain("?", field.DefaultValueInterface)), table, column).Error
	default:
		return m.DB.Exec("ALTER TABLE ? ALTER COLUMN ? SET DEFAULT "+field.DefaultValue, table, column).Error
	}
}
