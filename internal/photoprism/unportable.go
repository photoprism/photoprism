package photoprism

import (
	"path/filepath"
	"strings"
	"sync"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
)

// unportableLimit is the number of skipped paths that are remembered, so each is reported once per process.
const unportableLimit = 10000

// unportablePaths remembers the paths that were skipped because their names are not portable.
var unportablePaths = &warnedPaths{limit: unportableLimit}

// warnedPaths remembers paths up to a limit, and counts other paths by prefix until they are flushed.
type warnedPaths struct {
	mu    sync.Mutex
	limit int
	seen  map[string]struct{}
	more  map[string]map[string]struct{}
	extra map[string]int
}

// First reports whether a path is not remembered yet, and remembers it if there is room. A path that does not fit
// is counted for its prefix instead, once as long as the counted paths stay within the limit.
func (w *warnedPaths) First(prefix, key string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.seen == nil {
		w.seen = make(map[string]struct{})
	} else if _, ok := w.seen[key]; ok {
		return false
	}

	if len(w.seen) < w.limit {
		w.seen[key] = struct{}{}
		return true
	}

	if w.more == nil {
		w.more = make(map[string]map[string]struct{})
	}

	if w.more[prefix] == nil {
		w.more[prefix] = make(map[string]struct{})
	}

	if _, ok := w.more[prefix][key]; ok {
		return false
	} else if len(w.more[prefix]) < w.limit {
		w.more[prefix][key] = struct{}{}
	} else {
		if w.extra == nil {
			w.extra = make(map[string]int)
		}

		w.extra[prefix]++
	}

	return false
}

// Flush returns the number of paths counted for a prefix since the last flush, and resets it.
func (w *warnedPaths) Flush(prefix string) int {
	w.mu.Lock()
	defer w.mu.Unlock()

	n := len(w.more[prefix]) + w.extra[prefix]
	delete(w.more, prefix)
	delete(w.extra, prefix)

	return n
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
	if unportablePaths.First(prefix, root+":"+rel) {
		log.Warnf("%s: skipped %s because its name contains characters that are not supported, rename it to %s",
			prefix, clean.Log(rel), action)
	}
}

// flushUnportable logs one warning with the number of paths a run skipped without reporting them, because the
// remembered paths reached their limit.
func flushUnportable(prefix string) {
	if n := unportablePaths.Flush(prefix); n > 0 {
		log.Warnf("%s: skipped %s that contain characters that are not supported", prefix,
			english.Plural(n, "more name", "more names"))
	}
}

// PortableRelated leaves the files whose names are not portable out of a group of related files and reports each
// of them. If the main file is left out, mf becomes the main file if it is in the group, otherwise there is none.
func PortableRelated(related RelatedFiles, mf *MediaFile, skipped func(f *MediaFile)) RelatedFiles {
	return portableGroup(related, mf, false, skipped)
}

// PortableSiblings is like PortableRelated, but keeps mf whatever its name, for actions on a file that is
// already indexed. Other files in originals must also be in folders whose names are portable.
func PortableSiblings(related RelatedFiles, mf *MediaFile, skipped func(f *MediaFile)) RelatedFiles {
	return portableGroup(related, mf, true, skipped)
}

// portableGroup leaves the files whose names are not portable out of a group, except mf if keepMf is set.
func portableGroup(related RelatedFiles, mf *MediaFile, keepMf bool, skipped func(f *MediaFile)) RelatedFiles {
	files := make(MediaFiles, 0, len(related.Files))
	mfInGroup := false

	// kept reports whether a file stays in the group.
	kept := func(f *MediaFile) bool {
		switch {
		case keepMf && mf != nil && f.FileName() == mf.FileName():
			return true
		case keepMf && f.Root() == entity.RootOriginals:
			return unportablePath(filepath.ToSlash(f.RootRelName())) == ""
		default:
			return clean.PortableName(filepath.Base(f.FileName()))
		}
	}

	for _, f := range related.Files {
		if f != nil && !kept(f) {
			skipped(f)
			continue
		} else if f != nil && mf != nil && f.FileName() == mf.FileName() {
			mfInGroup = true
		}

		files = append(files, f)
	}

	related.Files = files

	if related.Main != nil && !kept(related.Main) {
		if mfInGroup {
			related.Main = mf
		} else {
			related.Main = nil
		}
	}

	return related
}
