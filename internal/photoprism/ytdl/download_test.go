package ytdl

import (
	"bytes"
	"context"
	"io"
	"os/exec"
	"strings"
	"testing"
)

func TestDownload(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	stderrBuf := &bytes.Buffer{}
	r, err := New(context.Background(), testVideoRawURL, Options{
		StderrFn: func(cmd *exec.Cmd) io.Writer {
			return stderrBuf
		},
	})
	if err != nil {
		t.Fatal(err)
	}
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
