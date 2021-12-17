package fs

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/photoprism/photoprism/pkg/fastwalk"
)

var OriginalPaths = []string{
	"/photoprism/library/originals",
	"/photoprism/storage/originals",
	"/photoprism/originals",
	"/opt/photoprism/library/originals",
	"/opt/photoprism/storage/originals",
	"/opt/photoprism/originals",
	"/library/originals",
	"/storage/originals",
	"/originals",
	"library/originals",
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
	"/photoprism/library/import",
	"/photoprism/storage/import",
	"/photoprism/import",
	"/opt/photoprism/library/import",
	"/opt/photoprism/storage/import",
	"/opt/photoprism/import",
	"/library/import",
	"/storage/import",
	"/import",
	"library/import",
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
	"/photoprism/assets",
	"/opt/photoprism/assets",
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

	symlinks := make(map[string]bool)
	symlinksMutex := sync.Mutex{}

	appendResult := func(fileName string) {
		fileName = strings.Replace(fileName, root, "", 1)
		mutex.Lock()
		defer mutex.Unlock()
		result = append(result, fileName)
	}

	// Ignore hidden folders as well as those listed in an optional ".ppignore" file.
	ignore := NewIgnoreList(IgnoreFile, true, false)
	_ = ignore.Dir(root)

	err = fastwalk.Walk(root, func(fileName string, typ os.FileMode) error {
		if typ.IsDir() || typ == os.ModeSymlink && followLinks {
			if ignore.Ignore(fileName) {
				return filepath.SkipDir
			}

			if fileName != root {
				if !recursive {
					appendResult(fileName)

					return filepath.SkipDir
				} else if typ != os.ModeSymlink {
					appendResult(fileName)

					return nil
				} else if resolved, err := filepath.EvalSymlinks(fileName); err == nil {
					symlinksMutex.Lock()
					defer symlinksMutex.Unlock()

					if _, ok := symlinks[resolved]; ok {
						return filepath.SkipDir
					} else {
						symlinks[resolved] = true
						appendResult(fileName)
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

func FindDir(dirs []string) string {
	for _, dir := range dirs {
		absDir := Abs(dir)
		if PathExists(absDir) {
			return absDir
		}
	}

	return ""
}
