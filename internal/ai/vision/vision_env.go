package vision

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/photoprism/photoprism/internal/ai/vision/ollama"
	"github.com/photoprism/photoprism/internal/ai/vision/openai"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

var ensureEnvOnce sync.Once

// ensureEnv loads environment-backed credentials once so adapters can look up
// OPENAI_API_KEY / OLLAMA_API_KEY even when operators rely on *_FILE fallbacks.
// Future engine integrations can reuse this hook to normalize additional
// secrets.
func ensureEnv() {
	ensureEnvOnce.Do(func() {
		loadEnvKeyFromFile(openai.APIKeyEnv, openai.APIKeyFileEnv)
		loadEnvKeyFromFile(ollama.APIKeyEnv, ollama.APIKeyFileEnv)

		// Init the Ollama base URL by trimming trailing slashes or using the default.
		initEnvUrl(ollama.BaseUrlEnv, ollama.DefaultBaseUrl)
	})
}

// initEnvUrl ensures that the variable contains no trailing
// slashes and sets a default value if it is missing.
func initEnvUrl(envName, defaultUrl string) {
	if base := strings.TrimSpace(os.Getenv(envName)); base != "" {
		if normalized := strings.TrimRight(base, "/"); normalized != base {
			_ = os.Setenv(envName, normalized)
		}
	} else if defaultUrl != "" {
		_ = os.Setenv(envName, defaultUrl)
	}
}

// loadEnvKeyFromFile populates envVar from fileVar when the environment value
// is empty and the referenced file exists and is non-empty.
func loadEnvKeyFromFile(envVar, fileVar string) {
	if os.Getenv(envVar) != "" {
		return
	}

	filePath := strings.TrimSpace(os.Getenv(fileVar))

	if !fs.FileExistsNotEmpty(filePath) {
		return
	}

	filePath = filepath.Clean(filePath)

	// #nosec G304,G703 path is validated and intended for local secret file loading.
	if data, err := os.ReadFile(filePath); err == nil {
		if key := clean.Auth(string(data)); key != "" {
			_ = os.Setenv(envVar, key)
		}
	}
}
