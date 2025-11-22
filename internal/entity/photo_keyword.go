package entity

import (
	"github.com/jinzhu/gorm"
)

// PhotoKeyword represents the many-to-many relation between Photo and Keyword.
type PhotoKeyword struct {
	PhotoID   uint `gorm:"primary_key;auto_increment:false"`
	KeywordID uint `gorm:"primary_key;auto_increment:false;index"`
}

// TableName returns the entity table name.
func (PhotoKeyword) TableName() string {
	return "photos_keywords"
}

// NewPhotoKeyword returns a new PhotoKeyword relation ready for persistence.
func NewPhotoKeyword(photoID, keywordID uint) *PhotoKeyword {
	result := &PhotoKeyword{
		PhotoID:   photoID,
		KeywordID: keywordID,
	}

	return result
}

// Create inserts a new row to the database.
func (m *PhotoKeyword) Create() error {
	return Db().Create(m).Error
}

// AfterCreate flushes the keyword cache once a relation has been persisted.
func (m *PhotoKeyword) AfterCreate(scope *gorm.Scope) error {
	FlushCachedPhotoKeyword(m)
	return nil
}

// AfterUpdate flushes the keyword cache after a relation change.
func (m *PhotoKeyword) AfterUpdate(tx *gorm.DB) (err error) {
	FlushCachedPhotoKeyword(m)
	return
}

// Delete removes the keyword reference and clears the cache.
func (m *PhotoKeyword) Delete() error {
	FlushCachedPhotoKeyword(m)
	return Db().Delete(m).Error
}

// AfterDelete flushes the keyword cache when the photo-keyword relation is removed.
func (m *PhotoKeyword) AfterDelete(tx *gorm.DB) (err error) {
	FlushCachedPhotoKeyword(m)
	return
}

// HasID reports whether both sides of the relation have identifiers assigned,
// meaning the join row exists (or is ready to be cached) in the database.
func (m *PhotoKeyword) HasID() bool {
	if m == nil {
		return false
	}
	return m.PhotoID > 0 && m.KeywordID > 0
}

// CacheKey returns a string key for caching the entity.
func (m *PhotoKeyword) CacheKey() string {
	if m == nil {
		return ""
	}
	return photoKeywordCacheKey(m.PhotoID, m.KeywordID)
}

// FirstOrCreatePhotoKeyword returns the existing row, inserts a new row or nil in case of errors.
func FirstOrCreatePhotoKeyword(m *PhotoKeyword) *PhotoKeyword {
	if result, err := FindPhotoKeyword(m.PhotoID, m.KeywordID, true); err == nil {
		return result
	} else if createErr := m.Create(); createErr == nil {
		return m
	} else if result, err = FindPhotoKeyword(m.PhotoID, m.KeywordID, false); err == nil {
		return result
	} else {
		log.Errorf("photo-keyword: %s (find or create)", createErr)
	}

	return nil
}
