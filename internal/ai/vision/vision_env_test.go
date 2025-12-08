package vision

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadEnvKeyFromFile verifies that loadEnvKeyFromFile reads API keys from
// *_FILE variables when the primary env var is empty.
func TestLoadEnvKeyFromFile(t *testing.T) {
	t.Run("ReadsFileWhenUnset", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "key.txt")
		if err := os.WriteFile(path, []byte("file-secret\n"), 0o600); err != nil {
			t.Fatalf("write key file: %v", err)
		}

		t.Setenv("TEST_KEY", "")
		t.Setenv("TEST_KEY_FILE", path)

		loadEnvKeyFromFile("TEST_KEY", "TEST_KEY_FILE")

		if got := os.Getenv("TEST_KEY"); got != "file-secret" {
			t.Fatalf("expected file-secret, got %q", got)
		}
	})
	t.Run("EnvWinsOverFile", func(t *testing.T) {
		t.Setenv("TEST_KEY", "keep-env")
		t.Setenv("TEST_KEY_FILE", "/nonexistent")

		loadEnvKeyFromFile("TEST_KEY", "TEST_KEY_FILE")

		if got := os.Getenv("TEST_KEY"); got != "keep-env" {
			t.Fatalf("expected keep-env, got %q", got)
		}
	})
}
