package entity

import (
	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/mutex"
)

type PhotoKeyword struct {
	PhotoID   uint `gorm:"primary_key;auto_increment:false"`
	KeywordID uint `gorm:"primary_key;auto_increment:false;index"`
}

func (PhotoKeyword) TableName() string {
	return "photos_keywords"
}

func NewPhotoKeyword(photoID, keywordID uint) *PhotoKeyword {
	result := &PhotoKeyword{
		PhotoID:   photoID,
		KeywordID: keywordID,
	}

	return result
}

func (m *PhotoKeyword) FirstOrCreate(db *gorm.DB) *PhotoKeyword {
	mutex.Db.Lock()
	defer mutex.Db.Unlock()

	if err := db.FirstOrCreate(m, "photo_id = ? AND keyword_id = ?", m.PhotoID, m.KeywordID).Error; err != nil {
		log.Errorf("photo keyword: %s", err)
	}

	return m
}
