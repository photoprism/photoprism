package entity

import (
	"errors"
	"strings"
	"sync"

	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/pkg/txt"
)

var keywordMutex = sync.Mutex{}

// Keyword represents a normalized word used for full-text search and tagging.
type Keyword struct {
	ID      uint   `gorm:"primary_key" json:"ID,omitempty" yaml:"ID,omitempty"`
	Keyword string `gorm:"type:VARCHAR(64);index;" json:"Keyword" yaml:"Keyword,omitempty"`
	Skip    bool   `json:"Skip" yaml:"Skip"`
}

// TableName returns the entity table name.
func (Keyword) TableName() string {
	return "keywords"
}

// NewKeyword returns a normalized keyword entity ready for persistence.
func NewKeyword(keyword string) *Keyword {
	keyword = strings.ToLower(txt.Clip(keyword, txt.ClipKeyword))

	result := &Keyword{
		Keyword: keyword,
	}

	return result
}

// Update modifies a single column on an already persisted keyword and relies on
// the standard GORM callback to evict the cached instance afterwards.
func (m *Keyword) Update(attr string, value interface{}) error {
	if m == nil {
		return errors.New("keyword must not be nil - you may have found a bug")
	} else if !m.HasID() {
		return errors.New("keyword ID must not be empty - you may have found a bug")
	}

	// Omit FlushCachedKeyword() because this should automatically trigger the AfterUpdate() hook.
	return UnscopedDb().Model(m).Update(attr, value).Error
}

// Updates applies a set of column changes to an existing keyword while keeping
// the cache consistent via the AfterUpdate hook.
func (m *Keyword) Updates(values interface{}) error {
	if values == nil {
		return nil
	} else if m == nil {
		return errors.New("keyword must not be nil - you may have found a bug")
	} else if !m.HasID() {
		return errors.New("keyword ID must not be empty - you may have found a bug")
	}

	// Omit FlushCachedKeyword() because this should automatically trigger the AfterUpdate() hook.
	return UnscopedDb().Model(m).Updates(values).Error
}

// Save updates the record in the database or inserts a new record if it does not already exist.
func (m *Keyword) Save() error {
	return Db().Save(m).Error
}

// Create inserts a new row to the database.
func (m *Keyword) Create() error {
	keywordMutex.Lock()
	defer keywordMutex.Unlock()

	return Db().Create(m).Error
}

// AfterUpdate flushes the cache when the entity is updated.
func (m *Keyword) AfterUpdate(tx *gorm.DB) (err error) {
	FlushCachedKeyword(m)
	return
}

// AfterDelete flushes the cache when the entity is deleted.
func (m *Keyword) AfterDelete(tx *gorm.DB) (err error) {
	FlushCachedKeyword(m)
	return
}

// AfterCreate flushes the cache when the entity is created.
func (m *Keyword) AfterCreate(scope *gorm.Scope) error {
	FlushCachedKeyword(m)
	return nil
}

// HasID reports whether the keyword has already been persisted and assigned
// a primary key so callers can skip duplicate writes or lookups.
func (m *Keyword) HasID() bool {
	if m == nil {
		return false
	}
	return m.ID > 0
}

// FirstOrCreateKeyword returns the existing row, inserts a new row or nil in case of errors.
func FirstOrCreateKeyword(m *Keyword) *Keyword {
	if result, err := FindKeyword(m.Keyword, true); err == nil {
		return result
	} else if createErr := m.Create(); createErr == nil {
		return m
	} else if result, err = FindKeyword(m.Keyword, false); err == nil {
		return result
	} else {
		log.Errorf("keyword: %s (find or create %s)", createErr, m.Keyword)
	}

	return nil
}
