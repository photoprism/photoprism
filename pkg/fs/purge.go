package fs

import (
	gofs "io/fs"
	"os"
	"path/filepath"
	"strings"
)

// PurgeTestDbFiles removes temporary SQLite-related test files in dir.
//
// Patterns (case-insensitive), aligned with `make reset-sqlite`:
//   - '.*.db'          (hidden SQLite database files)
//   - '.*.db-journal'  (SQLite journal files)
//   - '.test.*'        (generic hidden test artifacts)
//
// If recursive is true, it traverses dir recursively. If false, it only checks
// files directly within dir so TestMain in a parent package won't affect
// sub-packages that may run in parallel.
//
// Errors from removing individual files are ignored; this is a best-effort
// cleanup helper for tests and local tooling.
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
