package models

import "github.com/jinzhu/gorm"

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
	db.FirstOrCreate(m, "photo_id = ? AND keyword_id = ?", m.PhotoID, m.KeywordID)

	return m
}
