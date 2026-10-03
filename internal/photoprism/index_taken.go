package photoprism

import (
	"time"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/meta"
)

// setTakenAtMeta sets the time a photo was taken from file metadata, or else the time the file was modified, which
// ranks below capture times and file names, so that a capture time of another file of the photo wins in any order.
func setTakenAtMeta(photo *entity.Photo, data meta.Data) {
	if utc, local, zone, modified := data.TakenOrModified(); modified {
		photo.SetTakenAt(utc, local, zone, entity.SrcModified)
	} else {
		photo.SetTakenAt(utc, local, zone, entity.SrcMeta)
	}
}

// mediaTimeUTC returns the time a media file was taken, or else modified, as stored for the file.
func mediaTimeUTC(data meta.Data) time.Time {
	utc, _, _, _ := data.TakenOrModified()
	return utc
}
