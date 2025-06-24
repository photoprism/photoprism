package dl

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestDownload(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	// Fetch metadata.
	stderrBuf := &bytes.Buffer{}
	r, err := NewMetadata(context.Background(), testVideoRawURL, Options{
		StderrFn: func(cmd *exec.Cmd) io.Writer {
			return stderrBuf
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Write metadata to "./testdata/download-test.json".
	jsonName := fs.Abs("./testdata/download-test.json")
	if json := r.Info.JSON(); json != nil {
		if err = fs.WriteFile(jsonName, json, fs.ModeFile); err != nil {
			t.Errorf("%s could not be created: %s", jsonName, err)
		}
	} else {
		t.Errorf("%s could not be created because json data is nil", jsonName)
	}

	// Remove metadata file.
	assert.FileExists(t, jsonName)
	_ = os.Remove(jsonName)

	// Download video based on metadata.
	dr, err := r.Download(context.Background(), r.Info.Formats[0].FormatID)
	if err != nil {
		t.Fatal(err)
	}
	downloadBuf := &bytes.Buffer{}
	n, err := io.Copy(downloadBuf, dr)
	if err != nil {
		t.Fatal(err)
	}
	dr.Close()

	if n != int64(downloadBuf.Len()) {
		t.Errorf("copy n not equal to download buffer: %d!=%d", n, downloadBuf.Len())
	}

	t.Logf("error: %s", stderrBuf.String())

	if n < 10000 {
		t.Errorf("should have copied at least 10000 bytes: %d", n)
	}

	if !strings.Contains(stderrBuf.String(), "Destination") {
		t.Errorf("did not find expected log message on stderr: %q", stderrBuf.String())
	}
}

func TestDownloadWithoutInfo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	stderrBuf := &bytes.Buffer{}
	dr, err := Download(context.Background(), testVideoRawURL, Options{
		StderrFn: func(cmd *exec.Cmd) io.Writer {
			return stderrBuf
		},
	}, "")
	if err != nil {
		t.Fatal(err)
	}
	downloadBuf := &bytes.Buffer{}
	n, err := io.Copy(downloadBuf, dr)
	if err != nil {
		t.Fatal(err)
	}
	dr.Close()

	if n != int64(downloadBuf.Len()) {
		t.Errorf("copy n not equal to download buffer: %d!=%d", n, downloadBuf.Len())
	}

	if n < 10000 {
		t.Errorf("should have copied at least 10000 bytes: %d", n)
	}

	if !strings.Contains(stderrBuf.String(), "Destination") {
		t.Errorf("did not find expected log message on stderr: %q", stderrBuf.String())
	}
}
