package photoprism

import (
	"fmt"
	"github.com/jinzhu/gorm"
	. "github.com/photoprism/photoprism/internal/models"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	IndexResultUpdated = "Updated"
	IndexResultAdded   = "Added"
)

type Indexer struct {
	originalsPath string
	tensorFlow    *TensorFlow
	db            *gorm.DB
}

func NewIndexer(originalsPath string, tensorFlow *TensorFlow, db *gorm.DB) *Indexer {
	instance := &Indexer{
		originalsPath: originalsPath,
		tensorFlow:    tensorFlow,
		db:            db,
	}

	return instance
}

func (i *Indexer) GetImageTags(jpeg *MediaFile) (results []*Tag) {
	tags, err := i.tensorFlow.GetImageTagsFromFile(jpeg.filename)

	if err != nil {
		return results
	}

	for _, tag := range tags {
		if tag.Probability > 0.15 { // TODO: Use config variable
			results = i.appendTag(results, tag.Label)
		}
	}

	return results
}

func (i *Indexer) appendTag(tags []*Tag, label string) []*Tag {
	if label == "" {
		return tags
	}

	label = strings.ToLower(label)

	for _, tag := range tags {
		if tag.TagLabel == label {
			return tags
		}
	}

	tag := NewTag(label).FirstOrCreate(i.db)

	return append(tags, tag)
}

func (i *Indexer) IndexMediaFile(mediaFile *MediaFile) string {
	var photo Photo
	var file, primaryFile File
	var isPrimary = false
	var colorNames []string
	var tags []*Tag

	canonicalName := mediaFile.GetCanonicalNameFromFile()
	fileHash := mediaFile.GetHash()
	relativeFileName := mediaFile.GetRelativeFilename(i.originalsPath)

	photoQuery := i.db.First(&photo, "photo_canonical_name = ?", canonicalName)

	if photoQuery.Error != nil {
		if jpeg, err := mediaFile.GetJpeg(); err == nil {
			// Geo Location
			if exifData, err := jpeg.GetExifData(); err == nil {
				photo.PhotoLat = exifData.Lat
				photo.PhotoLong = exifData.Long
				photo.PhotoArtist = exifData.Artist
			}

			// PhotoColors
			colorNames, photo.PhotoVibrantColor, photo.PhotoMutedColor = jpeg.GetColors()

			photo.PhotoColors = strings.Join(colorNames, ", ")

			// Tags (TensorFlow)
			tags = i.GetImageTags(jpeg)
		}

		if location, err := mediaFile.GetLocation(); err == nil {
			i.db.FirstOrCreate(location, "id = ?", location.ID)
			photo.Location = location
			photo.Country = NewCountry(location.LocCountryCode, location.LocCountry).FirstOrCreate(i.db)

			tags = i.appendTag(tags, location.LocCity)
			tags = i.appendTag(tags, location.LocCounty)
			tags = i.appendTag(tags, location.LocCountry)
			tags = i.appendTag(tags, location.LocCategory)
			tags = i.appendTag(tags, location.LocName)
			tags = i.appendTag(tags, location.LocType)

			if photo.PhotoTitle == "" && location.LocName != "" { // TODO: User defined title format
				if len(location.LocName) > 40 {
					photo.PhotoTitle = fmt.Sprintf("%s / %s", strings.Title(location.LocName), mediaFile.GetDateCreated().Format("2006"))
				} else {
					photo.PhotoTitle = fmt.Sprintf("%s / %s / %s", strings.Title(location.LocName), location.LocCity, mediaFile.GetDateCreated().Format("2006"))
				}
			} else if photo.PhotoTitle == "" && location.LocCity != "" {
				photo.PhotoTitle = fmt.Sprintf("%s / %s / %s", location.LocCity, location.LocCountry, mediaFile.GetDateCreated().Format("2006"))
			} else if photo.PhotoTitle == "" && location.LocCounty != "" {
				photo.PhotoTitle = fmt.Sprintf("%s / %s / %s", location.LocCounty, location.LocCountry, mediaFile.GetDateCreated().Format("2006"))
			}
		} else {
			var recentPhoto Photo

			if result := i.db.Order(gorm.Expr("ABS(DATEDIFF(taken_at, ?)) ASC", mediaFile.GetDateCreated())).Preload("Country").First(&recentPhoto); result.Error == nil {
				if recentPhoto.Country != nil {
					photo.Country = recentPhoto.Country
				}
			}
		}

		photo.Tags = tags

		if photo.PhotoTitle == "" {
			if len(photo.Tags) > 0 { // TODO: User defined title format
				photo.PhotoTitle = fmt.Sprintf("%s / %s", strings.Title(photo.Tags[0].TagLabel), mediaFile.GetDateCreated().Format("2006"))
			} else if photo.Country != nil && photo.Country.CountryName != "" {
				photo.PhotoTitle = fmt.Sprintf("%s / %s", strings.Title(photo.Country.CountryName), mediaFile.GetDateCreated().Format("2006"))
			} else {
				photo.PhotoTitle = fmt.Sprintf("Unknown / %s", mediaFile.GetDateCreated().Format("2006"))
			}
		}

		photo.Camera = NewCamera(mediaFile.GetCameraModel(), mediaFile.GetCameraMake()).FirstOrCreate(i.db)
		photo.Lens = NewLens(mediaFile.GetLensModel(), mediaFile.GetLensMake()).FirstOrCreate(i.db)
		photo.PhotoFocalLength = mediaFile.GetFocalLength()
		photo.PhotoAperture = mediaFile.GetAperture()

		photo.TakenAt = mediaFile.GetDateCreated()
		photo.PhotoCanonicalName = canonicalName
		photo.PhotoFavorite = false

		i.db.Create(&photo)
	} else if time.Now().Sub(photo.UpdatedAt).Minutes() > 10 { // If updated more than 10 minutes ago
		if jpeg, err := mediaFile.GetJpeg(); err == nil {
			// PhotoColors
			colorNames, photo.PhotoVibrantColor, photo.PhotoMutedColor = jpeg.GetColors()

			photo.PhotoColors = strings.Join(colorNames, ", ")

			photo.Camera = NewCamera(mediaFile.GetCameraModel(), mediaFile.GetCameraMake()).FirstOrCreate(i.db)
			photo.Lens = NewLens(mediaFile.GetLensModel(), mediaFile.GetLensMake()).FirstOrCreate(i.db)
			photo.PhotoFocalLength = mediaFile.GetFocalLength()
			photo.PhotoAperture = mediaFile.GetAperture()
		}

		if photo.LocationID == 0 {
			var recentPhoto Photo

			if result := i.db.Order(gorm.Expr("ABS(DATEDIFF(taken_at, ?)) ASC", photo.TakenAt)).Preload("Country").First(&recentPhoto); result.Error == nil {
				if recentPhoto.Country != nil {
					photo.Country = recentPhoto.Country
				}
			}
		}

		i.db.Save(&photo)
	}

	if result := i.db.Where("file_type = 'jpg' AND file_primary = 1 AND photo_id = ?", photo.ID).First(&primaryFile); result.Error != nil {
		isPrimary = mediaFile.GetType() == FileTypeJpeg
	} else {
		isPrimary = mediaFile.GetType() == FileTypeJpeg && (relativeFileName == primaryFile.FileName || fileHash == primaryFile.FileHash)
	}

	fileQuery := i.db.First(&file, "file_hash = ? OR file_name = ?", fileHash, relativeFileName)

	file.PhotoID = photo.ID
	file.FilePrimary = isPrimary
	file.FileName = relativeFileName
	file.FileHash = fileHash
	file.FileType = mediaFile.GetType()
	file.FileMime = mediaFile.GetMimeType()
	file.FileOrientation = mediaFile.GetOrientation()

	// Perceptual Hash
	if file.FilePerceptualHash == "" && mediaFile.IsJpeg() {
		if perceptualHash, err := mediaFile.GetPerceptualHash(); err == nil {
			file.FilePerceptualHash = perceptualHash
		}
	}

	if mediaFile.GetWidth() > 0 && mediaFile.GetHeight() > 0 {
		file.FileWidth = mediaFile.GetWidth()
		file.FileHeight = mediaFile.GetHeight()
		file.FileAspectRatio = mediaFile.GetAspectRatio()
	}

	if fileQuery.Error == nil {
		i.db.Save(&file)
		return IndexResultUpdated
	} else {
		i.db.Create(&file)
		return IndexResultAdded
	}
}

func (i *Indexer) IndexRelated(mediaFile *MediaFile) map[string]bool {
	indexed := make(map[string]bool)

	relatedFiles, mainFile, err := mediaFile.GetRelatedFiles()

	if err != nil {
		log.Printf("Could not index \"%s\": %s", mediaFile.GetRelativeFilename(i.originalsPath), err.Error())

		return indexed
	}

	mainIndexResult := i.IndexMediaFile(mainFile)
	indexed[mainFile.GetFilename()] = true

	log.Printf("%s main %s file \"%s\"", mainIndexResult, mainFile.GetType(), mainFile.GetRelativeFilename(i.originalsPath))

	for _, relatedMediaFile := range relatedFiles {
		if indexed[relatedMediaFile.GetFilename()] {
			continue
		}

		indexResult := i.IndexMediaFile(relatedMediaFile)
		indexed[relatedMediaFile.GetFilename()] = true

		log.Printf("%s related %s file \"%s\"", indexResult, relatedMediaFile.GetType(), relatedMediaFile.GetRelativeFilename(i.originalsPath))
	}

	return indexed
}

func (i *Indexer) IndexAll() map[string]bool {
	indexed := make(map[string]bool)

	err := filepath.Walk(i.originalsPath, func(filename string, fileInfo os.FileInfo, err error) error {
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
		log.Print(err.Error())
	}

	return indexed
}
