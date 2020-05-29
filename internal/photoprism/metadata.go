package photoprism

import (
	"errors"
	"path/filepath"

	"github.com/photoprism/photoprism/internal/meta"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

// MetaData returns exif meta data of a media file.
func (m *MediaFile) MetaData() (result meta.Data) {
	m.metaDataOnce.Do(func() {
		var err error

		if m.IsPhoto() {
			err = m.metaData.Exif(m.FileName())
		} else {
			err = errors.New("not a photo")
		}

		if jsonFile := fs.TypeJson.FindSub(m.FileName(), fs.HiddenPath, false); jsonFile == "" {
			log.Debugf("mediafile: no json sidecar file found for %s", txt.Quote(filepath.Base(m.FileName())))
		} else if jsonErr := m.metaData.JSON(jsonFile); jsonErr != nil {
			log.Warn(jsonErr)
		} else {
			err = nil
		}

		if err != nil {
			m.metaData.Error = err
			log.Debugf("mediafile: %s", err.Error())
		}
	})

	return m.metaData
}
