package entity

import (
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/maps"
	"github.com/photoprism/photoprism/internal/util"
)

// Photo location
type Location struct {
	ID          uint64 `gorm:"type:BIGINT;primary_key;auto_increment:false;"`
	LocLat      float64
	LocLng      float64
	LocName     string `gorm:"type:varchar(100);"`
	LocCategory string `gorm:"type:varchar(50);"`
	LocSuburb   string `gorm:"type:varchar(100);"`
	LocPlace    string `gorm:"type:varbinary(500);index;"`
	LocCity     string `gorm:"type:varchar(100);"`
	LocState    string `gorm:"type:varchar(100);"`
	LocCountry  string `gorm:"type:binary(2);"`
	LocSource   string `gorm:"type:varbinary(16);"`
	LocNotes    string `gorm:"type:text;"`
	LocFavorite bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewLocation(lat, lng float64) *Location {
	result := &Location{}

	result.ID = maps.S2Encode(lat, lng)
	result.LocLat = lat
	result.LocLng = lng

	return result
}

func (m *Location) Find(db *gorm.DB) error {
	if err := db.First(m, "id = ?", m.ID).Error; err == nil {
		return nil
	}

	l := &maps.Location{
		ID:     m.ID,
		LocLat: m.LocLat,
		LocLng: m.LocLng,
	}

	if err := l.Query(); err != nil {
		return err
	}

	m.LocName = l.LocName
	m.LocCategory = l.LocCategory
	m.LocSuburb = l.LocSuburb
	m.LocPlace = l.LocPlace
	m.LocCity = l.LocCity
	m.LocState = l.LocState
	m.LocCountry = l.LocCountry
	m.LocSource = l.LocSource

	if err := db.Create(m).Error; err != nil {
		log.Errorf("location: %s", err)
		return err
	}

	return nil
}

func (m *Location) Keywords() []string {
	result := []string{
		strings.ToLower(m.LocCity),
		strings.ToLower(m.LocSuburb),
		strings.ToLower(m.LocState),
		strings.ToLower(m.CountryName()),
		strings.ToLower(m.LocCategory),
	}

	result = append(result, util.Keywords(m.LocName)...)
	result = append(result, util.Keywords(m.LocPlace)...)
	result = append(result, util.Keywords(m.LocNotes)...)

	return result
}

func (m *Location) Unknown() bool {
	return m.ID == 0
}

func (m *Location) Latitude() float64 {
	return m.LocLat
}

func (m *Location) Longitude() float64 {
	return m.LocLng
}

func (m *Location) Name() string {
	return m.LocName
}

func (m *Location) Category() string {
	return m.LocCategory
}

func (m *Location) Suburb() string {
	return m.LocSuburb
}

func (m *Location) Place() string {
	return m.LocPlace
}

func (m *Location) City() string {
	return m.LocCity
}

func (m *Location) State() string {
	return m.LocState
}

func (m *Location) CountryCode() string {
	return m.LocCountry
}

func (m *Location) CountryName() string {
	return maps.CountryNames[m.LocCountry]
}

func (m *Location) Source() string {
	return m.LocSource
}
