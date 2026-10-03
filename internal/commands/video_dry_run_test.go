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
