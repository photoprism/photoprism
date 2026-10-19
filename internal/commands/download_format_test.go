package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/photoprism/photoprism/internal/photoprism/dl"
	"github.com/photoprism/photoprism/internal/photoprism/get"
)

func createArgsLoggingYtDlp(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "yt-dlp")

	if runtime.GOOS == "windows" {
		var b strings.Builder
		b.WriteString("@echo off\r\n")
		b.WriteString("if not \"%YTDLP_ARGS_LOG%\"==\"\" echo %* >> %YTDLP_ARGS_LOG%\r\n")
		b.WriteString("for %%A in (%*) do (\r\n")
		b.WriteString("  if \"%%~A\"==\"--version\" ( echo 2025.09.23 & goto :eof )\r\n")
		b.WriteString("  if \"%%~A\"==\"--dump-single-json\" ( echo {\"id\":\"abc\",\"title\":\"Test\",\"url\":\"http://example.com\",\"_type\":\"video\"} & goto :eof )\r\n")
		b.WriteString(")\r\n")
		if err := os.WriteFile(path, []byte(b.String()), 0o600); err != nil {
			t.Fatalf("failed to write fake yt-dlp: %v", err)
		}
		return path
	}

	var b strings.Builder
	b.WriteString("#!/usr/bin/env bash\n")
	b.WriteString("set -euo pipefail\n")
	b.WriteString("ARGS_LOG=\"${YTDLP_ARGS_LOG:-}\"\n")
	b.WriteString("if [[ -n \"$ARGS_LOG\" ]]; then echo \"$*\" >> \"$ARGS_LOG\"; fi\n")
	b.WriteString("for a in \"$@\"; do if [[ \"$a\" == \"--version\" ]]; then echo '2025.09.23'; exit 0; fi; done\n")
	b.WriteString("OUT_TPL=\"\"\n")
	b.WriteString("i=0; while [[ $i -lt $# ]]; do i=$((i+1)); arg=\"${!i}\"; if [[ \"$arg\" == \"--output\" ]]; then i=$((i+1)); OUT_TPL=\"${!i}\"; fi; done\n")
	b.WriteString("for a in \"$@\"; do if [[ \"$a\" == \"--dump-single-json\" ]]; then echo '{\"id\":\"abc\",\"title\":\"Test\",\"url\":\"http://example.com\",\"_type\":\"video\"}'; exit 0; fi; done\n")
	b.WriteString("for a in \"$@\"; do if [[ \"$a\" == \"--print\" ]]; then OUT=\"$OUT_TPL\"; OUT=${OUT//%(id)s/abc}; OUT=${OUT//%(ext)s/mp4}; mkdir -p \"$(dirname \"$OUT\")\"; CONTENT=\"${YTDLP_DUMMY_CONTENT:-dummy}\"; printf \"%s\" \"$CONTENT\" > \"$OUT\"; echo \"$OUT\"; exit 0; fi; done\n")
	b.WriteString("echo '[download]' 1>&2\n")
	b.WriteString("echo 'DATA'\n")
	b.WriteString("if [[ \"${YTDLP_FAIL_PIPE:-}\" == \"1\" ]]; then echo 'ERROR: simulated stream failure' 1>&2; exit 1; fi\n")

	if err := os.WriteFile(path, []byte(b.String()), 0o600); err != nil {
		t.Fatalf("failed to write fake yt-dlp: %v", err)
	}
	return path
}

func TestRunDownload_FileMethod_OmitsFormatSort(t *testing.T) {
	t.Setenv("YTDLP_FORCE_SHELL", "1")
	argsLog := filepath.Join(t.TempDir(), "args.log")
	t.Setenv("YTDLP_ARGS_LOG", argsLog)
	outDir := t.TempDir()
	outFile := filepath.Join(outDir, "ppdl_test.mp4")
	t.Setenv("YTDLP_OUTPUT_FILE", outFile)
	t.Setenv("YTDLP_DUMMY_CONTENT", "quality-test")
	dl.ResetVersionWarningForTest()

	fake := createArgsLoggingYtDlp(t)
	origBin := dl.YtDlpBin
	dl.YtDlpBin = fake
	defer func() { dl.YtDlpBin = origBin }()

	conf := get.Config()
	if conf == nil {
		t.Fatalf("missing test config")
	}
	conf.RegisterDb()

	// Avoid background ffmpeg work that could interfere with the test environment.
	opt := conf.Options()
	origFFmpeg := opt.FFmpegBin
	opt.FFmpegBin = "/bin/false"
	settings := conf.Settings()
	origConvert := settings.Index.Convert
	settings.Index.Convert = false

	dest := "dl-quality"
	if err := runDownload(conf, DownloadOpts{
		Dest:      dest,
		Method:    "file",
		FileRemux: "skip",
	}, []string{"https://example.com/video"}); err != nil {
		t.Fatalf("runDownload failed: %v", err)
	}

	// Ensure Originals cleanup so subsequent tests stay isolated.
	t.Cleanup(func() {
		_ = os.RemoveAll(filepath.Join(conf.OriginalsPath(), dest))
		settings.Index.Convert = origConvert
		opt.FFmpegBin = origFFmpeg
	})

	// Give the logging script a moment to flush in slower environments.
	time.Sleep(20 * time.Millisecond)

	data, err := os.ReadFile(argsLog) //nolint:gosec // test temp file
	if err != nil {
		t.Fatalf("reading args log failed: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "--print") {
		t.Fatalf("expected file download invocation to be logged, got: %q", content)
	}
	if strings.Contains(content, "--format-sort") {
		t.Fatalf("file method should not pass --format-sort; args: %q", content)
	}
	if !strings.Contains(content, "-f "+dl.FormatBest+" ") {
		t.Fatalf("file method should request %s; args: %q", dl.FormatBest, content)
	}
}

func TestRunDownload_PipeMethod_StreamFailure(t *testing.T) {
	t.Setenv("YTDLP_FORCE_SHELL", "1")
	argsLog := filepath.Join(t.TempDir(), "args.log")
	t.Setenv("YTDLP_ARGS_LOG", argsLog)
	t.Setenv("YTDLP_FAIL_PIPE", "1")
	dl.ResetVersionWarningForTest()

	fake := createArgsLoggingYtDlp(t)
	origBin := dl.YtDlpBin
	dl.YtDlpBin = fake
	defer func() { dl.YtDlpBin = origBin }()

	conf := get.Config()
	if conf == nil {
		t.Fatalf("missing test config")
	}
	conf.RegisterDb()

	opt := conf.Options()
	origFFmpeg := opt.FFmpegBin
	opt.FFmpegBin = "/bin/false"
	settings := conf.Settings()
	origConvert := settings.Index.Convert
	settings.Index.Convert = false

	var logOutput bytes.Buffer
	log.SetOutput(&logOutput)

	dest := "dl-pipe-failure"
	t.Cleanup(func() {
		log.SetOutput(os.Stderr)
		_ = os.RemoveAll(filepath.Join(conf.OriginalsPath(), dest))
		settings.Index.Convert = origConvert
		opt.FFmpegBin = origFFmpeg
	})

	err := runDownload(conf, DownloadOpts{Dest: dest, Method: "pipe"}, []string{"https://example.com/video"})
	if err == nil || err.Error() != "1 download failed" {
		t.Fatalf("expected the failed stream to count as a failed download, got %v", err)
	}

	logs := logOutput.String()
	if !strings.Contains(logs, "download failed: simulated stream failure") {
		t.Fatalf("expected the yt-dlp error to be logged, got: %q", logs)
	}
	if strings.Contains(logs, "remux failed") {
		t.Fatalf("expected no remux of an incomplete download, got: %q", logs)
	}
	if files, _ := filepath.Glob(filepath.Join(conf.OriginalsPath(), dest, "*")); len(files) > 0 {
		t.Fatalf("expected no imported files, got %v", files)
	}

	data, readErr := os.ReadFile(argsLog) //nolint:gosec // test temp file
	if readErr != nil {
		t.Fatalf("reading args log failed: %v", readErr)
	}
	if !strings.Contains(string(data), "-f "+dl.FormatBest+" ") {
		t.Fatalf("pipe method should request %s; args: %q", dl.FormatBest, string(data))
	}
}

func TestRunDownload_FileMethod_WithFormatSort(t *testing.T) {
	t.Setenv("YTDLP_FORCE_SHELL", "1")
	argsLog := filepath.Join(t.TempDir(), "args.log")
	t.Setenv("YTDLP_ARGS_LOG", argsLog)
	outDir := t.TempDir()
	outFile := filepath.Join(outDir, "ppdl_fmt.mp4")
	t.Setenv("YTDLP_OUTPUT_FILE", outFile)
	t.Setenv("YTDLP_DUMMY_CONTENT", "quality-test-override")
	dl.ResetVersionWarningForTest()

	fake := createArgsLoggingYtDlp(t)
	origBin := dl.YtDlpBin
	dl.YtDlpBin = fake
	defer func() { dl.YtDlpBin = origBin }()

	conf := get.Config()
	if conf == nil {
		t.Fatalf("missing test config")
	}
	conf.RegisterDb()

	opt := conf.Options()
	origFFmpeg := opt.FFmpegBin
	opt.FFmpegBin = "/bin/false"
	settings := conf.Settings()
	origConvert := settings.Index.Convert
	settings.Index.Convert = false

	dest := "dl-format-sort"
	if err := runDownload(conf, DownloadOpts{
		Dest:       dest,
		Method:     "file",
		FileRemux:  "skip",
		FormatSort: "res,fps,size",
	}, []string{"https://example.com/video"}); err != nil {
		t.Fatalf("runDownload failed with custom sort: %v", err)
	}

	t.Cleanup(func() {
		_ = os.RemoveAll(filepath.Join(conf.OriginalsPath(), dest))
		settings.Index.Convert = origConvert
		opt.FFmpegBin = origFFmpeg
	})

	time.Sleep(20 * time.Millisecond)

	data, err := os.ReadFile(argsLog) //nolint:gosec // test temp file
	if err != nil {
		t.Fatalf("reading args log failed: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "--format-sort") {
		t.Fatalf("expected --format-sort to be passed; args: %q", content)
	}
	if !strings.Contains(content, "res,fps,size") {
		t.Fatalf("expected custom format-sort expression to be present; args: %q", content)
	}
}
