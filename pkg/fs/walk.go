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
		// Symlinks are skipped by default unless they are links to directories
		skip = true

		// Symlink points to directory?
		if link, err := os.Stat(fileName); err == nil && link.IsDir() {
			resolved, err := filepath.EvalSymlinks(fileName)

			if err != nil || isIgnored || isDone || done[resolved].Exists() {
				result = filepath.SkipDir
			} else {
				// Don't traverse symlinks if they are hidden or already done...
				done[resolved] = Found
			}
		} else {
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
