package fs

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/karrick/godirwalk"
)

// SkipGodirwalk returns true if the file or directory should be skipped in godirwalk.Walk()
func SkipGodirwalk(fileName string, info *godirwalk.Dirent, done map[string]bool) (skip bool, result error) {
	isDone := done[fileName]
	isHidden := strings.HasPrefix(filepath.Base(fileName), ".")
	isDir := info.IsDir()
	isSymlink := info.IsSymlink()

	done[fileName] = true

	if isSymlink {
		skip = true

		// Symlink points to directory?
		if link, err := os.Stat(fileName); err == nil && link.IsDir() && (isHidden || isDone || done[link.Name()]) {
			// Don't traverse symlinks if they are hidden or already done...
			done[link.Name()] = true
			result = filepath.SkipDir
		}
	} else if isDir {
		skip = true

		if isHidden || isDone {
			// Don't traverse directories if they are hidden or already done...
			result = filepath.SkipDir
		}
	} else if isHidden || isDone {
		// Skip files that are hidden or already done...
		skip = true
	}

	return skip, result
}
