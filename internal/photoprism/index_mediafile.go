package photoprism

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/classify"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/meta"
	"github.com/photoprism/photoprism/pkg/txt"
)

const (
	indexResultUpdated IndexResult = "updated"
	indexResultAdded   IndexResult = "added"
	indexResultSkipped IndexResult = "skipped"
	indexResultFailed  IndexResult = "failed"
)

type IndexResult string

func (ind *Index) MediaFile(m *MediaFile, o IndexOptions) IndexResult {
	if m == nil {
		log.Error("index: media file is nil - you might have found a bug")
		return indexResultFailed
	}

	start := time.Now()

	var photo entity.Photo
	var file, primaryFile entity.File
	var metaData meta.Data
	var photoQuery, fileQuery *gorm.DB
	var keywords []string

	labels := classify.Labels{}
	fileBase := m.Basename()
	filePath := m.RelativePath(ind.originalsPath())
	fileName := m.RelativeFilename(ind.originalsPath())
	fileHash := m.Hash()
	fileChanged := true
	fileExists := false
	photoExists := false

	event.Publish("index.indexing", event.Data{
		"fileHash": fileHash,
		"fileName": fileName,
		"baseName": filepath.Base(fileName),
	})

	fileQuery = ind.db.Unscoped().First(&file, "file_hash = ? OR file_name = ?", fileHash, fileName)
	fileExists = fileQuery.Error == nil

	if !fileExists {
		photoQuery = ind.db.Unscoped().First(&photo, "photo_path = ? AND photo_name = ?", filePath, fileBase)

		if photoQuery.Error != nil && m.HasTimeAndPlace() {
			metaData, _ = m.MetaData()
			photoQuery = ind.db.Unscoped().First(&photo, "photo_lat = ? AND photo_lng = ? AND taken_at = ?", metaData.Lat, metaData.Lng, metaData.TakenAt)
		}
	} else {
		photoQuery = ind.db.Unscoped().First(&photo, "id = ?", file.PhotoID)
		fileChanged = file.FileHash != fileHash
	}

	photoExists = photoQuery.Error == nil

	if !fileChanged && photoExists && o.SkipUnchanged() {
		return indexResultSkipped
	}

	if !file.FilePrimary {
		if photoExists {
			if q := ind.db.Where("file_type = 'jpg' AND file_primary = 1 AND photo_id = ?", photo.ID).First(&primaryFile); q.Error != nil {
				file.FilePrimary = m.IsJpeg()
			}
		} else {
			file.FilePrimary = m.IsJpeg()
		}
	}

	if file.FilePrimary {
		primaryFile = file
	}

	photo.PhotoPath = filePath
	photo.PhotoName = fileBase

	if file.FilePrimary {
		if !ind.conf.TensorFlowDisabled() && (fileChanged || o.UpdateKeywords || o.UpdateLabels || o.UpdateTitle) {
			// Image classification via TensorFlow
			labels = ind.classifyImage(m)
			photo.PhotoNSFW = ind.isNSFW(m)
		}

		if fileChanged || o.UpdateExif {
			// Read UpdateExif data
			if metaData, err := m.MetaData(); err == nil {
				photo.PhotoLat = metaData.Lat
				photo.PhotoLng = metaData.Lng
				photo.TakenAt = metaData.TakenAt
				photo.TakenAtLocal = metaData.TakenAtLocal
				photo.TimeZone = metaData.TimeZone
				photo.PhotoAltitude = metaData.Altitude
				photo.PhotoArtist = metaData.Artist

				if len(metaData.UniqueID) > 15 {
					log.Debugf("index: file uuid \"%s\"", metaData.UniqueID)

					file.FileUUID = metaData.UniqueID
				}
			}
		}

		if fileChanged || o.UpdateCamera {
			// Set UpdateCamera, Lens, Focal Length and F Number
			photo.Camera = entity.NewCamera(m.CameraModel(), m.CameraMake()).FirstOrCreate(ind.db)
			photo.Lens = entity.NewLens(m.LensModel(), m.LensMake()).FirstOrCreate(ind.db)
			photo.PhotoFocalLength = m.FocalLength()
			photo.PhotoFNumber = m.FNumber()
			photo.PhotoIso = m.Iso()
			photo.PhotoExposure = m.Exposure()
		}

		if fileChanged || o.UpdateKeywords || o.UpdateLocation || o.UpdateTitle {
			locKeywords, locLabels := ind.indexLocation(m, &photo, labels, fileChanged, o)
			keywords = append(keywords, locKeywords...)
			labels = append(labels, locLabels...)
		}

		if photo.NoTitle() || (fileChanged || o.UpdateTitle) && photo.ModifiedTitle == false && photo.NoLocation() {
			if len(labels) > 0 && labels[0].Priority >= -1 && labels[0].Uncertainty <= 85 && labels[0].Name != "" {
				photo.PhotoTitle = fmt.Sprintf("%s / %s", txt.Title(labels[0].Name), m.DateCreated().Format("2006"))
			} else if !photo.TakenAtLocal.IsZero() {
				var daytimeString string
				hour := photo.TakenAtLocal.Hour()

				switch {
				case hour < 17:
					daytimeString = "Unknown"
				case hour < 20:
					daytimeString = "Sunset"
				default:
					daytimeString = "Unknown"
				}

				photo.PhotoTitle = fmt.Sprintf("%s / %s", daytimeString, photo.TakenAtLocal.Format("2006"))
			} else {
				photo.PhotoTitle = "Unknown"
			}

			log.Infof("index: changed empty photo title to \"%s\"", photo.PhotoTitle)
		}

		if photo.TakenAt.IsZero() || photo.TakenAtLocal.IsZero() {
			photo.TakenAt = m.DateCreated()
			photo.TakenAtLocal = photo.TakenAt
		}
	} else if m.IsXMP() {
		// TODO: Proof-of-concept for indexing XMP sidecar files
		if data, err := meta.XMP(m.Filename()); err == nil {
			if data.Title != "" && !photo.ModifiedTitle {
				photo.PhotoTitle = data.Title
			}

			if data.Copyright != "" {
				photo.PhotoCopyright = data.Copyright
			}

			if data.Artist != "" {
				photo.PhotoArtist = data.Artist
			}

			if data.Description != "" {
				photo.PhotoDescription = data.Description
			}
		}
	}

	photo.PhotoYear = photo.TakenAt.Year()
	photo.PhotoMonth = int(photo.TakenAt.Month())

	if photoExists {
		// Estimate location
		if o.UpdateLocation && photo.NoLocation() {
			ind.estimateLocation(&photo)
		}

		if err := ind.db.Unscoped().Save(&photo).Error; err != nil {
			log.Errorf("index: %s", err)
			return indexResultFailed
		}
	} else {
		event.Publish("count.photos", event.Data{
			"count": 1,
		})

		photo.PhotoFavorite = false

		if err := ind.db.Create(&photo).Error; err != nil {
			log.Errorf("index: %s", err)
			return indexResultFailed
		}
	}

	if len(labels) > 0 {
		log.Infof("index: adding labels %+v", labels)
		ind.addLabels(photo.ID, labels)
	}

	file.PhotoID = photo.ID
	file.PhotoUUID = photo.PhotoUUID
	file.FileSidecar = m.IsSidecar()
	file.FileVideo = m.IsVideo()
	file.FileMissing = false
	file.FileName = fileName
	file.FileHash = fileHash
	file.FileType = string(m.Type())
	file.FileMime = m.MimeType()
	file.FileOrientation = m.Orientation()

	if m.IsJpeg() && (fileChanged || o.UpdateColors) {
		// Color information
		if p, err := m.Colors(ind.thumbnailsPath()); err == nil {
			file.FileMainColor = p.MainColor.Name()
			file.FileColors = p.Colors.Hex()
			file.FileLuminance = p.Luminance.Hex()
			file.FileChroma = p.Chroma.Uint()
		}
	}

	if m.IsJpeg() && (fileChanged || o.UpdateSize) {
		if m.Width() > 0 && m.Height() > 0 {
			file.FileWidth = m.Width()
			file.FileHeight = m.Height()
			file.FileAspectRatio = m.AspectRatio()
			file.FilePortrait = m.Width() < m.Height()
		}
	}

	if file.FilePrimary && (fileChanged || o.UpdateKeywords || o.UpdateTitle) {
		keywords = append(keywords, file.FileMainColor)
		keywords = append(keywords, labels.Keywords()...)
		photo.IndexKeywords(keywords, ind.db)
	}

	if fileQuery.Error == nil {
		file.UpdatedIn = int64(time.Since(start))

		if err := ind.db.Unscoped().Save(&file).Error; err != nil {
			log.Errorf("index: %s", err)
			return indexResultFailed
		}

		return indexResultUpdated
	}

	file.CreatedIn = int64(time.Since(start))

	if err := ind.db.Create(&file).Error; err != nil {
		log.Errorf("index: %s", err)
		return indexResultFailed
	}

	return indexResultAdded
}

// isNSFW returns true if media file might be offensive and detection is enabled.
func (ind *Index) isNSFW(jpeg *MediaFile) bool {
	if !ind.conf.DetectNSFW() {
		return false
	}

	filename, err := jpeg.Thumbnail(ind.thumbnailsPath(), "fit_720")

	if err != nil {
		log.Error(err)
		return false
	}

	if nsfwLabels, err := ind.nsfwDetector.File(filename); err != nil {
		log.Error(err)
		return false
	} else {
		if nsfwLabels.NSFW() {
			log.Warnf("index: \"%s\" might contain offensive content", jpeg.Filename())
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
		filename, err := jpeg.Thumbnail(ind.thumbnailsPath(), thumb)

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

func (ind *Index) addLabels(photoId uint, labels classify.Labels) {
	for _, label := range labels {
		lm := entity.NewLabel(txt.Title(label.Name), label.Priority).FirstOrCreate(ind.db)

		if lm.New && label.Priority >= 0 {
			event.Publish("count.labels", event.Data{
				"count": 1,
			})
		}

		if lm.LabelPriority != label.Priority {
			lm.LabelPriority = label.Priority

			if err := ind.db.Save(&lm).Error; err != nil {
				log.Errorf("index: %s", err)
			}
		}

		plm := entity.NewPhotoLabel(photoId, lm.ID, label.Uncertainty, label.Source).FirstOrCreate(ind.db)

		// Add categories
		for _, category := range label.Categories {
			sn := entity.NewLabel(txt.Title(category), -3).FirstOrCreate(ind.db)
			if err := ind.db.Model(&lm).Association("LabelCategories").Append(sn).Error; err != nil {
				log.Errorf("index: %s", err)
			}
		}

		if plm.LabelUncertainty > label.Uncertainty {
			plm.LabelUncertainty = label.Uncertainty
			plm.LabelSource = label.Source
			if err := ind.db.Save(&plm).Error; err != nil {
				log.Errorf("index: %s", err)
			}
		}
	}
}

func (ind *Index) indexLocation(mediaFile *MediaFile, photo *entity.Photo, labels classify.Labels, fileChanged bool, o IndexOptions) ([]string, classify.Labels) {
	var keywords []string

	location, err := mediaFile.Location()

	if err == nil {
		location.Lock()
		defer location.Unlock()

		err = location.Find(ind.db, ind.conf.GeoCodingApi())
	}

	if err == nil {
		if location.Place.New {
			event.Publish("count.places", event.Data{
				"count": 1,
			})
		}

		photo.Location = location
		photo.LocationID = location.ID
		photo.Place = location.Place
		photo.PlaceID = location.PlaceID
		photo.LocationEstimated = false

		country := entity.NewCountry(location.CountryCode(), location.CountryName()).FirstOrCreate(ind.db)

		if country.New {
			event.Publish("count.countries", event.Data{
				"count": 1,
			})
		}

		locCategory := location.Category()
		keywords = append(keywords, location.Keywords()...)

		// Append category from reverse location lookup
		if locCategory != "" {
			keywords = append(keywords, locCategory)
			labels = append(labels, classify.LocationLabel(locCategory, 0, -1))
		}

		if (fileChanged || o.UpdateTitle) && photo.ModifiedTitle == false {
			if title := labels.Title(location.Name()); title != "" { // TODO: User defined title format
				log.Infof("index: using label \"%s\" to create photo title", title)
				if location.NoCity() || location.LongCity() || location.CityContains(title) {
					photo.PhotoTitle = fmt.Sprintf("%s / %s / %s", txt.Title(title), location.CountryName(), photo.TakenAt.Format("2006"))
				} else {
					photo.PhotoTitle = fmt.Sprintf("%s / %s / %s", txt.Title(title), location.City(), photo.TakenAt.Format("2006"))
				}
			} else if location.Name() != "" && location.City() != "" {
				if len(location.Name()) > 45 {
					photo.PhotoTitle = txt.Title(location.Name())
				} else if len(location.Name()) > 20 || len(location.City()) > 16 || strings.Contains(location.Name(), location.City()) {
					photo.PhotoTitle = fmt.Sprintf("%s / %s", location.Name(), photo.TakenAt.Format("2006"))
				} else {
					photo.PhotoTitle = fmt.Sprintf("%s / %s / %s", location.Name(), location.City(), photo.TakenAt.Format("2006"))
				}
			} else if location.City() != "" && location.CountryName() != "" {
				if len(location.City()) > 20 {
					photo.PhotoTitle = fmt.Sprintf("%s / %s", location.City(), photo.TakenAt.Format("2006"))
				} else {
					photo.PhotoTitle = fmt.Sprintf("%s / %s / %s", location.City(), location.CountryName(), photo.TakenAt.Format("2006"))
				}
			}

			if photo.NoTitle() {
				log.Warnf("index: could not set photo title based on location or labels for \"%s\"", filepath.Base(mediaFile.Filename()))
			} else {
				log.Infof("index: new photo title is \"%s\"", photo.PhotoTitle)
			}
		}
	} else {
		log.Warn(err)

		photo.Place = entity.UnknownPlace
		photo.PlaceID = entity.UnknownPlace.ID
	}

	photo.PhotoCountry = photo.Place.LocCountry

	return keywords, labels
}

func (ind *Index) estimateLocation(photo *entity.Photo) {
	var recentPhoto entity.Photo

	if result := ind.db.Unscoped().Order(gorm.Expr("ABS(DATEDIFF(taken_at, ?)) ASC", photo.TakenAt)).Preload("Place").First(&recentPhoto); result.Error == nil {
		if recentPhoto.HasPlace() {
			photo.Place = recentPhoto.Place
			photo.PhotoCountry = photo.Place.LocCountry
			photo.LocationEstimated = true
			log.Debugf("index: approximate location is \"%s\"", recentPhoto.Place.Label())
		}
	}
}
