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

// ensureEnv loads the engine API keys from their *_FILE variables and normalizes the engine base URLs once,
// so that adapters and variable expansion read the final values.
func ensureEnv() {
	ensureEnvOnce.Do(func() {
		loadEnvKeyFromFile(openai.APIKeyEnv, openai.APIKeyFileEnv)
		loadEnvKeyFromFile(ollama.APIKeyEnv, ollama.APIKeyFileEnv)

		// Init the base URLs by trimming trailing slashes or using the defaults.
		initEnvUrl(ollama.BaseUrlEnv, ollama.DefaultBaseUrl)
		initEnvUrl(openai.BaseUrlEnv, openai.DefaultBaseUrl)
	})
}

// initEnvUrl removes surrounding whitespace and the trailing slashes of the path from the variable, and
// sets the default value if nothing is left.
func initEnvUrl(envName, defaultUrl string) {
	raw := os.Getenv(envName)
	normalized := strings.TrimSpace(raw)

	if i := strings.IndexAny(normalized, "?#"); i >= 0 {
		normalized = strings.TrimRight(normalized[:i], "/") + normalized[i:]
	} else {
		normalized = strings.TrimRight(normalized, "/")
	}

	if normalized != "" {
		if normalized != raw {
			_ = os.Setenv(envName, normalized)
		}
	} else if defaultUrl != "" {
		_ = os.Setenv(envName, defaultUrl)
	}
}

// envKeys holds the API keys loaded from *_FILE variables. They are kept here rather than in the
// environment, so child processes do not inherit them.
var envKeys sync.Map

// loadEnvKeyFromFile loads the key for envVar from the file named in fileVar when envVar is empty and
// the referenced file exists and is not empty.
func loadEnvKeyFromFile(envVar, fileVar string) {
	envKeys.Delete(envVar)

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
			envKeys.Store(envVar, key)
		}
	}
}

// lookupEnv returns the value of the environment variable, or the key loaded for it from a file, and
// whether either is set.
func lookupEnv(name string) (string, bool) {
	if value, set := os.LookupEnv(name); set && value != "" {
		return value, true
	} else if key, found := envKeys.Load(name); found {
		return key.(string), true
	} else {
		return value, set
	}
}

// getEnv returns the value of the environment variable, or the key loaded for it from a file.
func getEnv(name string) string {
	value, _ := lookupEnv(name)
	return value
}

// expandEnv replaces ${var} or $var in s like os.ExpandEnv, including the keys loaded from files.
func expandEnv(s string) string {
	return os.Expand(s, getEnv)
}

// ExpandEnv replaces ${var} or $var in s like os.ExpandEnv, including the API keys loaded from files.
func ExpandEnv(s string) string {
	ensureEnv()
	return expandEnv(s)
}

// envModel returns the model identifier set in the environment variable, or an empty string.
func envModel(name string) string {
	return cleanModelId(strings.TrimSpace(os.Getenv(name)))
}

// envModelTagged returns the model identifier set in the environment variable without shortening it, so
// Model.GetModel can keep its tag, or an empty string.
func envModelTagged(name string) string {
	return modelIdText(os.Getenv(name))
}
