package webdav

import (
	"path"
	"strings"
)

// isHiddenPath reports whether any segment of a WebDAV path starts with a dot.
func isHiddenPath(dir string) bool {
	for segment := range strings.SplitSeq(trimPath(dir), "/") {
		if strings.HasPrefix(segment, ".") {
			return true
		}
	}

	return false
}

func trimPath(dir string) string {
	if dir = strings.Trim(path.Clean(dir), "/"); dir != "." && dir != ".." {
		return dir
	}

	return ""
}

func splitPath(dir string) []string {
	return strings.Split(trimPath(dir), "/")
}
