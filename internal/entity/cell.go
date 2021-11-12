package entity

import (
	"strings"
	"sync"
	"time"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/maps"

	"github.com/photoprism/photoprism/pkg/s2"
	"github.com/photoprism/photoprism/pkg/txt"
)

var cellMutex = sync.Mutex{}

// Cell represents a S2 cell with location data.
type Cell struct {
	ID           string    `gorm:"type:VARBINARY(42);primary_key;auto_increment:false;" json:"ID" yaml:"ID"`
	CellName     string    `gorm:"type:VARCHAR(200);" json:"Name" yaml:"Name,omitempty"`
	CellCategory string    `gorm:"type:VARCHAR(50);" json:"Category" yaml:"Category,omitempty"`
	PlaceID      string    `gorm:"type:VARBINARY(42);default:'zz'" json:"-" yaml:"PlaceID"`
	Place        *Place    `gorm:"PRELOAD:true" json:"Place" yaml:"-"`
	CreatedAt    time.Time `json:"CreatedAt" yaml:"-"`
	UpdatedAt    time.Time `json:"UpdatedAt" yaml:"-"`
}

// TableName returns the entity database table name.
func (Cell) TableName() string {
	return "cells"
}

// UnknownLocation is PhotoPrism's default location.
var UnknownLocation = Cell{
	ID:           UnknownID,
	Place:        &UnknownPlace,
	PlaceID:      UnknownID,
	CellName:     "",
	CellCategory: "",
}

// CreateUnknownLocation creates the default location if not exists.
func CreateUnknownLocation() {
	FirstOrCreateCell(&UnknownLocation)
}

// NewCell creates a location using a token extracted from coordinate
func NewCell(lat, lng float32) *Cell {
	result := &Cell{}

	result.ID = s2.PrefixedToken(float64(lat), float64(lng))

	return result
}

// Refresh updates the index by retrieving the latest data from an external API.
func (m *Cell) Refresh(api string) (err error) {
	start := time.Now()

	// Unknown?
	if m.Unknown() {
		// Skip.
		return nil
	}

	// Initialize.
	l := &maps.Location{
		ID: s2.NormalizeToken(m.ID),
	}

	// Query geodata API.
	if err = l.QueryApi(api); err != nil {
		return err
	}

	// Unknown location or label missing?
	if l.Unknown() || l.Label() == "" {
		// Ignore.
		return nil
	}

	place := &Place{}

	// Find existing place by label.
	if err := UnscopedDb().Where("place_label = ?", l.Label()).First(&place).Error; err != nil {
		log.Tracef("places: %s for cell %s", err, m.ID)
		place = &Place{ID: m.ID}
	} else {
		log.Tracef("places: found matching place %s for cell %s", place.ID, m.ID)
	}

	// Update place.
	if res := UnscopedDb().Model(place).Updates(Values{
		"PlaceLabel":    l.Label(),
		"PlaceCity":     l.City(),
		"PlaceDistrict": l.District(),
		"PlaceState":    l.State(),
		"PlaceCountry":  l.CountryCode(),
		"PlaceKeywords": l.KeywordString(),
	}); res.Error == nil && res.RowsAffected == 1 {
		// Update cell place id, name, and category.
		log.Tracef("places: updating place, name, and category for cell %s", m.ID)
		m.PlaceID = place.ID
		err = UnscopedDb().Model(m).Updates(Values{"PlaceID": m.PlaceID, "CellName": l.Name(), "CellCategory": l.Category()}).Error

	} else {
		// Update cell name and category.
		log.Tracef("places: updating name and category for cell %s", m.ID)
		err = UnscopedDb().Model(m).Updates(Values{"CellName": l.Name(), "CellCategory": l.Category()}).Error
	}

	log.Debugf("places: refreshed cell %s [%s]", txt.Quote(m.ID), time.Since(start))

	return err
}

// Find retrieves location data from the database or an external api if not known already.
func (m *Cell) Find(api string) error {
	start := time.Now()
	db := Db()

	if err := db.Preload("Place").First(m, "id = ?", m.ID).Error; err == nil {
		log.Debugf("location: found cell %s", m.ID)
		return nil
	}

	l := &maps.Location{
		ID: s2.NormalizeToken(m.ID),
	}

	if err := l.QueryApi(api); err != nil {
		return err
	}

	if found := FindPlace(l.PrefixedToken(), l.Label()); found != nil {
		m.Place = found
	} else {
		place := &Place{
			ID:            l.PrefixedToken(),
			PlaceLabel:    l.Label(),
			PlaceCity:     l.City(),
			PlaceState:    l.State(),
			PlaceCountry:  l.CountryCode(),
			PlaceKeywords: l.KeywordString(),
			PhotoCount:    1,
		}

		if createErr := place.Create(); createErr == nil {
			event.Publish("count.places", event.Data{
				"count": 1,
			})

			log.Infof("location: added place %s [%s]", place.ID, time.Since(start))

			m.Place = place
		} else if found := FindPlace(l.PrefixedToken(), l.Label()); found != nil {
			m.Place = found
		} else {
			log.Errorf("location: %s (create place %s)", createErr, place.ID)
			m.Place = &UnknownPlace
		}
	}

	m.PlaceID = m.Place.ID
	m.CellName = l.Name()
	m.CellCategory = l.Category()

	cellMutex.Lock()
	defer cellMutex.Unlock()

	if createErr := db.Create(m).Error; createErr == nil {
		log.Debugf("location: added cell %s [%s]", m.ID, time.Since(start))
		return nil
	} else if findErr := db.Preload("Place").First(m, "id = ?", m.ID).Error; findErr != nil {
		log.Errorf("location: %s (create cell %s)", createErr, m.ID)
		log.Errorf("location: %s (find cell %s)", findErr, m.ID)
		return createErr
	} else {
		log.Debugf("location: found cell %s [%s]", m.ID, time.Since(start))
	}

	return nil
}

// Create inserts a new row to the database.
func (m *Cell) Create() error {
	return Db().Create(m).Error
}

// FirstOrCreateCell fetches an existing row, inserts a new row or nil in case of errors.
func FirstOrCreateCell(m *Cell) *Cell {
	if m.ID == "" {
		log.Errorf("location: cell must not be empty")
		return nil
	}

	if m.PlaceID == "" {
		log.Errorf("location: place must not be empty (find or create cell %s)", m.ID)
		return nil
	}

	result := Cell{}

	if findErr := Db().Where("id = ?", m.ID).Preload("Place").First(&result).Error; findErr == nil {
		return &result
	} else if createErr := m.Create(); createErr == nil {
		return m
	} else if err := Db().Where("id = ?", m.ID).Preload("Place").First(&result).Error; err == nil {
		return &result
	} else {
		log.Errorf("location: %s (find or create cell %s)", createErr, m.ID)
	}

	return nil
}

// Keywords returns search keywords for a location.
func (m *Cell) Keywords() (result []string) {
	if m.Place == nil {
		log.Errorf("location: place for cell %s is nil - you might have found a bug", m.ID)
		return result
	}

	result = append(result, txt.Keywords(txt.ReplaceSpaces(m.City(), "-"))...)
	result = append(result, txt.Keywords(txt.ReplaceSpaces(m.State(), "-"))...)
	result = append(result, txt.Keywords(txt.ReplaceSpaces(m.CountryName(), "-"))...)
	result = append(result, txt.Keywords(m.Category())...)
	result = append(result, txt.Keywords(m.Name())...)
	result = append(result, txt.Words(m.Place.PlaceKeywords)...)

	result = txt.UniqueWords(result)

	return result
}

// Unknown checks if the location has no id
func (m *Cell) Unknown() bool {
	return m.ID == "" || m.ID == UnknownLocation.ID
}

// Name returns name of location
func (m *Cell) Name() string {
	return m.CellName
}

// NoName checks if the location has no name
func (m *Cell) NoName() bool {
	return m.CellName == ""
}

// Category returns the location category
func (m *Cell) Category() string {
	return m.CellCategory
}

// NoCategory checks id the location has no category
func (m *Cell) NoCategory() bool {
	return m.CellCategory == ""
}

// Label returns the location place label
func (m *Cell) Label() string {
	return m.Place.Label()
}

// City returns the location place city
func (m *Cell) City() string {
	return m.Place.City()
}

// LongCity checks if the city name is more than 16 char
func (m *Cell) LongCity() bool {
	return len(m.City()) > 16
}

// NoCity checks if the location has no city
func (m *Cell) NoCity() bool {
	return m.City() == ""
}

// CityContains checks if the location city contains the text string
func (m *Cell) CityContains(text string) bool {
	return strings.Contains(text, m.City())
}

// State returns the location place state
func (m *Cell) State() string {
	return m.Place.State()
}

// NoState checks if the location place has no state
func (m *Cell) NoState() bool {
	return m.Place.State() == ""
}

// CountryCode returns the location place country code
func (m *Cell) CountryCode() string {
	return m.Place.CountryCode()
}

// CountryName returns the location place country name
func (m *Cell) CountryName() string {
	return m.Place.CountryName()
}
