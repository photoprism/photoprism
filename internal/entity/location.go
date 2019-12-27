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
	maps.Location
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

func (m *Location) Category() string {
	return m.LocCategory
}

func (m *Location) Find(db *gorm.DB) error {
	if err := db.First(m, "id = ?", m.ID).Error; err == nil {
		return err
	}

	if err := m.Query(); err != nil {
		return err
	}

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
