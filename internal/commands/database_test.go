package commands

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/migrate"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestDatabaseCommands(t *testing.T) {
	t.Run("Registered", func(t *testing.T) {
		pos := map[*cli.Command]int{}

		for i, cmd := range PhotoPrism {
			pos[cmd] = i
		}

		require.Contains(t, pos, DatabaseCommands)
		assert.Equal(t, pos[MigrationsCommands]+1, pos[DatabaseCommands])
		assert.Equal(t, pos[DatabaseCommands]+1, pos[BackupCommand])
	})
	t.Run("Alias", func(t *testing.T) {
		assert.Equal(t, "database", DatabaseCommands.Name)
		assert.Equal(t, []string{"db"}, DatabaseCommands.Aliases)
	})
	t.Run("Subcommands", func(t *testing.T) {
		names := make([]string, 0, len(DatabaseCommands.Subcommands))

		for _, cmd := range DatabaseCommands.Subcommands {
			names = append(names, cmd.Name)
		}

		assert.Equal(t, []string{"backup", "restore", "transfer"}, names)
	})
	t.Run("HiddenPathAlias", func(t *testing.T) {
		for _, cmd := range []*cli.Command{DatabaseBackupCommand, DatabaseRestoreCommand} {
			var visible, hidden []string

			for _, f := range cmd.VisibleFlags() {
				visible = append(visible, f.Names()...)
			}

			for _, f := range cmd.Flags {
				if vf, ok := f.(cli.VisibleFlag); ok && !vf.IsVisible() {
					hidden = append(hidden, f.Names()...)
				}
			}

			assert.Contains(t, visible, "path", cmd.Name)
			assert.NotContains(t, visible, "database-path", cmd.Name)
			assert.Equal(t, []string{"database-path", "index-path"}, hidden, cmd.Name)
		}
	})
}

func TestDatabasePathFlags(t *testing.T) {
	flags := databasePathFlags()

	require.Len(t, flags, 2)
	assert.Equal(t, []string{"path"}, flags[0].Names())
	assert.Equal(t, []string{"database-path", "index-path"}, flags[1].Names())
	assert.True(t, flags[0].(cli.VisibleFlag).IsVisible())
	assert.False(t, flags[1].(cli.VisibleFlag).IsVisible())
}

func TestDatabasePathArg(t *testing.T) {
	dir := t.TempDir()
	other := t.TempDir()

	t.Run("Path", func(t *testing.T) {
		assert.Equal(t, dir, databasePathArg(newCommandContext(t, DatabaseBackupCommand, "--path", dir)))
	})
	t.Run("Alias", func(t *testing.T) {
		assert.Equal(t, dir, databasePathArg(newCommandContext(t, DatabaseBackupCommand, "--database-path", dir)))
	})
	t.Run("PathTakesPrecedence", func(t *testing.T) {
		assert.Equal(t, dir, databasePathArg(newCommandContext(t, DatabaseRestoreCommand, "--database-path", other, "--path", dir)))
	})
	t.Run("Relative", func(t *testing.T) {
		abs, err := filepath.Abs("backup")
		require.NoError(t, err)
		assert.Equal(t, abs, databasePathArg(newCommandContext(t, DatabaseRestoreCommand, "--path", "backup")))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", databasePathArg(newCommandContext(t, DatabaseRestoreCommand)))
	})
}

func TestDatabaseBackupArgs(t *testing.T) {
	dir := t.TempDir()

	t.Run("NoArgs", func(t *testing.T) {
		path, fileName, err := databaseBackupArgs(newCommandContext(t, DatabaseBackupCommand))

		require.NoError(t, err)
		assert.Equal(t, "", path)
		assert.Equal(t, "", fileName)
	})
	t.Run("RelativeInPath", func(t *testing.T) {
		path, fileName, err := databaseBackupArgs(newCommandContext(t, DatabaseBackupCommand, "--path", dir, "index.sql"))

		require.NoError(t, err)
		assert.Equal(t, dir, path)
		assert.Equal(t, filepath.Join(dir, "index.sql"), fileName)
	})
	t.Run("RelativeWithoutPath", func(t *testing.T) {
		path, fileName, err := databaseBackupArgs(newCommandContext(t, DatabaseRestoreCommand, "index.sql"))

		require.NoError(t, err)
		assert.Equal(t, "", path)
		assert.Equal(t, "index.sql", fileName)
	})
	t.Run("Stdout", func(t *testing.T) {
		path, fileName, err := databaseBackupArgs(newCommandContext(t, DatabaseBackupCommand, "--path", dir, "-"))

		require.NoError(t, err)
		assert.Equal(t, dir, path)
		assert.Equal(t, "-", fileName)
	})
	t.Run("AbsoluteWithPath", func(t *testing.T) {
		_, _, err := databaseBackupArgs(newCommandContext(t, DatabaseRestoreCommand, "--path", dir, filepath.Join(dir, "index.sql")))

		require.Error(t, err)
		assert.Equal(t, 2, ExitCode(err))
	})
	t.Run("OutsidePath", func(t *testing.T) {
		_, _, err := databaseBackupArgs(newCommandContext(t, DatabaseBackupCommand, "--path", dir, "../index.sql"))

		require.Error(t, err)
		assert.Equal(t, 2, ExitCode(err))
	})
	t.Run("Subfolder", func(t *testing.T) {
		_, fileName, err := databaseBackupArgs(newCommandContext(t, DatabaseBackupCommand, "--path", dir, "sub/../sub/index.sql"))

		require.NoError(t, err)
		assert.Equal(t, filepath.Join(dir, "sub", "index.sql"), fileName)
	})
	t.Run("EmptyPath", func(t *testing.T) {
		_, _, err := databaseBackupArgs(newCommandContext(t, DatabaseRestoreCommand, "--path", ""))

		require.Error(t, err)
		assert.Equal(t, 2, ExitCode(err))
	})
	t.Run("EmptyPathAlias", func(t *testing.T) {
		_, _, err := databaseBackupArgs(newCommandContext(t, DatabaseBackupCommand, "--database-path="))

		require.Error(t, err)
		assert.Equal(t, 2, ExitCode(err))
	})
	t.Run("Empty", func(t *testing.T) {
		_, _, err := databaseBackupArgs(newCommandContext(t, DatabaseRestoreCommand, ""))

		require.Error(t, err)
		assert.Equal(t, 2, ExitCode(err))
	})
	t.Run("TrailingFlag", func(t *testing.T) {
		_, _, err := databaseBackupArgs(newCommandContext(t, DatabaseBackupCommand, "index.sql", "--force"))

		require.Error(t, err)
		assert.Equal(t, 2, ExitCode(err))
	})
}

func TestDatabaseBackupCommand(t *testing.T) {
	t.Run("FileNameInPath", func(t *testing.T) {
		dir := t.TempDir()

		_, err := RunWithTestContext(DatabaseBackupCommand, []string{"backup", "--path", dir, "index.sql"})

		require.NoError(t, err)
		assert.True(t, fs.FileExistsNotEmpty(filepath.Join(dir, "index.sql")))
	})
	t.Run("NewPath", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "new", "backup")

		_, err := RunWithTestContext(DatabaseBackupCommand, []string{"backup", "--path", dir})

		require.NoError(t, err)
		assert.True(t, fs.FileExistsNotEmpty(filepath.Join(dir, time.Now().UTC().Format("2006-01-02")+".sql")))
	})
	t.Run("FileName", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "index.sql")

		_, err := RunWithTestContext(DatabaseBackupCommand, []string{"backup", fileName})

		require.NoError(t, err)
		assert.True(t, fs.FileExistsNotEmpty(fileName))
	})
	t.Run("Path", func(t *testing.T) {
		dir := t.TempDir()

		_, err := RunWithTestContext(DatabaseBackupCommand, []string{"backup", "--path", dir})

		require.NoError(t, err)
		assert.True(t, fs.FileExistsNotEmpty(filepath.Join(dir, time.Now().UTC().Format("2006-01-02")+".sql")))
	})
	t.Run("PathAlias", func(t *testing.T) {
		dir := t.TempDir()

		_, err := RunWithTestContext(DatabaseBackupCommand, []string{"backup", "--database-path", dir})

		require.NoError(t, err)
		assert.True(t, fs.FileExistsNotEmpty(filepath.Join(dir, time.Now().UTC().Format("2006-01-02")+".sql")))
	})
	t.Run("ExistingFile", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "index.sql")
		require.NoError(t, os.WriteFile(fileName, []byte("keep"), fs.ModeFile))

		_, err := RunWithTestContext(DatabaseBackupCommand, []string{"backup", fileName})

		require.Error(t, err)
		assert.Equal(t, 1, ExitCode(err))

		data, readErr := os.ReadFile(fileName)
		require.NoError(t, readErr)
		assert.Equal(t, "keep", string(data))
	})
	t.Run("TrailingFlag", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "index.sql")

		_, err := RunWithTestContext(DatabaseBackupCommand, []string{"backup", fileName, "--force"})

		require.Error(t, err)
		assert.Equal(t, 2, ExitCode(err))
		assert.Contains(t, err.Error(), "--force")
		assert.False(t, fs.FileExists(fileName))
	})
}

// restoreTestPhoto is the fixture whose title restore tests change after a backup.
const restoreTestPhoto = "ps6sg6be2lvl0yh7"

// prepareRestoreTest removes the schema version rows, so that only the migration after a restore records
// one, and returns a function that changes a photo title the restore must revert and one that checks both.
func prepareRestoreTest(t *testing.T) (change, check func()) {
	t.Helper()

	var photo entity.Photo
	require.NoError(t, entity.UnscopedDb().Where("photo_uid = ?", restoreTestPhoto).Take(&photo).Error)
	require.NoError(t, entity.UnscopedDb().Where("1 = 1").Delete(&migrate.Version{}).Error)

	title := photo.PhotoTitle

	t.Cleanup(func() {
		entity.UnscopedDb().Model(&entity.Photo{}).Where("photo_uid = ?", restoreTestPhoto).UpdateColumn("photo_title", title)
	})

	change = func() {
		require.NoError(t, entity.UnscopedDb().Model(&entity.Photo{}).Where("photo_uid = ?", restoreTestPhoto).
			UpdateColumn("photo_title", "Changed After Backup").Error)
	}

	check = func() {
		var restored entity.Photo
		require.NoError(t, entity.UnscopedDb().Where("photo_uid = ?", restoreTestPhoto).Take(&restored).Error)
		assert.Equal(t, title, restored.PhotoTitle)

		var versions int64
		require.NoError(t, entity.UnscopedDb().Model(&migrate.Version{}).Where("version = ?", get.Config().Version()).Count(&versions).Error)
		assert.Equal(t, int64(1), versions)
	}

	return change, check
}

func TestDatabaseRestoreCommand(t *testing.T) {
	t.Run("RoundTrip", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "index.sql")
		change, check := prepareRestoreTest(t)

		_, err := RunWithTestContext(DatabaseBackupCommand, []string{"backup", fileName})
		require.NoError(t, err)

		change()

		_, err = RunWithTestContext(DatabaseRestoreCommand, []string{"restore", "--force", fileName})
		require.NoError(t, err)

		check()
	})
	t.Run("FileNameInPath", func(t *testing.T) {
		dir := t.TempDir()
		change, check := prepareRestoreTest(t)

		_, err := RunWithTestContext(DatabaseBackupCommand, []string{"backup", "--path", dir, "index.sql"})
		require.NoError(t, err)

		change()

		_, err = RunWithTestContext(DatabaseRestoreCommand, []string{"restore", "--force", "--path", dir, "index.sql"})
		require.NoError(t, err)

		check()
	})
	t.Run("Path", func(t *testing.T) {
		dir := t.TempDir()
		change, check := prepareRestoreTest(t)

		_, err := RunWithTestContext(DatabaseBackupCommand, []string{"backup", "--path", dir})
		require.NoError(t, err)

		change()

		_, err = RunWithTestContext(DatabaseRestoreCommand, []string{"restore", "--force", "--index-path", dir})
		require.NoError(t, err)

		check()
	})
	t.Run("ExistingIndex", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "index.sql")
		require.NoError(t, os.WriteFile(fileName, []byte("SELECT 1;\n"), fs.ModeFile))

		_, err := RunWithTestContext(DatabaseRestoreCommand, []string{"restore", fileName})

		require.Error(t, err)
		assert.Equal(t, 1, ExitCode(err))
		assert.Contains(t, err.Error(), "found an existing index")
	})
	t.Run("NoBackup", func(t *testing.T) {
		_, err := RunWithTestContext(DatabaseRestoreCommand, []string{"restore", "--path", t.TempDir()})

		require.Error(t, err)
		assert.Equal(t, 1, ExitCode(err))
		assert.Contains(t, err.Error(), "failed to find a backup")
	})
	t.Run("MissingFile", func(t *testing.T) {
		_, err := RunWithTestContext(DatabaseRestoreCommand, []string{"restore", filepath.Join(t.TempDir(), "missing.sql")})

		require.Error(t, err)
		assert.Equal(t, 1, ExitCode(err))
		assert.Contains(t, err.Error(), "missing.sql, index cannot be restored")
	})
	t.Run("TrailingFlag", func(t *testing.T) {
		_, err := RunWithTestContext(DatabaseRestoreCommand, []string{"restore", "missing.sql", "--force"})

		require.Error(t, err)
		assert.Equal(t, 2, ExitCode(err))
		assert.Contains(t, err.Error(), "--force")
	})
}
