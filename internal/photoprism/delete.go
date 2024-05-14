package photoprism

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// DeletePhoto removes a photo from the index and optionally all related media files.
func DeletePhoto(p *entity.Photo, mediaFiles bool, originals bool) (numFiles int, err error) {
	if p == nil {
		return 0, errors.New("photo is nil")
	}

	yamlFileName, yamlRelName, err := p.YamlFileName(Config().OriginalsPath(), Config().SidecarPath())

	if err != nil {
		log.Warnf("photo: %s (delete %s)", err, clean.Log(yamlRelName))
	}

	// Permanently remove photo from index.
	files, err := p.DeletePermanently()

	if err != nil {
		return 0, err
	}

	if mediaFiles {
		numFiles = DeleteFiles(files, originals)
	}

	// Remove sidecar backup.
	if !fs.FileExists(yamlFileName) {
		return numFiles, nil
	} else if err := os.Remove(yamlFileName); err != nil {
		log.Warnf("photo: failed deleting sidecar file %s", clean.Log(yamlRelName))
	} else {
		numFiles++
		log.Infof("photo: deleted sidecar file %s", clean.Log(yamlRelName))
	}

	return numFiles, nil
}

// DeleteFiles permanently deletes media and related sidecar files.
func DeleteFiles(files entity.Files, originals bool) (numFiles int) {
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
			log.Tracef("files: %s", err)
		}

		// Remove original JSON sidecar file, if any.
		if jsonFile := f.FileName() + ".json"; !originals && f.Root() == entity.RootOriginals || !fs.FileExists(jsonFile) {
			// Do nothing.
		} else if err = os.Remove(jsonFile); err != nil {
			log.Warnf("files: failed deleting sidecar %s", clean.Log(filepath.Base(jsonFile)))
		} else {
			numFiles++
			log.Infof("files: deleted sidecar %s", clean.Log(filepath.Base(jsonFile)))
		}

		// Remove Exiftool JSON file in cache folder.
		if exifJson, _ := ExifToolCacheName(file.FileHash); !fs.FileExists(exifJson) {
			// Do nothing.
		} else if err = os.Remove(exifJson); err != nil {
			log.Warnf("files: failed deleting sidecar %s", clean.Log(filepath.Base(exifJson)))
		} else {
			numFiles++
			log.Infof("files: deleted sidecar %s", clean.Log(filepath.Base(exifJson)))
		}

		// Remove any other files in the sidecar folder.
		if n, _ := f.RemoveSidecarFiles(); n > 0 {
			numFiles += n
		}

		// Continue if the media file does not exist or should be preserved.
		if !fs.FileExists(fileName) {
			continue
		}

		// Remove the original media file, if it exists and is allowed.
		if relName := f.RootRelName(); relName == "" {
			log.Warnf("files: relative filename of %s must not be empty - bug?", clean.Log(fileName))
			continue
		} else if !originals && f.Root() == entity.RootOriginals {
			log.Debugf("files: skipped deleting %s", clean.Log(relName))
			continue
		} else if err = f.Remove(); err != nil {
			log.Errorf("files: failed deleting %s", clean.Log(relName))
		} else {
			numFiles++
			log.Infof("files: deleted %s", clean.Log(relName))
		}
	}

	return numFiles
}
