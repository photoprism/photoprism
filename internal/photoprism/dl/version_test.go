package dl

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestVersionWarning_OldVersion(t *testing.T) {
	ResetVersionWarningForTest()

	bin := writeVersionScript(t, "2025.09.05")
	orig := YtDlpBin
	YtDlpBin = bin
	defer func() { YtDlpBin = orig }()

	os.Unsetenv("YTDLP_FAKE")

	msg, ok := VersionWarning()
	if !ok {
		t.Fatalf("expected warning for old version")
	}
	if !strings.Contains(msg, minYtDlpVersion) {
		t.Fatalf("warning missing minimum version: %s", msg)
	}
}

func TestVersionWarning_NewEnough(t *testing.T) {
	ResetVersionWarningForTest()

	bin := writeVersionScript(t, "2025.09.23")
	orig := YtDlpBin
	YtDlpBin = bin
	defer func() { YtDlpBin = orig }()

	os.Unsetenv("YTDLP_FAKE")

	if _, ok := VersionWarning(); ok {
		t.Fatalf("did not expect warning for up-to-date version")
	}
}

func writeVersionScript(t *testing.T, version string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "yt-dlp")

	if runtime.GOOS == "windows" {
		content := "@echo off\r\n" +
			"echo " + version + "\r\n"
		// #nosec G306 executable test helper script
		if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
			t.Fatalf("failed to write fake yt-dlp: %v", err)
		}
		return path
	}

	content := "#!/usr/bin/env bash\n" +
		"set -euo pipefail\n" +
		"echo '" + version + "'\n"
	// #nosec G306 executable test helper script
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("failed to write fake yt-dlp: %v", err)
	}
	return path
}
