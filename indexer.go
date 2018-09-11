package photoprism

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/recognize"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Indexer struct {
	originalsPath string
	db            *gorm.DB
}

func NewIndexer(originalsPath string, db *gorm.DB) *Indexer {
	instance := &Indexer{
		originalsPath: originalsPath,
		db:            db,
	}

	return instance
}

func (i *Indexer) GetImageTags(jpeg *MediaFile) (result []Tag) {
	if imageBuffer, err := ioutil.ReadFile(jpeg.filename); err == nil {
		tags, err := recognize.GetImageTags(string(imageBuffer))

		if err != nil {
			return result
		}

		for _, tag := range tags {
			if tag.Probability > 0.2 { // TODO: Use config variable
				var tagModel Tag

				if res := i.db.First(&tagModel, "label = ?", tag.Label); res.Error != nil {
					tagModel.Label = tag.Label
				}

				result = append(result, tagModel)
			}
		}
	}

	return result
}

func (i *Indexer) IndexMediaFile(mediaFile *MediaFile) {
	var photo Photo
	var file, primaryFile File
	var isPrimary = false
	var colorNames []string
	var keywords []string

	canonicalName := mediaFile.GetCanonicalNameFromFile()
	fileHash := mediaFile.GetHash()

	if result := i.db.First(&photo, "canonical_name = ?", canonicalName); result.Error != nil {
		if jpeg, err := mediaFile.GetJpeg(); err == nil {
			if perceptualHash, err := jpeg.GetPerceptualHash(); err == nil {
				photo.PerceptualHash = perceptualHash
			}

			if exifData, err := jpeg.GetExifData(); err == nil {
				photo.Lat = exifData.Lat
				photo.Long = exifData.Long
				photo.Artist = exifData.Artist
			}

			colorNames, photo.VibrantColor, photo.MutedColor = jpeg.GetColors()

			photo.ColorNames = strings.Join(colorNames, ", ")

			photo.Tags = i.GetImageTags(jpeg)

			for _, tag := range photo.Tags {
				keywords = append(keywords, tag.Label)
			}
		}

		if location, err := mediaFile.GetLocation(); err == nil {
			i.db.FirstOrCreate(location, "id = ?", location.ID)
			photo.Location = location
			keywords = append(keywords, location.City, location.County, location.Country, location.LocationCategory)

			if location.Name != "" {
				photo.Title = fmt.Sprintf("%s / %s / %s", location.Name, location.Country, mediaFile.GetDateCreated().Format("2006"))
				keywords = append(keywords, location.Name)
			} else if location.City != "" {
				photo.Title = fmt.Sprintf("%s / %s / %s", location.City, location.Country, mediaFile.GetDateCreated().Format("2006"))
			} else if location.County != "" {
				photo.Title = fmt.Sprintf("%s / %s / %s", location.County, location.Country, mediaFile.GetDateCreated().Format("2006"))
			}

			if location.LocationType != "" {
				keywords = append(keywords, location.LocationType)
			}
		}

		if photo.Title == "" {
			if len(photo.Tags) > 0 {
				photo.Title = fmt.Sprintf("%s / %s", strings.Title(photo.Tags[0].Label), mediaFile.GetDateCreated().Format("2006"))
			} else {
				photo.Title = fmt.Sprintf("Unknown / %s", mediaFile.GetDateCreated().Format("2006"))
			}
		}

		photo.Keywords = strings.ToLower(strings.Join(keywords, ", "))
		photo.Camera = NewCamera(mediaFile.GetCameraModel()).FirstOrCreate(i.db)
		photo.TakenAt = mediaFile.GetDateCreated()
		photo.CanonicalName = canonicalName
		photo.Files = []File{}
		photo.Albums = []Album{}

		photo.Favorite = false
		photo.Private = true
		photo.Deleted = false

		i.db.Create(&photo)
	}

	if result := i.db.Where("file_type = 'jpg' AND primary_file = 1 AND photo_id = ?", photo.ID).First(&primaryFile); result.Error != nil {
		isPrimary = mediaFile.GetType() == FileTypeJpeg
	}

	if result := i.db.First(&file, "hash = ?", fileHash); result.Error != nil {
		file.PhotoID = photo.ID
		file.PrimaryFile = isPrimary
		file.Filename = mediaFile.GetFilename()
		file.Hash = fileHash
		file.FileType = mediaFile.GetType()
		file.MimeType = mediaFile.GetMimeType()
		file.Orientation = mediaFile.GetOrientation()

		if mediaFile.GetWidth() > 0 && mediaFile.GetHeight() > 0 {
			file.Width = mediaFile.GetWidth()
			file.Height = mediaFile.GetHeight()
			file.AspectRatio = mediaFile.GetAspectRatio()
		}

		i.db.Create(&file)
	}
}

func (i *Indexer) IndexAll() {
	err := filepath.Walk(i.originalsPath, func(filename string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if fileInfo.IsDir() || strings.HasPrefix(filepath.Base(filename), ".") {
			return nil
		}

		mediaFile := NewMediaFile(filename)

		if !mediaFile.Exists() || !mediaFile.IsPhoto() {
			return nil
		}

		relatedFiles, _, _ := mediaFile.GetRelatedFiles()

		for _, relatedMediaFile := range relatedFiles {
			log.Printf("Indexing %s", relatedMediaFile.GetFilename())
			i.IndexMediaFile(relatedMediaFile)
		}

		return nil
	})

	if err != nil {
		log.Print(err.Error())
	}
}
