package photoprism

import (
	"os"
	"path/filepath"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/sanitize"
)

// Delete permanently removes a photo and all its files.
func Delete(p entity.Photo) error {
	yamlFileName := p.YamlFileName(Config().OriginalsPath(), Config().SidecarPath())

	// Permanently remove photo from index.
	files, err := p.DeletePermanently()

	if err != nil {
		return err
	}

	// Delete related files.
	for _, file := range files {
		fileName := FileName(file.FileRoot, file.FileName)

		log.Debugf("delete: removing file %s", sanitize.Log(file.FileName))

		if f, err := NewMediaFile(fileName); err == nil {
			if sidecarJson := f.SidecarJsonName(); fs.FileExists(sidecarJson) {
				log.Debugf("delete: removing json sidecar %s", sanitize.Log(filepath.Base(sidecarJson)))
				logWarn("delete", os.Remove(sidecarJson))
			}

			if exifJson, err := f.ExifToolJsonName(); err == nil && fs.FileExists(exifJson) {
				log.Debugf("delete: removing exiftool sidecar %s", sanitize.Log(filepath.Base(exifJson)))
				logWarn("delete", os.Remove(exifJson))
			}

			logWarn("delete", f.RemoveSidecars())
		}

		if fs.FileExists(fileName) {
			logWarn("delete", os.Remove(fileName))
		}
	}

	// Remove sidecar backup.
	if fs.FileExists(yamlFileName) {
		log.Debugf("delete: removing yaml sidecar %s", sanitize.Log(filepath.Base(yamlFileName)))
		logWarn("delete", os.Remove(yamlFileName))
	}

	return nil
}
