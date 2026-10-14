package commands

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/migrate"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestDatabaseTransferTarget(t *testing.T) {
	t.Run("MariaDB", func(t *testing.T) {
		target, err := databaseTransferTarget(newCommandContext(t, DatabaseTransferCommand,
			"--target-driver", "mariadb", "--target-server", "mariadb:3306", "--target-name", "photoprism",
			"--target-user", "photoprism", "--target-password", "insecure"))

		require.NoError(t, err)
		assert.Equal(t, config.DatabaseOptions{Driver: dsn.DriverMySQL, Server: "mariadb:3306", Name: "photoprism",
			User: "photoprism", Password: "insecure"}, target)
	})
	t.Run("DriverFromDSN", func(t *testing.T) {
		target, err := databaseTransferTarget(newCommandContext(t, DatabaseTransferCommand,
			"--target-dsn", "postgres://photoprism:insecure@postgres:5432/photoprism"))

		require.NoError(t, err)
		assert.Equal(t, dsn.DriverPostgres, target.Driver)
	})
	t.Run("RelativeSQLiteFile", func(t *testing.T) {
		abs, err := filepath.Abs("target.db")
		require.NoError(t, err)

		target, err := databaseTransferTarget(newCommandContext(t, DatabaseTransferCommand,
			"--target-driver", "sqlite", "--target-dsn", "target.db?_busy_timeout=5000"))

		require.NoError(t, err)
		assert.Equal(t, dsn.DriverSQLite3, target.Driver)
		assert.Equal(t, abs+"?_busy_timeout=5000", target.DSN)
	})
	t.Run("SQLiteFileURI", func(t *testing.T) {
		target, err := databaseTransferTarget(newCommandContext(t, DatabaseTransferCommand,
			"--target-driver", "sqlite", "--target-dsn", "file:target.db"))

		require.NoError(t, err)
		assert.Equal(t, "file:target.db", target.DSN)
	})
	t.Run("PasswordFile", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "password")
		require.NoError(t, os.WriteFile(fileName, []byte("from-file\n"), fs.ModeSecretFile))

		target, err := databaseTransferTarget(newCommandContext(t, DatabaseTransferCommand,
			"--target-driver", "postgres", "--target-password-file", fileName))

		require.NoError(t, err)
		assert.Equal(t, "from-file", target.Password)
	})
	t.Run("PasswordTakesPrecedence", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "password")
		require.NoError(t, os.WriteFile(fileName, []byte("from-file"), fs.ModeSecretFile))

		target, err := databaseTransferTarget(newCommandContext(t, DatabaseTransferCommand,
			"--target-driver", "postgres", "--target-password", "from-flag", "--target-password-file", fileName))

		require.NoError(t, err)
		assert.Equal(t, "from-flag", target.Password)
	})
	t.Run("MissingPasswordFile", func(t *testing.T) {
		_, err := databaseTransferTarget(newCommandContext(t, DatabaseTransferCommand,
			"--target-driver", "postgres", "--target-password-file", filepath.Join(t.TempDir(), "missing")))

		assert.ErrorContains(t, err, "target password file")
	})
	t.Run("PasswordFileNotUsed", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "missing")

		target, err := databaseTransferTarget(newCommandContext(t, DatabaseTransferCommand,
			"--target-driver", "sqlite", "--target-dsn", "target.db", "--target-password-file", missing))
		require.NoError(t, err)
		assert.Equal(t, "", target.Password)

		target, err = databaseTransferTarget(newCommandContext(t, DatabaseTransferCommand,
			"--target-dsn", "postgres://photoprism:secret@postgres:5432/photoprism", "--target-password-file", missing))
		require.NoError(t, err)
		assert.Equal(t, "", target.Password)
	})
	t.Run("SQLitePassword", func(t *testing.T) {
		target, err := databaseTransferTarget(newCommandContext(t, DatabaseTransferCommand,
			"--target-driver", "sqlite", "--target-dsn", "target.db", "--target-password", "ignored"))

		require.NoError(t, err)
		assert.Equal(t, "", target.Password)
	})
	t.Run("DSNWithServer", func(t *testing.T) {
		_, err := databaseTransferTarget(newCommandContext(t, DatabaseTransferCommand,
			"--target-dsn", "postgres://photoprism@postgres:5432/photoprism", "--target-password", "secret"))

		assert.ErrorContains(t, err, "cannot be combined")
	})
	t.Run("SQLiteWithName", func(t *testing.T) {
		_, err := databaseTransferTarget(newCommandContext(t, DatabaseTransferCommand,
			"--target-driver", "sqlite", "--target-name", "target.db"))

		assert.ErrorContains(t, err, "--target-dsn only")
	})
	t.Run("MissingDriver", func(t *testing.T) {
		_, err := databaseTransferTarget(newCommandContext(t, DatabaseTransferCommand, "--target-name", "photoprism"))

		assert.ErrorContains(t, err, "--target-driver")
	})
	t.Run("UnsupportedDriver", func(t *testing.T) {
		_, err := databaseTransferTarget(newCommandContext(t, DatabaseTransferCommand, "--target-driver", "oracle"))

		assert.ErrorContains(t, err, "unsupported target database driver")
	})
}

func TestDatabaseTransferCommand(t *testing.T) {
	var photos int64
	require.NoError(t, entity.UnscopedDb().Model(&entity.Photo{}).Count(&photos).Error)
	require.Positive(t, photos)

	// countTargetRows returns the number of photos and version rows in a SQLite target database.
	countTargetRows := func(t *testing.T, fileName string) (photos, versions int64) {
		t.Helper()

		db, err := entity.OpenDb(dsn.DriverSQLite3, fileName)
		require.NoError(t, err)
		defer func() { _ = entity.CloseDb(db) }()

		require.NoError(t, db.Model(&entity.Photo{}).Unscoped().Count(&photos).Error)
		require.NoError(t, db.Model(&migrate.Version{}).Where("version = ?", get.Config().Version()).Count(&versions).Error)

		return photos, versions
	}

	fileName := filepath.Join(t.TempDir(), "target.db")
	args := []string{"transfer", "--target-driver", "sqlite", "--target-dsn", fileName, "--batch", "13"}

	t.Run("SQLite", func(t *testing.T) {
		_, err := RunWithTestContext(DatabaseTransferCommand, args)

		require.NoError(t, err)

		targetPhotos, versions := countTargetRows(t, fileName)
		assert.Equal(t, photos, targetPhotos)
		assert.Equal(t, int64(1), versions)
	})
	t.Run("NotEmpty", func(t *testing.T) {
		_, err := RunWithTestContext(DatabaseTransferCommand, args)

		require.Error(t, err)
		assert.Equal(t, 2, ExitCode(err))
		assert.Contains(t, err.Error(), "not empty")
		assert.Contains(t, err.Error(), "already been started")
	})
	t.Run("Force", func(t *testing.T) {
		_, err := RunWithTestContext(DatabaseTransferCommand, append(args, "--force"))

		require.NoError(t, err)

		targetPhotos, _ := countTargetRows(t, fileName)
		assert.Equal(t, photos, targetPhotos)
	})
	t.Run("OtherTables", func(t *testing.T) {
		otherFile := filepath.Join(t.TempDir(), "other.db")
		db, err := entity.OpenDb(dsn.DriverSQLite3, otherFile)
		require.NoError(t, err)
		require.NoError(t, db.Exec("CREATE TABLE other (id INTEGER)").Error)
		require.NoError(t, entity.CloseDb(db))

		_, err = RunWithTestContext(DatabaseTransferCommand, []string{"transfer", "--target-driver", "sqlite", "--target-dsn", otherFile})

		require.Error(t, err)
		assert.Equal(t, 2, ExitCode(err))
		assert.Contains(t, err.Error(), "not empty (other)")
	})
	t.Run("FailedMigration", func(t *testing.T) {
		m := failedTestMigration(t)

		_, err := RunWithTestContext(DatabaseTransferCommand, []string{"transfer", "--target-driver", "sqlite",
			"--target-dsn", filepath.Join(t.TempDir(), "target.db")})

		require.Error(t, err)
		assert.Equal(t, 1, ExitCode(err))
		assert.Contains(t, err.Error(), m.ID)
	})
	t.Run("FailedMigrationAllowed", func(t *testing.T) {
		failedTestMigration(t)
		targetFile := filepath.Join(t.TempDir(), "target.db")

		_, err := RunWithTestContext(DatabaseTransferCommand, []string{"transfer", "--allow-failed-migrations",
			"--target-driver", "sqlite", "--target-dsn", targetFile})

		require.NoError(t, err)

		targetPhotos, _ := countTargetRows(t, targetFile)
		assert.Equal(t, photos, targetPhotos)
	})
	t.Run("SameDatabase", func(t *testing.T) {
		conf := get.Config()

		_, err := RunWithTestContext(DatabaseTransferCommand, []string{"transfer", "--force",
			"--target-driver", conf.DatabaseDriver(), "--target-dsn", conf.DatabaseDSN()})

		require.Error(t, err)
		assert.Equal(t, 2, ExitCode(err))
		assert.Contains(t, err.Error(), "the target is the index database")

		var after int64
		require.NoError(t, entity.UnscopedDb().Model(&entity.Photo{}).Count(&after).Error)
		assert.Equal(t, photos, after)
	})
	t.Run("MissingDriver", func(t *testing.T) {
		_, err := RunWithTestContext(DatabaseTransferCommand, []string{"transfer"})

		require.Error(t, err)
		assert.Equal(t, 2, ExitCode(err))
	})
	t.Run("Argument", func(t *testing.T) {
		_, err := RunWithTestContext(DatabaseTransferCommand, []string{"transfer", "--target-driver", "sqlite", "target.db"})

		require.Error(t, err)
		assert.Equal(t, 2, ExitCode(err))
	})
	t.Run("Batch", func(t *testing.T) {
		_, err := RunWithTestContext(DatabaseTransferCommand, []string{"transfer", "--target-driver", "sqlite", "--batch", "0"})

		require.Error(t, err)
		assert.Equal(t, 2, ExitCode(err))
	})
}

func TestMigrationsTransferCommand(t *testing.T) {
	var found *cli.Command

	for _, cmd := range MigrationsCommands.Subcommands {
		if cmd.Name == "transfer" {
			found = cmd
		}
	}

	require.NotNil(t, found)
	assert.True(t, found.Hidden)
	assert.Equal(t, actionName(DatabaseTransferCommand), actionName(found))
	assert.Equal(t, flagNames(DatabaseTransferCommand.Flags), flagNames(found.Flags))
}

// flagNames returns the names of the flags, except the help flag that the CLI adds when it runs a command.
func flagNames(flags []cli.Flag) (names []string) {
	for _, f := range flags {
		if f != cli.HelpFlag {
			names = append(names, f.Names()...)
		}
	}

	return names
}

func TestDatabaseTransferFlags(t *testing.T) {
	flags := databaseTransferFlags()

	assert.NotSame(t, &flags[0], &databaseTransferFlags()[0])
	assert.Equal(t, []string{"target-driver", "target-dsn", "target-server", "target-name", "target-user", "target-password",
		"target-password-file", "force", "f", "allow-failed-migrations", "batch", "b"}, flagNames(flags))
}

// failedTestMigration records the first migration of the test database's dialect as failed until the test ends,
// and skips the test if the dialect has none.
func failedTestMigration(t *testing.T) migrate.Migration {
	t.Helper()

	db := entity.Db()
	known := migrate.Dialects[dsn.ParseDriver(db.Name())]

	if len(known) == 0 {
		t.Skip("the dialect of the test database has no migrations")
	}

	var m migrate.Migration
	require.NoError(t, db.Where("id = ?", known[0].ID).First(&m).Error)
	require.NoError(t, db.Model(&m).UpdateColumn("error", "test").Error)
	t.Cleanup(func() { entity.Db().Model(&m).UpdateColumn("error", "") })

	return m
}

func TestFailedMigrations(t *testing.T) {
	db := entity.Db()

	t.Run("None", func(t *testing.T) {
		ids, err := failedMigrations(db)

		require.NoError(t, err)
		assert.Empty(t, ids)
	})
	t.Run("Unknown", func(t *testing.T) {
		m := migrate.Migration{ID: "29991231-000001", Dialect: db.Name(), Error: "test", StartedAt: time.Now().UTC()}
		require.NoError(t, db.Create(&m).Error)
		t.Cleanup(func() { db.Delete(&m) })

		ids, err := failedMigrations(db)

		require.NoError(t, err)
		assert.Empty(t, ids)
	})
	t.Run("Known", func(t *testing.T) {
		m := failedTestMigration(t)

		ids, err := failedMigrations(db)

		require.NoError(t, err)
		assert.Equal(t, []string{m.ID}, ids)
	})
}

func TestRemoveSQLiteFiles(t *testing.T) {
	t.Run("Files", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "target.db")
		other := filepath.Join(filepath.Dir(fileName), "other.db")

		for _, name := range []string{fileName, fileName + "-journal", fileName + "-wal", other} {
			require.NoError(t, os.WriteFile(name, []byte("x"), fs.ModeFile))
		}

		removeSQLiteFiles(fileName)

		assert.NoFileExists(t, fileName)
		assert.NoFileExists(t, fileName+"-journal")
		assert.NoFileExists(t, fileName+"-wal")
		assert.FileExists(t, other)
	})
	t.Run("NotRegular", func(t *testing.T) {
		dir := t.TempDir()
		link := filepath.Join(dir, "link.db")
		folder := filepath.Join(dir, "folder.db")
		require.NoError(t, os.Symlink(filepath.Join(dir, "missing.db"), link))
		require.NoError(t, os.Mkdir(folder, fs.ModeDir))

		removeSQLiteFiles(link)
		removeSQLiteFiles(folder)

		_, linkErr := os.Lstat(link)
		assert.NoError(t, linkErr)
		assert.DirExists(t, folder)
	})
	t.Run("Empty", func(t *testing.T) {
		removeSQLiteFiles("")
	})
}
