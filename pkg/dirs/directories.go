package dirs

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/fs/fastwalk"
	"github.com/photoprism/photoprism/pkg/media"
)

// Dirs returns a slice of directories in a path with fileCounts where applicable and optionally recursively and/or with symlinks.
// Note: If recursive is false, then sub directories will not have fileCounts.
//
// Warning: Following symlinks can make the result non-deterministic and hard to test!
func Dirs(root string, recursive, followLinks bool) (result []string, counts map[string]int, err error) {
	result = []string{}
	counts = make(map[string]int)
	mutex := sync.Mutex{}

	// Ignore hidden folders as well as those listed in an optional ".ppignore" file.
	ignore := fs.NewIgnoreList(fs.PPIgnoreFilename, true, false)

	symlinks := make(map[string]bool)
	symlinksMutex := sync.Mutex{}

	// appendResult adds the relative path of a subdirectory to the results.
	appendResult := func(dir string) {
		mutex.Lock()
		defer mutex.Unlock()
		result = append(result, strings.Replace(dir, root, "", 1))
		if tmpStr := strings.Replace(dir, root, "", 1); len(tmpStr) == 0 {
			counts["/"] = 0
		} else {
			counts[tmpStr] = 0
		}
	}

	// incrementFiles increments the file counter for the subdirectory in the results.
	incrementFiles := func(dir string) {
		mutex.Lock()
		defer mutex.Unlock()
		if tmpStr := strings.Replace(dir, root, "", 1); len(tmpStr) == 0 {
			counts["/"]++
		} else {
			counts[tmpStr]++
		}
	}

	result = append(result, "/")

	err = fastwalk.Walk(root, func(dir string, mode os.FileMode) error {
		if mode.IsDir() || mode == os.ModeSymlink && followLinks {
			// Skip if symlink does not point to existing directory.
			if mode == os.ModeSymlink {
				if info, statErr := os.Stat(dir); statErr != nil || !info.IsDir() {
					return filepath.SkipDir
				}
			}

			// Skip if directory should be ignored.
			if _ = ignore.Path(dir); ignore.Ignore(dir) {
				return filepath.SkipDir
			} else if fs.FileExists(filepath.Join(dir, fs.PPStorageFilename)) {
				return filepath.SkipDir
			}

			// Only add subdirectories.
			if dir != root {
				if !recursive {
					appendResult(dir)

					return filepath.SkipDir
				} else if mode != os.ModeSymlink {
					appendResult(dir)

					return nil
				} else if resolved, resolveErr := fs.Resolve(dir); resolveErr == nil {
					symlinksMutex.Lock()
					defer symlinksMutex.Unlock()

					if _, ok := symlinks[resolved]; ok {
						return filepath.SkipDir
					}
					symlinks[resolved] = true
					appendResult(dir)

					return fastwalk.ErrTraverseLink
				}
			}
		} else if mode.IsRegular() {
			if media.FromName(dir).IsMain() {
				incrementFiles(filepath.Dir(dir))
			}
		}

		return nil
	})

	sort.Strings(result)

	return result, counts, err
}
