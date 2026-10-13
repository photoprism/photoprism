package webdav

import (
	"path"
	"strings"

	"github.com/photoprism/photoprism/pkg/fs"
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
// ("..") segment. The raw path is inspected before normalization, since
// path.Clean collapses a rooted interior ".." into a dotless path.
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

// UnsafeSyncPath reports whether a logical transfer path contains a parent-directory segment.
func UnsafeSyncPath(name string) bool {
	return isUnsafePath(name)
}

// hasControl reports whether a path contains a C0 control character or DEL.
func hasControl(name string) bool {
	for i := 0; i < len(name); i++ {
		if b := name[i]; b < 0x20 || b == 0x7F {
			return true
		}
	}

	return false
}

// SkipSyncPath reports whether a logical transfer path is excluded from WebDAV sync, which includes
// paths with control characters. Unsafe paths remain subject to traversal validation instead of being
// skipped as benign.
func SkipSyncPath(name string) bool {
	if fs.HasReservedComponent(name) {
		return true
	}

	if isUnsafePath(name) {
		return false
	}

	return hasControl(name) || isHiddenPath(strings.ReplaceAll(name, "\\", "/"))
}
