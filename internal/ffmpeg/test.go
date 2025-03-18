package ffmpeg

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
	"github.com/photoprism/photoprism/pkg/fs"
)

func RunCommandTest(t *testing.T, encoder encode.Encoder, srcName, destName string, cmd *exec.Cmd, deleteAfterTest bool) {
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	cmd.Env = append(cmd.Env, []string{
		fmt.Sprintf("HOME=%s", fs.Abs("./testdata")),
	}...)

	// Transcode source media file to AVC.
	start := time.Now()
	if err := cmd.Run(); err != nil {
		if stderr.String() != "" {
			err = errors.New(stderr.String())
		}

		// Remove broken video file.
		if !deleteAfterTest || !fs.FileExists(destName) {
			// Do nothing.
		} else if removeErr := os.Remove(destName); removeErr != nil {
			t.Logf("%s: failed to remove %s after error (%s)", encoder, srcName, removeErr)
		}

		// Log ffmpeg output for debugging.
		if err.Error() != "" {
			t.Error(err)
			t.Fatalf("%s: failed to transcode %s [%s]", encoder, srcName, time.Since(start))
		}
	}

	// Log filename and transcoding time.
	t.Logf("%s: created %s [%s]", encoder, destName, time.Since(start))

	// Return if destination file should not be deleted.
	if !deleteAfterTest {
		return
	}

	// Delete destination file after test.
	if removeErr := os.Remove(destName); removeErr != nil {
		t.Fatalf("%s: failed to remove %s after successful test (%s)", encoder, srcName, removeErr)
	}
}
