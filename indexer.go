package photoprism

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/recognize"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
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

func (i *Indexer) GetImageTags(jpeg *MediaFile) (results []*Tag) {
	if imageBuffer, err := ioutil.ReadFile(jpeg.filename); err == nil {
		tags, err := recognize.GetImageTags(string(imageBuffer))

		if err != nil {
			return results
		}

		for _, tag := range tags {
			if tag.Probability > 0.2 { // TODO: Use config variable
				results = i.appendTag(results, tag.Label)
			}
		}
	}

	return results
}

func getKeywordWithSynonyms(keyword string) []string {
	var result []string

	// TODO: Just a proof-of-concept for now, needs implementation via config file or dictionary
	switch keyword {
	case "tabby":
		result = []string{keyword, "cat"}
	case "lynx":
		result = []string{keyword, "cat"}
	case "tiger":
		result = []string{keyword, "cat"}
	default:
		result = []string{keyword}
	}

	return result
}

func getKeywordsAsString(keywords []string) string {
	var result []string

	for _, keyword := range keywords {
		result = append(result, getKeywordWithSynonyms(keyword)...)
	}

	result = uniqueStrings(result)
	sort.Strings(result)

	return strings.ToLower(strings.Join(result, ", "))
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

func (i *Indexer) IndexMediaFile(mediaFile *MediaFile) {
	var photo Photo
	var file, primaryFile File
	var isPrimary = false
	var colorNames []string
	var tags []*Tag

	canonicalName := mediaFile.GetCanonicalNameFromFile()
	fileHash := mediaFile.GetHash()

	if result := i.db.First(&photo, "photo_canonical_name = ?", canonicalName); result.Error != nil {
		if jpeg, err := mediaFile.GetJpeg(); err == nil {
			// Perceptual Hash
			if perceptualHash, err := jpeg.GetPerceptualHash(); err == nil {
				photo.PhotoPerceptualHash = perceptualHash
			}

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

			tags = i.appendTag(tags, location.LocCity)
			tags = i.appendTag(tags, location.LocCounty)
			tags = i.appendTag(tags, location.LocCountry)
			tags = i.appendTag(tags, location.LocCategory)
			tags = i.appendTag(tags, location.LocName)
			tags = i.appendTag(tags, location.LocType)

			if location.LocName != "" { // TODO: User defined title format
				photo.PhotoTitle = fmt.Sprintf("%s / %s / %s", location.LocName, location.LocCountry, mediaFile.GetDateCreated().Format("2006"))
			} else if location.LocCity != "" {
				photo.PhotoTitle = fmt.Sprintf("%s / %s / %s", location.LocCity, location.LocCountry, mediaFile.GetDateCreated().Format("2006"))
			} else if location.LocCounty != "" {
				photo.PhotoTitle = fmt.Sprintf("%s / %s / %s", location.LocCounty, location.LocCountry, mediaFile.GetDateCreated().Format("2006"))
			}
		}

		if photo.PhotoTitle == "" {
			if len(photo.Tags) > 0 { // TODO: User defined title format
				photo.PhotoTitle = fmt.Sprintf("%s / %s", strings.Title(photo.Tags[0].TagLabel), mediaFile.GetDateCreated().Format("2006"))
			} else {
				photo.PhotoTitle = fmt.Sprintf("Unknown / %s", mediaFile.GetDateCreated().Format("2006"))
			}
		}

		photo.Tags = tags
		photo.Camera = NewCamera(mediaFile.GetCameraModel()).FirstOrCreate(i.db)
		photo.TakenAt = mediaFile.GetDateCreated()
		photo.PhotoCanonicalName = canonicalName

		photo.PhotoFavorite = false

		i.db.Create(&photo)
	}

	if result := i.db.Where("file_type = 'jpg' AND file_primary = 1 AND photo_id = ?", photo.ID).First(&primaryFile); result.Error != nil {
		isPrimary = mediaFile.GetType() == FileTypeJpeg
	}

	if result := i.db.First(&file, "file_hash = ?", fileHash); result.Error != nil {
		file.PhotoID = photo.ID
		file.FilePrimary = isPrimary
		file.FileName = mediaFile.GetFilename()
		file.FileHash = fileHash
		file.FileType = mediaFile.GetType()
		file.FileMime = mediaFile.GetMimeType()
		file.FileOrientation = mediaFile.GetOrientation()

		if mediaFile.GetWidth() > 0 && mediaFile.GetHeight() > 0 {
			file.FileWidth = mediaFile.GetWidth()
			file.FileHeight = mediaFile.GetHeight()
			file.FileAspectRatio = mediaFile.GetAspectRatio()
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

		mediaFile, err := NewMediaFile(filename)

		if err != nil || !mediaFile.IsPhoto() {
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
