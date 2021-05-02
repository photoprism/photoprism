package query

import (
	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

func filter(in *gorm.DB, f form.DbSearch) (out *gorm.DB) {
	out = in.Where("(deleted_at IS NOT NULL AND deleted_at >= ?) OR (updated_at >= ?)", f.Since, f.Since)
	return out.Offset(f.Offset).Limit(f.Count)
}

// TablePhotos gets all entries from photos table
func TablePhotos(f form.DbSearch) (photos []entity.Photo, err error) {
	s := Db().Unscoped().Table("photos")
	s = filter(s, f)
	result := s.Find(&photos)
	return photos, result.Error
}

// TableFiles gets all entries from files table
func TableFiles(f form.DbSearch) (files []entity.File, err error) {
	s := Db().Unscoped().Table("files")
	s = filter(s, f)
	result := s.Find(&files)
	return files, result.Error
}

// TableAlbums gets all entries from albums table
func TableAlbums(f form.DbSearch) (albums []entity.Album, err error) {
	s := Db().Unscoped().Table("albums")
	s = filter(s, f)
	result := s.Find(&albums)
	return albums, result.Error
}

// TablePhotosAlbums gets all entries from photos_albums table
func TablePhotosAlbums(f form.DbSearch) (photosAlbums []entity.PhotoAlbum, err error) {
	s := Db().Unscoped().Table("photos_albums")
	s = s.Where("updated_at >= ?", f.Since).Offset(f.Offset).Limit(f.Count)
	result := s.Find(&photosAlbums)
	return photosAlbums, result.Error
}
