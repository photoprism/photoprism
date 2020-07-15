package fs

import (
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/karrick/godirwalk"
)

var OriginalPaths = []string{
	"/photoprism/photos/originals",
	"/photoprism/storage/originals",
	"/photoprism/originals",
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
	"storage/originals",
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
}

var ImportPaths = []string{
	"/photoprism/photos/import",
	"/photoprism/storage/import",
	"/photoprism/import",
	"photoprism/import",
	"PhotoPrism/Import",
	"pictures/import",
	"Pictures/Import",
	"photos/import",
	"Photos/Import",
	"storage/import",
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
}

func Dirs(root string, recursive bool) (result []string, err error) {
	result = []string{}
	ignore := NewIgnoreList(".ppignore", true, false)
	mutex := sync.Mutex{}

	err = godirwalk.Walk(root, &godirwalk.Options{Callback: func(fileName string, info *godirwalk.Dirent) error {
		isDirlike, err := info.IsDirOrSymlinkToDir()
		if err != nil {
			return err
		}
		if isDirlike {
			if ignore.Ignore(fileName) {
				return filepath.SkipDir
			}

			if fileName != root {
				mutex.Lock()
				fileName = strings.Replace(fileName, root, "", 1)
				result = append(result, fileName)
				mutex.Unlock()

				if !recursive {
					return filepath.SkipDir
				}
			}
		}

		return nil
	},
		Unsorted:            false,
		FollowSymbolicLinks: true,
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
