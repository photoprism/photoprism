package dl

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

// createFakeYtDlp writes a small script that logs args to YTDLP_ARGS_LOG,
// optionally writes JSON when --dump-single-json is present, prints a file path
// when --print is present, emits a download prefix on stderr, and writes small
// data to stdout for pipe mode.
func createFakeYtDlp(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "yt-dlp")
	var script bytes.Buffer
	if runtime.GOOS == "windows" {
		// not needed in our CI/dev container; keep placeholder
		script.WriteString("@echo off\r\n")
		script.WriteString("setlocal enabledelayedexpansion\r\n")
		script.WriteString("echo %* >> %YTDLP_ARGS_LOG%\r\n")
		script.WriteString("echo {\"id\":\"abc\",\"title\":\"Test\",\"url\":\"http://example.com\",\"_type\":\"video\"} \r\n")
	} else {
		script.WriteString("#!/usr/bin/env bash\n")
		script.WriteString("set -euo pipefail\n")
		script.WriteString("ARGS_LOG=\"${YTDLP_ARGS_LOG:-}\"\n")
		script.WriteString("OUT_FILE=\"${YTDLP_OUTPUT_FILE:-}\"\n")
		script.WriteString("if [[ -n \"$ARGS_LOG\" ]]; then echo \"$*\" >> \"$ARGS_LOG\"; fi\n")
		script.WriteString("if [[ -n \"${YTDLP_ARGV_LOG:-}\" ]]; then printf '%s\\n' \"$@\" >> \"$YTDLP_ARGV_LOG\"; fi\n")
		// If metadata mode, print minimal JSON to stdout
		script.WriteString("for a in \"$@\"; do if [[ \"$a\" == \"--dump-single-json\" ]]; then echo '{\"id\":\"abc\",\"title\":\"Test\",\"url\":\"http://example.com\",\"_type\":\"video\"}'; exit 0; fi; done\n")
		// If print mode (file download), create file and print path(s)
		script.WriteString("for a in \"$@\"; do if [[ \"$a\" == \"--print\" ]]; then if [[ -n \"$OUT_FILE\" ]]; then mkdir -p \"$(dirname \"$OUT_FILE\")\"; echo 'dummy' > \"$OUT_FILE\"; echo \"$OUT_FILE\"; fi; exit 0; fi; done\n")
		// Pipe mode: emit download prefix on stderr then write some bytes to stdout
		script.WriteString("echo '[download]' 1>&2\n")
		script.WriteString("echo 'DATA'\n")
	}
	// #nosec G306 executable test helper script needs exec permissions
	if err := os.WriteFile(path, script.Bytes(), 0o755); err != nil {
		t.Fatalf("failed to write fake yt-dlp: %v", err)
	}
	return path
}

func TestInfoFromURL_IncludesHeadersAndCookies(t *testing.T) {
	fake := createFakeYtDlp(t)
	orig := YtDlpBin
	YtDlpBin = fake
	defer func() { YtDlpBin = orig }()

	argsLog := filepath.Join(t.TempDir(), "args.log")
	t.Setenv("YTDLP_ARGS_LOG", argsLog)

	_, _, err := infoFromURL(context.Background(), "https://example.com/video", Options{
		Cookies:            "cookies.txt",
		CookiesFromBrowser: "chrome:Default",
		AddHeaders:         []string{"Authorization: Bearer X", "Origin: https://example.com"},
		Type:               TypeSingle,
	})
	if err != nil {
		t.Fatalf("infoFromURL error: %v", err)
	}

	data, err := os.ReadFile(filepath.Clean(argsLog))
	if err != nil {
		t.Fatalf("reading args log failed: %v", err)
	}
	s := string(data)
	for _, expect := range []string{"--cookies cookies.txt", "--cookies-from-browser chrome:Default", "--add-header Authorization: Bearer X", "--add-header Origin: https://example.com"} {
		if !strings.Contains(s, expect) {
			t.Fatalf("missing expected arg %q in %q", expect, s)
		}
	}
}

func TestDownloadWithOptions_IncludesHeadersAndCookies_Pipe(t *testing.T) {
	fake := createFakeYtDlp(t)
	orig := YtDlpBin
	YtDlpBin = fake
	defer func() { YtDlpBin = orig }()
	argsLog := filepath.Join(t.TempDir(), "args.log")
	t.Setenv("YTDLP_ARGS_LOG", argsLog)

	r := Metadata{
		RawURL: "https://example.com/v",
		Options: Options{
			noInfoDownload:     true,
			Cookies:            "cookies.txt",
			CookiesFromBrowser: "firefox:Profile",
			AddHeaders:         []string{"Authorization: Bearer Y"},
		},
	}
	dr, err := r.DownloadWithOptions(context.Background(), DownloadOptions{})
	if err != nil {
		t.Fatalf("DownloadWithOptions error: %v", err)
	}
	// Read a bit and close
	buf := make([]byte, 4)
	_, _ = dr.Read(buf)
	_ = dr.Close()

	data, err := os.ReadFile(filepath.Clean(argsLog))
	if err != nil {
		t.Fatalf("reading args log failed: %v", err)
	}
	s := string(data)
	for _, expect := range []string{"--cookies cookies.txt", "--cookies-from-browser firefox:Profile", "--add-header Authorization: Bearer Y"} {
		if !strings.Contains(s, expect) {
			t.Fatalf("missing expected arg %q in %q", expect, s)
		}
	}
}

func TestDownloadWithOptions_FormatArgs(t *testing.T) {
	fake := createFakeYtDlp(t)
	orig := YtDlpBin
	YtDlpBin = fake
	defer func() { YtDlpBin = orig }()

	// downloadArgv runs a download with the given output and returns the yt-dlp argv, one element per entry.
	downloadArgv := func(t *testing.T, output string) []string {
		argvLog := filepath.Join(t.TempDir(), "argv.log")
		t.Setenv("YTDLP_ARGV_LOG", argvLog)
		r := Metadata{RawURL: "https://example.com/v", Options: Options{noInfoDownload: true}}
		dr, err := r.DownloadWithOptions(context.Background(), DownloadOptions{Filter: FormatBest, Output: output})
		if err != nil {
			t.Fatalf("DownloadWithOptions error: %v", err)
		}
		_, _ = io.Copy(io.Discard, dr)
		_ = dr.Close()
		data, err := os.ReadFile(filepath.Clean(argvLog))
		if err != nil {
			t.Fatalf("reading argv log failed: %v", err)
		}
		return strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	}

	t.Run("Stdout", func(t *testing.T) {
		argv := downloadArgv(t, "")
		i := slices.Index(argv, "--downloader-args")
		if i < 0 || i+1 >= len(argv) || argv[i+1] != "ffmpeg_i:"+ffmpegReconnectArgs {
			t.Fatalf("expected --downloader-args with FFmpeg reconnect args as one element; argv: %q", argv)
		}
	})
	t.Run("OutputFile", func(t *testing.T) {
		argv := downloadArgv(t, filepath.Join(t.TempDir(), "out.mp4"))
		if slices.Contains(argv, "--downloader-args") {
			t.Fatalf("expected no downloader args for file output; argv: %q", argv)
		}
	})
}

// writeFakeYtDlp writes an executable yt-dlp stub with the given bash body and returns its path.
func writeFakeYtDlp(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "yt-dlp")
	// #nosec G306 executable test helper script needs exec permissions
	if err := os.WriteFile(path, []byte("#!/usr/bin/env bash\n"+body), 0o755); err != nil {
		t.Fatalf("failed to write fake yt-dlp: %v", err)
	}
	return path
}

func TestDownloadResult_Close(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("requires bash")
	}

	orig := YtDlpBin
	defer func() { YtDlpBin = orig }()

	// download streams a pipe download with the given yt-dlp stub and returns the data and Close error.
	download := func(t *testing.T, body string) (string, error) {
		YtDlpBin = writeFakeYtDlp(t, body)
		r := Metadata{RawURL: "https://example.com/v", Options: Options{noInfoDownload: true}}
		dr, err := r.DownloadWithOptions(context.Background(), DownloadOptions{Filter: FormatBest})
		if err != nil {
			t.Fatalf("DownloadWithOptions error: %v", err)
		}
		data, _ := io.ReadAll(dr)
		return string(data), dr.Close()
	}

	t.Run("Success", func(t *testing.T) {
		data, err := download(t, "echo '[download]' 1>&2\necho DATA\n")
		if err != nil || data != "DATA\n" {
			t.Fatalf("expected complete stream without error; data %q, err %v", data, err)
		}
	})
	t.Run("StreamError", func(t *testing.T) {
		_, err := download(t, "echo '[download] Destination: -' 1>&2\necho DATA\necho 'ERROR: ffmpeg exited with code 251' 1>&2\nexit 1\n")
		if err == nil || err.Error() != "ffmpeg exited with code 251" {
			t.Fatalf("expected yt-dlp error message, got %v", err)
		}
	})
	t.Run("ExitWithoutMessage", func(t *testing.T) {
		_, err := download(t, "echo '[download]' 1>&2\necho DATA\nexit 3\n")
		if err == nil || !strings.Contains(err.Error(), "yt-dlp failed") {
			t.Fatalf("expected yt-dlp exit error, got %v", err)
		}
	})
	t.Run("ErrorBeforeDownload", func(t *testing.T) {
		YtDlpBin = writeFakeYtDlp(t, "echo 'ERROR: Requested format is not available' 1>&2\nexit 1\n")
		r := Metadata{RawURL: "https://example.com/v", Options: Options{noInfoDownload: true}}
		dr, err := r.DownloadWithOptions(context.Background(), DownloadOptions{Filter: FormatBest})
		if err == nil || err.Error() != "Requested format is not available" {
			t.Fatalf("expected yt-dlp error message, got %v", err)
		}
		if closeErr := dr.Close(); closeErr == nil {
			t.Fatalf("expected Close to report the failed run")
		}
	})
}

func TestDownloadWithOptions_OmitsFilterWhenDirect(t *testing.T) {
	fake := createFakeYtDlp(t)
	orig := YtDlpBin
	YtDlpBin = fake
	defer func() { YtDlpBin = orig }()

	argsLog := filepath.Join(t.TempDir(), "args.log")
	t.Setenv("YTDLP_ARGS_LOG", argsLog)

	r := Metadata{
		RawURL:  "https://example.com/direct",
		Info:    Info{Direct: true},
		Options: Options{noInfoDownload: true},
	}
	_, err := r.DownloadWithOptions(context.Background(), DownloadOptions{Filter: FormatBest})
	if err != nil {
		t.Fatalf("DownloadWithOptions error: %v", err)
	}
	data, err := os.ReadFile(filepath.Clean(argsLog))
	if err != nil {
		t.Fatalf("reading args log failed: %v", err)
	}
	s := string(data)
	if strings.Contains(s, "-f "+FormatBest) {
		t.Fatalf("expected -f not to be present for direct downloads; args: %s", s)
	}
}

func TestDownloadToFileWithOptions_PrintsAndCreatesFiles(t *testing.T) {
	fake := createFakeYtDlp(t)
	orig := YtDlpBin
	YtDlpBin = fake
	defer func() { YtDlpBin = orig }()
	argsLog := filepath.Join(t.TempDir(), "args.log")
	t.Setenv("YTDLP_ARGS_LOG", argsLog)
	outDir := t.TempDir()
	outFile := filepath.Join(outDir, "ppdl_test.mp4")
	t.Setenv("YTDLP_OUTPUT_FILE", outFile)

	r := Metadata{
		RawURL: "https://example.com/v",
		Options: Options{
			noInfoDownload: true,
		},
	}
	files, err := r.DownloadToFileWithOptions(context.Background(), DownloadOptions{Output: filepath.Join(outDir, "ppdl_%(id)s.%(ext)s")})
	if err != nil {
		t.Fatalf("DownloadToFileWithOptions error: %v", err)
	}
	if len(files) == 0 {
		t.Fatalf("expected at least one file path returned")
	}
	if _, statErr := os.Stat(outFile); statErr != nil {
		t.Fatalf("expected file to exist: %v", statErr)
	}
}

func TestDownloadToFileWithOptions_IncludesPostprocessorArgs(t *testing.T) {
	fake := createFakeYtDlp(t)
	orig := YtDlpBin
	YtDlpBin = fake
	defer func() { YtDlpBin = orig }()

	argsLog := filepath.Join(t.TempDir(), "args.log")
	t.Setenv("YTDLP_ARGS_LOG", argsLog)

	outDir := t.TempDir()
	outFile := filepath.Join(outDir, "ppdl_test.mp4")
	t.Setenv("YTDLP_OUTPUT_FILE", outFile)

	r := Metadata{
		RawURL: "https://example.com/v",
		Options: Options{
			noInfoDownload: true,
			FFmpegPostArgs: "-metadata creation_time=2021-11-20T00:00:00Z",
		},
	}
	_, err := r.DownloadToFileWithOptions(context.Background(), DownloadOptions{Output: filepath.Join(outDir, "ppdl_%(id)s.%(ext)s")})
	if err != nil {
		t.Fatalf("DownloadToFileWithOptions error: %v", err)
	}

	data, err := os.ReadFile(filepath.Clean(argsLog))
	if err != nil {
		t.Fatalf("reading args log failed: %v", err)
	}
	s := string(data)
	if !strings.Contains(s, "--postprocessor-args") || !strings.Contains(s, "ffmpeg:-metadata creation_time=2021-11-20T00:00:00Z") {
		t.Fatalf("missing postprocessor args in yt-dlp invocation: %s", s)
	}
}

func TestDownloadWithOptions_IncludesPostprocessorArgs_Pipe(t *testing.T) {
	fake := createFakeYtDlp(t)
	orig := YtDlpBin
	YtDlpBin = fake
	defer func() { YtDlpBin = orig }()

	argsLog := filepath.Join(t.TempDir(), "args.log")
	t.Setenv("YTDLP_ARGS_LOG", argsLog)

	r := Metadata{
		RawURL: "https://example.com/v",
		Options: Options{
			noInfoDownload: true,
			FFmpegPostArgs: "-metadata creation_time=2021-11-20T00:00:00Z",
		},
	}
	dr, err := r.DownloadWithOptions(context.Background(), DownloadOptions{})
	if err != nil {
		t.Fatalf("DownloadWithOptions error: %v", err)
	}
	// Read a bit and close to finish the process
	buf := make([]byte, 4)
	_, _ = dr.Read(buf)
	_ = dr.Close()

	data, err := os.ReadFile(filepath.Clean(argsLog))
	if err != nil {
		t.Fatalf("reading args log failed: %v", err)
	}
	s := string(data)
	if !strings.Contains(s, "--postprocessor-args") || !strings.Contains(s, "ffmpeg:-metadata creation_time=2021-11-20T00:00:00Z") {
		t.Fatalf("missing postprocessor args in yt-dlp invocation: %s", s)
	}
}
