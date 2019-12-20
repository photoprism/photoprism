package photoprism

import (
	"fmt"
	"math"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/util"
)

const (
	indexResultUpdated IndexResult = "updated"
	indexResultAdded   IndexResult = "added"
	indexResultSkipped IndexResult = "skipped"
	indexResultFailed  IndexResult = "failed"
)

type IndexResult string

func (i *Indexer) indexMediaFile(m *MediaFile, o IndexerOptions) IndexResult {
	start := time.Now()

	var photo entity.Photo
	var file, primaryFile entity.File
	var exifData *Exif
	var photoQuery, fileQuery *gorm.DB
	var keywords []string
	var isNSFW bool

	labels := Labels{}
	fileBase := m.Basename()
	filePath := m.RelativePath(i.originalsPath())
	fileName := m.RelativeFilename(i.originalsPath())
	fileHash := m.Hash()
	fileChanged := true
	fileExists := false
	photoExists := false

	event.Publish("index.indexing", event.Data{
		"fileHash": fileHash,
		"fileName": fileName,
		"baseName": filepath.Base(fileName),
	})

	fileQuery = i.db.Unscoped().First(&file, "file_hash = ? OR file_name = ?", fileHash, fileName)
	fileExists = fileQuery.Error == nil

	if !fileExists {
		photoQuery = i.db.Unscoped().First(&photo, "photo_path = ? AND photo_name = ?", filePath, fileBase)

		if photoQuery.Error != nil && m.HasTimeAndPlace() {
			exifData, _ = m.Exif()
			photoQuery = i.db.Unscoped().First(&photo, "photo_lat = ? AND photo_lng = ? AND taken_at = ?", exifData.Lat, exifData.Lng, exifData.TakenAt)
		}
	} else {
		photoQuery = i.db.Unscoped().First(&photo, "id = ?", file.PhotoID)
		fileChanged = file.FileHash != fileHash
	}

	photoExists = photoQuery.Error == nil

	if !fileChanged && photoExists && o.SkipUnchanged() {
		return indexResultSkipped
	}

	if !file.FilePrimary {
		if photoExists {
			if q := i.db.Where("file_type = 'jpg' AND file_primary = 1 AND photo_id = ?", photo.ID).First(&primaryFile); q.Error != nil {
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
		if fileChanged || o.UpdateKeywords || o.UpdateLabels || o.UpdateTitle {
			// Image classification labels
			labels, isNSFW = i.classifyImage(m)
			photo.PhotoNSFW = isNSFW
		}

		if fileChanged || o.UpdateExif {
			// Read UpdateExif data
			if exifData, err := m.Exif(); err == nil {
				photo.PhotoLat = exifData.Lat
				photo.PhotoLng = exifData.Lng
				photo.TakenAt = exifData.TakenAt
				photo.TakenAtLocal = exifData.TakenAtLocal
				photo.TimeZone = exifData.TimeZone
				photo.PhotoAltitude = exifData.Altitude
				photo.PhotoArtist = exifData.Artist

				if len(exifData.UUID) > 15 {
					log.Debugf("index: file uuid \"%s\"", exifData.UUID)

					file.FileUUID = exifData.UUID
				}
			}
		}

		if fileChanged || o.UpdateCamera {
			// Set UpdateCamera, Lens, Focal Length and F Number
			photo.Camera = entity.NewCamera(m.CameraModel(), m.CameraMake()).FirstOrCreate(i.db)
			photo.Lens = entity.NewLens(m.LensModel(), m.LensMake()).FirstOrCreate(i.db)
			photo.PhotoFocalLength = m.FocalLength()
			photo.PhotoFNumber = m.FNumber()
			photo.PhotoIso = m.Iso()
			photo.PhotoExposure = m.Exposure()
		}

		if fileChanged || o.UpdateKeywords || o.UpdateLocation || o.UpdateTitle {
			locKeywords, locLabels := i.indexLocation(m, &photo, labels, fileChanged, o)
			keywords = append(keywords, locKeywords...)
			labels = append(labels, locLabels...)
		}

		if photo.PhotoTitle == "" || (fileChanged || o.UpdateTitle) && photo.PhotoTitleChanged == false && photo.LocationID == "" {
			if len(labels) > 0 && labels[0].Priority >= -1 && labels[0].Uncertainty <= 85 && labels[0].Name != "" {
				photo.PhotoTitle = fmt.Sprintf("%s / %s", util.Title(labels[0].Name), m.DateCreated().Format("2006"))
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
	}

	if photoExists {
		// Estimate location
		if o.UpdateLocation && photo.LocationID == "" {
			i.estimateLocation(&photo)
		}

		if err := i.db.Unscoped().Save(&photo).Error; err != nil {
			log.Errorf("index: %s", err)
			return indexResultFailed
		}
	} else {
		event.Publish("count.photos", event.Data{
			"count": 1,
		})

		photo.PhotoFavorite = false

		if err := i.db.Create(&photo).Error; err != nil {
			log.Errorf("index: %s", err)
			return indexResultFailed
		}
	}

	if len(labels) > 0 {
		log.Infof("index: adding labels %+v", labels)
		i.addLabels(photo.ID, labels)
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
		if p, err := m.Colors(i.thumbnailsPath()); err == nil {
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
		photo.IndexKeywords(keywords, i.db)
	}

	if fileQuery.Error == nil {
		file.UpdatedIn = int64(time.Since(start))

		if err := i.db.Unscoped().Save(&file).Error; err != nil {
			log.Errorf("index: %s", err)
			return indexResultFailed
		}

		return indexResultUpdated
	}

	file.CreatedIn = int64(time.Since(start))

	if err := i.db.Create(&file).Error; err != nil {
		log.Errorf("index: %s", err)
		return indexResultFailed
	}

	return indexResultAdded
}

// classifyImage returns all matching labels for a media file.
func (i *Indexer) classifyImage(jpeg *MediaFile) (results Labels, isNSFW bool) {
	start := time.Now()

	var thumbs []string

	if jpeg.AspectRatio() == 1 {
		thumbs = []string{"tile_224"}
	} else {
		thumbs = []string{"tile_224", "left_224", "right_224"}
	}

	var labels Labels

	for _, thumb := range thumbs {
		filename, err := jpeg.Thumbnail(i.thumbnailsPath(), thumb)

		if err != nil {
			log.Error(err)
			continue
		}

		imageLabels, err := i.tensorFlow.LabelsFromFile(filename)

		if err != nil {
			log.Error(err)
			continue
		}

		labels = append(labels, imageLabels...)
	}

	if filename, err := jpeg.Thumbnail(i.thumbnailsPath(), "fit_720"); err != nil {
		log.Error(err)
	} else {
		if nsfwLabels, err := i.nsfwDetector.LabelsFromFile(filename); err != nil {
			log.Error(err)
		} else {
			log.Infof("nsfw: %+v", nsfwLabels)

			if nsfwLabels.NSFW() {
				isNSFW = true
			}

			if nsfwLabels.Sexy > 0.85 {
				uncertainty := 100 - int(math.Round(float64(nsfwLabels.Sexy*100)))
				labels = append(labels, Label{Name: "sexy", Source: "nsfw", Uncertainty: uncertainty, Priority: -1})
			}

			if nsfwLabels.Drawing > 0.85 {
				uncertainty := 100 - int(math.Round(float64(nsfwLabels.Drawing*100)))
				categories := []string{"painting", "art"}
				labels = append(labels, Label{Name: "drawing", Source: "nsfw", Uncertainty: uncertainty, Priority: -1, Categories: categories})
			}
		}
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

	if isNSFW {
		log.Info("index: image might contain offensive content")
	}

	elapsed := time.Since(start)

	log.Debugf("index: image classification took %s", elapsed)

	return results, isNSFW
}

func (i *Indexer) addLabels(photoId uint, labels Labels) {
	for _, label := range labels {
		lm := entity.NewLabel(label.Name, label.Priority).FirstOrCreate(i.db)

		if lm.New && label.Priority >= 0 {
			event.Publish("count.labels", event.Data{
				"count": 1,
			})
		}

		if lm.LabelPriority != label.Priority {
			lm.LabelPriority = label.Priority

			if err := i.db.Save(&lm).Error; err != nil {
				log.Errorf("index: %s", err)
			}
		}

		plm := entity.NewPhotoLabel(photoId, lm.ID, label.Uncertainty, label.Source).FirstOrCreate(i.db)

		// Add categories
		for _, category := range label.Categories {
			sn := entity.NewLabel(category, -3).FirstOrCreate(i.db)
			if err := i.db.Model(&lm).Association("LabelCategories").Append(sn).Error; err != nil {
				log.Errorf("index: %s", err)
			}
		}

		if plm.LabelUncertainty > label.Uncertainty {
			plm.LabelUncertainty = label.Uncertainty
			plm.LabelSource = label.Source
			if err := i.db.Save(&plm).Error; err != nil {
				log.Errorf("index: %s", err)
			}
		}
	}
}

func (i *Indexer) indexLocation(mediaFile *MediaFile, photo *entity.Photo, labels Labels, fileChanged bool, o IndexerOptions) ([]string, Labels) {
	var keywords []string

	if location, err := mediaFile.Location(); err == nil {
		err := location.Find(i.db)

		if err != nil {
			log.Error(err)
			return keywords, labels
		}

		photo.Location = location
		photo.LocationID = location.ID
		photo.LocationEstimated = false

		photo.Country = entity.NewCountry(location.CountryCode(), location.CountryName()).FirstOrCreate(i.db)

		if photo.Country.New {
			event.Publish("count.countries", event.Data{
				"count": 1,
			})
		}

		countryName := photo.Country.CountryName
		locCategory := location.Category()

		keywords = append(keywords, location.Keywords()...)

		// Append category from reverse location lookup
		if locCategory != "" {
			keywords = append(keywords, locCategory)
			labels = append(labels, NewLocationLabel(locCategory, 0, -1))
		}

		if (fileChanged || o.UpdateTitle) && photo.PhotoTitleChanged == false {
			if title := labels.Title(location.Title()); title != "" { // TODO: User defined title format
				log.Infof("index: using label \"%s\" to create photo title", title)
				if location.LocCity == "" || len(location.LocCity) > 16 || strings.Contains(title, location.LocCity) {
					photo.PhotoTitle = fmt.Sprintf("%s / %s / %s", util.Title(title), countryName, photo.TakenAt.Format("2006"))
				} else {
					photo.PhotoTitle = fmt.Sprintf("%s / %s / %s", util.Title(title), location.LocCity, photo.TakenAt.Format("2006"))
				}
			} else if location.Title() != "" && location.City() != "" {
				if len(location.Title()) > 45 {
					photo.PhotoTitle = util.Title(location.Title())
				} else if len(location.Title()) > 20 || len(location.City()) > 16 || strings.Contains(location.Title(), location.City()) {
					photo.PhotoTitle = fmt.Sprintf("%s / %s", location.Title(), photo.TakenAt.Format("2006"))
				} else {
					photo.PhotoTitle = fmt.Sprintf("%s / %s / %s", location.Title(), location.LocCity, photo.TakenAt.Format("2006"))
				}
			} else if location.City() != "" && countryName != "" {
				if len(location.City()) > 20 {
					photo.PhotoTitle = fmt.Sprintf("%s / %s", location.LocCity, photo.TakenAt.Format("2006"))
				} else {
					photo.PhotoTitle = fmt.Sprintf("%s / %s / %s", location.LocCity, countryName, photo.TakenAt.Format("2006"))
				}
			}

			if photo.PhotoTitle == "" {
				log.Warnf("index: could not set photo title based on location or labels for \"%s\"", filepath.Base(mediaFile.Filename()))
			} else {
				log.Infof("index: new photo title is \"%s\"", photo.PhotoTitle)
			}
		}
	} else {
		log.Debugf("index: location cannot be determined precisely (%s)", err.Error())
	}

	return keywords, labels
}

func (i *Indexer) estimateLocation(photo *entity.Photo) {
	var recentPhoto entity.Photo

	if result := i.db.Unscoped().Order(gorm.Expr("ABS(DATEDIFF(taken_at, ?)) ASC", photo.TakenAt)).Preload("Country").First(&recentPhoto); result.Error == nil {
		if recentPhoto.Country != nil {
			photo.Country = recentPhoto.Country
			photo.LocationEstimated = true
			log.Debugf("index: approximate location is \"%s\"", recentPhoto.Country.CountryName)
		}
	}
}
