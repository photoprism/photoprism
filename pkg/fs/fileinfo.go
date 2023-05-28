package fs

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/emersion/go-webdav"
)

// FileInfo represents a file system entry.
type FileInfo struct {
	Name string    `json:"name"`
	Abs  string    `json:"abs"`
	Size int64     `json:"size"`
	Date time.Time `json:"date"`
	Dir  bool      `json:"dir"`
}

func fileDir(dir, sep string) string {
	if dir != sep && len(dir) > 0 {
		if dir[len(dir)-1:] == sep {
			dir = dir[:len(dir)-1]
		}

		if dir[0:1] != sep {
			dir = sep + dir
		}
	} else {
		dir = sep
	}

	return dir
}

// NewFileInfo creates a FileInfo struct from the os.FileInfo record.
func NewFileInfo(file os.FileInfo, dir string) FileInfo {
	dir = fileDir(dir, PathSeparator)

	result := FileInfo{
		Name: file.Name(),
		Abs:  filepath.Join(dir, file.Name()),
		Size: file.Size(),
		Date: file.ModTime(),
		Dir:  file.IsDir(),
	}

	return result
}

// WebFileInfo creates a FileInfo struct from a webdav.FileInfo record.
func WebFileInfo(file webdav.FileInfo, dir string) FileInfo {
	filePath := strings.Trim(file.Path, "/")
	dir = strings.Trim(dir, "/")
	result := FileInfo{
		Name: path.Base(filePath),
		Abs:  "/" + RelName(filePath, dir),
		Size: file.Size,
		Date: file.ModTime,
		Dir:  file.IsDir,
	}

	return result
}

type FileInfos []FileInfo

func (infos FileInfos) Len() int      { return len(infos) }
func (infos FileInfos) Swap(i, j int) { infos[i], infos[j] = infos[j], infos[i] }
func (infos FileInfos) Less(i, j int) bool {
	return strings.Compare(infos[i].Abs, infos[j].Abs) == -1
}
func (infos FileInfos) Abs() (result []string) {
	for _, info := range infos {
		result = append(result, info.Abs)
	}

	return result
}

func NewFileInfos(infos []os.FileInfo, dir string) FileInfos {
	var result FileInfos

	for _, info := range infos {
		result = append(result, NewFileInfo(info, dir))
	}

	return result
}
