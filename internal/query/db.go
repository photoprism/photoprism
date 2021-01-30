package query

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

func filterAndSort(in *gorm.DB, deleted bool, since time.Time, count uint16) (out *gorm.DB) {
	if deleted {
		out = in.Where("deleted_at IS NOT NULL AND deleted_at >= ?", since).Order("deleted_at ASC")
	} else {
		out = in.Where("updated_at >= ?", since).Order("updated_at ASC")
	}
	return out.Limit(count)
}

// TablePhotos gets all entries from photos table
func TablePhotos(f form.DbSearch) (photos []entity.Photo, err error) {
	s := Db().Unscoped().Table("photos")
	s = filterAndSort(s, f.Deleted, f.Since, f.Count)
	result := s.Find(&photos)
	return photos, result.Error
}

// TableFiles gets all entries from files table
func TableFiles(f form.DbSearch) (files []entity.File, err error) {
	s := Db().Unscoped().Table("files")
	s = filterAndSort(s, f.Deleted, f.Since, f.Count)
	result := s.Find(&files)
	return files, result.Error
}

// TableAlbums gets all entries from albums table
func TableAlbums(f form.DbSearch) (albums []entity.Album, err error) {
	s := Db().Unscoped().Table("albums")
	s = filterAndSort(s, f.Deleted, f.Since, f.Count)
	result := s.Find(&albums)
	return albums, result.Error
}

// TablePhotosAlbums gets all entries from photos_albums table
func TablePhotosAlbums(f form.DbSearch) (photosAlbums []entity.PhotoAlbum, err error) {
	s := Db().Unscoped().Table("photos_albums")
	s = filterAndSort(s, f.Deleted, f.Since, f.Count)
	result := s.Find(&photosAlbums)
	return photosAlbums, result.Error
}
