package download

import (
	"errors"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/log/status"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Register generated an event to make the specified file available
// for download until the cache expires, or the server is restarted.
func Register(fileUuid, fileName string) error {
	if !rnd.IsUUID(fileUuid) {
		event.AuditWarn([]string{"api", "download", "create temporary token for %s", status.Failed}, fileName)
		return errors.New("invalid file uuid")
	}

	if fileName = fs.Abs(fileName); !fs.FileExists(fileName) {
		event.AuditWarn([]string{"api", "download", "create temporary token for %s", status.Failed}, fileName)
		return errors.New("file not found")
	} else if Deny(fileName) {
		event.AuditErr([]string{"api", "download", "create temporary token for %s", status.Denied}, fileName)
		return errors.New("forbidden file path")
	}

	event.AuditInfo([]string{"api", "download", "create temporary token for %s", status.Succeeded}, fileName)

	cache.SetDefault(fileUuid, fileName)

	return nil
}
