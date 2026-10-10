package entity

import (
	"path/filepath"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// OriginalsPath is the absolute originals folder in which default base paths of new accounts are checked.
var OriginalsPath = ""

// WarnExistingBasePath reports a new contributor account whose default base path already exists in the
// originals folder, and returns true if it does.
func (m *User) WarnExistingBasePath() bool {
	if m == nil || OriginalsPath == "" || !m.RequiresBasePath() {
		return false
	}

	basePath := m.DefaultBasePath()

	if basePath == "" || m.BasePath != "" && m.BasePath != basePath {
		return false
	} else if !fs.PathExists(filepath.Join(OriginalsPath, basePath)) {
		return false
	}

	event.AuditWarn([]string{"user %s", "default base path %s", "already exists"}, clean.LogQuote(m.Username()), clean.Log(basePath))

	return true
}
