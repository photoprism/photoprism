package commands

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/photoprism/get"
)

// TestFaceMigrationExit pins the exit error of a command that waits for a face migration.
func TestFaceMigrationExit(t *testing.T) {
	t.Run("NoConfig", func(t *testing.T) {
		err := faceMigrationExit("index", nil)

		var exit cli.ExitCoder
		require.ErrorAs(t, err, &exit)
		assert.Equal(t, 1, exit.ExitCode())
		assert.Equal(t, "index: waiting for the face migration to complete, run the command again once it has", err.Error())
	})
	t.Run("NamesTheHolder", func(t *testing.T) {
		conf := get.Config()
		lock, err := mutex.AcquireFileLock(conf.FacesLockFile(), "faces migration")
		require.NoError(t, err)
		t.Cleanup(lock.Release)

		err = faceMigrationExit("import", conf)
		assert.Contains(t, err.Error(), "import: waiting for the faces migration, running as process")
	})
}

// TestFaceMigrationCommands pins that indexing, importing and downloading exit with status 1 while
// a face migration holds the lock, rather than reporting a run that did nothing.
func TestFaceMigrationCommands(t *testing.T) {
	conf := get.Config()
	require.NotNil(t, conf)

	lock, err := mutex.AcquireFileLock(conf.FacesLockFile(), "faces migration")
	require.NoError(t, err)
	t.Cleanup(lock.Release)

	// The commands close the database when they return, so each subtest opens it again.
	t.Cleanup(conf.RegisterDb)

	exitCode := func(t *testing.T, err error) int {
		t.Helper()

		var exit cli.ExitCoder
		require.ErrorAs(t, err, &exit)

		return exit.ExitCode()
	}

	t.Run("Index", func(t *testing.T) {
		conf.RegisterDb()

		_, err := RunWithTestContext(IndexCommand, []string{"index"})
		require.Error(t, err)
		assert.Equal(t, 1, exitCode(t, err))
		assert.Contains(t, err.Error(), "index: waiting for the faces migration")
	})
	t.Run("Import", func(t *testing.T) {
		conf.RegisterDb()

		_, err := RunWithTestContext(ImportCommand, []string{"import", t.TempDir()})
		require.Error(t, err)
		assert.Equal(t, 1, exitCode(t, err))
		assert.Contains(t, err.Error(), "import: waiting for the faces migration")
	})
	t.Run("Copy", func(t *testing.T) {
		conf.RegisterDb()

		_, err := RunWithTestContext(CopyCommand, []string{"cp", t.TempDir()})
		require.Error(t, err)
		assert.Equal(t, 1, exitCode(t, err))
		assert.Contains(t, err.Error(), "copy: waiting for the faces migration")
	})
	t.Run("FacesIndex", func(t *testing.T) {
		conf.RegisterDb()

		var index *cli.Command

		for _, cmd := range FacesCommands.Subcommands {
			if cmd.Name == "index" {
				index = cmd
			}
		}

		require.NotNil(t, index)

		_, err := RunWithTestContext(index, []string{"index"})
		require.Error(t, err)
		assert.Equal(t, 1, exitCode(t, err))
		assert.Contains(t, err.Error(), "faces index: waiting for the faces migration")
	})
	t.Run("Download", func(t *testing.T) {
		conf.RegisterDb()

		var logOutput bytes.Buffer
		log.SetOutput(&logOutput)
		t.Cleanup(func() { log.SetOutput(os.Stderr) })

		err := runDownload(conf, DownloadOpts{}, []string{"not a url"})
		require.Error(t, err)
		assert.Equal(t, 1, exitCode(t, err))
		assert.Contains(t, err.Error(), "download: waiting for the faces migration")

		// Refused before anything is downloaded, rather than after the downloads that are then removed.
		assert.NotContains(t, logOutput.String(), "importing downloads")
	})
}
