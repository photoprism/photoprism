package photoprism

import (
	"github.com/jinzhu/gorm"
	"os"
	"strings"
	"path/filepath"
	"log"
	"io/ioutil"
	"github.com/photoprism/photoprism/recognize"
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
			if tag.Probability > 0.2 {
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
	var file File

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
			}

			photo.Tags = i.GetImageTags(jpeg)
		}

		if location, err := mediaFile.GetLocation(); err == nil {
			i.db.FirstOrCreate(location, "id = ?", location.ID)
			photo.Location = location
		}

		photo.Camera = NewCamera(mediaFile.GetCameraModel()).FirstOrCreate(i.db)
		photo.TakenAt = mediaFile.GetDateCreated()
		photo.CanonicalName = canonicalName
		photo.Files = []File{}
		photo.Albums = []Album{}
		photo.Author = ""

		photo.Favorite = false
		photo.Private = true
		photo.Deleted = false

		i.db.Create(&photo)
	}

	if result := i.db.First(&file, "hash = ?", fileHash); result.Error != nil {
		file.PhotoID = photo.ID
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
