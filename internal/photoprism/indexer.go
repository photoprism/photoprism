package photoprism

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/models"
	"github.com/photoprism/photoprism/internal/util"
)

const (
	indexResultUpdated = "updated"
	indexResultAdded   = "added"
)

// Indexer defines an indexer with originals path tensorflow and a db.
type Indexer struct {
	conf       *config.Config
	tensorFlow *TensorFlow
	db         *gorm.DB
}

// NewIndexer returns a new indexer.
// TODO: Is it really necessary to return a pointer?
func NewIndexer(conf *config.Config, tensorFlow *TensorFlow) *Indexer {
	instance := &Indexer{
		conf:       conf,
		tensorFlow: tensorFlow,
		db:         conf.Db(),
	}

	return instance
}

func (i *Indexer) originalsPath() string {
	return i.conf.OriginalsPath()
}

func (i *Indexer) thumbnailsPath() string {
	return i.conf.ThumbnailsPath()
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

	log.Debugf("finding %+v labels for %s took %s", results, jpeg.Filename(), elapsed)

	return results
}

func (i *Indexer) indexMediaFile(mediaFile *MediaFile) string {
	var photo models.Photo
	var file, primaryFile models.File
	var isPrimary = false
	var exifData *Exif
	var photoQuery, fileQuery *gorm.DB

	labels := Labels{}
	fileBase := mediaFile.Basename()
	filePath := mediaFile.RelativePath(i.originalsPath())
	fileName := mediaFile.RelativeFilename(i.originalsPath())
	fileHash := mediaFile.Hash()

	event.Publish("index.file", event.Data{
		"fileHash": fileHash,
		"fileName": fileName,
		"baseName": filepath.Base(fileName),
	})

	exifData, err := mediaFile.Exif()

	if err != nil {
		log.Debug(err)
	}

	fileQuery = i.db.Unscoped().First(&file, "file_hash = ? OR file_name = ?", fileHash, fileName)

	if fileQuery.Error != nil {
		photoQuery = i.db.Unscoped().First(&photo, "photo_path = ? AND photo_name = ?", filePath, fileBase)

		if photoQuery.Error != nil && mediaFile.HasTimeAndPlace() {
			photoQuery = i.db.Unscoped().First(&photo, "photo_lat = ? AND photo_long = ? AND taken_at = ?", exifData.Lat, exifData.Long, exifData.TakenAt)
		}
	} else {
		photoQuery = i.db.Unscoped().First(&photo, "id = ?", file.PhotoID)
	}

	photo.PhotoPath = filePath
	photo.PhotoName = fileBase

	if jpeg, err := mediaFile.Jpeg(); err == nil {
		// Image classification labels
		labels = i.classifyImage(jpeg)

		// Read Exif data
		if exifData, err := jpeg.Exif(); err == nil {
			photo.PhotoLat = exifData.Lat
			photo.PhotoLong = exifData.Long
			photo.TakenAt = exifData.TakenAt
			photo.TakenAtLocal = exifData.TakenAtLocal
			photo.TimeZone = exifData.TimeZone
			photo.PhotoAltitude = exifData.Altitude
			photo.PhotoArtist = exifData.Artist

			if exifData.UUID != "" {
				log.Debugf("photo uuid: %s", exifData.UUID)
				photo.PhotoUUID = exifData.UUID
			} else {
				log.Debug("no photo uuid")
			}
		}

		// Set Camera, Lens, Focal Length and F Number
		photo.Camera = models.NewCamera(mediaFile.CameraModel(), mediaFile.CameraMake()).FirstOrCreate(i.db)
		photo.Lens = models.NewLens(mediaFile.LensModel(), mediaFile.LensMake()).FirstOrCreate(i.db)
		photo.PhotoFocalLength = mediaFile.FocalLength()
		photo.PhotoFNumber = mediaFile.FNumber()
		photo.PhotoIso = mediaFile.Iso()
		photo.PhotoExposure = mediaFile.Exposure()
	}

	if photo.TakenAt.IsZero() || photo.TakenAtLocal.IsZero() {
		photo.TakenAt = mediaFile.DateCreated()
		photo.TakenAtLocal = photo.TakenAt
	}

	if location, err := mediaFile.Location(); err == nil {
		i.db.FirstOrCreate(location, "id = ?", location.ID)
		photo.Location = location
		photo.LocationEstimated = false

		photo.Country = models.NewCountry(location.LocCountryCode, location.LocCountry).FirstOrCreate(i.db)

		// Append labels from OpenStreetMap
		if location.LocCity != "" {
			labels = append(labels, NewLocationLabel(location.LocCity, 0, -3))
		}

		if location.LocCounty != "" {
			labels = append(labels, NewLocationLabel(location.LocCounty, 0, -3))
		}

		if location.LocCountry != "" {
			labels = append(labels, NewLocationLabel(location.LocCountry, 0, -3))
		}

		if location.LocCategory != "" {
			labels = append(labels, NewLocationLabel(location.LocCategory, 0, -2))
		}

		if location.LocName != "" && len(location.LocName) <= 25 {
			labels = append(labels, NewLocationLabel(location.LocName, 50, 0))
		}

		if location.LocType != "" {
			labels = append(labels, NewLocationLabel(location.LocType, 0, -1))
		}

		// Sort by priority and uncertainty
		sort.Sort(labels)

		if photo.PhotoTitleChanged == false {
			log.Infof("setting title based on the following labels: %#v", labels)
			if len(labels) > 0 && labels[0].Priority >= -1 && labels[0].Uncertainty <= 60 && labels[0].Name != "" { // TODO: User defined title format
				log.Infof("label for title: %#v", labels[0])
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
		}
	} else {
		log.Debugf("location cannot be determined precisely: %s", err)
	}

	if photo.PhotoTitleChanged == false && photo.PhotoTitle == "" {
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
		}
	}

	log.Debugf("title: \"%s\"", photo.PhotoTitle)

	if photoQuery.Error != nil {
		photo.PhotoFavorite = false

		i.db.Create(&photo)
	} else {
		// Estimate location
		if photo.LocationID == 0 {
			var recentPhoto models.Photo

			if result := i.db.Unscoped().Order(gorm.Expr("ABS(DATEDIFF(taken_at, ?)) ASC", photo.TakenAt)).Preload("Country").First(&recentPhoto); result.Error == nil {
				if recentPhoto.Country != nil {
					photo.Country = recentPhoto.Country
					photo.LocationEstimated = true
					log.Debugf("approximate location: %s", recentPhoto.Country.CountryName)
				}
			}
		}

		i.db.Unscoped().Save(&photo)
	}

	log.Infof("adding labels: %+v", labels)

	for _, label := range labels {
		lm := models.NewLabel(label.Name, label.Priority).FirstOrCreate(i.db)

		if lm.LabelPriority != label.Priority {
			lm.LabelPriority = label.Priority
			i.db.Save(&lm)
		}

		plm := models.NewPhotoLabel(photo.ID, lm.ID, label.Uncertainty, label.Source).FirstOrCreate(i.db)

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

	if result := i.db.Where("file_type = 'jpg' AND file_primary = 1 AND photo_id = ?", photo.ID).First(&primaryFile); result.Error != nil {
		isPrimary = mediaFile.IsJpeg()
	} else {
		isPrimary = mediaFile.IsJpeg() && (fileName == primaryFile.FileName || fileHash == primaryFile.FileHash)
	}

	file.PhotoID = photo.ID
	file.FilePrimary = isPrimary
	file.FileMissing = false
	file.FileName = fileName
	file.FileHash = fileHash
	file.FileType = mediaFile.Type()
	file.FileMime = mediaFile.MimeType()
	file.FileOrientation = mediaFile.Orientation()

	// Color information
	if p, err := mediaFile.Colors(i.thumbnailsPath()); err == nil {
		file.FileMainColor = p.MainColor.Name()
		file.FileColors = p.Colors.Hex()
		file.FileLuminance = p.Luminance.Hex()
		file.FileChroma = p.Chroma.Uint()
	}

	if mediaFile.Width() > 0 && mediaFile.Height() > 0 {
		file.FileWidth = mediaFile.Width()
		file.FileHeight = mediaFile.Height()
		file.FileAspectRatio = mediaFile.AspectRatio()
		file.FilePortrait = mediaFile.Width() < mediaFile.Height()
	}

	if fileQuery.Error == nil {
		i.db.Unscoped().Save(&file)
		return indexResultUpdated
	}

	i.db.Create(&file)
	return indexResultAdded
}

// IndexRelated will index all mediafiles which has relate to a given mediafile.
func (i *Indexer) IndexRelated(mediaFile *MediaFile) map[string]bool {
	indexed := make(map[string]bool)

	relatedFiles, mainFile, err := mediaFile.RelatedFiles()

	if err != nil {
		log.Warnf("could not index \"%s\": %s", mediaFile.RelativeFilename(i.originalsPath()), err.Error())

		return indexed
	}

	mainIndexResult := i.indexMediaFile(mainFile)
	indexed[mainFile.Filename()] = true

	log.Infof("%s main %s file \"%s\"", mainIndexResult, mainFile.Type(), mainFile.RelativeFilename(i.originalsPath()))

	for _, relatedMediaFile := range relatedFiles {
		if indexed[relatedMediaFile.Filename()] {
			continue
		}

		indexResult := i.indexMediaFile(relatedMediaFile)
		indexed[relatedMediaFile.Filename()] = true

		log.Infof("%s related %s file \"%s\"", indexResult, relatedMediaFile.Type(), relatedMediaFile.RelativeFilename(i.originalsPath()))
	}

	return indexed
}

// IndexAll will index all mediafiles.
func (i *Indexer) IndexAll() map[string]bool {
	indexed := make(map[string]bool)

	err := filepath.Walk(i.originalsPath(), func(filename string, fileInfo os.FileInfo, err error) error {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Could not index file %s due to an unexpected error: %s", filename, err)
			}
		}()
		if err != nil || indexed[filename] {
			return nil
		}

		if fileInfo.IsDir() || strings.HasPrefix(filepath.Base(filename), ".") {
			return nil
		}

		mediaFile, err := NewMediaFile(filename)

		if err != nil || !mediaFile.IsPhoto() {
			return nil
		}

		for relatedFilename := range i.IndexRelated(mediaFile) {
			indexed[relatedFilename] = true
		}

		return nil
	})

	if err != nil {
		log.Warn(err.Error())
	}

	return indexed
}
