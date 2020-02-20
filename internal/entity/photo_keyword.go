package entity

import (
	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/mutex"
)

// PhotoKeyword represents the many-to-many relation between Photo and Keyword
type PhotoKeyword struct {
	PhotoID   uint `gorm:"primary_key;auto_increment:false"`
	KeywordID uint `gorm:"primary_key;auto_increment:false;index"`
}

// TableName returns PhotoKeyword table identifier "photos_keywords"
func (PhotoKeyword) TableName() string {
	return "photos_keywords"
}

// NewPhotoKeyword register a new PhotoKeyword relation
func NewPhotoKeyword(photoID, keywordID uint) *PhotoKeyword {
	result := &PhotoKeyword{
		PhotoID:   photoID,
		KeywordID: keywordID,
	}

	return result
}

// FirstOrCreate check wether the PhotoKeywords relation already exist in the database before the creation
func (m *PhotoKeyword) FirstOrCreate(db *gorm.DB) *PhotoKeyword {
	mutex.Db.Lock()
	defer mutex.Db.Unlock()

	if err := db.FirstOrCreate(m, "photo_id = ? AND keyword_id = ?", m.PhotoID, m.KeywordID).Error; err != nil {
		log.Errorf("photo keyword: %s", err)
	}

	return m
}
