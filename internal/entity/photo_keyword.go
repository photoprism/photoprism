package entity

// PhotoKeyword represents the many-to-many relation between Photo and Keyword
type PhotoKeyword struct {
	PhotoID   uint `gorm:"primary_key;auto_increment:false"`
	KeywordID uint `gorm:"primary_key;auto_increment:false;index"`
}

// TableName returns the entity table name.
func (PhotoKeyword) TableName() string {
	return "photos_keywords"
}

// NewPhotoKeyword registers a new PhotoKeyword relation
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

// FirstOrCreatePhotoKeyword returns the existing row, inserts a new row or nil in case of errors.
func FirstOrCreatePhotoKeyword(m *PhotoKeyword) *PhotoKeyword {
	result := PhotoKeyword{}

	if err := Db().Where("photo_id = ? AND keyword_id = ?", m.PhotoID, m.KeywordID).First(&result).Error; err == nil {
		return &result
	} else if createErr := m.Create(); createErr == nil {
		return m
	} else if err := Db().Where("photo_id = ? AND keyword_id = ?", m.PhotoID, m.KeywordID).First(&result).Error; err == nil {
		return &result
	} else {
		log.Errorf("photo-keyword: %s (find or create)", createErr)
	}

	return nil
}
