package photoprism

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/meta"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/time/tz"
	"github.com/photoprism/photoprism/pkg/txt"
	"github.com/photoprism/photoprism/pkg/txt/clip"
)

// MediaFile indexes a single media file on behalf of the default owner.
func (ind *Index) MediaFile(m *MediaFile, o IndexOptions, originalName, photoUID string) (result IndexResult) {
	return ind.UserMediaFile(m, o, originalName, photoUID, entity.OwnerUnknown)
}

// UserMediaFile indexes a single media file for the provided owner, performing duplicate detection,
// metadata extraction, and database updates before returning an IndexResult describing the outcome.
func (ind *Index) UserMediaFile(m *MediaFile, o IndexOptions, originalName, photoUID, userUID string) (result IndexResult) {
	if m == nil {
		result.Status = IndexFailed
		result.Err = errors.New("index: no media file provided for processing - you may have found a bug")
		return result
	}

	// Skip file?
	if ind.files.Ignore(m.RootRelName(), m.Root(), m.ModTime(), o.Rescan) {
		// Skip known file.
		result.Status = IndexSkipped
		return result
	} else if o.FacesOnly && !m.IsJpeg() {
		// Skip non-jpeg file when indexing faces only.
		result.Status = IndexSkipped
		return result
	}

	start := time.Now()

	var photoQuery, fileQuery *gorm.DB
	var locKeywords []string

	file, primaryFile := entity.File{}, entity.File{}

	photo := entity.NewUserPhoto(o.Stack, userUID)
	metaData := meta.NewData()
	labels := classify.Labels{}
	isNSFW := false
	stripSequence := Config().Settings().StackSequences() && o.Stack

	fileRoot, fileBase, filePath, fileName := m.PathNameInfo(stripSequence)
	fullBase := m.BasePrefix(false)
	logName := clean.Log(fileName)
	fileSize, modTime, err := m.Stat()

	if err != nil {
		result.Status = IndexFailed
		result.Err = fmt.Errorf("index: %s not found (%s)", logName, err)
		return result
	}

	fileHash := ""
	fileChanged := true
	fileRenamed := false
	fileExists := false
	fileStacked := false

	photoExists := false

	event.Publish("index.indexing", event.Data{
		"uid":      o.UID,
		"action":   o.Action,
		"fileHash": fileHash,
		"fileSize": fileSize,
		"fileName": fileName,
		"fileRoot": fileRoot,
		"baseName": filepath.Base(fileName),
	})

	// Try to find existing file by path and name.
	fileQuery = entity.UnscopedDb().First(&file, "file_name = ? AND (file_root = ? OR file_root = '')", fileName, fileRoot)
	fileExists = fileQuery.Error == nil

	// Try to find existing file by hash. Skip this for sidecar files, and files outside the originals folder.
	if !fileExists && !m.IsSidecar() && m.Root() == entity.RootOriginals {
		fileHash = m.Hash()
		fileQuery = entity.UnscopedDb().First(&file, "file_hash = ?", fileHash)

		indFileName := ""

		if fileQuery.Error == nil {
			fileExists = true
			indFileName = FileName(file.FileRoot, file.FileName)
		}

		if !fileExists {
			// Do nothing.
		} else if fs.FileExists(indFileName) {
			if err = entity.AddDuplicate(m.RootRelName(), m.Root(), m.Hash(), m.FileSize(), m.ModTime().Unix()); err != nil {
				log.Errorf("index: %s in %s", err, m.RootRelName())
			}

			result.Status = IndexDuplicate
			return result
		} else if err = file.Rename(m.RootRelName(), m.Root(), filePath, fileBase); err != nil {
			result.Status = IndexFailed
			result.Err = fmt.Errorf("index: %s in %s (rename)", err, logName)
			return result
		} else if renamedSidecars, err := m.RenameSidecarFiles(indFileName); err != nil {
			log.Errorf("index: %s in %s (rename sidecars)", err.Error(), logName)

			fileRenamed = true
		} else {
			for srcName, destName := range renamedSidecars {
				if err := query.RenameFile(entity.RootSidecar, srcName, entity.RootSidecar, destName); err != nil {
					log.Errorf("index: %s in %s (update sidecar index)", err.Error(), filepath.Join(entity.RootSidecar, srcName))
				}
			}

			fileRenamed = true
		}
	}

	// Find existing photo if a photo uid was provided or file has not been indexed yet...
	if !fileExists && photoUID != "" {
		// Find existing photo by UID.
		photoQuery = entity.UnscopedDb().First(&photo, "photo_uid = ?", photoUID)

		if photoQuery.Error == nil {
			// Found.
			fileStacked = true
		} else {
			// Log and return error if photo uid was not found.
			result.Status = IndexFailed
			result.Err = fmt.Errorf("index: failed indexing %s, unknown photo uid %s (%s)", logName, photoUID, photoQuery.Error)
			return result
		}
	} else if !fileExists {
		// Find existing photo by matching path and name.
		if photoQuery = entity.UnscopedDb().First(&photo, "photo_path = ? AND photo_name = ?", filePath, fullBase); photoQuery.Error == nil || fileBase == fullBase || !o.Stack {
			// Skip next query.
		} else if photoQuery = entity.UnscopedDb().First(&photo, "photo_path = ? AND photo_name = ? AND photo_stack > -1", filePath, fileBase); photoQuery.Error == nil {
			// Found.
			fileStacked = true
		} else if photoQuery = entity.UnscopedDb().First(&photo, "id IN (SELECT photo_id FROM files WHERE file_name = LIKE ? AND file_root = ? AND file_sidecar = 0 AND file_missing = 0) AND photo_path = ? AND photo_stack > -1", fs.StripKnownExt(fileName)+".%", entity.RootOriginals, filePath); photoQuery.Error == nil {
			// Found.
			fileStacked = true
		}

		// Find existing photo by unique id or time and location?
		if o.Stack {
			// Same unique ID?
			if photoQuery.Error != nil && Config().Settings().StackUUID() && m.MetaData().HasDocumentID() {
				photoQuery = entity.UnscopedDb().First(&photo, "uuid <> '' AND uuid = ?", clean.Log(m.MetaData().DocumentID))

				if photoQuery.Error == nil {
					// Found.
					fileStacked = true
				}
			}

			// Matching location and time metadata?
			if photoQuery.Error != nil && Config().Settings().StackMeta() && m.MetaData().HasTimeAndPlace() {
				metaData = m.MetaData()
				photoQuery = entity.UnscopedDb().First(&photo, "photo_lat = ? AND photo_lng = ? AND taken_at = ? AND taken_src = 'meta' AND camera_serial = ?", metaData.Lat, metaData.Lng, metaData.TakenAt, metaData.CameraSerial)

				if photoQuery.Error == nil {
					// Found.
					fileStacked = true
				}
			}
		}
	} else if fileExists {
		// Find photo by the id or uid assigned to the file.
		if file.PhotoID > 0 {
			photoQuery = entity.UnscopedDb().First(&photo, "id = ?", file.PhotoID)
		} else if rnd.IsUID(file.PhotoUID, entity.PhotoUID) {
			photoQuery = entity.UnscopedDb().First(&photo, "photo_uid = ?", file.PhotoUID)
		} else {
			// Should never happen.
			result.Status = IndexFailed
			result.Err = fmt.Errorf("index: file %s has no photo id or uid assigned - you may have found a bug, please report", logName)
			return result
		}
	} else {
		// Should never happen.
		result.Status = IndexFailed
		result.Err = fmt.Errorf("index: unexpectedly failed indexing %s - you may have found a bug, please report", logName)
		return result
	}

	// Found a photo?
	photoExists = photoQuery.Error == nil

	// Detect changes in existing files.
	if fileExists {
		// Detect and report changed photo UID.
		if photoExists && photoUID != "" && photoUID != file.PhotoUID {
			fileChanged = true
			log.Debugf("index: %s has new photo uid %s", clean.Log(m.BaseName()), photoUID)
		}

		// Detect and report file changes.
		if fileRenamed {
			fileChanged = true
			log.Debugf("index: %s was renamed", clean.Log(m.BaseName()))
		} else if file.Changed(fileSize, modTime) {
			fileChanged = true
			log.Debugf("index: %s was modified (new size %d, old size %d, new timestamp %d, old timestamp %d)", clean.Log(m.BaseName()), fileSize, file.FileSize, modTime.Unix(), file.ModTime)
		} else if file.Missing() {
			fileChanged = true
			log.Debugf("index: %s was missing", clean.Log(m.BaseName()))
		}
	}

	// Update file <=> photo relationship if needed.
	if photoExists && (file.PhotoID != photo.ID || file.PhotoUID != photo.PhotoUID) {
		file.PhotoID = photo.ID
		file.PhotoUID = photo.PhotoUID
		file.PhotoTakenAt = photo.TakenAtLocal
	}

	// Skip unchanged files.
	if !fileChanged && photoExists && o.SkipUnchanged() || !photoExists && m.IsSidecar() {
		result.Status = IndexSkipped
		return result
	}

	// Remove file from duplicates table if exists.
	if err = entity.PurgeDuplicate(m.RootRelName(), m.Root()); err != nil {
		log.Errorf("index: %s in %s (purge duplicate)", err, m.RootRelName())
	}

	// Create default thumbnails if needed.
	if err = m.GenerateThumbnails(ind.thumbPath(), false); err != nil {
		result.Status = IndexFailed
		result.Err = fmt.Errorf("index: failed to generate thumbnails for %s (%s)", clean.Log(m.RootRelName()), err.Error())
		return result
	}

	// Fetch photo details such as keywords, subject, and artist.
	details := photo.GetDetails()

	// Try to recover photo metadata from backup if not exists.
	if !photoExists {
		photo.PhotoQuality = -1

		if o.Stack {
			photo.PhotoStack = entity.IsStackable
		}

		if yamlName := fs.SidecarYaml.FindFirst(m.FileName(), []string{Config().SidecarPath(), fs.PPHiddenPathname}, Config().OriginalsPath(), stripSequence); yamlName != "" {
			if err = photo.LoadFromYaml(yamlName); err != nil {
				log.Errorf("index: %s in %s (restore from yaml)", err.Error(), logName)
			} else if photo.HasUID() {
				photoExists = true
				log.Infof("index: metadata of photo uid %s restored from %s", photo.PhotoUID, clean.Log(filepath.Base(yamlName)))
			}
		}
	}

	// Calculate SHA1 file hash if not exists.
	if fileHash == "" {
		fileHash = m.Hash()
	}

	// Update file hash references?
	if !fileExists || file.FileHash == "" || file.FileHash == fileHash {
		// Do nothing.
	} else if err := file.ReplaceHash(fileHash); err != nil {
		log.Errorf("index: %s while updating covers of %s", err, logName)
	}

	// Clear (previous) file error.
	file.FileError = ""

	// Flag first JPEG as primary file for this photo.
	if !file.FilePrimary {
		if photoExists {
			if res := entity.UnscopedDb().Where("photo_id = ? AND file_primary = 1 AND file_type IN (?) AND file_error = ''", photo.ID, media.PreviewExpr).First(&primaryFile); res.Error != nil {
				file.FilePrimary = m.IsPreviewImage()
			}
		} else {
			file.FilePrimary = m.IsPreviewImage()
		}
	}

	// Update photo path and name based on the main filename.
	if !fileStacked && (file.FilePrimary || photo.PhotoName == "") {
		photo.PhotoPath = filePath

		if !o.Stack || !stripSequence || photo.PhotoStack == entity.IsUnstacked {
			photo.PhotoName = fullBase
		} else {
			photo.PhotoName = fileBase
		}
	}

	// Set basic file information.
	file.FileRoot = fileRoot
	file.FileName = fileName
	file.FileHash = fileHash
	file.FileSize = fileSize

	// Set file original name if available.
	if originalName != "" {
		file.OriginalName = originalName
	}

	// Set photo original name based on file original name.
	if file.OriginalName != "" {
		photo.OriginalName = fs.StripKnownExt(file.OriginalName)
	}

	if photo.PhotoQuality == -1 && (file.FilePrimary || fileChanged) {
		// Restore pictures that have been purged automatically.
		photo.DeletedAt = nil
	} else if o.SkipArchived && photo.DeletedAt != nil {
		// Skip archived pictures for faster indexing.
		result.Status = IndexArchived
		return result
	}

	// Extra labels to ba added when new files have a photo id.
	extraLabels := classify.Labels{}

	// Detect faces in images?
	if o.FacesOnly && (!photoExists || !fileExists || !file.FilePrimary || file.FileError != "") {
		// New and non-primary files can be skipped when updating faces only.
		result.Status = IndexSkipped
		return result
	} else if o.DetectFaces && file.FilePrimary {
		// Run face detection on primary files when enabled for this indexing run.
		if markers := file.Markers(); markers != nil {
			// Detect faces.
			faces := ind.Faces(m, markers.DetectedFaceCount())

			// Create markers from faces and add them.
			if len(faces) > 0 {
				file.AddFaces(faces)
			}

			// Skip when indexing faces only and no new markers were found.
			if !file.UnsavedMarkers() && o.FacesOnly {
				result.Status = IndexSkipped
				return result
			}

			// Update photo face count.
			photo.PhotoFaces = markers.ValidFaceCount()
		} else {
			log.Errorf("index: failed loading markers for %s", logName)
		}
	}

	// Reset photo metadata if this is a forced rescan.
	if o.Rescan && photoUID == "" {
		// Reset video duration.
		photo.ResetDuration()

		// Reset media type.
		photo.ResetMediaType(entity.SrcFile)
	}

	// Reset file perceptive diff and chroma percent.
	file.FileDiff = -1
	file.FileChroma = -1
	file.FileVideo = m.IsVideo()
	file.MediaType = m.MediaType().String()

	// Handle file types.
	switch {
	case m.IsPreviewImage():
		// Update color information, if available.
		if color, colorErr := m.Colors(Config().ThumbCachePath()); colorErr != nil {
			log.Debugf("%s while detecting colors", colorErr.Error())
			file.FileError = clip.Chars(colorErr.Error(), txt.ClipError)
			file.FilePrimary = false
		} else {
			file.FileMainColor = color.MainColor.Name()
			file.FileColors = color.Colors.Hex()
			file.FileLuminance = color.Luminance.Hex()
			file.FileDiff = color.Luminance.Diff()
			file.FileChroma = color.Chroma.Percent()

			if file.FilePrimary {
				photo.PhotoColor = color.MainColor.ID()
			}
		}

		// Update resolution and aspect ratio.
		if m.Width() > 0 && m.Height() > 0 {
			file.FileWidth = m.Width()
			file.FileHeight = m.Height()
			file.FileAspectRatio = m.AspectRatio()
			file.FilePortrait = m.Portrait()

			// Set photo resolution based on the largest media file.
			if res := m.Megapixels(); res > photo.PhotoResolution {
				photo.PhotoResolution = res
			}
		}

		// Update file metadata.
		if data := m.MetaData(); data.Error == nil {
			file.FileCodec = data.Codec
			file.SetMediaUTC(data.TakenAt)
			file.SetProjection(data.Projection)
			file.SetHDR(data.IsHDR())
			file.SetColorProfile(data.ColorProfile)
			file.SetSoftware(data.Software)

			// Get video metadata from embedded file?
			if !data.HasVideoEmbedded {
				file.SetDuration(data.Duration)
				file.SetFPS(data.FPS)
				file.SetFrames(data.Frames)
			} else if info := m.VideoInfo(); info.Compatible {
				file.SetDuration(info.Duration)
				file.SetFPS(info.FPS)
				file.SetFrames(info.Frames)

				if file.FileDuration > photo.PhotoDuration {
					photo.PhotoDuration = file.FileDuration
				}

				// Change file and photo type to "live" if the file has a video embedded.
				file.FileVideo = true
				file.MediaType = entity.MediaLive

				// Set photo media type to "live".
				photo.SetMediaType(media.Live, entity.SrcFile)
			}

			if file.OriginalName == "" && filepath.Base(file.FileName) != data.FileName {
				file.OriginalName = data.FileName
				if photo.OriginalName == "" {
					photo.OriginalName = fs.StripKnownExt(data.FileName)
				}
			}

			if data.HasInstanceID() {
				log.Infof("index: %s has instance_id %s", logName, clean.Log(data.InstanceID))

				file.InstanceID = data.InstanceID
			}

			if m.IsAnimatedImage() && file.FileDuration > photo.PhotoDuration {
				photo.PhotoDuration = file.FileDuration
			}
		}

		// If the file contains multiple images for an animation,
		// change the media type to "animated".
		if photo.HasMediaType(media.Image) && m.IsAnimatedImage() {
			photo.SetMediaType(media.Animated, entity.SrcAuto)
		} else if photo.PhotoType == "" {
			photo.SetMediaType(media.Image, entity.SrcAuto)
		}
	case m.IsXMP():
		if data, dataErr := meta.XMP(m.FileName()); dataErr == nil {
			// Update basic metadata.
			photo.SetTitle(data.Title, entity.SrcXmp)
			photo.SetCaption(data.Caption, entity.SrcXmp)
			photo.SetTakenAt(data.TakenAt, data.TakenAtLocal, data.TimeZone, entity.SrcXmp)
			photo.SetCoordinates(data.Lat, data.Lng, data.Altitude, entity.SrcXmp)

			// Update metadata details.
			details.SetKeywords(data.Keywords.String(), entity.SrcXmp)
			details.SetNotes(data.Notes, entity.SrcXmp)
			details.SetSubject(data.Subject, entity.SrcXmp)
			details.SetArtist(data.Artist, entity.SrcXmp)
			details.SetCopyright(data.Copyright, entity.SrcXmp)
			details.SetLicense(data.License, entity.SrcXmp)
			details.SetSoftware(data.Software, entity.SrcXmp)

			// Update externally marked as favorite.
			if data.Favorite {
				_ = photo.SetFavorite(data.Favorite)
			}
		} else {
			log.Warn(dataErr.Error())
			file.FileError = clip.Chars(dataErr.Error(), txt.ClipError)
		}
	case m.IsRaw(), m.IsImage():
		if data := m.MetaData(); data.Error == nil {
			// Update basic metadata.
			photo.SetTitle(data.Title, entity.SrcMeta)
			photo.SetCaption(data.Caption, entity.SrcMeta)
			photo.SetTakenAt(data.TakenAt, data.TakenAtLocal, data.TimeZone, entity.SrcMeta)
			photo.SetCoordinates(data.Lat, data.Lng, data.Altitude, entity.SrcMeta)
			photo.SetCameraSerial(data.CameraSerial)

			// Update metadata details.
			details.SetKeywords(data.Keywords.String(), entity.SrcMeta)
			details.SetNotes(data.Notes, entity.SrcMeta)
			details.SetSubject(data.Subject, entity.SrcMeta)
			details.SetArtist(data.Artist, entity.SrcMeta)
			details.SetCopyright(data.Copyright, entity.SrcMeta)
			details.SetLicense(data.License, entity.SrcMeta)
			details.SetSoftware(data.Software, entity.SrcMeta)

			if data.HasDocumentID() && photo.UUID == "" {
				log.Infof("index: %s has document_id %s", logName, clean.Log(data.DocumentID))

				photo.UUID = data.DocumentID
			}

			if data.HasInstanceID() {
				log.Infof("index: %s has instance_id %s", logName, clean.Log(data.InstanceID))

				file.InstanceID = data.InstanceID
			}

			if file.OriginalName == "" && filepath.Base(file.FileName) != data.FileName {
				file.OriginalName = data.FileName
				if photo.OriginalName == "" {
					photo.OriginalName = fs.StripKnownExt(data.FileName)
				}
			}

			file.FileCodec = data.Codec
			file.FileWidth = m.Width()
			file.FileHeight = m.Height()
			file.FileAspectRatio = m.AspectRatio()
			file.FilePortrait = m.Portrait()
			file.SetMediaUTC(data.TakenAt)
			file.SetPages(data.Pages)
			file.SetProjection(data.Projection)
			file.SetHDR(data.IsHDR())
			file.SetColorProfile(data.ColorProfile)
			file.SetSoftware(data.Software)

			// Get video metadata from embedded file?
			if !m.IsHeic() || !data.HasVideoEmbedded {
				file.SetDuration(data.Duration)
				file.SetFPS(data.FPS)
				file.SetFrames(data.Frames)
			} else if info := m.VideoInfo(); info.Compatible {
				file.SetDuration(info.Duration)
				file.SetFPS(info.FPS)
				file.SetFrames(info.Frames)

				if file.FileDuration > photo.PhotoDuration {
					photo.PhotoDuration = file.FileDuration
				}

				// Change file and photo type to "live" if the file has a video embedded.
				file.FileVideo = true
				file.MediaType = entity.MediaLive

				// If the file also contains a video, set photo media type to "live".
				photo.SetMediaType(media.Live, entity.SrcFile)
			} else {
				// If the file does not contain a video, set the media type to "image".
				photo.SetMediaType(media.Image, entity.SrcAuto)
			}

			// Set photo resolution based on the largest media file.
			if res := m.Megapixels(); res > photo.PhotoResolution {
				photo.PhotoResolution = res
			}

			photo.SetCamera(entity.FirstOrCreateCamera(entity.NewCamera(m.CameraMake(), m.CameraModel())), entity.SrcMeta)
			photo.SetLens(entity.FirstOrCreateLens(entity.NewLens(m.LensMake(), m.LensModel())), entity.SrcMeta)
			photo.SetExposure(m.FocalLength(), m.FNumber(), m.Iso(), m.Exposure(), entity.SrcMeta)
		}

		// If the media type is still set to "image" and has not been
		// manually modified, then check and update it as needed.
		if photo.HasMediaType(media.Image) {
			if m.IsAnimatedImage() {
				photo.SetMediaType(media.Animated, entity.SrcAuto)
			} else if m.IsRaw() {
				photo.SetMediaType(media.Raw, entity.SrcAuto)
			} else if m.IsLive(photo.PhotoDuration) {
				photo.SetMediaType(media.Live, entity.SrcAuto)
			} else if m.IsVector() {
				photo.SetMediaType(media.Vector, entity.SrcAuto)
			}
		}
	case m.IsVector():
		if data := m.MetaData(); data.Error == nil {
			// Update basic metadata.
			photo.SetTitle(data.Title, entity.SrcMeta)
			photo.SetCaption(data.Caption, entity.SrcMeta)
			photo.SetTakenAt(data.TakenAt, data.TakenAtLocal, data.TimeZone, entity.SrcMeta)

			// Update metadata details.
			details.SetKeywords(data.Keywords.String(), entity.SrcMeta)
			details.SetNotes(data.Notes, entity.SrcMeta)
			details.SetSubject(data.Subject, entity.SrcMeta)
			details.SetArtist(data.Artist, entity.SrcMeta)
			details.SetCopyright(data.Copyright, entity.SrcMeta)
			details.SetLicense(data.License, entity.SrcMeta)
			details.SetSoftware(data.Software, entity.SrcMeta)

			if data.HasDocumentID() && photo.UUID == "" {
				log.Infof("index: %s has document_id %s", logName, clean.Log(data.DocumentID))

				photo.UUID = data.DocumentID
			}

			if data.HasInstanceID() {
				log.Infof("index: %s has instance_id %s", logName, clean.Log(data.InstanceID))

				file.InstanceID = data.InstanceID
			}

			if file.OriginalName == "" && filepath.Base(file.FileName) != data.FileName {
				file.OriginalName = data.FileName
				if photo.OriginalName == "" {
					photo.OriginalName = fs.StripKnownExt(data.FileName)
				}
			}

			file.FileCodec = data.Codec
			file.FileWidth = m.Width()
			file.FileHeight = m.Height()
			file.FileAspectRatio = m.AspectRatio()
			file.FilePortrait = m.Portrait()
			file.SetMediaUTC(data.TakenAt)
			file.SetPages(data.Pages)
			file.SetProjection(data.Projection)
			file.SetHDR(data.IsHDR())
			file.SetColorProfile(data.ColorProfile)
			file.SetSoftware(data.Software)

			// Set photo resolution based on the largest media file.
			if res := m.Megapixels(); res > photo.PhotoResolution {
				photo.PhotoResolution = res
			}
		}

		// Set photo media type to "vector".
		photo.SetMediaType(media.Vector, entity.SrcAuto)
	case m.IsDocument():
		if data := m.MetaData(); data.Error == nil {
			photo.SetTitle(data.Title, entity.SrcMeta)
			photo.SetCaption(data.Caption, entity.SrcMeta)
			photo.SetTakenAt(data.TakenAt, data.TakenAtLocal, data.TimeZone, entity.SrcMeta)

			// Update metadata details.
			details.SetKeywords(data.Keywords.String(), entity.SrcMeta)
			details.SetNotes(data.Notes, entity.SrcMeta)
			details.SetSubject(data.Subject, entity.SrcMeta)
			details.SetArtist(data.Artist, entity.SrcMeta)
			details.SetCopyright(data.Copyright, entity.SrcMeta)
			details.SetLicense(data.License, entity.SrcMeta)
			details.SetSoftware(data.Software, entity.SrcMeta)

			if data.HasDocumentID() && photo.UUID == "" {
				log.Infof("index: %s has document_id %s", logName, clean.Log(data.DocumentID))

				photo.UUID = data.DocumentID
			}

			if data.HasInstanceID() {
				log.Infof("index: %s has instance_id %s", logName, clean.Log(data.InstanceID))

				file.InstanceID = data.InstanceID
			}

			if file.OriginalName == "" && filepath.Base(file.FileName) != data.FileName {
				file.OriginalName = data.FileName
				if photo.OriginalName == "" {
					photo.OriginalName = fs.StripKnownExt(data.FileName)
				}
			}

			file.FileCodec = data.Codec
			file.FileWidth = m.Width()
			file.FileHeight = m.Height()
			file.FileAspectRatio = m.AspectRatio()
			file.FilePortrait = m.Portrait()
			file.SetMediaUTC(data.TakenAt)
			file.SetPages(data.Pages)
			file.SetColorProfile(data.ColorProfile)
			file.SetSoftware(data.Software)

			// Set photo resolution based on the largest media file.
			if res := m.Megapixels(); res > photo.PhotoResolution {
				photo.PhotoResolution = res
			}
		}

		// Set photo media type to "document".
		photo.SetMediaType(media.Document, entity.SrcAuto)
	case m.IsVideo():
		if data := m.MetaData(); data.Error == nil {
			photo.SetTitle(data.Title, entity.SrcMeta)
			photo.SetCaption(data.Caption, entity.SrcMeta)
			photo.SetTakenAt(data.TakenAt, data.TakenAtLocal, data.TimeZone, entity.SrcMeta)
			photo.SetCoordinates(data.Lat, data.Lng, data.Altitude, entity.SrcMeta)
			photo.SetCameraSerial(data.CameraSerial)

			// Update metadata details.
			details.SetKeywords(data.Keywords.String(), entity.SrcMeta)
			details.SetNotes(data.Notes, entity.SrcMeta)
			details.SetSubject(data.Subject, entity.SrcMeta)
			details.SetArtist(data.Artist, entity.SrcMeta)
			details.SetCopyright(data.Copyright, entity.SrcMeta)
			details.SetLicense(data.License, entity.SrcMeta)
			details.SetSoftware(data.Software, entity.SrcMeta)

			if data.HasDocumentID() && photo.UUID == "" {
				log.Infof("index: %s has document_id %s", logName, clean.Log(data.DocumentID))

				photo.UUID = data.DocumentID
			}

			if data.HasInstanceID() {
				log.Infof("index: %s has instance_id %s", logName, clean.Log(data.InstanceID))

				file.InstanceID = data.InstanceID
			}

			if file.OriginalName == "" && filepath.Base(file.FileName) != data.FileName {
				file.OriginalName = data.FileName
				if photo.OriginalName == "" {
					photo.OriginalName = fs.StripKnownExt(data.FileName)
				}
			}

			file.FileCodec = data.Codec
			file.FileWidth = m.Width()
			file.FileHeight = m.Height()
			file.FileAspectRatio = m.AspectRatio()
			file.FilePortrait = m.Portrait()
			file.SetMediaUTC(data.TakenAt)
			file.SetDuration(data.Duration)
			file.SetFPS(data.FPS)
			file.SetFrames(data.Frames)
			file.SetProjection(data.Projection)
			file.SetHDR(data.IsHDR())
			file.SetColorProfile(data.ColorProfile)
			file.SetSoftware(data.Software)

			// Set photo resolution based on the largest media file.
			if res := m.Megapixels(); res > photo.PhotoResolution {
				photo.PhotoResolution = res
			}

			if file.FileDuration > photo.PhotoDuration {
				photo.PhotoDuration = file.FileDuration
			}

			photo.SetCamera(entity.FirstOrCreateCamera(entity.NewCamera(m.CameraMake(), m.CameraModel())), entity.SrcMeta)
			photo.SetLens(entity.FirstOrCreateLens(entity.NewLens(m.LensMake(), m.LensModel())), entity.SrcMeta)
			photo.SetExposure(m.FocalLength(), m.FNumber(), m.Iso(), m.Exposure(), entity.SrcMeta)
		}

		// Set the media type to "live" instead of "video" if the video duration
		// is less than 3.1 seconds and a JPEG or HEIC image exists.
		if photo.PhotoDuration > 0 && m.IsLive(photo.PhotoDuration) {
			photo.SetMediaType(media.Live, entity.SrcAuto)
		} else {
			photo.SetMediaType(media.Video, entity.SrcAuto)
		}

		// Set the video dimensions from the primary image if it could not be determined from the video metadata.
		// If there is no primary image yet, File.UpdateVideoInfos() sets the fields in retrospect when there is one.
		if file.FileWidth == 0 && primaryFile.FileWidth > 0 {
			file.FileWidth = primaryFile.FileWidth
			file.FileHeight = primaryFile.FileHeight
			file.FileAspectRatio = primaryFile.FileAspectRatio
			file.FilePortrait = primaryFile.FilePortrait
		}

		// Set the video appearance from the primary image. In a future version, a still image extracted from the
		// video could be used for this purpose if the primary image is not directly derived from the video file,
		// e.g. in live photo stacks, see https://github.com/photoprism/photoprism/pull/3588#issuecomment-1683429455
		if primaryFile.FileDiff > 0 {
			file.FileMainColor = primaryFile.FileMainColor
			file.FileColors = primaryFile.FileColors
			file.FileLuminance = primaryFile.FileLuminance
			file.FileDiff = primaryFile.FileDiff
			file.FileChroma = primaryFile.FileChroma
		}
	}

	// Set taken date based on file mod time or name if other metadata is missing.
	if m.IsMedia() && entity.SrcPriority[photo.TakenSrc] <= entity.SrcPriority[entity.SrcName] {
		// Try to extract time from original file name first.
		if taken := txt.DateFromFilePath(photo.OriginalName); !taken.IsZero() {
			photo.SetTakenAt(taken, taken, tz.Local, entity.SrcName)
		} else if takenAt, takenAtLocal, takenSrc := m.TakenAt(); takenSrc == entity.SrcName {
			photo.SetTakenAt(takenAt, takenAtLocal, tz.Local, entity.SrcName)
		} else if !takenAt.IsZero() && !takenAtLocal.IsZero() {
			photo.SetTakenAt(takenAt, takenAtLocal, tz.Local, takenSrc)
		}
	}

	// File obviously exists: remove deleted and missing flags.
	file.DeletedAt = nil
	file.FileMissing = false

	// Previews files are used for rendering thumbnails and image classification, plus sidecar files if they exist.
	if file.FilePrimary {
		primaryFile = file

		// Classify images with TensorFlow if the run enables automatic labels.
		if o.GenerateLabels {
			labels = m.GenerateLabels(entity.SrcAuto)

			// Append labels from other sources such as face detection.
			if len(extraLabels) > 0 {
				labels = append(labels, extraLabels...)
			}

			isNSFW = labels.IsNSFW(vision.Config.Thresholds.GetNSFW())
		}

		// Decouple NSFW detection from label generation.
		if !photoExists {
			if isNSFW {
				photo.PhotoPrivate = true
			} else if o.DetectNsfw {
				photo.PhotoPrivate = m.DetectNSFW()
			}
		}

		// Read metadata from embedded Exif and JSON sidecar file, if exists.
		if data := m.MetaData(); data.Error == nil {
			// Update basic metadata.
			photo.SetTitle(data.Title, entity.SrcMeta)
			photo.SetCaption(data.Caption, entity.SrcMeta)
			photo.SetTakenAt(data.TakenAt, data.TakenAtLocal, data.TimeZone, entity.SrcMeta)
			photo.SetCoordinates(data.Lat, data.Lng, data.Altitude, entity.SrcMeta)
			photo.SetCameraSerial(data.CameraSerial)

			// Update metadata details.
			details.SetKeywords(data.Keywords.String(), entity.SrcMeta)
			details.SetNotes(data.Notes, entity.SrcMeta)
			details.SetSubject(data.Subject, entity.SrcMeta)
			details.SetArtist(data.Artist, entity.SrcMeta)
			details.SetCopyright(data.Copyright, entity.SrcMeta)
			details.SetLicense(data.License, entity.SrcMeta)
			details.SetSoftware(data.Software, entity.SrcMeta)

			if data.HasDocumentID() && photo.UUID == "" {
				log.Debugf("index: %s has document_id %s", logName, clean.Log(data.DocumentID))

				photo.UUID = data.DocumentID
			}
		}

		photo.SetCamera(entity.FirstOrCreateCamera(entity.NewCamera(m.CameraMake(), m.CameraModel())), entity.SrcMeta)
		photo.SetLens(entity.FirstOrCreateLens(entity.NewLens(m.LensMake(), m.LensModel())), entity.SrcMeta)
		photo.SetExposure(m.FocalLength(), m.FNumber(), m.Iso(), m.Exposure(), entity.SrcMeta)

		var locLabels classify.Labels

		locKeywords, locLabels = photo.UpdateLocation()
		labels = append(labels, locLabels...)
	}

	if photo.UnknownLocation() {
		photo.Cell = &entity.UnknownLocation
		photo.CellID = entity.UnknownLocation.ID
	}

	if photo.UnknownPlace() {
		photo.Place = &entity.UnknownPlace
		photo.PlaceID = entity.UnknownPlace.ID
	}

	photo.UpdateDateFields()

	// Panorama?
	if file.Panorama() {
		photo.PhotoPanorama = true
	}

	// Update file properties.
	file.FileSidecar = m.IsSidecar()
	file.FileType = m.FileType().String()
	file.FileMime = m.ContentType()
	file.SetOrientation(m.Orientation(), entity.SrcMeta)
	file.ModTime = modTime.UTC().Truncate(time.Second).Unix()

	// Detect ICC color profile for JPEGs if still unknown at this point.
	if file.FileColorProfile == "" && fs.ImageJpeg.Equal(file.FileType) {
		file.SetColorProfile(m.ColorProfile())
	}

	// Update existing photo entity?
	if photo.HasID() {
		if err = photo.Save(); err != nil {
			result.Status = IndexFailed
			result.Err = fmt.Errorf("index: %s in %s (update existing photo)", err, logName)
			return result
		}
	} else {
		// Create a new photo entity or load the existing entity if it exists.
		if p := photo.FirstOrCreate(); p == nil {
			result.Status = IndexFailed
			result.Err = fmt.Errorf("index: failed to create %s", logName)
			return result
		} else {
			photo = *p
		}

		if photo.PhotoPrivate {
			event.Publish("count.private", event.Data{
				"count": 1,
			})
		}

		switch photo.MediaType() {
		case media.Animated:
			event.Publish("count.animated", event.Data{
				"count": 1,
			})
		case media.Live:
			event.Publish("count.live", event.Data{
				"count": 1,
			})
		case media.Audio:
			event.Publish("count.audio", event.Data{
				"count": 1,
			})
		case media.Video:
			event.Publish("count.videos", event.Data{
				"count": 1,
			})
		case media.Document:
			event.Publish("count.documents", event.Data{
				"count": 1,
			})
		default:
			event.Publish("count.photos", event.Data{
				"count": 1,
			})
		}

		event.EntitiesCreated("photos", []entity.Photo{photo})
	}

	photo.AddLabels(labels)

	file.PhotoID = photo.ID
	result.PhotoID = photo.ID

	file.PhotoUID = photo.PhotoUID
	result.PhotoUID = photo.PhotoUID

	// Set photo properties based on primary file.
	if file.FilePrimary {
		photoLabels := photo.ClassifyLabels()

		if err = photo.GenerateTitle(photoLabels); err != nil {
			log.Debugf("%s in %s (update title)", err, logName)
		}

		w := txt.Words(details.Keywords)

		if !fs.IsGenerated(fileBase) {
			w = append(w, txt.FilenameKeywords(fileBase)...)
		}

		if photo.OriginalName == "" {
			// Do nothing.
		} else if fs.IsGenerated(photo.OriginalName) {
			w = append(w, txt.FilenameKeywords(filepath.Dir(photo.OriginalName))...)
		} else {
			w = append(w, txt.FilenameKeywords(photo.OriginalName)...)
		}

		w = append(w, txt.FilenameKeywords(filePath)...)
		w = append(w, locKeywords...)
		w = append(w, file.FileMainColor)
		w = append(w, photoLabels.Keywords()...)

		details.Keywords = strings.Join(txt.UniqueWords(w), ", ")

		if details.Keywords != "" {
			log.Tracef("index: %s has keywords %s", logName, details.Keywords)
		} else {
			log.Tracef("index: found no keywords for %s", logName)
		}

		photo.PhotoQuality = photo.QualityScore()

		if err = photo.Save(); err != nil {
			result.Status = IndexFailed
			result.Err = fmt.Errorf("index: %s in %s (update metadata)", err, logName)
			return result
		}

		if err = photo.UpdateLabels(); err != nil {
			log.Errorf("index: %s in %s (update labels)", err, logName)
		}

		if err = photo.IndexKeywords(); err != nil {
			log.Errorf("index: %s in %s (save keywords)", err, logName)
		}

		if err = query.AlbumEntryFound(photo.PhotoUID); err != nil {
			log.Errorf("index: %s in %s (remove missing flag from album entry)", err, logName)
		}
	} else if err = photo.UpdateQuality(); err != nil {
		result.Status = IndexFailed
		result.Err = fmt.Errorf("index: %s in %s (update quality)", err, logName)
		return result
	}

	result.Status = IndexUpdated

	if fileQuery.Error == nil {
		file.UpdatedIn = int64(time.Since(start))

		if err = file.Save(); err != nil {
			result.Status = IndexFailed
			result.Err = fmt.Errorf("index: %s in %s (update existing file)", err, logName)
			return result
		}
	} else {
		file.CreatedIn = int64(time.Since(start))

		if err = file.Create(); err != nil {
			result.Status = IndexFailed
			result.Err = fmt.Errorf("index: %s in %s (add new file)", err, logName)
			return result
		}

		event.Publish("count.files", event.Data{
			"count": 1,
		})

		if fileStacked {
			result.Status = IndexStacked
		} else {
			result.Status = IndexAdded
		}
	}

	// Update related video files so they are properly grouped with the primary image in search results.
	if photo.HasMediaType(media.Video, media.Live) && file.FilePrimary {
		if updateErr := file.UpdateVideoInfos(); updateErr != nil {
			log.Errorf("index: %s in %s (update video infos)", updateErr, logName)
		}
	}

	result.FileID = file.ID
	result.FileUID = file.FileUID

	downloadedAs := fileName

	if originalName != "" {
		downloadedAs = originalName
	}

	if err = query.SetDownloadFileID(downloadedAs, file.ID); err != nil {
		log.Errorf("index: %s in %s (set download id)", err, logName)
	}

	if !o.Stack || photo.PhotoStack == entity.IsUnstacked {
		// Do nothing.
	} else if original, merged, err := photo.Merge(Config().Settings().StackMeta(), Config().Settings().StackUUID()); err != nil {
		log.Errorf("index: %s in %s (merge)", err.Error(), logName)
	} else if len(merged) == 1 && original.ID == photo.ID {
		log.Infof("index: merged one existing photo with %s", logName)
	} else if len(merged) > 1 && original.ID == photo.ID {
		log.Infof("index: merged %d existing photos with %s", len(merged), logName)
	} else if len(merged) > 0 && original.ID != photo.ID {
		log.Infof("index: merged %s with existing photo id %d", logName, original.ID)
		result.Status = IndexStacked
		return result
	}

	// Create backup of picture metadata in sidecar YAML file.
	if file.FilePrimary && Config().SidecarYaml() {
		if err = photo.SaveSidecarYaml(Config().OriginalsPath(), Config().SidecarPath()); err != nil {
			log.Errorf("index: %s in %s (save as yaml)", err, logName)
		}
	}

	return result
}
