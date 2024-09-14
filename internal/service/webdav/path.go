package webdav

import (
	"path"
	"strings"
)

func trimPath(dir string) string {
	if dir = strings.Trim(path.Clean(dir), "/"); dir != "." && dir != ".." {
		return dir
	}

	return ""
}

func splitPath(dir string) []string {
	return strings.Split(trimPath(dir), "/")
}
