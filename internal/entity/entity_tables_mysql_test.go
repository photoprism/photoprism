package entity

import (
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/dsn"
)

func TestMysqlTable_SmallIndexes(t *testing.T) {
	for _, c := range []struct {
		engine, format string
		small          bool
	}{
		{"InnoDB", "Compact", true},
		{"InnoDB", "redundant", true},
		{"InnoDB", "Dynamic", false},
		{"InnoDB", "Compressed", false},
		{"InnoDB", "", false},
		{"MyISAM", "Compact", false},
	} {
		assert.Equal(t, c.small, mysqlTable{Engine: c.engine, RowFormat: c.format}.SmallIndexes(), c.engine+" "+c.format)
	}
}

func TestMysqlTable_Charset(t *testing.T) {
	t.Run("Utf8mb4", func(t *testing.T) {
		assert.Equal(t, "utf8mb4", mysqlTable{Collation: "utf8mb4_uca1400_ai_ci"}.Charset())
	})
	t.Run("Latin1", func(t *testing.T) {
		assert.Equal(t, "latin1", mysqlTable{Collation: "latin1_swedish_ci"}.Charset())
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", mysqlTable{}.Charset())
	})
}

func TestMysqlTables(t *testing.T) {
	if DbDialect() != dsn.DialectMySQL {
		t.Run("Error", func(t *testing.T) {
			_, err := mysqlTables(Db(), Entities)
			assert.Error(t, err)
		})
		return
	}

	t.Run("Success", func(t *testing.T) {
		tables, err := mysqlTables(Db(), Tables{Photo{}.TableName(): &Photo{}, "no_such_table": &Photo{}})
		require.NoError(t, err)
		require.Len(t, tables, 1)
		assert.Equal(t, mysqlTable{Name: "photos", Engine: "InnoDB", Collation: mysqlDefaultCollation, RowFormat: "Dynamic"}, tables[0])
	})
	t.Run("SchemaTables", func(t *testing.T) {
		tables, err := mysqlTables(Db(), Entities)
		require.NoError(t, err)
		require.NotEmpty(t, tables)
		for _, table := range tables {
			assert.False(t, schemaTables[table.Name], table.Name)
		}
	})
	t.Run("Empty", func(t *testing.T) {
		tables, err := mysqlTables(Db(), Tables{})
		require.NoError(t, err)
		assert.Empty(t, tables)
	})
}

func TestMysqlTableCollation(t *testing.T) {
	t.Run("NoTables", func(t *testing.T) {
		assert.Equal(t, mysqlDefaultCollation, mysqlTableCollation(nil))
	})
	t.Run("Shared", func(t *testing.T) {
		tables := []mysqlTable{{Name: "a", Collation: "utf8mb4_uca1400_ai_ci"}, {Name: "b", Collation: "utf8mb4_uca1400_ai_ci"}}
		assert.Equal(t, "utf8mb4_uca1400_ai_ci", mysqlTableCollation(tables))
	})
	t.Run("MostCommon", func(t *testing.T) {
		tables := []mysqlTable{{Name: "a", Collation: "utf8mb4_uca1400_ai_ci"}, {Name: "b", Collation: "utf8mb4_uca1400_ai_ci"}, {Name: "c", Collation: mysqlDefaultCollation}}
		assert.Equal(t, "utf8mb4_uca1400_ai_ci", mysqlTableCollation(tables))
	})
	// Ties are checked repeatedly, as the counts are visited in random order.
	t.Run("TiePrefersDefault", func(t *testing.T) {
		tables := []mysqlTable{{Name: "a", Collation: "utf8mb4_uca1400_ai_ci"}, {Name: "b", Collation: mysqlDefaultCollation}}
		for i := 0; i < 100; i++ {
			require.Equal(t, mysqlDefaultCollation, mysqlTableCollation(tables))
		}
	})
	t.Run("TieWithoutDefault", func(t *testing.T) {
		tables := []mysqlTable{{Name: "a", Collation: "utf8mb4_uca1400_ai_ci"}, {Name: "b", Collation: "utf8mb4_general_ci"},
			{Name: "c", Collation: "utf8mb4_bin"}}
		for i := 0; i < 100; i++ {
			require.Equal(t, "utf8mb4_bin", mysqlTableCollation(tables))
		}
	})
	t.Run("TiePrefersUtf8mb4", func(t *testing.T) {
		tables := []mysqlTable{{Name: "a", Collation: "latin1_swedish_ci"}, {Name: "b", Collation: "latin1_swedish_ci"},
			{Name: "c", Collation: "utf8mb4_general_ci"}, {Name: "d", Collation: "utf8mb4_general_ci"}}
		for i := 0; i < 100; i++ {
			require.Equal(t, "utf8mb4_general_ci", mysqlTableCollation(tables))
		}
	})
	t.Run("MostCommonNotUtf8mb4", func(t *testing.T) {
		tables := []mysqlTable{{Name: "a", Collation: "latin1_swedish_ci"}, {Name: "b", Collation: "latin1_swedish_ci"}, {Name: "c", Collation: "utf8mb4_general_ci"}}
		assert.Equal(t, mysqlDefaultCollation, mysqlTableCollation(tables))
	})
	t.Run("InvalidName", func(t *testing.T) {
		tables := []mysqlTable{{Name: "a", Collation: "utf8mb4_bin; DROP"}}
		assert.Equal(t, mysqlDefaultCollation, mysqlTableCollation(tables))
	})
}

func TestMysqlPreferCollation(t *testing.T) {
	t.Run("TargetIsDefault", func(t *testing.T) {
		assert.False(t, mysqlPreferCollation("utf8mb4_bin", mysqlDefaultCollation))
	})
	t.Run("Default", func(t *testing.T) {
		assert.True(t, mysqlPreferCollation(mysqlDefaultCollation, "utf8mb4_bin"))
	})
	t.Run("Utf8mb4", func(t *testing.T) {
		assert.True(t, mysqlPreferCollation("utf8mb4_general_ci", "latin1_swedish_ci"))
		assert.False(t, mysqlPreferCollation("latin1_swedish_ci", "utf8mb4_general_ci"))
	})
	t.Run("Name", func(t *testing.T) {
		assert.True(t, mysqlPreferCollation("utf8mb4_bin", "utf8mb4_general_ci"))
		assert.False(t, mysqlPreferCollation("utf8mb4_general_ci", "utf8mb4_bin"))
	})
}

func TestMysqlTableOptions(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		assert.Equal(t, "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci", mysqlTableOptions(mysqlDefaultCollation))
	})
}

// captureWarnings returns the warning messages fn logs.
func captureWarnings(t *testing.T, fn func()) (result []string) {
	t.Helper()

	logger, ok := log.(*logrus.Logger)
	require.True(t, ok)

	hook := logtest.NewLocal(logger)
	t.Cleanup(hook.Reset)

	fn()

	for _, entry := range hook.AllEntries() {
		if entry.Level == logrus.WarnLevel {
			result = append(result, entry.Message)
		}
	}

	return result
}

func TestWarnMysqlTables(t *testing.T) {
	t.Run("Shared", func(t *testing.T) {
		tables := []mysqlTable{{Name: "a", Engine: "InnoDB", Collation: "utf8mb4_uca1400_ai_ci"}, {Name: "b", Engine: "InnoDB", Collation: "utf8mb4_uca1400_ai_ci"}}
		assert.Empty(t, captureWarnings(t, func() { warnMysqlTables(tables) }))
	})
	t.Run("NoTables", func(t *testing.T) {
		assert.Empty(t, captureWarnings(t, func() { warnMysqlTables(nil) }))
	})
	t.Run("Engine", func(t *testing.T) {
		tables := []mysqlTable{{Name: "a", Engine: "MyISAM", Collation: mysqlDefaultCollation}, {Name: "b", Engine: "InnoDB", Collation: mysqlDefaultCollation}}
		warnings := captureWarnings(t, func() { warnMysqlTables(tables) })
		require.Len(t, warnings, 1)
		assert.Contains(t, warnings[0], "without InnoDB (a)")
		assert.Contains(t, warnings[0], "ENGINE=InnoDB")
	})
	t.Run("RowFormat", func(t *testing.T) {
		tables := []mysqlTable{{Name: "a", Engine: "InnoDB", Collation: mysqlDefaultCollation, RowFormat: "Compact"},
			{Name: "b", Engine: "InnoDB", Collation: mysqlDefaultCollation, RowFormat: "Dynamic"},
			{Name: "c", Engine: "InnoDB", Collation: mysqlDefaultCollation, RowFormat: "REDUNDANT"},
			{Name: "d", Engine: "InnoDB", Collation: mysqlDefaultCollation, RowFormat: "Compressed"},
			{Name: "e", Engine: "MyISAM", Collation: mysqlDefaultCollation, RowFormat: "Compact"}}
		warnings := captureWarnings(t, func() { warnMysqlTables(tables) })
		require.Len(t, warnings, 2)
		assert.Contains(t, warnings[0], "without InnoDB (e)")
		assert.Contains(t, warnings[1], "Compact or Redundant row format (a, c)")
		assert.Contains(t, warnings[1], `"ALTER TABLE <table> ROW_FORMAT=DYNAMIC;"`)
	})
	t.Run("MixedCollations", func(t *testing.T) {
		tables := []mysqlTable{{Name: "a", Engine: "InnoDB", Collation: "utf8mb4_general_ci"}, {Name: "b", Engine: "InnoDB", Collation: mysqlDefaultCollation}}
		warnings := captureWarnings(t, func() { warnMysqlTables(tables) })
		require.Len(t, warnings, 1)
		assert.Contains(t, warnings[0], "without a shared utf8mb4 collation (a)")
		assert.Contains(t, warnings[0], "CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci")
	})
	t.Run("MostCommonCollation", func(t *testing.T) {
		tables := []mysqlTable{{Name: "a", Engine: "InnoDB", Collation: "utf8mb4_general_ci"}, {Name: "b", Engine: "InnoDB", Collation: "utf8mb4_uca1400_ai_ci"},
			{Name: "c", Engine: "InnoDB", Collation: "utf8mb4_uca1400_ai_ci"}}
		warnings := captureWarnings(t, func() { warnMysqlTables(tables) })
		require.Len(t, warnings, 1)
		assert.Contains(t, warnings[0], "without a shared utf8mb4 collation (a)")
		assert.Contains(t, warnings[0], "COLLATE utf8mb4_uca1400_ai_ci;")
	})
	t.Run("FollowWarning", func(t *testing.T) {
		tables := []mysqlTable{{Name: "a", Engine: "InnoDB", Collation: "latin1_swedish_ci"}, {Name: "b", Engine: "InnoDB", Collation: "utf8mb4_general_ci"},
			{Name: "c", Engine: "InnoDB", Collation: "utf8mb4_uca1400_ai_ci"}, {Name: "d", Engine: "InnoDB", Collation: "utf8mb4_uca1400_ai_ci"}}
		warnings := captureWarnings(t, func() { warnMysqlTables(tables) })
		require.Len(t, warnings, 1)

		// Converting the tables the warning names to the collation it names leaves nothing to report.
		_, list, found := strings.Cut(warnings[0], "collation (")
		require.True(t, found)
		list, _, _ = strings.Cut(list, ")")
		_, target, found := strings.Cut(warnings[0], "COLLATE ")
		require.True(t, found)
		target, _, _ = strings.Cut(target, ";")
		assert.Equal(t, "a, b", list)
		assert.Equal(t, "utf8mb4_uca1400_ai_ci", target)
		for i := range tables {
			if strings.Contains(", "+list+", ", ", "+tables[i].Name+", ") {
				tables[i].Collation = target
			}
		}
		assert.Empty(t, captureWarnings(t, func() { warnMysqlTables(tables) }))
	})
	t.Run("Charset", func(t *testing.T) {
		tables := []mysqlTable{{Name: "a", Engine: "InnoDB", Collation: "utf8mb3_general_ci"}, {Name: "b", Engine: "InnoDB", Collation: "utf8mb3_general_ci"}}
		warnings := captureWarnings(t, func() { warnMysqlTables(tables) })
		require.Len(t, warnings, 1)
		assert.True(t, strings.Contains(warnings[0], "collation (a, b)"), warnings[0])
	})
}

func TestWarnMysqlTables_RowFormat(t *testing.T) {
	if DbDialect() != dsn.DialectMySQL {
		t.Skip("requires the MariaDB test database")
	}

	// A table with the Compact row format cannot hold the widened primary key until it is converted as the warning says.
	db := Db()
	require.NoError(t, db.Migrator().DropTable(fileSyncWidthTable))
	require.NoError(t, db.Exec("CREATE TABLE "+fileSyncWidthTable+" (remote_name VARBINARY(255) NOT NULL, "+
		"service_id INT UNSIGNED NOT NULL, status VARBINARY(16), PRIMARY KEY (remote_name, service_id)) "+
		"ENGINE=InnoDB ROW_FORMAT=COMPACT").Error)
	t.Cleanup(func() { assert.NoError(t, db.Migrator().DropTable(fileSyncWidthTable)) })

	assert.Error(t, db.AutoMigrate(&fileSyncWidthCurrent{}))

	tables, err := mysqlTables(db, Tables{fileSyncWidthTable: &fileSyncWidthCurrent{}})
	require.NoError(t, err)
	warnings := captureWarnings(t, func() { warnMysqlTables(tables) })
	require.Len(t, warnings, 1)
	assert.Contains(t, warnings[0], "row format ("+fileSyncWidthTable+")")

	_, command, found := strings.Cut(warnings[0], `run "`)
	require.True(t, found)
	command, _, _ = strings.Cut(command, `"`)
	require.NoError(t, db.Exec(strings.Replace(command, "<table>", fileSyncWidthTable, 1)).Error)

	require.NoError(t, db.AutoMigrate(&fileSyncWidthCurrent{}))
	tables, err = mysqlTables(db, Tables{fileSyncWidthTable: &fileSyncWidthCurrent{}})
	require.NoError(t, err)
	assert.Empty(t, captureWarnings(t, func() { warnMysqlTables(tables) }))
}
