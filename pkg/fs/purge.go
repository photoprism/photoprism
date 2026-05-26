package fs

import (
	gofs "io/fs"
	"os"
	"path/filepath"
	"strings"
)

// PurgeTestDbFiles removes hidden SQLite test artifacts (`.*.db`,
// `.*.db-journal`, `.test.*`) from dir, optionally recursively. Aligned with
// `make reset-sqlite`. Removal errors are ignored — best-effort cleanup.
func PurgeTestDbFiles(dir string, recursive bool) {
	if dir == "" {
		return
	}

	// Common predicate used by both modes.
	matchAndRemove := func(path, name string, info os.FileInfo) {
		if info == nil || !info.Mode().IsRegular() {
			return
		}
		lower := strings.ToLower(name)
		if strings.HasPrefix(name, ".") {
			if strings.HasSuffix(lower, ".db") || strings.HasSuffix(lower, ".db-journal") || strings.HasPrefix(lower, ".test.") {
				_ = os.Remove(path)
			}
		}
	}

	if recursive {
		_ = filepath.WalkDir(dir, func(path string, d gofs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				return nil
			}
			// Gather FileInfo to ensure regular file.
			if info, statErr := d.Info(); statErr == nil {
				matchAndRemove(path, d.Name(), info)
			}
			return nil
		})
		return
	}

	// Non-recursive: only immediate entries in dir.
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if info, statErr := e.Info(); statErr == nil {
			matchAndRemove(filepath.Join(dir, e.Name()), e.Name(), info)
		}
	}
}
