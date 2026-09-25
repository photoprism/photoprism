package commands

import (
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/pkg/clean"
)

// sanitizeSubfolderArg normalizes a subfolder argument and returns an exit-coded
// error when a non-empty value is invalid.
func sanitizeSubfolderArg(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	subPath := clean.UserPath(raw)

	if raw != "" && subPath == "" {
		return "", cli.Exit("invalid subfolder path", 2)
	}

	return subPath, nil
}

// sanitizeDestinationArg normalizes a destination flag value and returns an
// exit-coded error when a non-empty value is invalid.
func sanitizeDestinationArg(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	dest := clean.UserPath(raw)

	if raw != "" && dest == "" {
		return "", cli.Exit("invalid destination path", 2)
	}

	return dest, nil
}

// absPathArg returns the absolute form of a path flag value, or the value as given if it is empty or
// cannot be resolved, so the command checks the same directory the backup package later uses.
func absPathArg(raw string) string {
	if raw == "" {
		return ""
	}

	if abs, err := filepath.Abs(raw); err == nil {
		return abs
	}

	return raw
}
