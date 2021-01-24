package photoprism

import (
	"os"
	"path/filepath"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Delete permanently removes a photo and all its files.
func Delete(p entity.Photo) error {
	yamlFileName := p.YamlFileName(Config().OriginalsPath(), Config().SidecarPath())

	files := p.AllFiles()

	for _, file := range files {
		fileName := FileName(file.FileRoot, file.FileName)

		log.Debugf("delete: removing file %s", txt.Quote(file.FileName))

		if f, err := NewMediaFile(fileName); err == nil {
			if sidecarJson := f.SidecarJsonName(); fs.FileExists(sidecarJson) {
				log.Debugf("delete: removing json sidecar %s", txt.Quote(filepath.Base(sidecarJson)))
				logWarn("delete", os.Remove(sidecarJson))
			}

			if exifJson, err := f.ExifToolJsonName(); err == nil && fs.FileExists(exifJson) {
				log.Debugf("delete: removing exiftool sidecar %s", txt.Quote(filepath.Base(exifJson)))
				logWarn("delete", os.Remove(exifJson))
			}
		}

		if fs.FileExists(fileName) {
			logWarn("delete", os.Remove(fileName))
		}
	}

	if fs.FileExists(yamlFileName) {
		log.Debugf("delete: removing yaml sidecar %s", txt.Quote(filepath.Base(yamlFileName)))
		logWarn("delete", os.Remove(yamlFileName))
	}

	return p.DeletePermanently()
}
