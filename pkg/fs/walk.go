package fs

import (
	"os"
	"path/filepath"
)

// SkipWalk returns true if the file or directory should be skipped in godirwalk.Walk()
func SkipWalk(fileName string, isDir, isSymlink bool, done Done, ignore *IgnoreList) (skip bool, result error) {
	isDone := done[fileName].Exists()
	isIgnored := ignore.Ignore(fileName)

	if isSymlink {
		// Symlink points to directory?
		if link, err := os.Stat(fileName); err == nil && link.IsDir() {
			// Skip directories.
			skip = true
			resolved, err := filepath.EvalSymlinks(fileName)

			if err != nil || isIgnored || isDone || done[resolved].Exists() {
				result = filepath.SkipDir
			} else {
				// Flag the symlink target as processed.
				done[resolved] = Found
			}
		} else if err != nil {
			// Also skip on error.
			skip = true
			result = filepath.SkipDir
		}
	} else if isDir {
		skip = true

		if isIgnored || isDone {
			// Don't traverse directories if they are hidden or already done...
			result = filepath.SkipDir
		}
	} else if isIgnored || isDone {
		// Skip files that are hidden or already done...
		skip = true
	}

	if skip {
		done[fileName] = Found
	}

	return skip, result
}
