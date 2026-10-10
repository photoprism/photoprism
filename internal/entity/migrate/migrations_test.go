package migrate

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/event"
)

// captureSystemLog redirects the system log to a buffer until the test ends.
func captureSystemLog(t *testing.T) *bytes.Buffer {
	t.Helper()

	var buf bytes.Buffer

	orig := event.SystemLog
	event.SystemLog = &logrus.Logger{
		Out:       &buf,
		Formatter: &logrus.TextFormatter{DisableTimestamp: true},
		Level:     logrus.TraceLevel,
	}
	t.Cleanup(func() { event.SystemLog = orig })

	return &buf
}

func TestMigrations_Start(t *testing.T) {
	db, err := gorm.Open("sqlite3", filepath.Join(t.TempDir(), "migrations.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	db.LogMode(false)

	require.NoError(t, db.AutoMigrate(&Migration{}).Error)
	require.NoError(t, db.Exec("CREATE TABLE start_test (id INTEGER PRIMARY KEY, name VARCHAR(16))").Error)

	// find returns the stored migration with the specified ID.
	find := func(id string) Migration {
		var m Migration
		require.NoError(t, db.Where("id = ?", id).First(&m).Error)
		return m
	}

	t.Run("Success", func(t *testing.T) {
		buf := captureSystemLog(t)
		list := Migrations{{ID: "20261010-000001", Statements: []string{"ALTER TABLE start_test ADD COLUMN added INT;"}}}
		list.Start(db, Opt(true, false, nil))

		m := find("20261010-000001")
		assert.True(t, m.Finished())
		assert.Empty(t, m.Error)
		assert.Empty(t, buf.String())
	})
	t.Run("Ignored", func(t *testing.T) {
		require.NoError(t, db.Exec("ALTER TABLE start_test ADD COLUMN existing INT").Error)

		buf := captureSystemLog(t)
		list := Migrations{{ID: "20261010-000002", Statements: []string{"ALTER TABLE start_test ADD COLUMN existing INT;"}}}
		list.Start(db, Opt(true, false, nil))

		m := find("20261010-000002")
		assert.True(t, m.Finished())
		assert.Empty(t, m.Error)
		assert.Empty(t, buf.String())
	})
	t.Run("Failed", func(t *testing.T) {
		buf := captureSystemLog(t)
		list := Migrations{
			{ID: "20261010-000003", Statements: []string{"UPDATE start_test_missing SET name = 'x';"}},
			{ID: "20261010-000004", Statements: []string{"UPDATE start_test SET name = 'y';"}},
		}
		list.Start(db, Opt(true, false, nil))

		failed := find("20261010-000003")
		assert.False(t, failed.Finished())
		assert.Contains(t, failed.Error, "no such table")
		later := find("20261010-000004")
		assert.True(t, later.Finished(), "later migrations still run")

		out := buf.String()
		assert.Contains(t, out, "level=error")
		assert.Contains(t, out, "migrate: executing 20261010-000003 failed with no such table: start_test_missing")
		assert.NotContains(t, out, "20261010-000004")
	})
}
