package photoprism

import (
	"fmt"
	"path/filepath"

	"github.com/photoprism/photoprism/internal/meta"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media/video"
)

// HasSidecarJson returns true if this file has or is a json sidecar file.
func (m *MediaFile) HasSidecarJson() bool {
	if m.IsJSON() {
		return true
	}

	return fs.SidecarJSON.FindFirst(m.FileName(), []string{Config().SidecarPath(), fs.PPHiddenPathname}, Config().OriginalsPath(), false) != ""
}

// SidecarJsonName returns the corresponding JSON sidecar file name as used by Google Photos (and potentially other apps).
func (m *MediaFile) SidecarJsonName() string {
	jsonName := m.fileName + ".json"

	if fs.FileExistsNotEmpty(jsonName) {
		return jsonName
	}

	return ""
}

// ExifToolJsonName returns the cached ExifTool metadata file name.
func (m *MediaFile) ExifToolJsonName() (string, error) {
	if Config().DisableExifTool() {
		return "", fmt.Errorf("media: exiftool json files disabled")
	}

	return ExifToolCacheName(m.Hash())
}

// NeedsExifToolJson tests if an ExifTool JSON file needs to be created.
func (m *MediaFile) NeedsExifToolJson() bool {
	if m.InSidecar() && m.IsImage() || !m.IsMedia() || m.Empty() {
		return false
	}

	jsonName, err := m.ExifToolJsonName()

	if err != nil {
		return false
	}

	return !fs.FileExists(jsonName)
}

// CreateExifToolJson extracts metadata to a JSON file using Exiftool.
func (m *MediaFile) CreateExifToolJson(convert *Convert) error {
	if !m.NeedsExifToolJson() {
		return nil
	} else if jsonName, err := convert.ToJson(m, false); err != nil {
		log.Tracef("exiftool: %s", clean.Error(err))
		log.Debugf("exiftool: failed parsing %s", clean.Log(m.RootRelName()))
	} else if err = m.metaData.JSON(jsonName, ""); err != nil {
		return fmt.Errorf("%s in %s (read json sidecar)", clean.Error(err), clean.Log(m.BaseName()))
	}

	return nil
}

// ReadExifToolJson reads metadata from a cached ExifTool JSON file.
func (m *MediaFile) ReadExifToolJson() error {
	jsonName, err := m.ExifToolJsonName()

	if err != nil {
		return err
	}

	return m.metaData.JSON(jsonName, "")
}

// MetaData returns exif meta data of a media file.
func (m *MediaFile) MetaData() (result meta.Data) {
	if !m.Ok() || !m.IsMedia() {
		// Not a main media file.
		return m.metaData
	}

	// Gather the data once and cache it.
	m.metaOnce.Do(func() {
		var err error

		if m.ExifSupported() {
			err = m.metaData.Exif(m.FileName(), m.FileType(), Config().ExifBruteForce())
		} else {
			err = fmt.Errorf("exif not supported")
		}

		// Parse regular JSON sidecar files ("img_1234.json")
		if !m.IsSidecar() {
			if jsonFiles := fs.SidecarJSON.FindAll(m.FileName(), []string{Config().SidecarPath(), fs.PPHiddenPathname}, Config().OriginalsPath(), false); len(jsonFiles) == 0 {
				log.Tracef("metadata: found no additional sidecar file for %s", clean.Log(filepath.Base(m.FileName())))
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

			if jsonErr := m.ReadExifToolJson(); jsonErr != nil {
				log.Debug(jsonErr)
			} else {
				err = nil
			}
		}

		if err != nil {
			m.metaData.Error = err
			log.Debugf("%s in %s", err, clean.Log(m.BaseName()))
		}
	})

	return m.metaData
}

// VideoInfo returns video information if this is a video file or has a video embedded.
func (m *MediaFile) VideoInfo() video.Info {
	if !m.Ok() || !m.IsMedia() {
		// Not a main media file.
		return m.videoInfo
	}

	// Gather the data once and cache it.
	m.videoOnce.Do(func() {
		if info, err := video.ProbeFile(m.FileName()); err != nil {
			log.Debugf("video: %s in %s", err, clean.Log(m.BaseName()))
		} else {
			m.videoInfo = info
		}
	})

	return m.videoInfo
}
