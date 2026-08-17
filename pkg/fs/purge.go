package fs

import (
	gofs "io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// PurgeExpired removes regular files in dir that end with ext and are older than maxAge, and reports how
// many were removed, how many matching files are left on disk (including any that could not be deleted),
// and how many deletions failed. It is not recursive, an empty ext matches every regular file, and a
// maxAge of zero or less or an unreadable dir removes nothing and reports zeros.
func PurgeExpired(dir, ext string, maxAge time.Duration) (removed, remaining, failed int) {
	if dir == "" || maxAge <= 0 {
		return 0, 0, 0
	}

	entries, err := os.ReadDir(dir)

	if err != nil {
		return 0, 0, 0
	}

	ext = strings.ToLower(ext)
	cutoff := time.Now().Add(-maxAge)

	for _, entry := range entries {
		if entry.IsDir() || ext != "" && !strings.HasSuffix(strings.ToLower(entry.Name()), ext) {
			continue
		}

		info, infoErr := entry.Info()

		if infoErr != nil || !info.Mode().IsRegular() {
			continue
		}

		if info.ModTime().After(cutoff) {
			remaining++
			continue
		}

		if os.Remove(filepath.Join(dir, entry.Name())) != nil {
			failed++
			remaining++
		} else {
			removed++
		}
	}

	return removed, remaining, failed
}

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
			if strings.HasSuffix(lower, ".db") || strings.HasSuffix(lower, ".db-journal") || strings.HasSuffix(lower, ".db-shm") || strings.HasSuffix(lower, ".db-wal") || strings.HasPrefix(lower, ".test.") {
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
