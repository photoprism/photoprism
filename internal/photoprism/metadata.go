package photoprism

import (
	"path/filepath"

	"github.com/photoprism/photoprism/internal/meta"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

// MetaData returns exif meta data of a media file.
func (m *MediaFile) MetaData() (result meta.Data, err error) {
	m.metaDataOnce.Do(func() {
		if jsonFile := fs.TypeJson.Find(m.FileName(), false); jsonFile == "" {
			log.Debugf("mediafile: no json sidecar file found for %s", txt.Quote(filepath.Base(m.FileName())))
		} else if err := m.metaData.JSON(jsonFile); err != nil {
			log.Warn(err)
		}

		err = m.metaData.Exif(m.FileName())
	})

	return m.metaData, err
}
