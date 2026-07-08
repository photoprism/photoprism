package server

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/http/proxy"
)

// IsWebDAVPath returns true when pathValue points to a WebDAV collection path.
//
// Optional basePaths are checked first and may include prefixed variants such as
// "/instance-a/originals" and "/instance-a/import".
func IsWebDAVPath(pathValue string, basePaths ...string) bool {
	if pathValue == "" {
		return false
	}

	for _, basePath := range basePaths {
		if hasCollectionPath(pathValue, basePath) {
			return true
		}
	}

	if hasCollectionPath(pathValue, "/originals") || hasCollectionPath(pathValue, "/import") {
		return true
	}

	requestParts := splitSlashPath(pathValue)
	prefixParts := splitSlashPath(proxy.PathPrefix)

	// WebDAV via path proxy: /<prefix>/<instance>/(originals|import)/...
	if len(prefixParts) == 0 || len(requestParts) < len(prefixParts)+2 {
		return false
	}

	for i, part := range prefixParts {
		if requestParts[i] != part {
			return false
		}
	}

	return isWebDAVCollection(requestParts[len(prefixParts)+1])
}

// hasCollectionPath reports whether pathValue equals basePath or any nested child.
func hasCollectionPath(pathValue, basePath string) bool {
	basePath = strings.TrimRight(basePath, "/")
	if basePath == "" {
		return false
	}

	return pathValue == basePath || strings.HasPrefix(pathValue, basePath+"/")
}

// isWebDAVCollection reports whether segment is a known WebDAV collection name.
func isWebDAVCollection(segment string) bool {
	return segment == "originals" || segment == "import"
}

// splitSlashPath returns a list of non-empty slash-separated path segments.
func splitSlashPath(pathValue string) []string {
	pathValue = strings.Trim(pathValue, "/")
	if pathValue == "" {
		return nil
	}

	return strings.Split(pathValue, "/")
}
