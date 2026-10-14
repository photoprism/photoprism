package clean

import (
	"path/filepath"
	"strings"

	"github.com/photoprism/photoprism/pkg/txt"
)

// FileNameRedacted returns a file name with its last dash-separated segment replaced by a placeholder,
// keeping the extension so entries stay recognizable. Use it for generated, single-use names.
func FileNameRedacted(s string) string {
	ext := filepath.Ext(s)
	name := strings.TrimSuffix(s, ext)

	if i := strings.LastIndex(name, "-"); i > 0 {
		return name[:i+1] + txt.Masked + ext
	}

	return txt.Masked + ext
}

// FileName removes reserved and unsafe characters from a file name and trims ASCII spaces, keeping the
// other characters as they are. It returns an empty string for a path, "." or a name containing "..".
func FileName(s string) string {
	if s == "" || reject(s, 512) || strings.Contains(s, "/") || strings.Contains(s, "..") {
		return ""
	}

	s = strings.Trim(pathString(s, `~/\:|"?*<>{}`), " ")

	if s == "." || reject(s, 0) || strings.Contains(s, "..") {
		return ""
	}

	return s
}
