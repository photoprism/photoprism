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

// Cell represents an S2 cell with metadata and reference to a place.
type Cell struct {
	ID           string    `gorm:"type:VARBINARY(42);primary_key;auto_increment:false;" json:"ID" yaml:"ID"`
	CellName     string    `gorm:"type:VARCHAR(200);" json:"Name" yaml:"Name,omitempty"`
	CellStreet   string    `gorm:"type:VARCHAR(100);" json:"Street" yaml:"Street,omitempty"`
	CellPostcode string    `gorm:"type:VARCHAR(50);" json:"Postcode" yaml:"Postcode,omitempty"`
	CellCategory string    `gorm:"type:VARCHAR(50);" json:"Category" yaml:"Category,omitempty"`
	PlaceID      string    `gorm:"type:VARBINARY(42);default:'zz'" json:"-" yaml:"PlaceID"`
	Place        *Place    `gorm:"PRELOAD:true" json:"Place" yaml:"-"`
	CreatedAt    time.Time `json:"CreatedAt" yaml:"-"`
	UpdatedAt    time.Time `json:"UpdatedAt" yaml:"-"`
}

// TableName returns the entity table name.
func (Cell) TableName() string {
	return "cells"
}

// UnknownLocation is PhotoPrism's default location.
var UnknownLocation = Cell{
	ID:           UnknownID,
	Place:        &UnknownPlace,
	PlaceID:      UnknownID,
	CellName:     "",
	CellStreet:   "",
	CellPostcode: "",
	CellCategory: "",
}

// CreateUnknownLocation creates the default location if not exists.
func CreateUnknownLocation() {
	UnknownLocation = *FirstOrCreateCell(&UnknownLocation)
}

// NewCell creates a location using a token extracted from coordinate
func NewCell(lat, lng float32) *Cell {
	result := &Cell{}

	result.ID = s2.PrefixedToken(float64(lat), float64(lng))

	return result
}

// Refresh updates the index by retrieving the latest data from an external API.
func (m *Cell) Refresh(api string) (err error) {
	// Unknown?
	if m.Unknown() {
		// Skip.
		return nil
	}

	start := time.Now()

	// Initialize.
	l := &maps.Location{
		ID: s2.NormalizeToken(m.ID),
	}

	cellMutex.Lock()
	defer cellMutex.Unlock()

	// Retrieve location details from Places API.
	if err = l.QueryApi(api); err != nil {
		return err
	}

	// Unknown location or label missing?
	if l.Unknown() || l.Label() == "" {
		// Ignore.
		return nil
	}

	oldPlaceID := m.PlaceID

	place := Place{
		ID:            l.PlaceID(),
		PlaceLabel:    l.Label(),
		PlaceDistrict: l.District(),
		PlaceCity:     l.City(),
		PlaceState:    l.State(),
		PlaceCountry:  l.CountryCode(),
		PlaceKeywords: l.KeywordString(),
		PhotoCount:    1,
	}

	m.Place = &place
	m.PlaceID = l.PlaceID()

	// Create or update place.
	if err = place.Save(); err != nil {
		log.Errorf("index: %s while saving place %s", err, place.ID)
	} else {
		log.Tracef("index: updated place %s", place.ID)
	}

	m.CellName = l.Name()
	m.CellStreet = l.Street()
	m.CellPostcode = l.Postcode()
	m.CellCategory = l.Category()

	// Update cell.
	err = m.Save()

	if err != nil {
		log.Errorf("index: %s while updating cell %s [%s]", err, m.ID, time.Since(start))
		return err
	} else if oldPlaceID == m.PlaceID {
		log.Tracef("index: cell %s keeps place_id %s", m.ID, m.PlaceID)
	} else if err := UnscopedDb().Table(Photo{}.TableName()).
		Where("place_id = ?", oldPlaceID).
		UpdateColumn("place_id", m.PlaceID).
		Error; err != nil {
		log.Warnf("index: %s while changing place_id from %s to %s", err, oldPlaceID, m.PlaceID)
	}

	log.Debugf("index: updated cell %s [%s]", m.ID, time.Since(start))

	return err
}

// Find retrieves location data from the database or an external api if not known already.
func (m *Cell) Find(api string) error {
	start := time.Now()
	db := Db()

	if err := db.Preload("Place").First(m, "id = ?", m.ID).Error; err == nil {
		log.Tracef("cell: found %s", m.ID)
		return nil
	}

	l := &maps.Location{
		ID: s2.NormalizeToken(m.ID),
	}

	cellMutex.Lock()
	defer cellMutex.Unlock()

	// Retrieve location details from Places API.
	if err := l.QueryApi(api); err != nil {
		return err
	}

	if found := FindPlace(l.PlaceID()); found != nil {
		m.Place = found
	} else {
		place := &Place{
			ID:            l.PlaceID(),
			PlaceLabel:    l.Label(),
			PlaceDistrict: l.District(),
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

			log.Infof("cell: added place %s [%s]", place.ID, time.Since(start))

			m.Place = place
		} else if found := FindPlace(l.PlaceID()); found != nil {
			m.Place = found
		} else {
			log.Errorf("cell: %s while creating place %s", createErr, place.ID)
			m.Place = &UnknownPlace
		}
	}

	m.PlaceID = m.Place.ID
	m.CellName = l.Name()
	m.CellStreet = l.Street()
	m.CellPostcode = l.Postcode()
	m.CellCategory = l.Category()

	if createErr := db.Create(m).Error; createErr == nil {
		log.Debugf("cell: added %s [%s]", m.ID, time.Since(start))
		return nil
	} else if findErr := db.Preload("Place").First(m, "id = ?", m.ID).Error; findErr != nil {
		log.Errorf("cell: %s (create %s)", createErr, m.ID)
		log.Errorf("cell: %s (find %s)", findErr, m.ID)
		return createErr
	} else {
		log.Tracef("cell: found %s [%s]", m.ID, time.Since(start))
	}

	return nil
}

// Create inserts a new row to the database.
func (m *Cell) Create() error {
	return Db().Create(m).Error
}

// Save updates the record in the database or inserts a new record if it does not already exist.
func (m *Cell) Save() error {
	return Db().Save(m).Error
}

// Delete removes the entity from the index.
func (m *Cell) Delete() (err error) {
	return UnscopedDb().Delete(m).Error
}

// FirstOrCreateCell fetches an existing row, inserts a new row or nil in case of errors.
func FirstOrCreateCell(m *Cell) *Cell {
	if m.ID == "" {
		log.Errorf("cell: id missing")
		return nil
	}

	if m.PlaceID == "" {
		log.Errorf("cell: place id missing (find or create %s)", m.ID)
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
		log.Errorf("cell: %s (find or create %s)", createErr, m.ID)
	}

	return nil
}

// Keywords returns search keywords for a location.
func (m *Cell) Keywords() (result []string) {
	if m.Place == nil {
		log.Errorf("cell: place for %s is missing - you may have found a bug", m.ID)
		return result
	}

	result = append(result, txt.Keywords(txt.ReplaceSpaces(m.District(), "-"))...)
	result = append(result, txt.Keywords(txt.ReplaceSpaces(m.City(), "-"))...)
	result = append(result, txt.Keywords(txt.ReplaceSpaces(m.State(), "-"))...)
	result = append(result, txt.Keywords(txt.ReplaceSpaces(m.CountryName(), "-"))...)
	result = append(result, txt.Keywords(m.Name())...)
	result = append(result, txt.Keywords(m.Street())...)
	result = append(result, txt.Keywords(m.Category())...)
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

// Street returns the street name if any.
func (m *Cell) Street() string {
	return m.CellStreet
}

// NoStreet checks if the location has a street.
func (m *Cell) NoStreet() bool {
	return m.CellStreet == ""
}

// Postcode returns the postcode if any.
func (m *Cell) Postcode() string {
	return m.CellPostcode
}

// NoPostcode checks if the location has a postcode.
func (m *Cell) NoPostcode() bool {
	return m.CellPostcode == ""
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

// District returns the district name if any.
func (m *Cell) District() string {
	return m.Place.District()
}

// City returns the location city name if any.
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
