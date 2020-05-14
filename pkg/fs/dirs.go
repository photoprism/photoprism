package fs

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/photoprism/photoprism/pkg/fastwalk"
)

func Dirs(root string, recursive bool) (result []string, err error) {
	result = []string{}
	ignore := NewIgnoreList(".ppignore", true, false)
	mutex := sync.Mutex{}

	err = fastwalk.Walk(root, func(fileName string, info os.FileMode) error {
		if info.IsDir() {
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
	})

	sort.Strings(result)

	return result, err
}
