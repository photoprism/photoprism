package photoprism

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/models"
	"github.com/photoprism/photoprism/internal/util"
)

const (
	indexResultUpdated IndexResult = "updated"
	indexResultAdded   IndexResult = "added"
	indexResultSkipped IndexResult = "skipped"
)

type IndexResult string

func (i *Indexer) indexMediaFile(mediaFile *MediaFile, o IndexerOptions) IndexResult {
	var photo models.Photo
	var file, primaryFile models.File
	var isPrimary = false
	var exifData *Exif
	var photoQuery, fileQuery *gorm.DB
	var keywords []string

	labels := Labels{}
	fileBase := mediaFile.Basename()
	filePath := mediaFile.RelativePath(i.originalsPath())
	fileName := mediaFile.RelativeFilename(i.originalsPath())
	fileHash := mediaFile.Hash()
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

		if photoQuery.Error != nil && mediaFile.HasTimeAndPlace() {
			exifData, _ = mediaFile.Exif()
			photoQuery = i.db.Unscoped().First(&photo, "photo_lat = ? AND photo_long = ? AND taken_at = ?", exifData.Lat, exifData.Long, exifData.TakenAt)
		}
	} else {
		photoQuery = i.db.Unscoped().First(&photo, "id = ?", file.PhotoID)
		fileChanged = file.FileHash != fileHash
		isPrimary = file.FilePrimary
	}

	photoExists = photoQuery.Error == nil

	if !fileChanged && photoExists && !photo.TakenAt.IsZero() && o.SkipUnchanged() {
		return indexResultSkipped
	}

	photo.PhotoPath = filePath
	photo.PhotoName = fileBase

	if isPrimary || !photoExists || photo.TakenAt.IsZero() {
		if jpeg, err := mediaFile.Jpeg(); err == nil {
			if fileChanged || o.UpdateLabels || o.UpdateTitle {
				// Image classification labels
				labels = i.classifyImage(jpeg)
			}

			if fileChanged || o.UpdateExif {
				// Read UpdateExif data
				if exifData, err := jpeg.Exif(); err == nil {
					photo.PhotoLat = exifData.Lat
					photo.PhotoLong = exifData.Long
					photo.TakenAt = exifData.TakenAt
					photo.TakenAtLocal = exifData.TakenAtLocal
					photo.TimeZone = exifData.TimeZone
					photo.PhotoAltitude = exifData.Altitude
					photo.PhotoArtist = exifData.Artist

					if exifData.UUID != "" {
						log.Debugf("index: photo uuid \"%s\"", exifData.UUID)
						photo.PhotoUUID = exifData.UUID
					} else {
						log.Debug("index: no photo uuid in exif data")
					}
				}
			}

			if fileChanged || o.UpdateCamera {
				// Set UpdateCamera, Lens, Focal Length and F Number
				photo.Camera = models.NewCamera(mediaFile.CameraModel(), mediaFile.CameraMake()).FirstOrCreate(i.db)
				photo.Lens = models.NewLens(mediaFile.LensModel(), mediaFile.LensMake()).FirstOrCreate(i.db)
				photo.PhotoFocalLength = mediaFile.FocalLength()
				photo.PhotoFNumber = mediaFile.FNumber()
				photo.PhotoIso = mediaFile.Iso()
				photo.PhotoExposure = mediaFile.Exposure()
			}
		}

		if fileChanged || o.UpdateLocation || o.UpdateTitle {
			keywords, labels = i.indexLocation(mediaFile, &photo, keywords, labels, fileChanged, o)
		}

		if (fileChanged || o.UpdateTitle) && photo.PhotoTitle == "" {
			if len(labels) > 0 && labels[0].Priority >= -1 && labels[0].Uncertainty <= 85 && labels[0].Name != "" {
				photo.PhotoTitle = fmt.Sprintf("%s / %s", util.Title(labels[0].Name), mediaFile.DateCreated().Format("2006"))
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
		}

		log.Debugf("index: changed photo title to \"%s\"", photo.PhotoTitle)
	}

	// This should never happen
	if photo.TakenAt.IsZero() || photo.TakenAtLocal.IsZero() {
		photo.TakenAt = mediaFile.DateCreated()
		photo.TakenAtLocal = photo.TakenAt

		log.Warnf("index: %s has invalid date, set to \"%s\"", filepath.Base(mediaFile.Filename()), photo.TakenAt.Format("2006-01-02 15:04:05"))
	}

	if photoExists {
		// Estimate location
		if o.UpdateLocation && photo.LocationID == 0 {
			i.estimateLocation(&photo)
		}

		i.db.Unscoped().Save(&photo)
	} else {
		event.Publish("count.photos", event.Data{
			"count": 1,
		})

		photo.PhotoFavorite = false

		i.db.Create(&photo)
	}

	if len(labels) > 0 {
		log.Infof("index: adding labels %+v", labels)
	}

	if fileChanged || o.UpdateLabels {
		i.addLabels(photo.ID, labels)
	}

	if result := i.db.Where("file_type = 'jpg' AND file_primary = 1 AND photo_id = ?", photo.ID).First(&primaryFile); result.Error != nil {
		isPrimary = mediaFile.IsJpeg()
	} else {
		isPrimary = mediaFile.IsJpeg() && (fileName == primaryFile.FileName || fileHash == primaryFile.FileHash)
	}

	if (fileChanged || o.UpdateKeywords || o.UpdateTitle) && isPrimary {
		photo.IndexKeywords(keywords, i.db)
	}

	file.PhotoID = photo.ID
	file.PhotoUUID = photo.PhotoUUID
	file.FilePrimary = isPrimary
	file.FileMissing = false
	file.FileName = fileName
	file.FileHash = fileHash
	file.FileType = mediaFile.Type()
	file.FileMime = mediaFile.MimeType()
	file.FileOrientation = mediaFile.Orientation()

	if fileChanged || o.UpdateColors {
		// Color information
		if p, err := mediaFile.Colors(i.thumbnailsPath()); err == nil {
			file.FileMainColor = p.MainColor.Name()
			file.FileColors = p.Colors.Hex()
			file.FileLuminance = p.Luminance.Hex()
			file.FileChroma = p.Chroma.Uint()
		}
	}

	if fileChanged || o.UpdateSize {
		if mediaFile.Width() > 0 && mediaFile.Height() > 0 {
			file.FileWidth = mediaFile.Width()
			file.FileHeight = mediaFile.Height()
			file.FileAspectRatio = mediaFile.AspectRatio()
			file.FilePortrait = mediaFile.Width() < mediaFile.Height()
		}
	}

	if fileQuery.Error == nil {
		i.db.Unscoped().Save(&file)
		return indexResultUpdated
	}

	i.db.Create(&file)
	return indexResultAdded
}

// classifyImage returns all matching labels for a media file.
func (i *Indexer) classifyImage(jpeg *MediaFile) (results Labels) {
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

func (i *Indexer) addLabels(photoId uint, labels Labels) {
	for _, label := range labels {
		lm := models.NewLabel(label.Name, label.Priority).FirstOrCreate(i.db)

		if lm.New {
			event.Publish("count.labels", event.Data{
				"count": 1,
			})
		}

		if lm.LabelPriority != label.Priority {
			lm.LabelPriority = label.Priority
			i.db.Save(&lm)
		}

		plm := models.NewPhotoLabel(photoId, lm.ID, label.Uncertainty, label.Source).FirstOrCreate(i.db)

		// Add categories
		for _, category := range label.Categories {
			sn := models.NewLabel(category, -1).FirstOrCreate(i.db)
			i.db.Model(&lm).Association("LabelCategories").Append(sn)
		}

		if plm.LabelUncertainty > label.Uncertainty {
			plm.LabelUncertainty = label.Uncertainty
			plm.LabelSource = label.Source
			i.db.Save(&plm)
		}
	}
}

func (i *Indexer) indexLocation (mediaFile *MediaFile, photo *models.Photo, keywords []string, labels Labels, fileChanged bool, o IndexerOptions) ([]string, Labels){
	if location, err := mediaFile.Location(); err == nil {
		i.db.FirstOrCreate(location, "id = ?", location.ID)
		photo.Location = location
		photo.LocationEstimated = false

		photo.Country = models.NewCountry(location.LocCountryCode, location.LocCountry).FirstOrCreate(i.db)

		if photo.Country.New {
			event.Publish("count.countries", event.Data{
				"count": 1,
			})
		}

		keywords = append(keywords, util.Keywords(location.LocDisplayName)...)

		// Append labels from OpenStreetMap
		if location.LocCity != "" {
			labels = append(labels, NewLocationLabel(location.LocCity, 0, -2))
		}

		if location.LocCountry != "" {
			labels = append(labels, NewLocationLabel(location.LocCountry, 0, -2))
		}

		if location.LocCategory != "" {
			labels = append(labels, NewLocationLabel(location.LocCategory, 0, -2))
		}

		if location.LocType != "" {
			labels = append(labels, NewLocationLabel(location.LocType, 0, -1))
		}

		// Sort by priority and uncertainty
		sort.Sort(labels)


		if (fileChanged || o.UpdateTitle) && photo.PhotoTitleChanged == false {
			if len(labels) > 0 && labels[0].Priority >= -1 && labels[0].Uncertainty <= 60 && labels[0].Name != "" { // TODO: User defined title format
				log.Infof("index: using label %s to create photo title (%d%% uncertainty)", labels[0].Name, labels[0].Uncertainty)
				if location.LocCity == "" || len(location.LocCity) > 16 || strings.Contains(labels[0].Name, location.LocCity) {
					photo.PhotoTitle = fmt.Sprintf("%s / %s / %s", util.Title(labels[0].Name), location.LocCountry, photo.TakenAt.Format("2006"))
				} else {
					photo.PhotoTitle = fmt.Sprintf("%s / %s / %s", util.Title(labels[0].Name), location.LocCity, photo.TakenAt.Format("2006"))
				}
			} else if location.LocName != "" && location.LocCity != "" {
				if len(location.LocName) > 45 {
					photo.PhotoTitle = util.Title(location.LocName)
				} else if len(location.LocName) > 20 || len(location.LocCity) > 16 || strings.Contains(location.LocName, location.LocCity) {
					photo.PhotoTitle = fmt.Sprintf("%s / %s", util.Title(location.LocName), photo.TakenAt.Format("2006"))
				} else {
					photo.PhotoTitle = fmt.Sprintf("%s / %s / %s", util.Title(location.LocName), location.LocCity, photo.TakenAt.Format("2006"))
				}
			} else if location.LocCity != "" && location.LocCountry != "" {
				if len(location.LocCity) > 20 {
					photo.PhotoTitle = fmt.Sprintf("%s / %s", location.LocCity, photo.TakenAt.Format("2006"))
				} else {
					photo.PhotoTitle = fmt.Sprintf("%s / %s / %s", location.LocCity, location.LocCountry, photo.TakenAt.Format("2006"))
				}
			} else if location.LocCounty != "" && location.LocCountry != "" {
				if len(location.LocCounty) > 20 {
					photo.PhotoTitle = fmt.Sprintf("%s / %s", location.LocCounty, photo.TakenAt.Format("2006"))
				} else {
					photo.PhotoTitle = fmt.Sprintf("%s / %s / %s", location.LocCounty, location.LocCountry, photo.TakenAt.Format("2006"))
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

func (i *Indexer) estimateLocation(photo *models.Photo) {
	var recentPhoto models.Photo

	if result := i.db.Unscoped().Order(gorm.Expr("ABS(DATEDIFF(taken_at, ?)) ASC", photo.TakenAt)).Preload("Country").First(&recentPhoto); result.Error == nil {
		if recentPhoto.Country != nil {
			photo.Country = recentPhoto.Country
			photo.LocationEstimated = true
			log.Debugf("index: approximate location is \"%s\"", recentPhoto.Country.CountryName)
		}
	}
}
