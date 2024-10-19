package backup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestDatabase(t *testing.T) {
	t.Run("DatabaseNotFoundToStdOut", func(t *testing.T) {
		c := get.Config()

		if c.DatabaseDriver() != config.SQLite3 {
			t.Skip("Not executing against sqlite")
		}
		backupPath, err := filepath.Abs("./testdata/sqlite")

		if err != nil {
			t.Fatal(err)
		}

		if err = os.MkdirAll(backupPath, fs.ModeDir); err != nil {
			t.Fatal(err)
		}

		err = Database(backupPath, "", true, true, 2)

		assert.Error(t, err)

		if err = os.RemoveAll(backupPath); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("DatabaseNotFound", func(t *testing.T) {
		c := get.Config()

		if c.DatabaseDriver() != config.SQLite3 {
			t.Skip("Not executing against sqlite")
		}
		backupPath, err := filepath.Abs("./testdata/sqlite")

		if err != nil {
			t.Fatal(err)
		}

		if err = os.MkdirAll(backupPath, fs.ModeDir); err != nil {
			t.Fatal(err)
		}

		err = Database(backupPath, "", false, true, 2)

		assert.Error(t, err)

		if err = os.RemoveAll(backupPath); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("Success", func(t *testing.T) {
		c := get.Config()

		if c.DatabaseDriver() == config.SQLite3 {
			t.Skip("Executing against sqlite")
		}
		backupPath, err := filepath.Abs("./testdata/sqlite")

		if err != nil {
			t.Fatal(err)
		}

		if err = os.MkdirAll(backupPath, fs.ModeDir); err != nil {
			t.Fatal(err)
		}

		err = Database(backupPath, "", false, true, 2)

		assert.Nil(t, err)

		if err = os.RemoveAll(backupPath); err != nil {
			t.Fatal(err)
		}
	})

}
