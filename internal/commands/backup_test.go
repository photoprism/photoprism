package commands

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/fs"
)

// newBackupContext parses args against the backup command's own flags. The command is attached
// because Context.IsSet resolves a name through Command.Flags. Pass the long form: urfave copies
// an alias onto its other names in normalizeFlags during a real run, which this set does not do,
// so "-r 2" would read as set with a zero value here.
func newBackupContext(t *testing.T, args ...string) *cli.Context {
	t.Helper()

	set := flag.NewFlagSet("backup", flag.ContinueOnError)

	for _, f := range BackupCommand.Flags {
		if err := f.Apply(set); err != nil {
			t.Fatal(err)
		}
	}

	if err := set.Parse(args); err != nil {
		t.Fatal(err)
	}

	ctx := cli.NewContext(cli.NewApp(), set, nil)
	ctx.Command = BackupCommand

	return ctx
}

func TestBackupRetain(t *testing.T) {
	t.Run("UnsetTakesConfiguredValue", func(t *testing.T) {
		assert.Equal(t, 7, backupRetain(newBackupContext(t), 7))
	})
	t.Run("UnsetTakesConfiguredKeepAll", func(t *testing.T) {
		assert.Equal(t, -1, backupRetain(newBackupContext(t), -1))
	})
	t.Run("FlagOverridesConfiguredValue", func(t *testing.T) {
		assert.Equal(t, 2, backupRetain(newBackupContext(t, "--retain", "2"), 7))
	})
	t.Run("FlagKeepAllOverridesConfiguredValue", func(t *testing.T) {
		assert.Equal(t, -1, backupRetain(newBackupContext(t, "--retain", "-1"), 3))
	})
	t.Run("FlagZeroOverridesConfiguredValue", func(t *testing.T) {
		assert.Equal(t, 0, backupRetain(newBackupContext(t, "--retain", "0"), 3))
	})
}

func TestBackupIndex(t *testing.T) {
	t.Run("FileName", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "index.sql")

		require.NoError(t, backupIndex(get.Config(), "", fileName, false, 0))
		assert.True(t, fs.FileExistsNotEmpty(fileName))
	})
	t.Run("Path", func(t *testing.T) {
		dir := t.TempDir()

		require.NoError(t, backupIndex(get.Config(), dir, "", false, 0))
		assert.True(t, fs.FileExistsNotEmpty(filepath.Join(dir, time.Now().UTC().Format("2006-01-02")+".sql")))
	})
	t.Run("FileNameKeepsBackups", func(t *testing.T) {
		dir := t.TempDir()
		older := []string{filepath.Join(dir, "2020-01-01.sql"), filepath.Join(dir, "2020-01-02.sql")}

		for _, fileName := range older {
			require.NoError(t, os.WriteFile(fileName, []byte("keep"), fs.ModeFile))
		}

		require.NoError(t, backupIndex(get.Config(), dir, filepath.Join(dir, "index.sql"), false, 1))

		for _, fileName := range older {
			assert.FileExists(t, fileName)
		}
	})
	t.Run("ExistingFile", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "index.sql")
		require.NoError(t, os.WriteFile(fileName, []byte("keep"), fs.ModeFile))

		err := backupIndex(get.Config(), "", fileName, false, 0)

		require.Error(t, err)
		assert.Equal(t, 1, ExitCode(err))
		assert.Contains(t, err.Error(), "failed to create database backup")
	})
}

func TestBackupCommand(t *testing.T) {
	t.Run("DatabasePath", func(t *testing.T) {
		dir := t.TempDir()

		_, err := RunWithTestContext(BackupCommand, []string{"backup", "-i", "--database-path", dir})

		require.NoError(t, err)
		assert.True(t, fs.FileExistsNotEmpty(filepath.Join(dir, time.Now().UTC().Format("2006-01-02")+".sql")))
	})
}
