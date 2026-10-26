package commands

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/dl"
	"github.com/photoprism/photoprism/internal/photoprism/get"
)

func TestMissingFormatsHint(t *testing.T) {
	hint, ok := missingFormatsHint(dl.YoutubedlError("Requested format is not available. Use --list-formats for a list of available formats"))
	if !ok {
		t.Fatalf("expected hint for missing formats error")
	}
	if hint == "" {
		t.Fatalf("hint should not be empty")
	}

	if _, ok := missingFormatsHint(dl.YoutubedlError("some other error")); ok {
		t.Fatalf("unexpected hint for unrelated error")
	}
}

func TestResolveDownloadMethodEnv(t *testing.T) {
	t.Setenv(downloadMethodEnv, "FILE")

	method, fromEnv, err := resolveDownloadMethod("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if method != "file" {
		t.Fatalf("expected file, got %s", method)
	}
	if !fromEnv {
		t.Fatalf("expected value to originate from env")
	}
}

func TestResolveDownloadMethodInvalidEnv(t *testing.T) {
	t.Setenv(downloadMethodEnv, "weird")

	if _, _, err := resolveDownloadMethod(""); err == nil {
		t.Fatalf("expected error for invalid env method")
	}
}

func TestResolveDownloadMethodFlagTakesPriority(t *testing.T) {
	t.Setenv(downloadMethodEnv, "file")

	method, fromEnv, err := resolveDownloadMethod("pipe")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if method != "pipe" {
		t.Fatalf("expected pipe, got %s", method)
	}
	if fromEnv {
		t.Fatalf("did not expect env to be used when flag provided")
	}
}

func TestRunDownload_InvalidURLsReportPluralFailures(t *testing.T) {
	conf := get.Config()
	if conf == nil {
		t.Fatalf("missing test config")
	}

	conf.RegisterDb()

	var logOutput bytes.Buffer
	log.SetOutput(&logOutput)
	t.Cleanup(func() {
		log.SetOutput(os.Stderr)
	})

	err := runDownload(conf, DownloadOpts{
		Dest: "dl-invalid",
	}, []string{
		"not a url",
		"ftp://example.com/video.mp4",
	})

	if err == nil {
		t.Fatalf("expected invalid URLs to fail")
	}
	if err.Error() != "2 downloads failed" {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(logOutput.String(), "completed with 2 errors in") {
		t.Fatalf("expected pluralized summary log, got: %q", logOutput.String())
	}
}

// TestRunDownload_UnsupportedDest checks that downloads are not started for a destination the index would skip.
func TestRunDownload_UnsupportedDest(t *testing.T) {
	conf := get.Config()
	require.NotNil(t, conf)
	conf.RegisterDb()

	err := runDownload(conf, DownloadOpts{Dest: "a:b"}, []string{"https://example.com/video.mp4"})
	assert.ErrorIs(t, err, photoprism.ErrImportDest)
}

// testStream is an io.ReadCloser that returns data and a fixed Close error.
type testStream struct {
	io.Reader
	closeErr error
	closed   bool
}

// Close records the call and returns the configured error.
func (s *testStream) Close() error {
	s.closed = true
	return s.closeErr
}

func TestSavePipeDownload(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		filePath := filepath.Join(t.TempDir(), "video.mp4")
		stream := &testStream{Reader: strings.NewReader("DATA")}
		if !savePipeDownload(stream, filePath) {
			t.Fatalf("expected download to be saved")
		}
		if data, err := os.ReadFile(filePath); err != nil || string(data) != "DATA" { //nolint:gosec // test temp file
			t.Fatalf("expected file content DATA, got %q (%v)", data, err)
		}
		if !stream.closed {
			t.Fatalf("expected stream to be closed")
		}
	})
	t.Run("StreamError", func(t *testing.T) {
		filePath := filepath.Join(t.TempDir(), "video.mp4")
		stream := &testStream{Reader: strings.NewReader("DATA"), closeErr: errors.New("ffmpeg exited with code 251")}
		if savePipeDownload(stream, filePath) {
			t.Fatalf("expected failed stream to be reported")
		}
		if _, err := os.Stat(filePath); !os.IsNotExist(err) {
			t.Fatalf("expected incomplete file to be removed, stat err: %v", err)
		}
	})
	t.Run("CreateError", func(t *testing.T) {
		filePath := filepath.Join(t.TempDir(), "missing", "video.mp4")
		stream := &testStream{Reader: strings.NewReader("DATA")}
		if savePipeDownload(stream, filePath) {
			t.Fatalf("expected create failure to be reported")
		}
		if !stream.closed {
			t.Fatalf("expected stream to be closed after create failure")
		}
	})
}
