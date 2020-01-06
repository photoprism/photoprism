package entity

import (
	"time"

	"github.com/jinzhu/gorm"
)

// Photos can be added to multiple albums
type PhotoAlbum struct {
	PhotoUUID string `gorm:"type:varbinary(36);primary_key;auto_increment:false"`
	AlbumUUID string `gorm:"type:varbinary(36);primary_key;auto_increment:false;index"`
	Order     int
	CreatedAt time.Time
	UpdatedAt time.Time
	Photo     *Photo
	Album     *Album
}

func (PhotoAlbum) TableName() string {
	return "photos_albums"
}

func NewPhotoAlbum(photoUUID, albumUUID string) *PhotoAlbum {
	result := &PhotoAlbum{
		PhotoUUID: photoUUID,
		AlbumUUID: albumUUID,
	}

	return result
}

func (m *PhotoAlbum) FirstOrCreate(db *gorm.DB) *PhotoAlbum {
	writeMutex.Lock()
	defer writeMutex.Unlock()
	if err := db.FirstOrCreate(m, "photo_uuid = ? AND album_uuid = ?", m.PhotoUUID, m.AlbumUUID).Error; err != nil {
		log.Errorf("photo album: %s", err)
	}

	return m
}
