package photoprism

import (
	"fmt"
	"path"
	"sync"
	"time"

	"github.com/photoprism/photoprism/internal/entity/query"
)

// Files represents a list of already indexed file names and their unix modification timestamps.
type Files struct {
	count int
	files query.FileMap
	mutex sync.RWMutex
}

// NewFiles returns a new Files instance.
func NewFiles() *Files {
	m := &Files{
		files: make(query.FileMap),
	}

	return m
}

// Init fetches the list from the database once.
func (m *Files) Init() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if len(m.files) > 0 {
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
		return nil
	}
}

// Done should be called after all files have been processed.
func (m *Files) Done() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if (len(m.files) - m.count) == 0 {
		return
	}

	m.count = 0
	m.files = make(query.FileMap)
}

// Remove a file from the lookup table.
func (m *Files) Remove(fileName, fileRoot string) {
	key := path.Join(fileRoot, fileName)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.files, key)
}

// Ignore tests of a file requires indexing, file name must be relative to the originals path.
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

// Indexed tests of a file was already indexed without modifying the files map.
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

// Exists tests of a file exists.
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
