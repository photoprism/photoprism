package entity

import (
	"time"
)

// PhotoAlbum represents the many_to_many relation between Photo and Album
type PhotoAlbum struct {
	PhotoUID  string `gorm:"type:varbinary(36);primary_key;auto_increment:false"`
	AlbumUID  string `gorm:"type:varbinary(36);primary_key;auto_increment:false;index"`
	Order     int
	Hidden    bool
	CreatedAt time.Time
	UpdatedAt time.Time
	Photo     *Photo `gorm:"PRELOAD:false"`
	Album     *Album `gorm:"PRELOAD:true"`
}

// TableName returns PhotoAlbum table identifier "photos_albums"
func (PhotoAlbum) TableName() string {
	return "photos_albums"
}

// NewPhotoAlbum registers an photo and album association using UID
func NewPhotoAlbum(photoUID, albumUID string) *PhotoAlbum {
	result := &PhotoAlbum{
		PhotoUID: photoUID,
		AlbumUID: albumUID,
	}

	return result
}

// FirstOrCreate checks if the PhotoAlbum relation already exist in the database before the creation
func (m *PhotoAlbum) FirstOrCreate() *PhotoAlbum {
	if err := Db().FirstOrCreate(m, "photo_uid = ? AND album_uid = ?", m.PhotoUID, m.AlbumUID).Error; err != nil {
		log.Errorf("photo album: %s", err)
	}

	return m
}
