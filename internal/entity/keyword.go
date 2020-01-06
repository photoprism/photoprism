package entity

import (
	"strings"

	"github.com/jinzhu/gorm"
)

// Keyword for full text search
type Keyword struct {
	ID      uint   `gorm:"primary_key"`
	Keyword string `gorm:"type:varchar(64);index;"`
	Skip    bool
}

func NewKeyword(keyword string) *Keyword {
	keyword = strings.ToLower(strings.TrimSpace(keyword))

	result := &Keyword{
		Keyword: keyword,
	}

	return result
}

func (m *Keyword) FirstOrCreate(db *gorm.DB) *Keyword {
	writeMutex.Lock()
	defer writeMutex.Unlock()
	if err := db.FirstOrCreate(m, "keyword = ?", m.Keyword).Error; err != nil {
		log.Errorf("keyword: %s", err)
	}

	return m
}
