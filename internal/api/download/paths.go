package download

import (
	"path/filepath"
	"strings"

	"github.com/photoprism/photoprism/pkg/fs"
)

// AllowedPaths lists absolute directories from which downloads may be registered.
var AllowedPaths []string

// Deny checks if the filename may not be registered for download.
func Deny(fileName string) bool {
	if len(AllowedPaths) == 0 || fileName == "" {
		return true
	} else if fileName = fs.Abs(fileName); strings.HasPrefix(fileName, "/etc") ||
		strings.HasPrefix(filepath.Base(fileName), ".") {
		return true
	}

	for _, dir := range AllowedPaths {
		if dir != "" && strings.HasPrefix(fileName, dir+"/") {
			return false
		}
	}

	return true
}
