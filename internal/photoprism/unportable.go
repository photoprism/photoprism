package photoprism

import (
	"path/filepath"
	"sync"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// unsupportedLimit is the number of skipped paths that are remembered, so each is reported once per process.
const unsupportedLimit = 10000

// unsupportedPaths remembers the paths that were skipped because their names are not portable or hidden.
var unsupportedPaths = &warnedPaths{limit: unsupportedLimit}

// warnedPaths remembers paths up to a limit, and counts other paths by prefix until they are flushed.
type warnedPaths struct {
	mu    sync.Mutex
	limit int
	seen  map[string]struct{}
	more  map[string]map[string]struct{}
	extra map[string]int
}

// First reports whether a path is not remembered yet for a prefix, and remembers it if there is room. A path that
// does not fit is counted for its prefix instead, once as long as the counted paths stay within the limit.
func (w *warnedPaths) First(prefix, key string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	id := prefix + "\x00" + key

	if w.seen == nil {
		w.seen = make(map[string]struct{})
	} else if _, ok := w.seen[id]; ok {
		return false
	}

	if len(w.seen) < w.limit {
		w.seen[id] = struct{}{}
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

// skipUnportableDir returns the part of a folder's relative path up to its first name that is not portable, and then
// removes the folder, and the folder a symbolic link points to, from the walked folders, so that a folder reached
// under another name is still walked under that name. It returns an empty string if all names are portable.
func skipUnportableDir(done fs.Done, fileName, rel string, isSymlink bool) string {
	p := clean.UnportablePath(rel)

	if p == "" {
		return ""
	}

	delete(done, fileName)

	if isSymlink {
		if resolved, err := filepath.EvalSymlinks(fileName); err == nil {
			delete(done, resolved)
		}
	}

	return p
}

// warnUnportable logs once per process that a path below a root was skipped because its name is not
// portable, using a prefix such as "index" and an action such as "index it".
func warnUnportable(prefix, root, rel, action string) {
	if unsupportedPaths.First(prefix, root+":"+rel) {
		log.Warnf("%s: skipped %s because of unsupported characters, rename it to %s",
			prefix, clean.Log(rel), action)
	}
}

// flushUnportable logs one warning with the number of paths a run skipped without reporting them, because the
// remembered paths reached their limit.
func flushUnportable(prefix string) {
	if n := unsupportedPaths.Flush(prefix); n > 0 {
		log.Warnf("%s: skipped %s with unsupported characters", prefix,
			english.Plural(n, "more name", "more names"))
	}
}

// warnHidden logs once per process that a hidden file below a root was skipped and left in place, using a prefix
// such as "import" and an action such as "import it".
func warnHidden(prefix, root, rel, action string) {
	if unsupportedPaths.First(prefix+" hidden", root+":"+rel) {
		log.Warnf("%s: skipped hidden file %s, rename it to %s", prefix, clean.Log(rel), action)
	}
}

// flushHidden logs one warning with the number of hidden files a run kept without reporting them, because the
// remembered paths reached their limit.
func flushHidden(prefix string) {
	if n := unsupportedPaths.Flush(prefix + " hidden"); n > 0 {
		log.Warnf("%s: skipped %s", prefix, english.Plural(n, "more hidden file", "more hidden files"))
	}
}

// PortableRelated leaves the files whose names are not portable out of a group of related files and reports each
// of them. If the main file is left out, mf becomes the main file if it is in the group, otherwise there is none.
func PortableRelated(related RelatedFiles, mf *MediaFile, skipped func(f *MediaFile)) RelatedFiles {
	return portableGroup(related, mf, false, skipped)
}

// PortableSiblings is like PortableRelated, but keeps mf whatever its name, for actions on a file that is
// already indexed. Other files in originals must also be in folders whose names are portable, except a main file
// that is indexed, and mf does not become the main file if it is in the sidecar folder.
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
			return clean.UnportablePath(filepath.ToSlash(f.RootRelName())) == ""
		default:
			return clean.PortableName(filepath.Base(f.FileName()))
		}
	}

	main := related.Main
	keepMain := keepMf && main != nil && !kept(main) && fileIndexed(main)

	// stays reports whether a file stays in the group, including a main file that is kept because it is indexed.
	stays := func(f *MediaFile) bool {
		return kept(f) || keepMain && f.FileName() == main.FileName()
	}

	for _, f := range related.Files {
		if f != nil && !stays(f) {
			skipped(f)
			continue
		} else if f != nil && mf != nil && f.FileName() == mf.FileName() {
			mfInGroup = true
		}

		files = append(files, f)
	}

	related.Files = files

	if main != nil && !stays(main) {
		if mfInGroup && (!keepMf || !fs.InDir(mf.FileName(), Config().SidecarPath())) {
			related.Main = mf
		} else {
			related.Main = nil
		}
	}

	return related
}

// fileIndexed reports whether a file has an index entry that is not deleted, matching it as the index does.
func fileIndexed(f *MediaFile) bool {
	var count int64

	if err := entity.Db().Model(&entity.File{}).Where("file_name = ? AND (file_root = ? OR file_root = '') AND deleted_at IS NULL",
		f.RootRelName(), f.Root()).Count(&count).Error; err != nil {
		log.Debugf("index: %s (find %s)", clean.Error(err), clean.Log(f.RootRelName()))
		return false
	}

	return count > 0
}
