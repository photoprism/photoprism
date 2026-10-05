package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileName returns the file path for a sidecar file with the specified extension and creates its folder.
func FileName(fileName, dirName, baseDir, fileExt string) (string, error) {
	result, err := FilePath(fileName, dirName, baseDir, fileExt)

	if err != nil {
		return "", err
	}

	// Create parent directories if they do not exist yet.
	if err = MkdirAll(filepath.Dir(result)); err != nil {
		return "", err
	}

	return result, nil
}

// FilePath returns the file path for a sidecar file with the specified extension, like FileName, but
// without creating its folder, e.g. to plan an output before any file is written. A file that already
// lies in an absolute dirName outside baseDir gets its sidecar next to it.
func FilePath(fileName, dirName, baseDir, fileExt string) (string, error) {
	if fileName == "" {
		return "", fmt.Errorf("file name is empty")
	} else if fileExt == "" {
		return "", fmt.Errorf("file extension is empty")
	}

	dir := filepath.Dir(fileName)

	switch {
	case dirName == "" || dirName == "." || dir == dirName:
		dirName = dir
	case filepath.IsAbs(dirName) && InDir(dir, dirName) && !InDir(dir, baseDir):
		// A file in the folder itself is written next to it, where Type.FindEach looks it up.
		dirName = dir
	case filepath.IsAbs(dirName):
		dirName = filepath.Join(dirName, RelName(dir, baseDir))
	default:
		dirName = filepath.Join(dir, dirName)
	}

	// Compose and return file path.
	return filepath.Join(dirName, filepath.Base(fileName)) + fileExt, nil
}

// InDir reports whether fileName is dir or a path below it. Names are compared as given, without
// resolving them, and a directory only contains names that continue after a path separator.
func InDir(fileName, dir string) bool {
	switch {
	case fileName == "" || dir == "":
		return false
	case fileName == dir:
		return true
	case strings.HasSuffix(dir, string(os.PathSeparator)):
		return strings.HasPrefix(fileName, dir)
	default:
		return strings.HasPrefix(fileName, dir+string(os.PathSeparator))
	}
}

// RelName returns the file name relative to a directory, or the file name itself if it is not below it.
func RelName(fileName, dir string) string {
	switch {
	case fileName == dir:
		return ""
	case !InDir(fileName, dir):
		return fileName
	case strings.HasSuffix(dir, string(os.PathSeparator)):
		return fileName[len(dir):]
	default:
		return fileName[len(dir)+1:]
	}
}

// FileNameHidden tests is a file name belongs to a hidden file.
func FileNameHidden(name string) bool {
	if len(name) == 0 {
		return false
	}

	name = filepath.Base(name)

	// Hidden files and folders starting with "." or "@" should be ignored.
	switch name[0:1] {
	case ".", "@":
		return true
	}

	if len(name) == 1 {
		return false
	}

	// File paths starting with _. and __ like __MACOSX should be ignored.
	switch name[0:2] {
	case "_.", "__":
		return true
	}

	return false
}
