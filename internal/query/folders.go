package query

import (
	"path/filepath"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
)

type Folders []entity.Folder

// FoldersByPath returns a slice of folders in a given directory incl sub directories in recursive mode.
func FoldersByPath(rootName, rootPath, path string, recursive bool) (folders Folders, err error) {
	dirs, err := fs.Dirs(filepath.Join(rootPath, path), recursive)

	if err != nil {
		return folders, err
	}

	folders = make(Folders, len(dirs))

	for i, dir := range dirs {
		folder := entity.NewFolder(rootName, filepath.Join(path, dir), nil)

		if f := entity.FirstOrCreateFolder(&folder); f != nil {
			folders[i] = *f
		}
	}

	return folders, nil
}
