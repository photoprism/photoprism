package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Report(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := NewConfig(CliTestContext())
		r, _ := m.Report()
		assert.GreaterOrEqual(t, len(r), 1)
	})
	t.Run("SQLiteDSN", func(t *testing.T) {
		m := NewConfig(CliTestContext())
		m.options.DatabaseDriver = SQLite3
		m.options.DatabaseDSN = "storage/testdata/test.db"
		m.options.DatabaseName = ""
		m.options.DatabaseUser = ""
		m.options.DatabasePassword = ""
		m.options.DatabaseServer = ""
		r, _ := m.Report()
		assert.GreaterOrEqual(t, len(r), 1)
		found := ""
		for _, row := range r {
			switch row[0] {
			case "database-driver":
				assert.Contains(t, row[1], SQLite3)
				found = found + ",database-driver"
			case "database-dsn":
				assert.Contains(t, row[1], "storage/testdata/test.db")
				found = found + ",database-dsn"
			case "database-dsn warning":
				assert.Fail(t, "database-dsn warning should not appear")
			}
		}
		assert.Contains(t, found, "database-driver")
		assert.Contains(t, found, "database-dsn")
	})
	t.Run("MariaDBDSN", func(t *testing.T) {
		m := NewConfig(CliTestContext())
		m.options.DatabaseDriver = MySQL
		m.options.DatabaseDSN = "foo:b@r@tcp(honeypot:1234)/baz?charset=utf8mb4,utf8&parseTime=true"
		m.options.DatabaseName = ""
		m.options.DatabaseUser = ""
		m.options.DatabasePassword = ""
		m.options.DatabaseServer = ""
		r, _ := m.Report()
		assert.GreaterOrEqual(t, len(r), 1)
		found := ""
		for _, row := range r {
			switch row[0] {
			case "database-driver":
				assert.Contains(t, row[1], MySQL)
				found = found + ",database-driver"
			case "database-dsn":
				assert.Contains(t, row[1], "foo:***@tcp(honeypot:1234)/baz?charset=utf8mb4,utf8&parseTime=true")
				found = found + ",database-dsn"
			case "database-dsn warning":
				assert.Contains(t, row[1], "database-dsn overides the following database settings")
				found = found + ",database-dsn warning"
			case "database-name":
				assert.Contains(t, row[1], "baz")
				found = found + ",database-name"
			case "database-server":
				assert.Contains(t, row[1], "honeypot:1234")
				found = found + ",database-server"
			case "database-user":
				assert.Contains(t, row[1], "foo")
				found = found + ",database-user"
			case "database-password":
				assert.Contains(t, row[1], "***")
				found = found + ",database-password"

			}
		}
		assert.Contains(t, found, "database-driver")
		assert.Contains(t, found, "database-dsn")
		assert.Contains(t, found, "database-dsn warning")
		assert.Contains(t, found, "database-name")
		assert.Contains(t, found, "database-server")
		assert.Contains(t, found, "database-user")
		assert.Contains(t, found, "database-password")
	})
}
