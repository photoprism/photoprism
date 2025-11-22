package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Report(t *testing.T) {
	m := NewConfig(CliTestContext())
	r, _ := m.Report()
	assert.GreaterOrEqual(t, len(r), 1)
}

func TestConfig_ReportDatabaseSection(t *testing.T) {
	collect := func(rows [][]string) map[string]string {
		result := make(map[string]string, len(rows))

		for _, row := range rows {
			if len(row) < 2 {
				continue
			}

			result[row[0]] = row[1]
		}

		return result
	}
	t.Run("SQLiteReportsDSN", func(t *testing.T) {
		conf := NewConfig(CliTestContext())
		resetDatabaseOptions(conf)

		rows, _ := conf.Report()
		values := collect(rows)

		assert.Equal(t, SQLite3, values["database-driver"])
		assert.Equal(t, conf.DatabaseDSN(), values["database-dsn"])
		_, hasName := values["database-name"]
		assert.False(t, hasName)
	})
	t.Run("MariaDBReportsIndividualFields", func(t *testing.T) {
		conf := NewConfig(CliTestContext())
		resetDatabaseOptions(conf)

		conf.options.DatabaseDriver = MySQL
		conf.options.DatabaseServer = "db.internal:3306"
		conf.options.DatabaseName = "photoprism"
		conf.options.DatabaseUser = "app"
		conf.options.DatabasePassword = "secret"

		rows, _ := conf.Report()
		values := collect(rows)

		assert.Equal(t, MySQL, values["database-driver"])
		assert.Equal(t, "photoprism", values["database-name"])
		assert.Equal(t, "db.internal:3306", values["database-server"])
		assert.Equal(t, "db.internal", values["database-host"])
		assert.Equal(t, "3306", values["database-port"])
		assert.Equal(t, "app", values["database-user"])
		assert.Equal(t, strings.Repeat("*", len("secret")), values["database-password"])
		_, hasDSN := values["database-dsn"]
		assert.False(t, hasDSN)
	})
	t.Run("MariaDBReportsDSNWhenConfigured", func(t *testing.T) {
		conf := NewConfig(CliTestContext())
		resetDatabaseOptions(conf)

		conf.options.DatabaseDriver = MySQL
		conf.options.DatabaseDSN = "user:pass@tcp(db.internal:3306)/photoprism"

		rows, _ := conf.Report()
		values := collect(rows)

		assert.Equal(t, MySQL, values["database-driver"])
		assert.Equal(t, "user:***@tcp(db.internal:3306)/photoprism?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true&timeout=15s", values["database-dsn"])
		_, hasName := values["database-name"]
		assert.False(t, hasName)
		_, hasPassword := values["database-password"]
		assert.False(t, hasPassword)
	})
}
