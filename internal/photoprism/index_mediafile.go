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
	Status    IndexStatus
	Error     error
	FileID    uint
	FileUUID  string
	PhotoID   uint
	PhotoUUID string
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

	photo := entity.Photo{}
	metaData := meta.Data{}
	description := entity.Description{}
	labels := classify.Labels{}

	fileBase := m.Base(ind.conf.Settings().Index.Group)
	filePath := m.RelativePath(ind.originalsPath())
	fileName := m.RelativeName(ind.originalsPath())
	fileHash := ""
	fileSize, fileModified := m.Stat()
	fileChanged := true
	fileExists := false
	photoExists := false

	event.Publish("index.indexing", event.Data{
		"fileHash": fileHash,
		"fileSize": fileSize,
		"fileName": fileName,
		"baseName": filepath.Base(fileName),
	})

	fileQuery = ind.db.Unscoped().First(&file, "file_name = ?", fileName)
	fileExists = fileQuery.Error == nil

	if !fileExists && !m.IsSidecar() {
		fileHash = m.Hash()
		fileQuery = ind.db.Unscoped().First(&file, "file_hash = ?", fileHash)
		fileExists = fileQuery.Error == nil

		if fileExists && fs.FileExists(filepath.Join(ind.conf.OriginalsPath(), file.FileName)) {
			result.Status = IndexDuplicate
			return result
		}
	}

	if !fileExists {
		photoQuery = ind.db.Unscoped().First(&photo, "photo_path = ? AND photo_name = ?", filePath, fileBase)

		if photoQuery.Error != nil && m.HasTimeAndPlace() {
			metaData, _ = m.MetaData()
			photoQuery = ind.db.Unscoped().First(&photo, "photo_lat = ? AND photo_lng = ? AND taken_at = ?", metaData.Lat, metaData.Lng, metaData.TakenAt)
		}
	} else {
		photoQuery = ind.db.Unscoped().First(&photo, "id = ?", file.PhotoID)

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
		ind.db.Model(&photo).Related(&description)
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
			if q := ind.db.Where("file_type = 'jpg' AND file_primary = 1 AND photo_id = ?", photo.ID).First(&primaryFile); q.Error != nil {
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

	if m.IsVideo() {
		photo.PhotoVideo = true
		metaData, _ = m.MetaData()

		file.FileCodec = metaData.Codec
		file.FileWidth = metaData.Width
		file.FileHeight = metaData.Height
		file.FileDuration = metaData.Duration
		file.FileAspectRatio = metaData.AspectRatio()
		file.FilePortrait = metaData.Portrait()

		if res := metaData.Megapixels(); res > photo.PhotoResolution {
			photo.PhotoResolution = res
		}

		if file.FileWidth == 0 && primaryFile.FileWidth > 0 {
			file.FileWidth = primaryFile.FileWidth
			file.FileHeight = primaryFile.FileHeight
			file.FileAspectRatio = primaryFile.FileAspectRatio
			file.FilePortrait = primaryFile.FilePortrait
		}

		file.FileDiff = primaryFile.FileDiff
		file.FileMainColor = primaryFile.FileMainColor
		file.FileChroma = primaryFile.FileChroma
		file.FileLuminance = primaryFile.FileLuminance
		file.FileColors = primaryFile.FileColors
	}

	// file obviously exists: remove deleted and missing flags
	file.DeletedAt = nil
	file.FileMissing = false

	// primary files are used for rendering thumbnails and image classification (plus sidecar files if they exist)
	if file.FilePrimary {
		primaryFile = file

		if !ind.conf.DisableTensorFlow() && (fileChanged || o.Rescan) {
			// Image classification via TensorFlow
			labels = ind.classifyImage(m)

			if !photoExists && ind.conf.Settings().Features.Private && ind.conf.DetectNSFW() {
				photo.PhotoPrivate = ind.NSFW(m)
			}
		}

		if fileChanged || o.Rescan {
			// read metadata from embedded Exif and JSON sidecar file (if exists)
			if metaData, err := m.MetaData(); err == nil {
				photo.SetTitle(metaData.Title, entity.SrcMeta)
				photo.SetDescription(metaData.Description, entity.SrcMeta)
				photo.SetTakenAt(metaData.TakenAt, metaData.TakenAtLocal, metaData.TimeZone, entity.SrcMeta)
				photo.SetCoordinates(metaData.Lat, metaData.Lng, metaData.Altitude, entity.SrcMeta)

				if photo.Description.NoNotes() {
					photo.Description.PhotoNotes = metaData.Comment
				}

				if photo.Description.NoSubject() {
					photo.Description.PhotoSubject = metaData.Subject
				}

				if photo.Description.NoKeywords() {
					photo.Description.PhotoKeywords = metaData.Keywords
				}

				if photo.Description.NoArtist() && metaData.Artist != "" {
					photo.Description.PhotoArtist = metaData.Artist
				}

				if photo.Description.NoArtist() && metaData.CameraOwner != "" {
					photo.Description.PhotoArtist = metaData.CameraOwner
				}

				if photo.NoCameraSerial() {
					photo.CameraSerial = metaData.CameraSerial
				}

				if len(metaData.UniqueID) > 15 {
					log.Debugf("index: file uuid %s", txt.Quote(metaData.UniqueID))

					file.FileUUID = metaData.UniqueID
				}
			}
		}

		if photo.CameraSrc == entity.SrcAuto && (fileChanged || o.Rescan) {
			// Set UpdateCamera, Lens, Focal Length and F Number
			photo.Camera = entity.NewCamera(m.CameraModel(), m.CameraMake()).FirstOrCreate()
			photo.Lens = entity.NewLens(m.LensModel(), m.LensMake()).FirstOrCreate()
			photo.PhotoFocalLength = m.FocalLength()
			photo.PhotoFNumber = m.FNumber()
			photo.PhotoIso = m.Iso()
			photo.PhotoExposure = m.Exposure()
		}

		if photo.TakenAt.IsZero() || photo.TakenAtLocal.IsZero() {
			photo.SetTakenAt(m.DateCreated(), m.DateCreated(), "", entity.SrcAuto)
		}

		if fileChanged || o.Rescan || photo.NoTitle() {
			if photo.HasLatLng() {
				var locLabels classify.Labels
				locKeywords, locLabels = photo.UpdateLocation(ind.conf.GeoCodingApi())
				labels = append(labels, locLabels...)
			} else {
				log.Info("index: no latitude and longitude in metadata")

				photo.Place = &entity.UnknownPlace
				photo.PlaceID = entity.UnknownPlace.ID
			}
		}
	} else if m.IsXMP() {
		// TODO: Proof-of-concept for indexing XMP sidecar files
		if data, err := meta.XMP(m.FileName()); err == nil {
			photo.SetTitle(data.Title, entity.SrcXmp)
			photo.SetDescription(data.Description, entity.SrcXmp)

			if photo.Description.NoNotes() && data.Comment != "" {
				photo.Description.PhotoNotes = data.Comment
			}

			if photo.Description.NoArtist() && data.Artist != "" {
				photo.Description.PhotoArtist = data.Artist
			}

			if photo.Description.NoCopyright() && data.Copyright != "" {
				photo.Description.PhotoCopyright = data.Copyright
			}
		}
	}

	if len(photo.PlaceID) < 2 {
		photo.Place = &entity.UnknownPlace
		photo.PlaceID = entity.UnknownPlace.ID
		photo.PhotoCountry = entity.UnknownPlace.CountryCode()
	}

	photo.UpdateYearMonth()

	if originalName != "" {
		file.OriginalName = originalName
	}

	file.FileSidecar = m.IsSidecar()
	file.FileVideo = m.IsVideo()
	file.FileName = fileName
	file.FileHash = fileHash
	file.FileSize = fileSize
	file.FileModified = fileModified
	file.FileType = string(m.FileType())
	file.FileMime = m.MimeType()
	file.FileOrientation = m.Orientation()

	if m.IsJpeg() && (fileChanged || o.Rescan) {
		// Color information
		if p, err := m.Colors(ind.thumbPath()); err != nil {
			log.Errorf("index: %s", err.Error())
		} else {
			file.FileMainColor = p.MainColor.Name()
			file.FileColors = p.Colors.Hex()
			file.FileLuminance = p.Luminance.Hex()
			file.FileDiff = p.Luminance.Diff()
			file.FileChroma = p.Chroma.Value()
		}
	}

	if m.IsJpeg() && (fileChanged || o.Rescan) {
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
	}

	if photoExists {
		// Estimate location
		if o.Rescan && photo.NoLocation() {
			ind.estimateLocation(&photo)
		}

		if err := ind.db.Unscoped().Save(&photo).Error; err != nil {
			log.Errorf("index: %s", err)
			result.Status = IndexFailed
			result.Error = err
			return result
		}
	} else {
		photo.PhotoFavorite = false

		if err := ind.db.Create(&photo).Error; err != nil {
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

		if photo.PhotoVideo {
			event.Publish("count.videos", event.Data{
				"count": 1,
			})
		}

		event.EntitiesCreated("photos", []entity.Photo{photo})
	}

	photo.AddLabels(labels)

	file.PhotoID = photo.ID
	result.PhotoID = photo.ID

	file.PhotoUUID = photo.PhotoUUID
	result.PhotoUUID = photo.PhotoUUID

	if file.FilePrimary && (fileChanged || o.Rescan) {
		labels := photo.ClassifyLabels()

		if err := photo.UpdateTitle(labels); err != nil {
			log.Warnf("%s (%s)", err.Error(), photo.PhotoUUID)
		}

		w := txt.Keywords(photo.Description.PhotoKeywords)

		if NonCanonical(fileBase) {
			w = append(w, txt.FilenameKeywords(filePath)...)
			w = append(w, txt.FilenameKeywords(fileBase)...)
		}

		w = append(w, locKeywords...)
		w = append(w, txt.FilenameKeywords(file.OriginalName)...)
		w = append(w, file.FileMainColor)
		w = append(w, labels.Keywords()...)

		photo.Description.PhotoKeywords = strings.Join(txt.UniqueWords(w), ", ")

		if photo.Description.PhotoKeywords != "" {
			log.Debugf("index: updated photo keywords (%s)", photo.Description.PhotoKeywords)
		} else {
			log.Debug("index: no photo keywords")
		}

		photo.PhotoQuality = photo.QualityScore()

		if err := ind.db.Unscoped().Save(&photo).Error; err != nil {
			log.Errorf("index: %s", err)
			result.Status = IndexFailed
			result.Error = err
			return result
		}

		if err := photo.IndexKeywords(); err != nil {
			log.Warnf("%s (%s)", err.Error(), photo.PhotoUUID)
		}
	} else {
		if photo.PhotoQuality >= 0 {
			photo.PhotoQuality = photo.QualityScore()
		}

		if err := ind.db.Unscoped().Save(&photo).Error; err != nil {
			log.Errorf("index: %s", err)
			result.Status = IndexFailed
			result.Error = err
			return result
		}
	}

	result.Status = IndexUpdated

	if fileQuery.Error == nil {
		file.UpdatedIn = int64(time.Since(start))

		if err := ind.db.Unscoped().Save(&file).Error; err != nil {
			log.Errorf("index: %s", err)
			result.Status = IndexFailed
			result.Error = err
			return result
		}
	} else {
		file.CreatedIn = int64(time.Since(start))

		if err := ind.db.Create(&file).Error; err != nil {
			log.Errorf("index: %s", err)
			result.Status = IndexFailed
			result.Error = err
			return result
		}

		result.Status = IndexAdded
	}

	if photo.PhotoVideo && file.FilePrimary {
		if err := file.UpdateVideoInfos(); err != nil {
			log.Errorf("index: %s", err)
		}
	}

	result.FileID = file.ID
	result.FileUUID = file.FileUUID

	downloadedAs := fileName

	if originalName != "" {
		downloadedAs = originalName
	}

	if err := query.SetDownloadFileID(downloadedAs, file.ID); err != nil {
		log.Errorf("index: %s", err)
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
			log.Warnf("index: %s might contain offensive content", jpeg.RelativeName(ind.originalsPath()))
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
