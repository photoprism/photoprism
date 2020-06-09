package entity

import (
	"time"
)

// PhotoAlbum represents the many_to_many relation between Photo and Album
type PhotoAlbum struct {
	PhotoUID  string `gorm:"type:varbinary(42);primary_key;auto_increment:false"`
	AlbumUID  string `gorm:"type:varbinary(42);primary_key;auto_increment:false;index"`
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

// Create inserts a new row to the database.
func (m *PhotoAlbum) Create() error {
	return Db().Create(m).Error
}

// Save updates or inserts a row.
func (m *PhotoAlbum) Save() error {
	return Db().Save(m).Error
}

// FirstOrCreatePhotoAlbum returns the existing row, inserts a new row or nil in case of errors.
func FirstOrCreatePhotoAlbum(m *PhotoAlbum) *PhotoAlbum {
	result := PhotoAlbum{}

	if err := Db().Where("photo_uid = ? AND album_uid = ?", m.PhotoUID, m.AlbumUID).First(&result).Error; err == nil {
		return &result
	} else if err := m.Create(); err != nil {
		log.Errorf("photo-album: %s", err)
		return nil
	}

	return m
}
