package entity

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"

	"github.com/photoprism/photoprism/pkg/dsn"
)

func TestTextDefaultColumn_DefaultValue(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		value, ok := textDefaultColumn{defaultValue: "image"}.DefaultValue()
		assert.True(t, ok)
		assert.Equal(t, "image", value)
	})
	t.Run("Empty", func(t *testing.T) {
		value, ok := textDefaultColumn{}.DefaultValue()
		assert.True(t, ok)
		assert.Equal(t, "", value)
	})
}

func TestMysqlHexDefault(t *testing.T) {
	assert.Equal(t, []string{"x'696d616765'", "696d616765"}, mysqlHexDefault.FindStringSubmatch("x'696d616765'"))
	assert.Equal(t, []string{"X'7A7A'", "7A7A"}, mysqlHexDefault.FindStringSubmatch("X'7A7A'"))
	assert.Nil(t, mysqlHexDefault.FindStringSubmatch("'zz'"))
	assert.Nil(t, mysqlHexDefault.FindStringSubmatch("x'zz'"))
}

func TestMysqlMigrator_ColumnTypes(t *testing.T) {
	if DbDialect() != dsn.DialectMySQL {
		t.Skip("requires the MariaDB test database")
	}

	t.Run("Migrator", func(t *testing.T) {
		_, ok := Db().Migrator().(mysqlMigrator)
		assert.True(t, ok)
	})
	t.Run("BinaryDefaults", func(t *testing.T) {
		columns, err := Db().Migrator().ColumnTypes(&Photo{})
		require.NoError(t, err)

		defaults := make(map[string]string)

		for _, c := range columns {
			if value, ok := c.DefaultValue(); ok {
				defaults[c.Name()] = value
			}
		}

		assert.Equal(t, "image", defaults["photo_type"])
		assert.Equal(t, "zz", defaults["place_id"])
		assert.Equal(t, "Local", defaults["time_zone"])
	})
}

// driftTest is a model whose table is created with the column types of an older release.
type driftTest struct {
	ID       uint   `gorm:"primaryKey"`
	Title    string `gorm:"size:200"`
	Serial   string `gorm:"type:bytes;size:160"`
	Label    string `gorm:"size:400"`
	Path     string `gorm:"type:bytes;size:1024"`
	TimeZone string `gorm:"type:bytes;size:64;default:'Local'"`
	Priority int    `gorm:"size:32;default:0"`
	Note     string `gorm:"type:bytes;size:16"`
	Kind     string `gorm:"size:100"`
	Slug     string `gorm:"type:bytes;size:160;default:'';"`
	Flag     bool   `gorm:"default:false;"`
	Legacy   string `gorm:"type:bytes;size:16"`
	Diff     int    `gorm:"size:32;default:-1;"`
	Words    string `gorm:"size:300"`
}

// TableName returns the entity table name.
func (driftTest) TableName() string {
	return "drift_test"
}

func TestMysqlMigrator_AlterColumn(t *testing.T) {
	if DbDialect() != dsn.DialectMySQL {
		t.Skip("requires the MariaDB test database")
	}

	t.Cleanup(func() { _ = Db().Migrator().DropTable(&driftTest{}) })
	require.NoError(t, Db().Exec("CREATE TABLE drift_test (id int unsigned AUTO_INCREMENT PRIMARY KEY, "+
		"title varchar(255), serial varbinary(255), label varbinary(755), path varbinary(500), "+
		"time_zone varbinary(64), priority int, note varbinary(16) NOT NULL, kind varbinary(50), slug varbinary(255), "+
		"flag tinyint(1), legacy varbinary(16) DEFAULT 'x', diff int unsigned, "+
		"words varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci) "+
		"ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci").Error)
	title, serial := strings.Repeat("t", 230), strings.Repeat("s", 200)
	require.NoError(t, Db().Exec("INSERT INTO drift_test (title, serial, label, path, note) VALUES (?, ?, ?, ?, '')",
		title, serial, "label", "path").Error)

	alters := func(statements []string) (result []string) {
		for _, s := range statements {
			if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(s)), "ALTER TABLE") {
				result = append(result, s)
			}
		}
		return result
	}

	first := alters(countStatements(t, func() {
		require.NoError(t, withTableOptions(Db(), Entities, false).AutoMigrate(&driftTest{}))
	}))

	t.Run("KeepsLongerColumns", func(t *testing.T) {
		for _, s := range first {
			assert.NotContains(t, s, "`title`", s)
			assert.NotContains(t, s, "`serial`", s)
			assert.NotContains(t, s, "`label`", s)
			assert.NotContains(t, s, "`kind`", s)
		}

		var row struct{ Title, Serial string }
		require.NoError(t, Db().Raw("SELECT title, serial FROM drift_test").Scan(&row).Error)
		assert.Len(t, row.Title, 230)
		assert.Len(t, row.Serial, 200)
	})
	t.Run("GrowsColumns", func(t *testing.T) {
		assert.Contains(t, strings.Join(first, "\n"), "MODIFY COLUMN `path` varbinary(1024)")
	})
	t.Run("SetsDefaults", func(t *testing.T) {
		joined := strings.Join(first, "\n")
		assert.Contains(t, joined, "ALTER COLUMN `time_zone` SET DEFAULT 'Local'")
		assert.Contains(t, joined, "ALTER COLUMN `priority` SET DEFAULT 0")
		assert.NotContains(t, joined, "MODIFY COLUMN `time_zone`")
	})
	t.Run("DefaultsOfKeptColumns", func(t *testing.T) {
		joined := strings.Join(first, "\n")
		assert.Contains(t, joined, "ALTER COLUMN `slug` SET DEFAULT ''")
		assert.NotContains(t, joined, "MODIFY COLUMN `slug`")
	})
	t.Run("Boolean", func(t *testing.T) {
		joined := strings.Join(first, "\n")
		assert.Contains(t, joined, "ALTER COLUMN `flag` SET DEFAULT false")
		assert.NotContains(t, joined, "MODIFY COLUMN `flag`")
	})
	t.Run("DropsDefault", func(t *testing.T) {
		assert.Contains(t, strings.Join(first, "\n"), "ALTER COLUMN `legacy` DROP DEFAULT")
	})
	t.Run("Unsigned", func(t *testing.T) {
		assert.NotContains(t, strings.Join(first, "\n"), "`diff`")
	})
	t.Run("GrowsWithCollation", func(t *testing.T) {
		assert.Contains(t, strings.Join(first, "\n"), "MODIFY COLUMN `words` varchar(300) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci")
		var collation string
		require.NoError(t, Db().Raw("SELECT collation_name FROM information_schema.columns WHERE table_schema = DATABASE() "+
			"AND table_name = 'drift_test' AND column_name = 'words'").Row().Scan(&collation))
		assert.Equal(t, "utf8mb4_general_ci", collation)
	})
	t.Run("AcceptsNull", func(t *testing.T) {
		assert.Contains(t, strings.Join(first, "\n"), "MODIFY COLUMN `note`")
	})
	t.Run("SecondRun", func(t *testing.T) {
		second := alters(countStatements(t, func() {
			require.NoError(t, withTableOptions(Db(), Entities, false).AutoMigrate(&driftTest{}))
		}))
		assert.Empty(t, second)
	})
}

func TestMysqlBaseType(t *testing.T) {
	t.Run("Boolean", func(t *testing.T) {
		assert.Equal(t, "tinyint", mysqlBaseType("boolean"))
		assert.Equal(t, "tinyint", mysqlBaseType("BOOL"))
	})
	t.Run("Other", func(t *testing.T) {
		assert.Equal(t, "varbinary", mysqlBaseType("VARBINARY"))
	})
}

func TestMysqlDataType(t *testing.T) {
	assert.Equal(t, []string{"varchar(160)", "varchar", "160"}, mysqlDataType.FindStringSubmatch("varchar(160)"))
	assert.Equal(t, []string{"int", "int", ""}, mysqlDataType.FindStringSubmatch("int unsigned"))
	assert.Equal(t, []string{"decimal(10,2)", "decimal", "10"}, mysqlDataType.FindStringSubmatch("decimal(10,2)"))
}

func TestMysqlDefaultDiffers(t *testing.T) {
	column := func(value string, valid bool) gorm.ColumnType {
		return migrator.ColumnType{DefaultValueValue: sql.NullString{String: value, Valid: valid}}
	}

	t.Run("NoDefault", func(t *testing.T) {
		assert.False(t, mysqlDefaultDiffers(column("", false), &schema.Field{}))
		assert.False(t, mysqlDefaultDiffers(column("NULL", true), &schema.Field{}))
		assert.True(t, mysqlDefaultDiffers(column("x", true), &schema.Field{}))
	})
	t.Run("String", func(t *testing.T) {
		field := &schema.Field{HasDefaultValue: true, DefaultValue: "'Local'"}
		assert.False(t, mysqlDefaultDiffers(column("Local", true), field))
		assert.True(t, mysqlDefaultDiffers(column("UTC", true), field))
		assert.True(t, mysqlDefaultDiffers(column("", false), field))
	})
	t.Run("Interface", func(t *testing.T) {
		field := &schema.Field{HasDefaultValue: true, DefaultValue: "0", DefaultValueInterface: int64(0)}
		assert.False(t, mysqlDefaultDiffers(column("0", true), field))
		assert.True(t, mysqlDefaultDiffers(column("1", true), field))
	})
	t.Run("Bool", func(t *testing.T) {
		field := &schema.Field{HasDefaultValue: true, DefaultValue: "false", DefaultValueInterface: false}
		assert.False(t, mysqlDefaultDiffers(column("0", true), field))
		assert.True(t, mysqlDefaultDiffers(column("1", true), field))
	})
}

func TestMysqlNegativeUnsigned(t *testing.T) {
	column := func(columnType string) gorm.ColumnType {
		return migrator.ColumnType{ColumnTypeValue: sql.NullString{String: columnType, Valid: true}}
	}

	t.Run("Unsigned", func(t *testing.T) {
		assert.True(t, mysqlNegativeUnsigned(column("int(10) unsigned"), &schema.Field{DefaultValue: "-1"}))
		assert.False(t, mysqlNegativeUnsigned(column("int(10) unsigned"), &schema.Field{DefaultValue: "0"}))
	})
	t.Run("Signed", func(t *testing.T) {
		assert.False(t, mysqlNegativeUnsigned(column("int(11)"), &schema.Field{DefaultValue: "-1"}))
	})
}
