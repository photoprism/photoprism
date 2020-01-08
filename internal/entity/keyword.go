package entity

import (
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/mutex"
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
	mutex.Db.Lock()
	defer mutex.Db.Unlock()

	if err := db.FirstOrCreate(m, "keyword = ?", m.Keyword).Error; err != nil {
		log.Errorf("keyword: %s", err)
	}

	return m
}
