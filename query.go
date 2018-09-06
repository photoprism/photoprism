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

func (s *Search) FindPhotos(query string, count int, offset int) (photos []Photo) {
	q := s.db.Preload("Tags").Preload("Files").Preload("Location").Preload("Albums")

	if query != "" {
		q = q.Joins("JOIN photo_tags ON photo_tags.photo_id=photos.id")
		q = q.Joins("JOIN tags ON photo_tags.tag_id=tags.id")
		q = q.Where("tags.label LIKE ?", "%"+query+"%")
	}

	q = q.Where(&Photo{Deleted: false}).Order("taken_at").Limit(count).Offset(offset)
	q = q.Find(&photos)

	return photos
}

func (s *Search) FindFiles(count int, offset int) (files []File) {
	s.db.Where(&File{}).Limit(count).Offset(offset).Find(&files)

	return files
}

func (s *Search) FindFile(id string) (file File) {
	s.db.Where("id = ?", id).First(&file)

	return file
}
