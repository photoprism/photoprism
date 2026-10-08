package photoprism

import (
	"path/filepath"
	"strings"
)

// magickBaseChars lists the characters ImageMagick may interpret in a base name as a pattern; "%" is
// interpreted in folder names as well.
const magickBaseChars = "*?[]{}"

// magickNames reports whether the files can be passed to ImageMagick by name, i.e. whether their paths
// contain no "%" and their base names none of magickBaseChars. ImageMagick is a fallback converter, so
// other files are skipped.
func magickNames(fileNames ...string) bool {
	for _, fileName := range fileNames {
		if strings.Contains(fileName, "%") || strings.ContainsAny(filepath.Base(fileName), magickBaseChars) {
			return false
		}
	}

	return true
}
