package entity

import (
	"time"
)

type PhotoAlbums []PhotoAlbum

// PhotoAlbum represents the many_to_many relation between Photo and Album
type PhotoAlbum struct {
	PhotoUID  string    `gorm:"type:VARBINARY(42);primary_key;auto_increment:false" json:"PhotoUID" yaml:"UID"`
	AlbumUID  string    `gorm:"type:VARBINARY(42);primary_key;auto_increment:false;index" json:"AlbumUID" yaml:"-"`
	Order     int       `json:"Order" yaml:"Order,omitempty"`
	Hidden    bool      `json:"Hidden" yaml:"Hidden,omitempty"`
	Missing   bool      `json:"Missing" yaml:"Missing,omitempty"`
	CreatedAt time.Time `json:"CreatedAt" yaml:"CreatedAt,omitempty"`
	UpdatedAt time.Time `json:"UpdatedAt" yaml:"-"`
	Photo     *Photo    `gorm:"PRELOAD:false" yaml:"-"`
	Album     *Album    `gorm:"PRELOAD:true" yaml:"-"`
}

// TableName returns the entity table name.
func (PhotoAlbum) TableName() string {
	return "photos_albums"
}

// NewPhotoAlbum creates a new photo and album mapping with UIDs.
func NewPhotoAlbum(photoUid, albumUid string) *PhotoAlbum {
	result := &PhotoAlbum{
		PhotoUID: photoUid,
		AlbumUID: albumUid,
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
