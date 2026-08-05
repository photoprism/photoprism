package media

import (
	"sort"
	"strings"

	"github.com/photoprism/photoprism/pkg/fs"
)

// MainExtensions returns the sorted, lowercase file extensions of all main media formats,
// i.e. formats whose Type reports IsMain (originals that can be indexed and displayed on their
// own, unlike sidecar or archive files). Callers that match case-preserved filenames should
// permute the case themselves, as the returned set is canonical lowercase.
func MainExtensions() []string {
	exts := make([]string, 0, len(fs.Extensions))
	seen := make(map[string]struct{}, len(fs.Extensions))

	for ext, fileType := range fs.Extensions {
		if !Formats[fileType].IsMain() {
			continue
		}

		lower := strings.ToLower(ext)

		if _, ok := seen[lower]; ok {
			continue
		}

		seen[lower] = struct{}{}
		exts = append(exts, lower)
	}

	sort.Strings(exts)

	return exts
}
