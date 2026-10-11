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
		assert.Equal(t, mysqlTable{Name: "photos", Engine: "InnoDB", Collation: mysqlDefaultCollation}, tables[0])
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
	t.Run("Mixed", func(t *testing.T) {
		tables := []mysqlTable{{Name: "a", Collation: "utf8mb4_uca1400_ai_ci"}, {Name: "b", Collation: "utf8mb4_general_ci"}}
		assert.Equal(t, mysqlDefaultCollation, mysqlTableCollation(tables))
	})
	t.Run("NotUtf8mb4", func(t *testing.T) {
		tables := []mysqlTable{{Name: "a", Collation: "latin1_swedish_ci"}}
		assert.Equal(t, mysqlDefaultCollation, mysqlTableCollation(tables))
	})
	t.Run("InvalidName", func(t *testing.T) {
		tables := []mysqlTable{{Name: "a", Collation: "utf8mb4_bin; DROP"}}
		assert.Equal(t, mysqlDefaultCollation, mysqlTableCollation(tables))
	})
}

func TestMysqlTableOptions(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		assert.Equal(t, "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci", mysqlTableOptions(mysqlDefaultCollation))
	})
}

func TestMysqlTargetCollation(t *testing.T) {
	t.Run("MostCommon", func(t *testing.T) {
		tables := []mysqlTable{{Name: "a", Collation: "utf8mb4_uca1400_ai_ci"}, {Name: "b", Collation: "utf8mb4_uca1400_ai_ci"}, {Name: "c", Collation: mysqlDefaultCollation}}
		assert.Equal(t, "utf8mb4_uca1400_ai_ci", mysqlTargetCollation(tables))
	})
	t.Run("TiePrefersDefault", func(t *testing.T) {
		tables := []mysqlTable{{Name: "a", Collation: "utf8mb4_uca1400_ai_ci"}, {Name: "b", Collation: mysqlDefaultCollation}}
		assert.Equal(t, mysqlDefaultCollation, mysqlTargetCollation(tables))
	})
	t.Run("TieWithoutDefault", func(t *testing.T) {
		tables := []mysqlTable{{Name: "a", Collation: "utf8mb4_uca1400_ai_ci"}, {Name: "b", Collation: "utf8mb4_general_ci"}}
		assert.Equal(t, "utf8mb4_general_ci", mysqlTargetCollation(tables))
	})
	t.Run("NotUtf8mb4", func(t *testing.T) {
		tables := []mysqlTable{{Name: "a", Collation: "latin1_swedish_ci"}, {Name: "b", Collation: "latin1_swedish_ci"}}
		assert.Equal(t, mysqlDefaultCollation, mysqlTargetCollation(tables))
	})
	t.Run("NoTables", func(t *testing.T) {
		assert.Equal(t, mysqlDefaultCollation, mysqlTargetCollation(nil))
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
	t.Run("Charset", func(t *testing.T) {
		tables := []mysqlTable{{Name: "a", Engine: "InnoDB", Collation: "utf8mb3_general_ci"}, {Name: "b", Engine: "InnoDB", Collation: "utf8mb3_general_ci"}}
		warnings := captureWarnings(t, func() { warnMysqlTables(tables) })
		require.Len(t, warnings, 1)
		assert.True(t, strings.Contains(warnings[0], "collation (a, b)"), warnings[0])
	})
}
