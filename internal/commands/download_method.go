package commands

import (
	"fmt"
	"os"
	"strings"
)

const downloadMethodEnv = "PHOTOPRISM_DL_METHOD"

// resolveDownloadMethod normalizes the download method by honoring the explicit flag
// value first, then the environment variable PHOTOPRISM_DL_METHOD, and finally
// defaulting to "pipe". It returns the resolved method, a boolean indicating whether
// the value originated from the environment, or an error when the input is invalid.
func resolveDownloadMethod(flagValue string) (string, bool, error) {
	trimmed := strings.TrimSpace(flagValue)
	method := strings.ToLower(trimmed)
	fromEnv := false

	if method == "" {
		envValue := strings.TrimSpace(os.Getenv(downloadMethodEnv))
		if envValue != "" {
			method = strings.ToLower(envValue)
			trimmed = envValue
			fromEnv = true
		}
	}

	if method == "" {
		return "pipe", false, nil
	}

	if method != "pipe" && method != "file" {
		if fromEnv {
			return "", true, fmt.Errorf("invalid %s value: %s (expected 'pipe' or 'file')", downloadMethodEnv, trimmed)
		}
		return "", false, fmt.Errorf("invalid download method: %s (expected 'pipe' or 'file')", trimmed)
	}

	return method, fromEnv, nil
}
