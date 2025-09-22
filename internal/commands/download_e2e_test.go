package commands

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/photoprism/photoprism/internal/photoprism/dl"
	"github.com/photoprism/photoprism/internal/photoprism/get"
)

// createFakeYtDlp writes a small script that:
//   - prints JSON when --dump-single-json (metadata)
//   - parses --output TEMPLATE and on --print creates a dummy file at TEMPLATE
//     with %(id)s -> abc and %(ext)s -> mp4, then prints the path
func createFakeYtDlp(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "yt-dlp")
	if runtime.GOOS == "windows" {
		// Not needed in CI/dev container. Keep simple stub.
		content := "@echo off\r\n" +
			"for %%A in (%*) do (\r\n" +
			"  if \"%%~A\"==\"--dump-single-json\" ( echo {\"id\":\"abc\",\"title\":\"Test\",\"url\":\"http://example.com\",\"_type\":\"video\"} & goto :eof )\r\n" +
			")\r\n"
		if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
			t.Fatalf("failed to write fake yt-dlp: %v", err)
		}
		return path
	}
	var b strings.Builder
	b.WriteString("#!/usr/bin/env bash\n")
	b.WriteString("set -euo pipefail\n")
	b.WriteString("OUT_TPL=\"\"\n")
	b.WriteString("i=0; while [[ $i -lt $# ]]; do i=$((i+1)); arg=\"${!i}\"; if [[ \"$arg\" == \"--dump-single-json\" ]]; then echo '{\"id\":\"abc\",\"title\":\"Test\",\"url\":\"http://example.com\",\"_type\":\"video\"}'; exit 0; fi; if [[ \"$arg\" == \"--output\" ]]; then i=$((i+1)); OUT_TPL=\"${!i}\"; fi; done\n")
	b.WriteString("if [[ $* == *'--print '* ]]; then OUT=\"$OUT_TPL\"; OUT=${OUT//%(id)s/abc}; OUT=${OUT//%(ext)s/mp4}; mkdir -p \"$(dirname \"$OUT\")\"; CONTENT=\"${YTDLP_DUMMY_CONTENT:-dummy}\"; echo \"$CONTENT\" > \"$OUT\"; echo \"$OUT\"; exit 0; fi\n")
	if err := os.WriteFile(path, []byte(b.String()), 0o755); err != nil {
		t.Fatalf("failed to write fake yt-dlp: %v", err)
	}
	return path
}

func TestDownloadImpl_FileMethod_AutoSkipsRemux(t *testing.T) {
	fake := createFakeYtDlp(t)
	orig := dl.YtDlpBin
	defer func() { dl.YtDlpBin = orig }()

	dest := "dl-e2e"
	// Force ffmpeg to an invalid path; with remux=auto the remux should be skipped for mp4
	if c := get.Config(); c != nil {
		c.Options().FFmpegBin = "/bin/false"
		// Disable convert (thumb generation) to avoid ffmpeg dependency in test
		s := c.Settings()
		s.Index.Convert = false
	}
	conf := get.Config()
	if conf == nil {
		t.Fatalf("missing test config")
	}
	// Ensure DB is initialized and registered (bypassing CLI InitConfig)
	_ = conf.Init()
	conf.RegisterDb()
	// Override yt-dlp after config init (config may set dl.YtDlpBin)
	dl.YtDlpBin = fake
	t.Logf("using yt-dlp binary: %s", dl.YtDlpBin)
	// Execute the implementation core directly
	err := runDownload(conf, DownloadOpts{
		Dest:      dest,
		Method:    "file",
		FileRemux: "auto",
	}, []string{"https://example.com/video"})
	if err != nil {
		t.Fatalf("runDownload failed (auto should skip remux): %v", err)
	}

	// Cleanup destination folder (best effort)
	if c := get.Config(); c != nil {
		outDir := filepath.Join(c.OriginalsPath(), dest)
		_ = os.RemoveAll(outDir)
	}
}

func TestDownloadImpl_FileMethod_Skip_NoRemux(t *testing.T) {
	fake := createFakeYtDlp(t)
	orig := dl.YtDlpBin
	defer func() { dl.YtDlpBin = orig }()

	dest := "dl-e2e-skip"
	// Ensure different file content so duplicate detection won't collapse into prior test's file
	t.Setenv("YTDLP_DUMMY_CONTENT", "dummy2")
	if c := get.Config(); c != nil {
		c.Options().FFmpegBin = "/bin/false" // would fail if remux attempted
		s := c.Settings()
		s.Index.Convert = false
	}
	conf := get.Config()
	if conf == nil {
		t.Fatalf("missing test config")
	}
	_ = conf.Init()
	conf.RegisterDb()
	dl.YtDlpBin = fake

	if err := runDownload(conf, DownloadOpts{
		Dest:      dest,
		Method:    "file",
		FileRemux: "skip",
	}, []string{"https://example.com/video"}); err != nil {
		t.Fatalf("runDownload failed with skip remux: %v", err)
	}

	// Verify an mp4 exists under Originals/dest
	c := get.Config()
	outDir := filepath.Join(c.OriginalsPath(), dest)
	found := false
	_ = filepath.WalkDir(outDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d == nil {
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(d.Name()), ".mp4") {
			found = true
			return filepath.SkipDir
		}
		return nil
	})
	if !found {
		t.Fatalf("expected at least one mp4 in %s", outDir)
	}
	_ = os.RemoveAll(outDir)
}

func TestDownloadImpl_FileMethod_Always_RemuxFails(t *testing.T) {
	fake := createFakeYtDlp(t)
	orig := dl.YtDlpBin
	defer func() { dl.YtDlpBin = orig }()

	dest := "dl-e2e-always"
	if c := get.Config(); c != nil {
		c.Options().FFmpegBin = "/bin/false" // force remux failure when called
		s := c.Settings()
		s.Index.Convert = false
	}
	conf := get.Config()
	if conf == nil {
		t.Fatalf("missing test config")
	}
	_ = conf.Init()
	conf.RegisterDb()
	dl.YtDlpBin = fake

	err := runDownload(conf, DownloadOpts{
		Dest:      dest,
		Method:    "file",
		FileRemux: "always",
	}, []string{"https://example.com/video"})
	if err == nil {
		t.Fatalf("expected failure when remux is required but ffmpeg is unavailable")
	}

	// Cleanup destination folder if anything was created
	c := get.Config()
	outDir := filepath.Join(c.OriginalsPath(), dest)
	_ = os.RemoveAll(outDir)
}
