package photoprism

import (
	"github.com/photoprism/photoprism/internal/meta"
)

// MetaData returns exif meta data of a media file.
func (m *MediaFile) MetaData() (result meta.Data, err error) {
	m.metaDataOnce.Do(func() { m.metaData, err = meta.Exif(m.FileName()) })
	return m.metaData, err
}
