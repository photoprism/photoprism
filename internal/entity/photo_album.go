package entity

import (
	"time"
)

// PhotoAlbums is a helper alias for collections of PhotoAlbum relations.
type PhotoAlbums []PhotoAlbum

// PhotoAlbum represents the many-to-many relation between Photo and Album.
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

// NewPhotoAlbum creates a new photo-to-album relation with the provided UIDs.
func NewPhotoAlbum(photoUid, albumUid string) *PhotoAlbum {
	result := &PhotoAlbum{
		PhotoUID: photoUid,
		AlbumUID: albumUid,
	}

	return result
}

// Create inserts a new row into the database.
func (m *PhotoAlbum) Create() error {
	return Db().Create(m).Error
}

// Save updates an existing relation or inserts a new one if needed.
func (m *PhotoAlbum) Save() error {
	return Db().Save(m).Error
}

// FirstOrCreatePhotoAlbum returns the persisted relation, creating it when necessary, or nil on failure.
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
