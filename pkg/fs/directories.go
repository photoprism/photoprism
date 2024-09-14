package fs

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/photoprism/photoprism/pkg/fs/fastwalk"
)

var OriginalPaths = []string{
	"/photoprism/storage/media/originals",
	"/photoprism/media/originals",
	"/photoprism/originals",
	"/srv/photoprism/storage/media/originals",
	"/srv/photoprism/media/originals",
	"/srv/photoprism/originals",
	"/opt/photoprism/storage/media/originals",
	"/opt/photoprism/media/originals",
	"/opt/photoprism/originals",
	"/media/originals",
	"/storage/originals",
	"/originals",
	"media/originals",
	"storage/originals",
	"photoprism/originals",
	"PhotoPrism/Originals",
	"photoprism/original",
	"PhotoPrism/Original",
	"pictures/originals",
	"Pictures/Originals",
	"pictures/original",
	"Pictures/Original",
	"photos/originals",
	"Photos/Originals",
	"photos/original",
	"Photos/Original",
	"originals",
	"Originals",
	"original",
	"Original",
	"pictures",
	"Pictures",
	"photos",
	"Photos",
	"images",
	"Images",
	"bilder",
	"Bilder",
	"fotos",
	"Fotos",
	"~/photoprism/originals",
	"~/PhotoPrism/Originals",
	"~/photoprism/original",
	"~/PhotoPrism/Original",
	"~/pictures/originals",
	"~/Pictures/Originals",
	"~/pictures/original",
	"~/Pictures/Original",
	"~/photos/originals",
	"~/Photos/Originals",
	"~/photos/original",
	"~/Photos/Original",
	"~/pictures",
	"~/Pictures",
	"~/photos",
	"~/Photos",
	"~/images",
	"~/Images",
	"~/bilder",
	"~/Bilder",
	"~/fotos",
	"~/Fotos",
	"/var/lib/photoprism/originals",
}

var ImportPaths = []string{
	"/photoprism/storage/media/import",
	"/photoprism/media/import",
	"/photoprism/import",
	"/srv/photoprism/storage/media/import",
	"/srv/photoprism/media/import",
	"/srv/photoprism/import",
	"/opt/photoprism/storage/media/import",
	"/opt/photoprism/media/import",
	"/opt/photoprism/import",
	"/media/import",
	"/storage/import",
	"/import",
	"media/import",
	"storage/import",
	"photoprism/import",
	"PhotoPrism/Import",
	"pictures/import",
	"Pictures/Import",
	"photos/import",
	"Photos/Import",
	"import",
	"Import",
	"~/pictures/import",
	"~/Pictures/Import",
	"~/photoprism/import",
	"~/PhotoPrism/Import",
	"~/photos/import",
	"~/Photos/Import",
	"~/import",
	"~/Import",
	"/var/lib/photoprism/import",
}

var AssetPaths = []string{
	"/opt/photoprism/assets",
	"/photoprism/assets",
	"~/.photoprism/assets",
	"~/photoprism/assets",
	"photoprism/assets",
	"assets",
	"/var/lib/photoprism/assets",
}

// Dirs returns a slice of directories in a path, optional recursively and with symlinks.
//
// Warning: Following symlinks can make the result non-deterministic and hard to test!
func Dirs(root string, recursive bool, followLinks bool) (result []string, err error) {
	result = []string{}
	mutex := sync.Mutex{}

	// Ignore hidden folders as well as those listed in an optional ".ppignore" file.
	ignore := NewIgnoreList(PPIgnoreFilename, true, false)

	symlinks := make(map[string]bool)
	symlinksMutex := sync.Mutex{}

	// appendResult adds the relative path of a subdirectory to the results.
	appendResult := func(dir string) {
		mutex.Lock()
		defer mutex.Unlock()
		result = append(result, strings.Replace(dir, root, "", 1))
	}

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
			} else if FileExists(filepath.Join(dir, PPStorageFilename)) {
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
				} else if resolved, resolveErr := Resolve(dir); resolveErr == nil {
					symlinksMutex.Lock()
					defer symlinksMutex.Unlock()

					if _, ok := symlinks[resolved]; ok {
						return filepath.SkipDir
					} else {
						symlinks[resolved] = true
						appendResult(dir)
					}

					return fastwalk.ErrTraverseLink
				}
			}
		}

		return nil
	})

	sort.Strings(result)

	return result, err
}

// FindDir checks if any of the specified directories exist and returns the absolute path of the first directory found.
func FindDir(dirs []string) string {
	for _, dir := range dirs {
		absDir := Abs(dir)
		if PathExists(absDir) {
			return absDir
		}
	}

	return ""
}

// MkdirAll creates a directory including all parent directories that might not yet exist.
// No error is returned if the directory already exists.
func MkdirAll(dir string) error {
	return os.MkdirAll(dir, ModeDir)
}
