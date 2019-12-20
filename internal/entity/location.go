package entity

import (
	"strings"
	"time"

	olc "github.com/google/open-location-code/go"
	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/maps"
	"github.com/photoprism/photoprism/internal/util"
)

// Photo location
type Location struct {
	maps.Location
	LocDescription string `gorm:"type:text;"`
	LocNotes       string `gorm:"type:text;"`
	LocFavorite    bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewLocation(lat, lng float64) *Location {
	result := &Location{}

	result.ID = olc.Encode(lat, lng, 11)
	result.LocLat = lat
	result.LocLng = lng

	return result
}

func (m *Location) Label() string {
	return m.LocLabel
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
		strings.ToLower(m.LocLabel),
	}

	result = append(result, util.Keywords(m.LocTitle)...)
	result = append(result, util.Keywords(m.LocDescription)...)
	result = append(result, util.Keywords(m.LocNotes)...)

	return result
}
