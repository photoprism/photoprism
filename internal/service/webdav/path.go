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

// isUnsafePath reports whether a remote WebDAV path contains a parent-directory
// ("..") segment. The raw path is inspected before normalization because
// path.Clean collapses a rooted interior ".." into a dotless path that would
// otherwise pass the hidden-path filter and later escape the local sync base
// directory once composed with a destination.
func isUnsafePath(dir string) bool {
	for segment := range strings.SplitSeq(strings.Trim(strings.ReplaceAll(dir, "\\", "/"), "/"), "/") {
		if segment == ".." {
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
