package photoprism

import (
	"github.com/jinzhu/gorm"
)

type Search struct {
	originalsPath string
	db            *gorm.DB
}

func NewQuery(originalsPath string, db *gorm.DB) *Search {
	instance := &Search{
		originalsPath: originalsPath,
		db:            db,
	}

	return instance
}

func (s *Search) FindPhotos (count int, offset int) (photos []Photo) {
	s.db.Preload("Tags").Preload("Files").Preload("Albums").Where(&Photo{Deleted: false}).Limit(count).Offset(offset).Find(&photos)

	return photos
}

func (s *Search) FindFiles (count int, offset int) (files []File) {
	s.db.Where(&File{}).Limit(count).Offset(offset).Find(&files)

	return files
}

func (s *Search) FindFile (id string) (file File) {
	s.db.Where("id = ?", id).First(&file)

	return file
}