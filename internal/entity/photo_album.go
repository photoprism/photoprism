package entity

import (
	"time"
)

// PhotoAlbum represents the many_to_many relation between Photo and Album
type PhotoAlbum struct {
	PhotoUUID string `gorm:"type:varbinary(36);primary_key;auto_increment:false"`
	AlbumUUID string `gorm:"type:varbinary(36);primary_key;auto_increment:false;index"`
	Order     int
	Hidden    bool
	CreatedAt time.Time
	UpdatedAt time.Time
	Photo     *Photo
	Album     *Album
}

// TableName returns PhotoAlbum table identifier "photos_albums"
func (PhotoAlbum) TableName() string {
	return "photos_albums"
}

// NewPhotoAlbum registers an photo and album association using UUID
func NewPhotoAlbum(photoUUID, albumUUID string) *PhotoAlbum {
	result := &PhotoAlbum{
		PhotoUUID: photoUUID,
		AlbumUUID: albumUUID,
	}

	return result
}

// FirstOrCreate checks if the PhotoAlbum relation already exist in the database before the creation
func (m *PhotoAlbum) FirstOrCreate() *PhotoAlbum {
	if err := Db().FirstOrCreate(m, "photo_uuid = ? AND album_uuid = ?", m.PhotoUUID, m.AlbumUUID).Error; err != nil {
		log.Errorf("photo album: %s", err)
	}

	return m
}
