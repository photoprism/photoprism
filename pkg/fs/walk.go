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
		// Check if symlink points to a directory.
		if link, err := os.Stat(fileName); err == nil && link.IsDir() {
			// Skip directories.
			skip = true
			resolved, evalErr := filepath.EvalSymlinks(fileName)

			if evalErr != nil || isIgnored || isDone || done[resolved].Exists() {
				// Skip symlinked directories that have errors, are hidden, or are already done.
				result = filepath.SkipDir
			} else if FileExists(filepath.Join(resolved, PPStorageFilename)) {
				// Skip symlinked directories that contain a .ppstorage file.
				isIgnored = true
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
			// Skip directories that are hidden or already done.
			result = filepath.SkipDir
		} else if FileExists(filepath.Join(fileName, PPStorageFilename)) {
			// Skip directories that contain a .ppstorage file.
			isIgnored = true
			result = filepath.SkipDir
		}
	} else if isIgnored || isDone {
		// Skip files that are hidden or already done.
		skip = true
	}

	if skip {
		done[fileName] = Found
	}

	return skip, result
}
