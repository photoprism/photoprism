package photoprism

import (
	"fmt"
	"path/filepath"

	"github.com/photoprism/photoprism/internal/meta"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

// MetaData returns exif meta data of a media file.
func (m *MediaFile) MetaData() (result meta.Data) {
	m.metaDataOnce.Do(func() {
		var err error

		if m.ExifSupported() {
			err = m.metaData.Exif(m.FileName(), m.FileType())
		} else {
			err = fmt.Errorf("exif not supported: %s", txt.Quote(m.BaseName()))
		}

		// Parse regular JSON sidecar files ("img_1234.json").
		if jsonFiles := fs.FormatJson.FindAll(m.FileName(), []string{Config().SidecarPath(), fs.HiddenPath}, Config().OriginalsPath(), false); len(jsonFiles) == 0 {
			log.Debugf("media: no json sidecar file found for %s", txt.Quote(filepath.Base(m.FileName())))
		} else {
			for _, jsonFile := range jsonFiles {
				jsonErr := m.metaData.JSON(jsonFile, m.BaseName())

				if jsonErr != nil {
					log.Debug(jsonErr)
				} else {
					err = nil
				}
			}
		}

		if err != nil {
			m.metaData.Error = err
			log.Debugf("media: %s", err.Error())
		}
	})

	return m.metaData
}
