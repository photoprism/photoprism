package vision

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

	// Case: trims surrounding whitespace.
	t.Setenv(envName, " https://llm.example.com/v1 ")
	initEnvUrl(envName, "http://default.local")
	if got := os.Getenv(envName); got != "https://llm.example.com/v1" {
		t.Fatalf("space: expected https://llm.example.com/v1, got %s", got)
	}

	// Case: sets default when only slashes remain.
	t.Setenv(envName, "///")
	initEnvUrl(envName, "http://default.local")
	if got := os.Getenv(envName); got != "http://default.local" {
		t.Fatalf("slashes: expected http://default.local, got %s", got)
	}

	// Case: trims the trailing slash of the path and keeps the query.
	t.Setenv(envName, "https://gw.example.com/v1/?token=abc")
	initEnvUrl(envName, "http://default.local")
	if got := os.Getenv(envName); got != "https://gw.example.com/v1?token=abc" {
		t.Fatalf("query: expected https://gw.example.com/v1?token=abc, got %s", got)
	}

	// Case: keeps a value whose path ends with a query unchanged.
	t.Setenv(envName, "https://gw.example.com/v1?api-version=preview/")
	initEnvUrl(envName, "http://default.local")
	if got := os.Getenv(envName); got != "https://gw.example.com/v1?api-version=preview/" {
		t.Fatalf("query slash: expected the value unchanged, got %s", got)
	}

	// Case: trims the trailing slash of the path before a fragment.
	t.Setenv(envName, "https://gw.example.com/v1/#top")
	initEnvUrl(envName, "http://default.local")
	if got := os.Getenv(envName); got != "https://gw.example.com/v1#top" {
		t.Fatalf("fragment: expected https://gw.example.com/v1#top, got %s", got)
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

		if got := getEnv("TEST_KEY"); got != "file-secret" {
			t.Fatalf("expected file-secret, got %q", got)
		}

		if got := os.Getenv("TEST_KEY"); got != "" {
			t.Fatalf("expected the key to stay out of the environment, got %q", got)
		}

		if got := expandEnv("Bearer ${TEST_KEY}"); got != "Bearer file-secret" {
			t.Fatalf("expected the key to expand, got %q", got)
		}
	})
	t.Run("ChildProcess", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "key.txt")
		if err := os.WriteFile(path, []byte("file-secret\n"), 0o600); err != nil {
			t.Fatalf("write key file: %v", err)
		}

		t.Setenv("TEST_KEY", "")
		t.Setenv("TEST_KEY_FILE", path)

		loadEnvKeyFromFile("TEST_KEY", "TEST_KEY_FILE")

		out, err := exec.Command("env").Output()

		if err != nil {
			t.Skipf("env not available: %v", err)
		} else if strings.Contains(string(out), "file-secret") {
			t.Fatal("expected child processes not to inherit the key")
		}
	})
	t.Run("EnvWinsOverFile", func(t *testing.T) {
		t.Setenv("TEST_KEY", "keep-env")
		t.Setenv("TEST_KEY_FILE", "/nonexistent")

		loadEnvKeyFromFile("TEST_KEY", "TEST_KEY_FILE")

		if got := getEnv("TEST_KEY"); got != "keep-env" {
			t.Fatalf("expected keep-env, got %q", got)
		}
	})
	t.Run("IgnoreDirectoryPath", func(t *testing.T) {
		t.Setenv("TEST_KEY", "")
		t.Setenv("TEST_KEY_FILE", t.TempDir())

		loadEnvKeyFromFile("TEST_KEY", "TEST_KEY_FILE")

		if got := getEnv("TEST_KEY"); got != "" {
			t.Fatalf("expected empty key, got %q", got)
		}
	})
}

func TestExpandEnv(t *testing.T) {
	t.Run("Environment", func(t *testing.T) {
		t.Setenv("TEST_EXPAND_KEY", "env-value")
		if got := ExpandEnv("Bearer ${TEST_EXPAND_KEY}"); got != "Bearer env-value" {
			t.Fatalf("expected the variable to expand, got %q", got)
		}
	})
	t.Run("FileKey", func(t *testing.T) {
		t.Setenv("TEST_EXPAND_KEY", "")
		envKeys.Store("TEST_EXPAND_KEY", "file-value")
		t.Cleanup(func() { envKeys.Delete("TEST_EXPAND_KEY") })
		if got := ExpandEnv("${TEST_EXPAND_KEY}"); got != "file-value" {
			t.Fatalf("expected the file key to expand, got %q", got)
		}
	})
}

func TestLookupEnv(t *testing.T) {
	t.Run("Environment", func(t *testing.T) {
		t.Setenv("TEST_LOOKUP_KEY", "env-value")
		value, set := lookupEnv("TEST_LOOKUP_KEY")
		if value != "env-value" || !set {
			t.Fatalf("expected env-value, got %q (%t)", value, set)
		}
	})
	t.Run("FileKey", func(t *testing.T) {
		t.Setenv("TEST_LOOKUP_KEY", "")
		envKeys.Store("TEST_LOOKUP_KEY", "file-value")
		t.Cleanup(func() { envKeys.Delete("TEST_LOOKUP_KEY") })
		value, set := lookupEnv("TEST_LOOKUP_KEY")
		if value != "file-value" || !set {
			t.Fatalf("expected file-value, got %q (%t)", value, set)
		}
	})
	t.Run("Unset", func(t *testing.T) {
		value, set := lookupEnv("TEST_LOOKUP_KEY_UNSET")
		if value != "" || set {
			t.Fatalf("expected no value, got %q (%t)", value, set)
		}
	})
}
