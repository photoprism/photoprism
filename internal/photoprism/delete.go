package photoprism

import (
	"os"
	"path/filepath"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// DeletePhoto removes a photo from the index and optionally all related media files.
func DeletePhoto(p entity.Photo, mediaFiles bool, originals bool) error {
	yamlFileName := p.YamlFileName(Config().OriginalsPath(), Config().SidecarPath())

	// Permanently remove photo from index.
	files, err := p.DeletePermanently()

	if err != nil {
		return err
	}

	if mediaFiles {
		DeleteFiles(files, originals)
	}

	// Remove sidecar backup.
	if fs.FileExists(yamlFileName) {
		log.Debugf("media: removing yaml sidecar %s", clean.Log(filepath.Base(yamlFileName)))
		logWarn("media", os.Remove(yamlFileName))
	}

	return nil
}

// DeleteFiles permanently deletes media and related sidecar files.
func DeleteFiles(files entity.Files, originals bool) {
	for _, file := range files {
		fileName := FileName(file.FileRoot, file.FileName)

		// Skip empty file names, just to be sure.
		if fileName == "" {
			continue
		}

		// Open media file.
		f, err := NewMediaFile(fileName)

		// Log media file error if any.
		if err != nil {
			log.Debugf("media: %s not found", clean.Log(file.FileName))
		}

		// Remove sidecar JSON files.
		if sidecarJson := f.SidecarJsonName(); fs.FileExists(sidecarJson) {
			log.Debugf("media: removing json sidecar %s", clean.Log(filepath.Base(sidecarJson)))
			logWarn("delete", os.Remove(sidecarJson))
		}
		if exifJson, err := f.ExifToolJsonName(); err == nil && fs.FileExists(exifJson) {
			log.Debugf("media: removing exiftool sidecar %s", clean.Log(filepath.Base(exifJson)))
			logWarn("media", os.Remove(exifJson))
		}

		// Remove any other sidecar files.
		logWarn("media", f.RemoveSidecars())

		// Continue if the media file does not exist or should be preserved.
		if !fs.FileExists(fileName) {
			continue
		} else if !originals && f.Root() == entity.RootOriginals {
			log.Debugf("media: skipped original %s", clean.Log(file.FileName))
			continue
		}

		log.Debugf("media: removing %s", clean.Log(file.FileName))

		// Remove media file.
		if err = f.Remove(); err != nil {
			log.Errorf("media: removed %s", clean.Log(file.FileName))
		} else {
			log.Infof("media: failed removing %s", clean.Log(file.FileName))
		}
	}
}
