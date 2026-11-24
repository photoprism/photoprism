package cluster

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// Theme archive bounds for node bootstrap and the theme pull command.
var (
	themeFileSizeLimit  int64 = 32 * fs.MB
	themeTotalSizeLimit int64 = 512 * fs.MB
)

// ExtractTheme extracts a theme archive into dir with fs.Unzip, skipping hidden and system names such as
// __MACOSX. Any other entry it cannot extract, including a name containing "..", fails the call, and the
// files it extracted are removed.
func ExtractTheme(zipName, dir string) error {
	ignored := make(map[string]bool)

	files, skipped, err := fs.Unzip(zipName, dir, themeFileSizeLimit, themeTotalSizeLimit, func(name string, _ bool) bool {
		if strings.Contains(name, "..") {
			return true
		} else if strings.HasPrefix(name, "__") || themeNameHidden(name) {
			ignored[name] = true
			return false
		}

		return true
	})

	if err == nil {
		for _, name := range skipped {
			if !ignored[name] {
				err = fmt.Errorf("theme archive entry %s cannot be extracted", clean.Log(name))
				break
			}
		}
	}

	if err == nil {
		return nil
	}

	// Removed in reverse order, so directories are empty when they are reached.
	for _, name := range slices.Backward(files) {
		_ = os.Remove(name)
	}

	return err
}

// themeNameHidden reports whether any component of an archive entry name is a hidden or system name, with a
// backslash read as a separator.
func themeNameHidden(name string) bool {
	for _, part := range strings.Split(strings.ReplaceAll(name, "\\", "/"), "/") {
		if part != "" && part != "." && fs.FileNameSystem(part) {
			return true
		}
	}

	return false
}
