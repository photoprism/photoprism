package photoprism

import (
	"fmt"
	"path"
	"sync"
	"time"

	"github.com/photoprism/photoprism/internal/entity/query"
)

// Files caches indexed originals keyed by root/name with their last-modified timestamps. The
// cache is only considered hydrated after Init() loads the database snapshot; partial writes
// before Init() must not trick callers into skipping a full reload.
type Files struct {
	count  int
	files  query.FileMap
	loaded bool
	mutex  sync.RWMutex
}

// NewFiles returns a new Files instance.
func NewFiles() *Files {
	m := &Files{
		files: make(query.FileMap),
	}

	return m
}

// Init loads the indexed file snapshot from the database on first use. If the cache only contains
// ad-hoc entries (for example, uploaded files recorded before Init was called), this forces a
// refresh so rescan=false jobs can safely skip unchanged files.
func (m *Files) Init() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.loaded {
		m.count = len(m.files)
		return nil
	}

	if err := query.PurgeOrphanDuplicates(); err != nil {
		return fmt.Errorf("%s (purge duplicates)", err.Error())
	}

	files, err := query.IndexedFiles()

	if err != nil {
		return fmt.Errorf("%s (find indexed files)", err.Error())
	} else {
		m.files = files
		m.count = len(files)
		m.loaded = true
		return nil
	}
}

// Done clears the cache and marks it stale so the next Init() call pulls a fresh database snapshot.
func (m *Files) Done() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.count = 0
	m.files = make(query.FileMap)
	m.loaded = false
}

// Remove evicts a file entry from the cache (e.g. after deletion or re-import).
func (m *Files) Remove(fileName, fileRoot string) {
	key := path.Join(fileRoot, fileName)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.files, key)
}

// Ignore determines whether a file should be skipped based on modification timestamp, updating the cache when needed.
func (m *Files) Ignore(fileName, fileRoot string, modTime time.Time, rescan bool) bool {
	timestamp := modTime.UTC().Truncate(time.Second).Unix()
	key := path.Join(fileRoot, fileName)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if rescan {
		m.files[key] = timestamp
		return false
	}

	mod, ok := m.files[key]

	if ok && mod == timestamp {
		return true
	} else {
		m.files[key] = timestamp
		return false
	}
}

// Indexed checks if a file has already been indexed without mutating the cache.
func (m *Files) Indexed(fileName, fileRoot string, modTime time.Time, rescan bool) bool {
	if rescan {
		return false
	}

	timestamp := modTime.Unix()
	key := path.Join(fileRoot, fileName)

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	mod, ok := m.files[key]

	if ok && mod == timestamp {
		return true
	} else {
		return false
	}
}

// Exists reports whether the given file key is present in the cache.
func (m *Files) Exists(fileName, fileRoot string) bool {
	key := path.Join(fileRoot, fileName)

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if _, ok := m.files[key]; ok {
		return true
	} else {
		return false
	}
}
