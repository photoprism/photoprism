package photoprism

import (
	"path/filepath"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
)

// publishConverting publishes an index.converting event for a file in one of the library folders.
// Cached media, such as videos extracted from an original, are named after the file hash and not reported.
func publishConverting(f *MediaFile, fileName, xmpName string) {
	if f == nil || f.Root() == entity.RootUnknown {
		return
	}

	event.Publish("index.converting", event.Data{
		"fileType": f.FileType(),
		"fileName": fileName,
		"baseName": filepath.Base(fileName),
		"xmpName":  xmpName,
	})
}
