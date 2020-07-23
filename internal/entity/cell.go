package entity

import (
	"strings"
	"time"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/maps"
	"github.com/photoprism/photoprism/pkg/s2"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Cell represents a S2 cell with location data.
type Cell struct {
	ID           string    `gorm:"type:varbinary(42);primary_key;auto_increment:false;" json:"ID" yaml:"ID"`
	CellName     string    `gorm:"type:varchar(255);" json:"Name" yaml:"Name,omitempty"`
	CellCategory string    `gorm:"type:varchar(64);" json:"Category" yaml:"Category,omitempty"`
	PlaceID      string    `gorm:"type:varbinary(42);default:'zz'" json:"-" yaml:"PlaceID"`
	Place        *Place    `gorm:"PRELOAD:true" json:"Place" yaml:"-"`
	CreatedAt    time.Time `json:"CreatedAt" yaml:"-"`
	UpdatedAt    time.Time `json:"UpdatedAt" yaml:"-"`
}

// UnknownLocation is PhotoPrism's default location.
var UnknownLocation = Cell{
	ID:           "zz",
	Place:        &UnknownPlace,
	PlaceID:      "zz",
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

// Find retrieves location data from the database or an external api if not known already.
func (m *Cell) Find(api string) error {
	start := time.Now()
	db := Db()

	if err := db.Preload("Place").First(m, "id = ?", m.ID).Error; err == nil {
		log.Debugf("places: found cell id %s", m.ID)
		return nil
	}

	l := &maps.Location{
		ID: s2.NormalizeToken(m.ID),
	}

	if err := l.QueryApi(api); err != nil {
		log.Errorf("places: %s (query cell id %s)", err, m.ID)
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

			log.Infof("places: added place id %s [%s]", place.ID, time.Since(start))

			m.Place = place
		} else if found := FindPlace(l.PrefixedToken(), l.Label()); found != nil {
			m.Place = found
		} else {
			log.Errorf("places: %s (create place id %s)", createErr, place.ID)
			m.Place = &UnknownPlace
		}
	}

	m.PlaceID = m.Place.ID
	m.CellName = l.Name()
	m.CellCategory = l.Category()

	if createErr := db.Create(m).Error; createErr == nil {
		log.Debugf("places: added cell id %s [%s]", m.ID, time.Since(start))
		return nil
	} else if findErr := db.Preload("Place").First(m, "id = ?", m.ID).Error; findErr != nil {
		log.Errorf("places: %s (create cell id %s)", createErr, m.ID)
		log.Errorf("places: %s (find cell id %s)", findErr, m.ID)
		return createErr
	} else {
		log.Debugf("places: found cell id %s after trying again [%s]", m.ID, time.Since(start))
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
		log.Errorf("places: cell id must not be empty")
		return nil
	}

	if m.PlaceID == "" {
		log.Errorf("places: place id must not be empty (first or create cell id %s)", m.ID)
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
		log.Errorf("places: %s (first or create cell id %s)", createErr, m.ID)
	}

	return nil
}

// Keywords returns search keywords for a location.
func (m *Cell) Keywords() (result []string) {
	if m.Place == nil {
		log.Errorf("places: place for cell id %s is nil - you might have found a bug", m.ID)
		return result
	}

	result = append(result, txt.Keywords(txt.ReplaceSpaces(m.City(), "-"))...)
	result = append(result, txt.Keywords(txt.ReplaceSpaces(m.State(), "-"))...)
	result = append(result, txt.Keywords(txt.ReplaceSpaces(m.CountryName(), "-"))...)
	result = append(result, txt.Keywords(m.Category())...)
	result = append(result, txt.Keywords(m.Name())...)
	result = append(result, txt.Keywords(m.Place.PlaceKeywords)...)

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
