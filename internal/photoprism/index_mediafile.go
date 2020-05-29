package photoprism

import (
	"errors"
	"math"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/classify"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/meta"
	"github.com/photoprism/photoprism/internal/nsfw"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

const (
	IndexUpdated   IndexStatus = "updated"
	IndexAdded     IndexStatus = "added"
	IndexSkipped   IndexStatus = "skipped"
	IndexDuplicate IndexStatus = "skipped duplicate"
	IndexArchived  IndexStatus = "skipped archived"
	IndexFailed    IndexStatus = "failed"
)

type IndexStatus string

type IndexResult struct {
	Status   IndexStatus
	Error    error
	FileID   uint
	FileUID  string
	PhotoID  uint
	PhotoUID string
}

func (r IndexResult) String() string {
	return string(r.Status)
}

func (r IndexResult) Success() bool {
	return r.Error == nil && r.FileID > 0
}

func (ind *Index) MediaFile(m *MediaFile, o IndexOptions, originalName string) (result IndexResult) {
	if m == nil {
		err := errors.New("index: media file is nil - you might have found a bug")
		log.Error(err)
		result.Error = err
		result.Status = IndexFailed
		return result
	}

	start := time.Now()

	var photoQuery, fileQuery *gorm.DB
	var locKeywords []string

	file, primaryFile := entity.File{}, entity.File{}

	photo := entity.Photo{PhotoType: entity.TypeImage, PhotoCountry: entity.UnknownCountry.ID}
	metaData := meta.Data{}
	description := entity.Details{}
	labels := classify.Labels{}

	fileBase := m.Base(ind.conf.Settings().Index.Group)
	filePath := m.RelativePath(ind.originalsPath())
	fileRoot := entity.RootDefault
	fileName := m.RelativeName(ind.originalsPath())
	quotedName := txt.Quote(m.RelativeName(ind.originalsPath()))
	fileHash := ""
	fileSize, fileModified := m.Stat()
	fileChanged := true
	fileExists := false
	photoExists := false

	event.Publish("index.indexing", event.Data{
		"fileHash": fileHash,
		"fileSize": fileSize,
		"fileName": fileName,
		"fileRoot": fileRoot,
		"baseName": filepath.Base(fileName),
	})

	fileQuery = entity.UnscopedDb().First(&file, "file_name = ?", fileName)
	fileExists = fileQuery.Error == nil

	if !fileExists && !m.IsSidecar() {
		fileHash = m.Hash()
		fileQuery = entity.UnscopedDb().First(&file, "file_hash = ?", fileHash)
		fileExists = fileQuery.Error == nil

		if fileExists && fs.FileExists(filepath.Join(ind.conf.OriginalsPath(), file.FileName)) {
			result.Status = IndexDuplicate
			return result
		}

		if !fileExists && m.MetaData().HasInstanceID() {
			fileQuery = entity.UnscopedDb().First(&file, "instance_id = ?", m.MetaData().InstanceID)
			fileExists = fileQuery.Error == nil
		}
	}

	if !fileExists {
		photoQuery = entity.UnscopedDb().First(&photo, "photo_path = ? AND photo_name = ?", filePath, fileBase)

		if photoQuery.Error != nil && m.MetaData().HasTimeAndPlace() {
			metaData = m.MetaData()
			photoQuery = entity.UnscopedDb().First(&photo, "photo_lat = ? AND photo_lng = ? AND taken_at = ?", metaData.Lat, metaData.Lng, metaData.TakenAt)
		}

		if photoQuery.Error != nil && m.MetaData().HasDocumentID() {
			photoQuery = entity.UnscopedDb().First(&photo, "document_id = ?", m.MetaData().DocumentID)
		}
	} else {
		photoQuery = entity.UnscopedDb().First(&photo, "id = ?", file.PhotoID)

		fileChanged = file.Changed(fileSize, fileModified)

		if fileChanged {
			log.Debugf("index: file was modified (new size %d, old size %d, new date %s, old date %s)", fileSize, file.FileSize, fileModified, file.FileModified)
		}
	}

	photoExists = photoQuery.Error == nil

	if !fileChanged && photoExists && o.SkipUnchanged() {
		result.Status = IndexSkipped
		return result
	}

	if photoExists {
		entity.UnscopedDb().Model(&photo).Related(&description)
	} else {
		photo.PhotoQuality = -1
	}

	if fileHash == "" {
		fileHash = m.Hash()
	}

	photo.PhotoPath = filePath
	photo.PhotoName = fileBase

	if !file.FilePrimary {
		if photoExists {
			if q := entity.UnscopedDb().Where("file_type = 'jpg' AND file_primary = 1 AND photo_id = ?", photo.ID).First(&primaryFile); q.Error != nil {
				file.FilePrimary = m.IsJpeg()
			}
		} else {
			file.FilePrimary = m.IsJpeg()
		}
	}

	if photo.PhotoQuality == -1 && file.FilePrimary {
		// restore photos that have been purged automatically
		photo.DeletedAt = nil
	} else if photo.DeletedAt != nil {
		// don't waste time indexing deleted / archived photos
		result.Status = IndexArchived
		return result
	}

	// Handle file types other than JPEG.
	switch {
	case m.IsJpeg():
		// Color information
		if p, err := m.Colors(ind.thumbPath()); err != nil {
			log.Errorf("index: %s for %s", err.Error(), quotedName)
		} else {
			file.FileMainColor = p.MainColor.Name()
			file.FileColors = p.Colors.Hex()
			file.FileLuminance = p.Luminance.Hex()
			file.FileDiff = p.Luminance.Diff()
			file.FileChroma = p.Chroma.Value()
		}

		if m.Width() > 0 && m.Height() > 0 {
			file.FileWidth = m.Width()
			file.FileHeight = m.Height()
			file.FileAspectRatio = m.AspectRatio()
			file.FilePortrait = m.Width() < m.Height()

			megapixels := int(math.Round(float64(file.FileWidth*file.FileHeight) / 1000000))

			if megapixels > photo.PhotoResolution {
				photo.PhotoResolution = megapixels
			}
		}
	case m.IsXMP():
		// TODO: Proof-of-concept for indexing XMP sidecar files
		if data, err := meta.XMP(m.FileName()); err == nil {
			photo.SetTitle(data.Title, entity.SrcXmp)
			photo.SetDescription(data.Description, entity.SrcXmp)

			if photo.Details.NoNotes() && data.Comment != "" {
				photo.Details.Notes = data.Comment
			}

			if photo.Details.NoArtist() && data.Artist != "" {
				photo.Details.Artist = data.Artist
			}

			if photo.Details.NoCopyright() && data.Copyright != "" {
				photo.Details.Copyright = data.Copyright
			}
		}
	case m.IsRaw():
		if photo.PhotoType == entity.TypeImage {
			photo.PhotoType = entity.TypeRaw
		}
	case m.IsVideo():
		metaData = m.MetaData()

		file.FileCodec = metaData.Codec
		file.FileWidth = metaData.Width
		file.FileHeight = metaData.Height
		file.FileDuration = metaData.Duration
		file.FileAspectRatio = metaData.AspectRatio()
		file.FilePortrait = metaData.Portrait()

		if file.FileDuration == 0 || file.FileDuration > time.Millisecond*3100 {
			photo.PhotoType = entity.TypeVideo
		} else {
			photo.PhotoType = entity.TypeLive
		}

		if res := metaData.Megapixels(); res > photo.PhotoResolution {
			photo.PhotoResolution = res
		}

		if file.FileWidth == 0 && primaryFile.FileWidth > 0 {
			file.FileWidth = primaryFile.FileWidth
			file.FileHeight = primaryFile.FileHeight
			file.FileAspectRatio = primaryFile.FileAspectRatio
			file.FilePortrait = primaryFile.FilePortrait
		}

		if primaryFile.FileDiff > 0 {
			file.FileDiff = primaryFile.FileDiff
			file.FileMainColor = primaryFile.FileMainColor
			file.FileChroma = primaryFile.FileChroma
			file.FileLuminance = primaryFile.FileLuminance
			file.FileColors = primaryFile.FileColors
		}
	}

	// file obviously exists: remove deleted and missing flags
	file.DeletedAt = nil
	file.FileMissing = false
	file.FileError = ""

	// primary files are used for rendering thumbnails and image classification (plus sidecar files if they exist)
	if file.FilePrimary {
		primaryFile = file

		if !ind.conf.DisableTensorFlow() {
			// Image classification via TensorFlow.
			labels = ind.classifyImage(m)

			if !photoExists && ind.conf.Settings().Features.Private && ind.conf.DetectNSFW() {
				photo.PhotoPrivate = ind.NSFW(m)
			}
		}

		// read metadata from embedded Exif and JSON sidecar file (if exists)
		if metaData := m.MetaData(); metaData.Error == nil {
			photo.SetTitle(metaData.Title, entity.SrcMeta)
			photo.SetDescription(metaData.Description, entity.SrcMeta)
			photo.SetTakenAt(metaData.TakenAt, metaData.TakenAtLocal, metaData.TimeZone, entity.SrcMeta)
			photo.SetCoordinates(metaData.Lat, metaData.Lng, metaData.Altitude, entity.SrcMeta)

			if photo.Details.NoNotes() {
				photo.Details.Notes = metaData.Comment
			}

			if photo.Details.NoSubject() {
				photo.Details.Subject = metaData.Subject
			}

			if photo.Details.NoKeywords() {
				photo.Details.Keywords = metaData.Keywords
			}

			if photo.Details.NoArtist() && metaData.Artist != "" {
				photo.Details.Artist = metaData.Artist
			}

			if photo.Details.NoArtist() && metaData.CameraOwner != "" {
				photo.Details.Artist = metaData.CameraOwner
			}

			if photo.NoCameraSerial() {
				photo.CameraSerial = metaData.CameraSerial
			}

			if metaData.HasDocumentID() && photo.DocumentID == "" {
				log.Debugf("index: %s has document id %s", quotedName, txt.Quote(metaData.DocumentID))

				photo.DocumentID = metaData.DocumentID
			}

			if metaData.HasInstanceID() && file.InstanceID == "" {
				log.Debugf("index: %s has instance id %s", quotedName, txt.Quote(metaData.InstanceID))

				file.InstanceID = metaData.InstanceID
			}
		}

		if photo.CameraSrc == entity.SrcAuto {
			// Set UpdateCamera, Lens, Focal Length and F Number.
			photo.Camera = entity.FirstOrCreateCamera(entity.NewCamera(m.CameraModel(), m.CameraMake()))

			if photo.Camera != nil {
				photo.CameraID = photo.Camera.ID
			} else {
				photo.CameraID = entity.UnknownCamera.ID
			}

			photo.Lens = entity.FirstOrCreateLens(entity.NewLens(m.LensModel(), m.LensMake()))

			if photo.Lens != nil {
				photo.LensID = photo.Lens.ID
			} else {
				photo.LensID = entity.UnknownLens.ID
			}

			photo.PhotoFocalLength = m.FocalLength()
			photo.PhotoFNumber = m.FNumber()
			photo.PhotoIso = m.Iso()
			photo.PhotoExposure = m.Exposure()
		}

		if photo.TakenAt.IsZero() || photo.TakenAtLocal.IsZero() {
			takenUtc, takenSrc := m.TakenAt()
			photo.SetTakenAt(takenUtc, takenUtc, "", takenSrc)
		}

		var locLabels classify.Labels
		locKeywords, locLabels = photo.UpdateLocation(ind.conf.GeoCodingApi())
		labels = append(labels, locLabels...)
	}

	if photo.UnknownLocation() {
		photo.Location = &entity.UnknownLocation
		photo.LocationID = entity.UnknownLocation.ID
	}

	if photo.UnknownPlace() {
		photo.Place = &entity.UnknownPlace
		photo.PlaceID = entity.UnknownPlace.ID
	}

	photo.UpdateYearMonth()

	if originalName != "" {
		file.OriginalName = originalName
	}

	file.FileSidecar = m.IsSidecar()
	file.FileVideo = m.IsVideo()
	file.FileRoot = fileRoot
	file.FileName = fileName
	file.FileHash = fileHash
	file.FileSize = fileSize
	file.FileModified = fileModified
	file.FileType = string(m.FileType())
	file.FileMime = m.MimeType()
	file.FileOrientation = m.Orientation()

	if photoExists {
		if err := entity.UnscopedDb().Save(&photo).Error; err != nil {
			log.Errorf("index: %s for %s", err.Error(), quotedName)
			result.Status = IndexFailed
			result.Error = err
			return result
		}
	} else {
		if yamlName := fs.TypeYaml.FindSub(m.FileName(), fs.HiddenPath, ind.conf.Settings().Index.Group); yamlName != "" {
			if err := photo.LoadFromYaml(yamlName); err != nil {
				log.Errorf("index: %s (restore from yaml) for %s", err.Error(), quotedName)
			} else {
				log.Infof("index: restored from %s", txt.Quote(fs.RelativeName(yamlName, ind.originalsPath())))
			}
		} else {
			photo.PhotoFavorite = false
		}

		if err := entity.UnscopedDb().Create(&photo).Error; err != nil {
			log.Errorf("index: %s", err)
			result.Status = IndexFailed
			result.Error = err
			return result
		}

		event.Publish("count.photos", event.Data{
			"count": 1,
		})

		if photo.PhotoPrivate {
			event.Publish("count.private", event.Data{
				"count": 1,
			})
		}

		if photo.PhotoType == entity.TypeVideo {
			event.Publish("count.videos", event.Data{
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

	// Main JPEG file.
	if file.FilePrimary {
		labels := photo.ClassifyLabels()

		if err := photo.UpdateTitle(labels); err != nil {
			log.Warnf("%s for %s", err.Error(), quotedName)
		}

		w := txt.Keywords(photo.Details.Keywords)

		if NonCanonical(fileBase) {
			w = append(w, txt.FilenameKeywords(filePath)...)
			w = append(w, txt.FilenameKeywords(fileBase)...)
		}

		w = append(w, locKeywords...)
		w = append(w, txt.FilenameKeywords(file.OriginalName)...)
		w = append(w, file.FileMainColor)
		w = append(w, labels.Keywords()...)

		photo.Details.Keywords = strings.Join(txt.UniqueWords(w), ", ")

		if photo.Details.Keywords != "" {
			log.Debugf("index: set keywords %s for %s", photo.Details.Keywords, quotedName)
		} else {
			log.Debugf("index: no keywords for %s", quotedName)
		}

		photo.PhotoQuality = photo.QualityScore()

		if err := entity.UnscopedDb().Save(&photo).Error; err != nil {
			log.Errorf("index: %s for %s", err, quotedName)
			result.Status = IndexFailed
			result.Error = err
			return result
		}

		if err := photo.SyncKeywordLabels(); err != nil {
			log.Errorf("index: %s for %s", err, quotedName)
		}

		if err := photo.IndexKeywords(); err != nil {
			log.Errorf("index: %s for %s", err, quotedName)
		}
	} else {
		if photo.PhotoQuality >= 0 {
			photo.PhotoQuality = photo.QualityScore()
		}

		if err := entity.UnscopedDb().Unscoped().Save(&photo).Error; err != nil {
			log.Errorf("index: %s for %s", err, quotedName)
			result.Status = IndexFailed
			result.Error = err
			return result
		}
	}

	result.Status = IndexUpdated

	if fileQuery.Error == nil {
		file.UpdatedIn = int64(time.Since(start))

		if err := entity.UnscopedDb().Save(&file).Error; err != nil {
			log.Errorf("index: %s for %s", err, quotedName)
			result.Status = IndexFailed
			result.Error = err
			return result
		}
	} else {
		file.CreatedIn = int64(time.Since(start))

		if err := entity.UnscopedDb().Create(&file).Error; err != nil {
			log.Errorf("index: %s for %s", err, quotedName)
			result.Status = IndexFailed
			result.Error = err
			return result
		}

		event.Publish("count.files", event.Data{
			"count": 1,
		})

		result.Status = IndexAdded
	}

	if (photo.PhotoType == entity.TypeVideo || photo.PhotoType == entity.TypeLive) && file.FilePrimary {
		if err := file.UpdateVideoInfos(); err != nil {
			log.Errorf("index: %s for %s", err, quotedName)
		}
	}

	result.FileID = file.ID
	result.FileUID = file.FileUID

	downloadedAs := fileName

	if originalName != "" {
		downloadedAs = originalName
	}

	if err := query.SetDownloadFileID(downloadedAs, file.ID); err != nil {
		log.Errorf("index: %s for %s", err, quotedName)
	}

	// Write YAML sidecar file (optional).
	if file.FilePrimary && ind.conf.SidecarYaml() {
		yamlFile := photo.YamlFileName(ind.originalsPath(), ind.conf.SidecarHidden())

		if err := photo.SaveAsYaml(yamlFile); err != nil {
			log.Errorf("index: %s (update yaml) for %s", err.Error(), quotedName)
		} else {
			log.Infof("index: updated yaml file %s", txt.Quote(fs.RelativeName(yamlFile, ind.originalsPath())))
		}
	}

	return result
}

// NSFW returns true if media file might be offensive and detection is enabled.
func (ind *Index) NSFW(jpeg *MediaFile) bool {
	filename, err := jpeg.Thumbnail(ind.thumbPath(), "fit_720")

	if err != nil {
		log.Error(err)
		return false
	}

	if nsfwLabels, err := ind.nsfwDetector.File(filename); err != nil {
		log.Error(err)
		return false
	} else {
		if nsfwLabels.NSFW(nsfw.ThresholdHigh) {
			log.Warnf("index: %s might contain offensive content", txt.Quote(jpeg.RelativeName(ind.originalsPath())))
			return true
		}
	}

	return false
}

// classifyImage returns all matching labels for a media file.
func (ind *Index) classifyImage(jpeg *MediaFile) (results classify.Labels) {
	start := time.Now()

	var thumbs []string

	if jpeg.AspectRatio() == 1 {
		thumbs = []string{"tile_224"}
	} else {
		thumbs = []string{"tile_224", "left_224", "right_224"}
	}

	var labels classify.Labels

	for _, thumb := range thumbs {
		filename, err := jpeg.Thumbnail(ind.thumbPath(), thumb)

		if err != nil {
			log.Error(err)
			continue
		}

		imageLabels, err := ind.tensorFlow.File(filename)

		if err != nil {
			log.Error(err)
			continue
		}

		labels = append(labels, imageLabels...)
	}

	// Sort by priority and uncertainty
	sort.Sort(labels)

	var confidence int

	for _, label := range labels {
		if confidence == 0 {
			confidence = 100 - label.Uncertainty
		}

		if (100 - label.Uncertainty) > (confidence / 3) {
			results = append(results, label)
		}
	}

	elapsed := time.Since(start)

	log.Debugf("index: image classification took %s", elapsed)

	return results
}
