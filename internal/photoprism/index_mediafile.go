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
	IndexStacked   IndexStatus = "stacked"
	IndexSkipped   IndexStatus = "skipped"
	IndexDuplicate IndexStatus = "skipped duplicate"
	IndexArchived  IndexStatus = "skipped archived"
	IndexFailed    IndexStatus = "failed"
)

type IndexStatus string

type IndexResult struct {
	Status   IndexStatus
	Err      error
	FileID   uint
	FileUID  string
	PhotoID  uint
	PhotoUID string
}

func (r IndexResult) String() string {
	return string(r.Status)
}

func (r IndexResult) Failed() bool {
	return r.Err != nil
}

func (r IndexResult) Success() bool {
	return r.Err == nil && r.FileID > 0
}

func (r IndexResult) Indexed() bool {
	return r.Status == IndexAdded || r.Status == IndexUpdated || r.Status == IndexStacked
}

func (r IndexResult) Stacked() bool {
	return r.Status == IndexStacked
}

func (ind *Index) MediaFile(m *MediaFile, o IndexOptions, originalName string) (result IndexResult) {
	if m == nil {
		err := errors.New("index: media file is nil - you might have found a bug")
		log.Error(err)
		result.Err = err
		result.Status = IndexFailed
		return result
	}

	start := time.Now()

	var photoQuery, fileQuery *gorm.DB
	var locKeywords []string

	file, primaryFile := entity.File{}, entity.File{}

	photo := entity.NewPhoto()
	metaData := meta.NewData()
	labels := classify.Labels{}

	fileRoot, fileBase, filePath, fileName := m.PathNameInfo()

	logName := txt.Quote(fileName)
	fileSize, fileModified := m.Stat()

	fileHash := ""
	fileChanged := true
	fileExists := false
	fileGrouped := false

	photoExists := false

	stripSequence := Config().Settings().Index.Sequences

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

		if fileExists && fs.FileExists(FileName(file.FileRoot, file.FileName)) {
			result.Status = IndexDuplicate
			return result
		}

		if !fileExists && m.MetaData().HasInstanceID() {
			fileQuery = entity.UnscopedDb().First(&file, "uuid = ?", m.MetaData().InstanceID)
			fileExists = fileQuery.Error == nil
		}
	}

	if !fileExists {
		photoQuery = entity.UnscopedDb().First(&photo, "photo_path = ? AND photo_name = ?", filePath, fileBase)

		if photoQuery.Error != nil && m.MetaData().HasTimeAndPlace() {
			metaData = m.MetaData()
			photoQuery = entity.UnscopedDb().First(&photo, "photo_lat = ? AND photo_lng = ? AND taken_at = ?", metaData.Lat, metaData.Lng, metaData.TakenAt)

			if photoQuery.Error == nil {
				fileGrouped = true
			}
		}

		if photoQuery.Error != nil && m.MetaData().HasDocumentID() {
			photoQuery = entity.UnscopedDb().First(&photo, "uuid = ?", m.MetaData().DocumentID)

			if photoQuery.Error == nil {
				fileGrouped = true
			}
		}
	} else {
		photoQuery = entity.UnscopedDb().First(&photo, "id = ?", file.PhotoID)

		fileChanged = file.Changed(fileSize, fileModified)

		if fileChanged {
			log.Debugf("index: file was modified (new size %d, old size %d, new date %s, old date %s)", fileSize, file.FileSize, fileModified, file.FileModified)
		}
	}

	photoExists = photoQuery.Error == nil

	if !fileChanged && photoExists && o.SkipUnchanged() || !photoExists && m.IsSidecar() {
		result.Status = IndexSkipped
		return result
	}

	details := photo.GetDetails()

	if !photoExists {
		photo.PhotoQuality = -1

		if yamlName := fs.TypeYaml.FindFirst(m.FileName(), []string{Config().SidecarPath(), fs.HiddenPath}, Config().OriginalsPath(), stripSequence); yamlName != "" {
			if err := photo.LoadFromYaml(yamlName); err != nil {
				log.Errorf("index: %s in %s (restore from yaml)", err.Error(), logName)
			} else if err := photo.Find(); err != nil {
				log.Infof("index: data restored from %s", txt.Quote(fs.RelName(yamlName, Config().OriginalsPath())))
			} else {
				photoExists = true
				log.Infof("index: uid %s restored from %s", photo.PhotoUID, txt.Quote(fs.RelName(yamlName, Config().OriginalsPath())))
			}
		}
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

	if originalName != "" {
		file.OriginalName = originalName

		if file.FilePrimary && photo.OriginalName == "" {
			photo.OriginalName = fs.BasePrefix(originalName, stripSequence)
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

	// Handle file types.
	switch {
	case m.IsJpeg():
		// Color information
		if p, err := m.Colors(Config().ThumbPath()); err != nil {
			log.Errorf("index: %s in %s (detect colors)", err.Error(), logName)
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

		if metaData := m.MetaData(); metaData.Error == nil {
			file.FileProjection = metaData.Projection
		}
	case m.IsXMP():
		// TODO: Proof-of-concept for indexing XMP sidecar files
		if data, err := meta.XMP(m.FileName()); err == nil {
			photo.SetTitle(data.Title, entity.SrcXmp)
			photo.SetDescription(data.Description, entity.SrcXmp)

			if details.NoNotes() && data.Comment != "" {
				details.Notes = data.Comment
			}

			if details.NoArtist() && data.Artist != "" {
				details.Artist = data.Artist
			}

			if details.NoCopyright() && data.Copyright != "" {
				details.Copyright = data.Copyright
			}
		}
	case m.IsRaw(), m.IsHEIF(), m.IsImageOther():
		if metaData := m.MetaData(); metaData.Error == nil {
			photo.SetTitle(metaData.Title, entity.SrcMeta)
			photo.SetDescription(metaData.Description, entity.SrcMeta)
			photo.SetTakenAt(metaData.TakenAt, metaData.TakenAtLocal, metaData.TimeZone, entity.SrcMeta)
			photo.SetCoordinates(metaData.Lat, metaData.Lng, metaData.Altitude, entity.SrcMeta)

			if details.NoNotes() {
				details.Notes = metaData.Comment
			}

			if details.NoSubject() {
				details.Subject = metaData.Subject
			}

			if details.NoKeywords() {
				details.Keywords = metaData.Keywords
			}

			if details.NoArtist() && metaData.Artist != "" {
				details.Artist = metaData.Artist
			}

			if details.NoArtist() && metaData.CameraOwner != "" {
				details.Artist = metaData.CameraOwner
			}

			if photo.NoCameraSerial() {
				photo.CameraSerial = metaData.CameraSerial
			}

			if metaData.HasDocumentID() && photo.UUID == "" {
				log.Debugf("index: %s has document id %s", logName, txt.Quote(metaData.DocumentID))

				photo.UUID = metaData.DocumentID
			}

			if metaData.HasInstanceID() && file.UUID == "" {
				log.Debugf("index: %s has instance id %s", logName, txt.Quote(metaData.InstanceID))

				file.UUID = metaData.InstanceID
			}

			file.FileCodec = metaData.Codec
			file.FileWidth = metaData.ActualWidth()
			file.FileHeight = metaData.ActualHeight()
			file.FileAspectRatio = metaData.AspectRatio()
			file.FilePortrait = metaData.Portrait()
			file.FileProjection = metaData.Projection

			if res := metaData.Megapixels(); res > photo.PhotoResolution {
				photo.PhotoResolution = res
			}
		}

		if photo.TypeSrc == entity.SrcAuto {
			// Update photo type only if not manually modified.
			if m.IsRaw() && photo.PhotoType == entity.TypeImage {
				photo.PhotoType = entity.TypeRaw
			}
		}
	case m.IsVideo():
		if metaData := m.MetaData(); metaData.Error == nil {
			photo.SetTitle(metaData.Title, entity.SrcMeta)
			photo.SetDescription(metaData.Description, entity.SrcMeta)
			photo.SetTakenAt(metaData.TakenAt, metaData.TakenAtLocal, metaData.TimeZone, entity.SrcMeta)
			photo.SetCoordinates(metaData.Lat, metaData.Lng, metaData.Altitude, entity.SrcMeta)

			if details.NoNotes() {
				details.Notes = metaData.Comment
			}

			if details.NoSubject() {
				details.Subject = metaData.Subject
			}

			if details.NoKeywords() {
				details.Keywords = metaData.Keywords
			}

			if details.NoArtist() && metaData.Artist != "" {
				details.Artist = metaData.Artist
			}

			if details.NoArtist() && metaData.CameraOwner != "" {
				details.Artist = metaData.CameraOwner
			}

			if photo.NoCameraSerial() {
				photo.CameraSerial = metaData.CameraSerial
			}

			if metaData.HasDocumentID() && photo.UUID == "" {
				log.Debugf("index: %s has document id %s", logName, txt.Quote(metaData.DocumentID))

				photo.UUID = metaData.DocumentID
			}

			if metaData.HasInstanceID() && file.UUID == "" {
				log.Debugf("index: %s has instance id %s", logName, txt.Quote(metaData.InstanceID))

				file.UUID = metaData.InstanceID
			}

			file.FileCodec = metaData.Codec
			file.FileWidth = metaData.ActualWidth()
			file.FileHeight = metaData.ActualHeight()
			file.FileDuration = metaData.Duration
			file.FileAspectRatio = metaData.AspectRatio()
			file.FilePortrait = metaData.Portrait()
			file.FileProjection = metaData.Projection

			if res := metaData.Megapixels(); res > photo.PhotoResolution {
				photo.PhotoResolution = res
			}
		}

		if photo.TypeSrc == entity.SrcAuto {
			// Update photo type only if not manually modified.
			if file.FileDuration == 0 || file.FileDuration > time.Millisecond*3100 {
				photo.PhotoType = entity.TypeVideo
			} else {
				photo.PhotoType = entity.TypeLive
			}
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

		if !Config().TensorFlowOff() {
			// Image classification via TensorFlow.
			labels = ind.classifyImage(m)

			if !photoExists && Config().Settings().Features.Private && Config().DetectNSFW() {
				photo.PhotoPrivate = ind.NSFW(m)
			}
		}

		// read metadata from embedded Exif and JSON sidecar file (if exists)
		if metaData := m.MetaData(); metaData.Error == nil {
			photo.SetTitle(metaData.Title, entity.SrcMeta)
			photo.SetDescription(metaData.Description, entity.SrcMeta)
			photo.SetTakenAt(metaData.TakenAt, metaData.TakenAtLocal, metaData.TimeZone, entity.SrcMeta)
			photo.SetCoordinates(metaData.Lat, metaData.Lng, metaData.Altitude, entity.SrcMeta)

			if details.NoNotes() {
				details.Notes = metaData.Comment
			}

			if details.NoSubject() {
				details.Subject = metaData.Subject
			}

			if details.NoKeywords() {
				details.Keywords = metaData.Keywords
			}

			if details.NoArtist() && metaData.Artist != "" {
				details.Artist = metaData.Artist
			}

			if details.NoArtist() && metaData.CameraOwner != "" {
				details.Artist = metaData.CameraOwner
			}

			if photo.NoCameraSerial() {
				photo.CameraSerial = metaData.CameraSerial
			}

			if metaData.HasDocumentID() && photo.UUID == "" {
				log.Debugf("index: %s has document id %s", logName, txt.Quote(metaData.DocumentID))

				photo.UUID = metaData.DocumentID
			}

			if metaData.HasInstanceID() && file.UUID == "" {
				log.Debugf("index: %s has instance id %s", logName, txt.Quote(metaData.InstanceID))

				file.UUID = metaData.InstanceID
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

		if photo.TakenSrc == entity.SrcAuto {
			takenUtc, takenSrc := m.TakenAt()
			photo.SetTakenAt(takenUtc, takenUtc, "", takenSrc)
		}

		var locLabels classify.Labels
		locKeywords, locLabels = photo.UpdateLocation(ind.conf.GeoCodingApi())
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

	if file.Panorama() {
		photo.PhotoPanorama = true
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
		if err := photo.Save(); err != nil {
			log.Errorf("index: %s in %s (save photo)", err.Error(), logName)
			result.Status = IndexFailed
			result.Err = err
			return result
		}
	} else {
		if err := photo.FirstOrCreate(); err != nil {
			log.Errorf("index: %s in %s (create photo)", err.Error(), logName)
			result.Status = IndexFailed
			result.Err = err
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
			log.Debugf("%s in %s (update title)", err.Error(), logName)
		}

		w := txt.Keywords(details.Keywords)

		if !fs.IsGenerated(fileBase) {
			w = append(w, txt.FilenameKeywords(filePath)...)
			w = append(w, txt.FilenameKeywords(fileBase)...)
		}

		w = append(w, locKeywords...)
		w = append(w, txt.FilenameKeywords(file.OriginalName)...)
		w = append(w, file.FileMainColor)
		w = append(w, labels.Keywords()...)

		details.Keywords = strings.Join(txt.UniqueWords(w), ", ")

		if details.Keywords != "" {
			log.Tracef("index: set keywords %s for %s", details.Keywords, logName)
		} else {
			log.Tracef("index: no keywords for %s", logName)
		}

		photo.PhotoQuality = photo.QualityScore()

		if err := photo.Save(); err != nil {
			log.Errorf("index: %s in %s (save photo)", err, logName)
			result.Status = IndexFailed
			result.Err = err
			return result
		}

		if err := photo.SyncKeywordLabels(); err != nil {
			log.Errorf("index: %s in %s (sync keywords and labels)", err, logName)
		}

		if err := photo.IndexKeywords(); err != nil {
			log.Errorf("index: %s in %s (save keywords)", err, logName)
		}
	} else {
		if photo.PhotoQuality >= 0 {
			photo.PhotoQuality = photo.QualityScore()
		}

		if err := photo.Save(); err != nil {
			log.Errorf("index: %s in %s (save photo)", err, logName)
			result.Status = IndexFailed
			result.Err = err
			return result
		}
	}

	result.Status = IndexUpdated

	if fileQuery.Error == nil {
		file.UpdatedIn = int64(time.Since(start))

		if err := file.Save(); err != nil {
			log.Errorf("index: %s in %s (save file)", err, logName)
			result.Status = IndexFailed
			result.Err = err
			return result
		}
	} else {
		file.CreatedIn = int64(time.Since(start))

		if err := file.Create(); err != nil {
			log.Errorf("index: %s in %s (create file)", err, logName)
			result.Status = IndexFailed
			result.Err = err
			return result
		}

		event.Publish("count.files", event.Data{
			"count": 1,
		})

		if fileGrouped {
			result.Status = IndexStacked
		} else {
			result.Status = IndexAdded
		}
	}

	if (photo.PhotoType == entity.TypeVideo || photo.PhotoType == entity.TypeLive) && file.FilePrimary {
		if err := file.UpdateVideoInfos(); err != nil {
			log.Errorf("index: %s in %s (update video infos)", err, logName)
		}
	}

	result.FileID = file.ID
	result.FileUID = file.FileUID

	downloadedAs := fileName

	if originalName != "" {
		downloadedAs = originalName
	}

	if err := query.SetDownloadFileID(downloadedAs, file.ID); err != nil {
		log.Errorf("index: %s in %s (set download id)", err, logName)
	}

	// Write YAML sidecar file (optional).
	if file.FilePrimary && Config().SidecarYaml() {
		yamlFile := photo.YamlFileName(Config().OriginalsPath(), Config().SidecarPath())

		if err := photo.SaveAsYaml(yamlFile); err != nil {
			log.Errorf("index: %s in %s (update yaml)", err.Error(), logName)
		} else {
			log.Infof("index: updated yaml file %s", txt.Quote(fs.RelName(yamlFile, Config().OriginalsPath())))
		}
	}

	return result
}

// NSFW returns true if media file might be offensive and detection is enabled.
func (ind *Index) NSFW(jpeg *MediaFile) bool {
	filename, err := jpeg.Thumbnail(Config().ThumbPath(), "fit_720")

	if err != nil {
		log.Error(err)
		return false
	}

	if nsfwLabels, err := ind.nsfwDetector.File(filename); err != nil {
		log.Error(err)
		return false
	} else {
		if nsfwLabels.NSFW(nsfw.ThresholdHigh) {
			log.Warnf("index: %s might contain offensive content", txt.Quote(jpeg.RelName(Config().OriginalsPath())))
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
		filename, err := jpeg.Thumbnail(Config().ThumbPath(), thumb)

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
