package commands

import (
	"bytes"
	"os"
	"strings"
	"testing"

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
