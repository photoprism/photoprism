package fs

import (
	"os"
	"path/filepath"
)

// SkipWalk returns true if the file or directory should be skipped in godirwalk.Walk()
func SkipWalk(name string, isDir, isSymlink bool, done Done, ignore *IgnoreList) (skip bool, result error) {
	isDone := done[name].Exists()

	if isSymlink {
		// Check if symlink points to a directory.
		if link, err := os.Stat(name); err == nil && link.IsDir() {
			// Skip directories.
			skip = true

			resolved, evalErr := filepath.EvalSymlinks(name)

			if evalErr == nil {
				_ = ignore.Path(name)
			}

			// Skip symlinked directories that cannot be resolved or are ignored, hidden, or already done.
			if ignore.Ignore(name) || evalErr != nil || isDone || done[resolved].Exists() {
				result = filepath.SkipDir
			} else if FileExists(filepath.Join(resolved, PPStorageFilename)) {
				// Skip symlinked directories that contain a .ppstorage file.
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

		if _ = ignore.Path(name); ignore.Ignore(name) || isDone {
			// Skip directories that are hidden or already done.
			result = filepath.SkipDir
		} else if FileExists(filepath.Join(name, PPStorageFilename)) {
			// Skip directories that contain a .ppstorage file.
			result = filepath.SkipDir
		}
	} else if ignore.Ignore(name) || isDone {
		// Skip files that are hidden or already done.
		skip = true
	}

	if skip {
		done[name] = Found
	}

	return skip, result
}
