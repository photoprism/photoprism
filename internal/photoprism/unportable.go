package photoprism

import (
	"path/filepath"
	"strings"
	"sync"

	"github.com/photoprism/photoprism/pkg/clean"
)

// unportableLimit is the number of skipped paths that are remembered, so each is reported once per process.
const unportableLimit = 10000

// unportablePaths remembers the paths that were skipped because their names are not portable.
var unportablePaths = &warnedPaths{limit: unportableLimit}

// warnedPaths remembers paths up to a limit, dropping the oldest ones above it.
type warnedPaths struct {
	mu    sync.Mutex
	limit int
	seen  map[string]struct{}
	order []string
}

// First reports whether a path is not remembered yet, and remembers it.
func (w *warnedPaths) First(key string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.seen == nil {
		w.seen = make(map[string]struct{})
	} else if _, ok := w.seen[key]; ok {
		return false
	}

	w.seen[key] = struct{}{}
	w.order = append(w.order, key)

	for len(w.order) > w.limit {
		delete(w.seen, w.order[0])
		w.order = w.order[1:]
	}

	return true
}

// unportablePath returns a slash-separated relative path up to and including its first name that is not
// portable, or an empty string if all of its names are.
func unportablePath(rel string) string {
	for i := 0; i <= len(rel); {
		j := strings.IndexByte(rel[i:], '/')

		if j < 0 {
			j = len(rel) - i
		}

		if !clean.PortableName(rel[i : i+j]) {
			return rel[:i+j]
		}

		i += j + 1
	}

	return ""
}

// warnUnportable logs once per process that a path below a root was skipped because its name is not
// portable, using a prefix such as "index" and an action such as "index it".
func warnUnportable(prefix, root, rel, action string) {
	if unportablePaths.First(root + ":" + rel) {
		log.Warnf("%s: skipped %s because its name contains characters that are not supported, rename it to %s",
			prefix, clean.Log(rel), action)
	}
}

// PortableRelated leaves the files whose names are not portable out of a group of related files and reports each
// of them. If the main file is left out, mf becomes the main file if it is in the group, otherwise there is none.
func PortableRelated(related RelatedFiles, mf *MediaFile, skipped func(f *MediaFile)) RelatedFiles {
	files := make(MediaFiles, 0, len(related.Files))
	mfInGroup := false

	for _, f := range related.Files {
		if f != nil && !clean.PortableName(filepath.Base(f.FileName())) {
			skipped(f)
			continue
		} else if f != nil && mf != nil && f.FileName() == mf.FileName() {
			mfInGroup = true
		}

		files = append(files, f)
	}

	related.Files = files

	if related.Main != nil && !clean.PortableName(filepath.Base(related.Main.FileName())) {
		if mfInGroup {
			related.Main = mf
		} else {
			related.Main = nil
		}
	}

	return related
}
