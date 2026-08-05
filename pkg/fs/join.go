package fs

import (
	"fmt"
	"path/filepath"
	"strings"
)

// SafeJoin joins a base directory with an untrusted relative name and returns the
// resulting path only if it stays within the base directory. Absolute paths,
// Windows-style volume names, drive-letter prefixes, and parent-directory
// traversal are rejected, with containment verified via filepath.Rel rather than
// a string prefix. It is the shared safe-join used for archive extraction and for
// resolving remote or user-supplied names to a local destination path.
func SafeJoin(baseDir, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("invalid path")
	}

	// Normalize separators so mixed '/' and '\\' are handled consistently.
	name = strings.ReplaceAll(name, "\\", "/")

	// Reject Windows-style drive-letter prefixes even on non-Windows platforms.
	if len(name) >= 2 && name[1] == ':' && ((name[0] >= 'A' && name[0] <= 'Z') || (name[0] >= 'a' && name[0] <= 'z')) {
		return "", fmt.Errorf("invalid path: absolute or volume path not allowed")
	}

	// Reject absolute or volume paths.
	if filepath.IsAbs(name) || filepath.VolumeName(name) != "" {
		return "", fmt.Errorf("invalid path: absolute or volume path not allowed")
	}

	cleaned := filepath.Clean(name)
	base := filepath.Clean(baseDir)

	// Compose destination and verify it stays inside base using filepath.Rel.
	dest := filepath.Join(base, cleaned)
	rel, err := filepath.Rel(base, dest)

	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	} else if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid path: outside base directory")
	}

	return dest, nil
}
