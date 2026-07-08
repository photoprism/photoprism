package vision

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitEnvUrl(t *testing.T) {
	const envName = "TEST_OLLAMA_BASE_URL"

	// Case: trims trailing slash.
	t.Setenv(envName, "http://example.com/")
	initEnvUrl(envName, "")
	if got := os.Getenv(envName); got != "http://example.com" {
		t.Fatalf("trim: expected http://example.com, got %s", got)
	}

	// Case: sets default when unset.
	t.Setenv(envName, "")
	initEnvUrl(envName, "http://default.local")
	if got := os.Getenv(envName); got != "http://default.local" {
		t.Fatalf("default: expected http://default.local, got %s", got)
	}

	// Case: leaves already-normalized value untouched.
	t.Setenv(envName, "http://kept.local")
	initEnvUrl(envName, "http://ignored.local")
	if got := os.Getenv(envName); got != "http://kept.local" {
		t.Fatalf("preserve: expected http://kept.local, got %s", got)
	}
}

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
	t.Run("IgnoreDirectoryPath", func(t *testing.T) {
		t.Setenv("TEST_KEY", "")
		t.Setenv("TEST_KEY_FILE", t.TempDir())

		loadEnvKeyFromFile("TEST_KEY", "TEST_KEY_FILE")

		if got := os.Getenv("TEST_KEY"); got != "" {
			t.Fatalf("expected empty key, got %q", got)
		}
	})
}
