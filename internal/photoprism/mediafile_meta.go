package photoprism

import (
	"fmt"
	"path/filepath"

	"github.com/photoprism/photoprism/internal/meta"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media/video"
	"github.com/photoprism/photoprism/pkg/time/tz"
)

// HasSidecarJson reports whether the media file already has a JSON sidecar in
// any of the configured lookup paths (or is itself a JSON sidecar).
func (m *MediaFile) HasSidecarJson() bool {
	if m.IsJSON() {
		return true
	}

	return fs.SidecarJson.FindFirst(m.FileName(), []string{Config().SidecarPath(), fs.PPHiddenPathname}, Config().OriginalsPath(), false) != ""
}

// SidecarJsonName returns the Google Photos style JSON sidecar path if it exists
// alongside the media file; otherwise it returns an empty string.
func (m *MediaFile) SidecarJsonName() string {
	jsonName := m.fileName + ".json"

	if fs.FileExistsNotEmpty(jsonName) {
		return jsonName
	}

	return ""
}

// ExifToolJsonName returns the path to the cached ExifTool JSON metadata file or
// an error when ExifTool integration is disabled.
func (m *MediaFile) ExifToolJsonName() (string, error) {
	if Config().DisableExifTool() {
		return "", fmt.Errorf("media: exiftool json files disabled")
	}

	return ExifToolCacheName(m.Hash())
}

// NeedsExifToolJson indicates whether a new ExifTool JSON export should be
// generated for this media file.
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

// CreateExifToolJson runs ExifTool via the provided Convert helper and merges
// its JSON output into the cached metadata. When nothing needs to be generated
// the call is a no-op.
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

// ReadExifToolJson loads cached ExifTool JSON metadata into the MediaFile
// metadata cache.
func (m *MediaFile) ReadExifToolJson() error {
	jsonName, err := m.ExifToolJsonName()

	if err != nil {
		return err
	}

	return m.metaData.JSON(jsonName, "")
}

// MetaData returns cached EXIF/sidecar metadata. On first access it probes the
// underlying file, merges JSON sidecars (including ExifTool exports) and
// normalises the time zone field.
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
			if jsonFiles := fs.SidecarJson.FindAll(m.FileName(), []string{Config().SidecarPath(), fs.PPHiddenPathname}, Config().OriginalsPath(), false); len(jsonFiles) == 0 {
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

		// Log error, if any.
		if err != nil {
			m.metaData.Error = err
			log.Debugf("%s in %s", err, clean.Log(m.BaseName()))
		}

		// Normalize time zone name.
		m.metaData.TimeZone = tz.Name(m.metaData.TimeZone)
	})

	return m.metaData
}

// VideoInfo probes the file with a built-in parser to retrieve video
// metadata; results are cached after the first successful call.
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
