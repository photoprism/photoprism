package entity

import (
	"errors"
	"reflect"
	"strings"
	"sync"

	"github.com/photoprism/photoprism/pkg/txt"
)

var keywordMutex = sync.Mutex{}

// Keyword used for full text search
type Keyword struct {
	ID      uint   `gorm:"primaryKey;"`
	Keyword string `gorm:"type:VARCHAR(64);index;"`
	Skip    bool
}

// TableName returns the entity table name.
func (Keyword) TableName() string {
	return "keywords"
}

// NewKeyword registers a new keyword in database
func NewKeyword(keyword string) *Keyword {
	keyword = strings.ToLower(txt.Clip(keyword, txt.ClipKeyword))

	result := &Keyword{
		Keyword: keyword,
	}

	return result
}

// Retrieves the ID value from the values interface
func getIDInValuesOrZero(values any) (result any) {
	switch reflect.TypeOf(values).Kind() {
	case reflect.Struct:
		t := reflect.TypeOf(values)
		v := reflect.ValueOf(values)
		for i := 0; i < t.NumField(); i++ {
			if t.Field(i).Name == "ID" {
				return v.Field(i).Interface()
			}
		}
	default:
		log.Errorf("Unsupported Type %v in getIDInValuesOrZero", reflect.TypeOf(values).Kind())
	}
	return 0
}

// Updates multiple columns in the database.
func (m *Keyword) Updates(values interface{}) error {
	if m.ID == 0 {
		id := getIDInValuesOrZero(values)
		if id != 0 {
			return UnscopedDb().Model(m).
				Where("id = ?", id).
				UpdateColumns(values).Error
		} else {
			return errors.New("PK value not provided")
		}
	} else {
		return UnscopedDb().Model(m).UpdateColumns(values).Error
	}
}

// Update a column in the database.
func (m *Keyword) Update(attr string, value interface{}) error {
	if m.ID == 0 {
		return errors.New("PK value not provided in model")
	} else {
		return UnscopedDb().Model(m).UpdateColumn(attr, value).Error
	}
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

// FirstOrCreateKeyword returns the existing row, inserts a new row or nil in case of errors.
func FirstOrCreateKeyword(m *Keyword) *Keyword {
	result := Keyword{}

	if err := Db().Where("keyword = ?", m.Keyword).First(&result).Error; err == nil {
		return &result
	} else if createErr := m.Create(); createErr == nil {
		return m
	} else if err := Db().Where("keyword = ?", m.Keyword).First(&result).Error; err == nil {
		return &result
	} else {
		log.Errorf("keyword: %s (find or create %s)", createErr, m.Keyword)
	}

	return nil
}
