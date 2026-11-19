package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/fs"
)

// TestVideoCommandsDryRunSummary checks planned counts at each command's action boundary.
func TestVideoCommandsDryRunSummary(t *testing.T) {
	for _, tc := range []struct {
		name            string
		command         *cli.Command
		fileType, codec string
	}{
		{"Trim", VideoTrimCommand, fs.VideoMp4.String(), "avc1"},
		{"Remux", VideoRemuxCommand, fs.VideoMkv.String(), "avc1"},
		{"Transcode", VideoTranscodeCommand, fs.VideoAVI.String(), "mp4v"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reopenConnection()
			photoUID, filename := videoConfirmFixture(t, "clip."+tc.fileType, tc.fileType, tc.codec, 10*time.Second)
			before, err := os.ReadFile(filename) //nolint:gosec // The fixture owns this isolated path.
			require.NoError(t, err)
			sidecarDir := filepath.Join(get.Config().SidecarPath(), filepath.Base(filepath.Dir(filename)))
			require.NoError(t, fs.MkdirAll(sidecarDir))
			t.Cleanup(func() { _ = os.RemoveAll(sidecarDir) })
			var output bytes.Buffer
			previous := log
			log = logrus.New()
			log.SetOutput(&output)
			t.Cleanup(func() { log = previous })
			args := []string{tc.command.Name, "--dry-run", "uid:" + photoUID}
			if tc.name == "Trim" {
				args = append(args, "2")
			}
			_, err = RunWithTestContext(tc.command, args)
			require.NoError(t, err)
			assert.Contains(t, output.String(), tc.command.Name+": would process 1 file, skipped 0 files")
			after, err := os.ReadFile(filename) //nolint:gosec // The fixture owns this isolated path.
			require.NoError(t, err)
			assert.Equal(t, before, after)
			files, err := os.ReadDir(filepath.Dir(filename))
			require.NoError(t, err)
			assert.Len(t, files, 1)
			sidecars, err := os.ReadDir(sidecarDir)
			require.NoError(t, err)
			assert.Empty(t, sidecars)
		})
	}
}

// TestVideoCommandsFaceMigration pins that the video commands change no file while a face migration holds
// the lock, and that a dry run still reports what it would do.
func TestVideoCommandsFaceMigration(t *testing.T) {
	conf := get.Config()

	lock, err := mutex.AcquireFileLock(conf.FacesLockFile(), "faces migration")
	require.NoError(t, err)
	t.Cleanup(lock.Release)

	for _, tc := range []struct {
		name            string
		command         *cli.Command
		fileType, codec string
	}{
		{"Trim", VideoTrimCommand, fs.VideoMp4.String(), "avc1"},
		{"Remux", VideoRemuxCommand, fs.VideoMkv.String(), "avc1"},
		{"Transcode", VideoTranscodeCommand, fs.VideoAVI.String(), "mp4v"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reopenConnection()
			photoUID, filename := videoConfirmFixture(t, "clip."+tc.fileType, tc.fileType, tc.codec, 10*time.Second)
			before, readErr := os.ReadFile(filename) //nolint:gosec // The fixture owns this isolated path.
			require.NoError(t, readErr)
			sidecarDir := filepath.Join(conf.SidecarPath(), filepath.Base(filepath.Dir(filename)))
			require.NoError(t, fs.MkdirAll(sidecarDir))
			t.Cleanup(func() { _ = os.RemoveAll(sidecarDir) })

			args := func(flag string) []string {
				result := []string{tc.command.Name, flag, "uid:" + photoUID}

				if tc.name == "Trim" {
					result = append(result, "2")
				}

				return result
			}

			_, runErr := RunWithTestContext(tc.command, args("--yes"))

			var exit cli.ExitCoder
			require.ErrorAs(t, runErr, &exit)
			assert.Equal(t, 1, exit.ExitCode())
			assert.Contains(t, runErr.Error(), tc.command.Name+": waiting for the faces migration")

			reopenConnection()
			_, runErr = RunWithTestContext(tc.command, args("--dry-run"))
			assert.NoError(t, runErr)

			after, readErr := os.ReadFile(filename) //nolint:gosec // The fixture owns this isolated path.
			require.NoError(t, readErr)
			assert.Equal(t, before, after)
			files, readErr := os.ReadDir(filepath.Dir(filename))
			require.NoError(t, readErr)
			assert.Len(t, files, 1, "no output was written")
			sidecars, readErr := os.ReadDir(sidecarDir)
			require.NoError(t, readErr)
			assert.Empty(t, sidecars)
		})
	}
}

// TestVideoBatchStopsForFaceMigration pins that a migration starting during a batch stops it before the
// next file, and that the exit then says to index the processed files rather than to repeat the command.
func TestVideoBatchStopsForFaceMigration(t *testing.T) {
	for _, tc := range []struct {
		name            string
		command         *cli.Command
		fileType, codec string
	}{
		{"Trim", VideoTrimCommand, fs.VideoMp4.String(), "avc1"},
		{"Remux", VideoRemuxCommand, fs.VideoMkv.String(), "avc1"},
		{"Transcode", VideoTranscodeCommand, fs.VideoAVI.String(), "mp4v"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// The lock reads as held from the second file on: one check before the prompt, one per file.
			var checks int
			locked := videoFacesLocked
			videoFacesLocked = func(*config.Config) string {
				checks++

				if checks > 2 {
					return "faces migration, running as process 1 on test"
				}

				return ""
			}
			t.Cleanup(func() { videoFacesLocked = locked })

			reopenConnection()
			firstUID, firstName := videoConfirmFixture(t, "first."+tc.fileType, tc.fileType, tc.codec, 10*time.Second)
			secondUID, secondName := videoConfirmFixture(t, "second."+tc.fileType, tc.fileType, tc.codec, 10*time.Second)

			// An indexed video has its preview image in the sidecar folder, so the folders exist.
			for _, name := range []string{firstName, secondName} {
				sidecarDir := filepath.Join(get.Config().SidecarPath(), filepath.Base(filepath.Dir(name)))
				require.NoError(t, fs.MkdirAll(sidecarDir))
				t.Cleanup(func() { _ = os.RemoveAll(sidecarDir) })
			}

			var output bytes.Buffer
			previous := log
			log = logrus.New()
			log.SetOutput(&output)
			t.Cleanup(func() { log = previous })

			args := []string{tc.command.Name, "--yes", "uid:" + firstUID + "|" + secondUID}

			if tc.name == "Trim" {
				args = append(args, "2")
			}

			_, runErr := RunWithTestContext(tc.command, args)

			var exit cli.ExitCoder
			require.ErrorAs(t, runErr, &exit, output.String())
			assert.Equal(t, 1, exit.ExitCode())
			assert.Contains(t, runErr.Error(), tc.command.Name+": stopped for the faces migration")
			assert.Contains(t, runErr.Error(), "after 1 file")
			assert.Equal(t, 3, checks, "the batch stops at the second file")

			// The fixtures are no real videos, so the first file fails rather than changing; the other is not attempted.
			assert.Contains(t, output.String(), tc.command.Name+": processed 0 files, skipped 0 files, 1 file failed")
		})
	}
}
