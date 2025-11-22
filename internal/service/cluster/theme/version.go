package theme

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// DetectVersion returns the sanitized theme version for the given directory.
// It prefers a version.txt file when present and falls back to the app.js
// modification timestamp when the version file is missing or empty.
func DetectVersion(themePath string) (string, error) {
	if themePath == "" {
		return "", errors.New("empty theme path")
	}

	info, err := os.Stat(filepath.Join(themePath, fs.AppJsFile))

	if err != nil {
		return "", err
	} else if info.IsDir() {
		return "", fmt.Errorf("%s must not be a directory", fs.AppJsFile)
	}

	versionFile := filepath.Join(themePath, fs.VersionTxtFile)

	if data, readErr := os.ReadFile(versionFile); readErr == nil {
		if v := clean.TypeUnicode(string(data)); v != "" {
			return v, nil
		}
	}

	return clean.TypeUnicode(info.ModTime().UTC().Format(time.RFC3339)), nil
}
