package entity

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/maps"
)

// Photo place
type Place struct {
	ID          uint64 `gorm:"type:BIGINT;primary_key;auto_increment:false;"`
	LocLabel    string `gorm:"type:varbinary(500);unique_index;"`
	LocCity     string `gorm:"type:varchar(100);"`
	LocState    string `gorm:"type:varchar(100);"`
	LocCountry  string `gorm:"type:binary(2);"`
	LocNotes    string `gorm:"type:text;"`
	LocFavorite bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	New         bool `gorm:"-"`
}

var UnknownPlace = NewPlace(1, "Unknown", "Unknown", "Unknown", "zz")

func CreateUnknownPlace(db *gorm.DB) {
	UnknownPlace.FirstOrCreate(db)
}

func (m *Place) AfterCreate(scope *gorm.Scope) error {
	return scope.SetColumn("New", true)
}

func FindPlace(id uint64, db *gorm.DB) *Place {
	place := &Place{}

	if err := db.First(place, "id = ?", id).Error; err != nil {
		log.Debugf("place: %s for id %d", err.Error(), id)
	}

	return place
}

func FindPlaceByLabel(id uint64, label string, db *gorm.DB) *Place {
	place := &Place{}

	if err := db.First(place, "id = ? OR loc_label = ?", id, label).Error; err != nil {
		log.Debugf("place: %s for id %d or label \"%s\"", err.Error(), id, label)
	}

	return place
}

func NewPlace(id uint64, label, city, state, countryCode string) *Place {
	result := &Place{
		ID:         id,
		LocLabel:   label,
		LocCity:    city,
		LocState:   state,
		LocCountry: countryCode,
	}

	return result
}

func (m *Place) Find(db *gorm.DB) error {
	if err := db.First(m, "id = ?", m.ID).Error; err != nil {
		return err
	}

	return nil
}

func (m *Place) FirstOrCreate(db *gorm.DB) *Place {
	if err := db.FirstOrCreate(m, "id = ? OR loc_label = ?", m.ID, m.LocLabel).Error; err != nil {
		log.Debugf("place: %s for id %d or label \"%s\"", err.Error(), m.ID, m.LocLabel)
	}

	return m
}

func (m *Place) NoID() bool {
	return m.ID == 0
}

func (m *Place) Label() string {
	return m.LocLabel
}

func (m *Place) City() string {
	return m.LocCity
}

func (m *Place) State() string {
	return m.LocState
}

func (m *Place) CountryCode() string {
	return m.LocCountry
}

func (m *Place) CountryName() string {
	return maps.CountryNames[m.LocCountry]
}

func (m *Place) Notes() string {
	return m.LocNotes
}
