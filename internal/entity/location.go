package entity

import (
	"strings"
	"time"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/maps"
	"github.com/photoprism/photoprism/pkg/s2"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Location used to associate photos to location
type Location struct {
	ID          string    `gorm:"type:varbinary(16);primary_key;auto_increment:false;" json:"ID" yaml:"ID"`
	PlaceID     string    `gorm:"type:varbinary(16);" json:"-" yaml:"PlaceID"`
	Place       *Place    `gorm:"PRELOAD:true" json:"Place" yaml:"-"`
	LocName     string    `gorm:"type:varchar(255);" json:"Name" yaml:"Name,omitempty"`
	LocCategory string    `gorm:"type:varchar(64);" json:"Category" yaml:"Category,omitempty"`
	LocSource   string    `gorm:"type:varbinary(16);" json:"Source" yaml:"Source,omitempty"`
	CreatedAt   time.Time `json:"CreatedAt" yaml:"-"`
	UpdatedAt   time.Time `json:"UpdatedAt" yaml:"-"`
}

// UnknownLocation is PhotoPrism's default location.
var UnknownLocation = Location{
	ID:          "zz",
	Place:       &UnknownPlace,
	PlaceID:     "zz",
	LocName:     "",
	LocCategory: "",
	LocSource:   SrcAuto,
}

// CreateUnknownLocation creates the default location if not exists.
func CreateUnknownLocation() {
	FirstOrCreateLocation(&UnknownLocation)
}

// NewLocation creates a location using a token extracted from coordinate
func NewLocation(lat, lng float32) *Location {
	result := &Location{}

	result.ID = s2.Token(float64(lat), float64(lng))

	return result
}

// Find retrieves location data from the database or an external api if not known already.
func (m *Location) Find(api string) error {
	start := time.Now()
	db := Db()

	if err := db.Preload("Place").First(m, "id = ?", m.ID).Error; err == nil {
		log.Infof("location: found %s (%+v)", m.ID, m)
		return nil
	}

	l := &maps.Location{
		ID: m.ID,
	}

	if err := l.QueryApi(api); err != nil {
		log.Errorf("location: %s failed %s", m.ID, err)
		return err
	}

	if place := FindPlace(l.S2Token(), l.Label()); place != nil {
		m.Place = place
	} else {
		place = &Place{
			ID:          l.S2Token(),
			LocLabel:    l.Label(),
			LocCity:     l.City(),
			LocState:    l.State(),
			LocCountry:  l.CountryCode(),
			LocKeywords: l.KeywordString(),
		}

		if err := place.Create(); err != nil {
			log.Errorf("place: failed adding %s %s", place.ID, err.Error())
			m.Place = &UnknownPlace
		} else {
			event.Publish("count.places", event.Data{
				"count": 1,
			})

			log.Infof("place: added %s [%s]", place.ID, time.Since(start))

			m.Place = place
		}
	}

	m.PlaceID = m.Place.ID
	m.LocName = l.Name()
	m.LocCategory = l.Category()
	m.LocSource = l.Source()

	if err := db.Create(m).Error; err == nil {
		log.Infof("location: added %s [%s]", m.ID, time.Since(start))
		return nil
	} else if err := db.Preload("Place").First(m, "id = ?", m.ID).Error; err != nil {
		log.Errorf("location: failed adding %s %s [%s]", m.ID, err.Error(), time.Since(start))
		return err
	} else {
		log.Infof("location: found %s after second try [%s]", m.ID, time.Since(start))
	}

	return nil
}

// Create inserts a new row to the database.
func (m *Location) Create() error {
	return Db().Create(m).Error
}

// FirstOrCreateLocation returns the existing row, inserts a new row or nil in case of errors.
func FirstOrCreateLocation(m *Location) *Location {
	if m.ID == "" {
		log.Errorf("location: id must not be empty")
		return nil
	}

	if m.PlaceID == "" {
		log.Errorf("location: place_id must not be empty (id %s)", m.ID)
		return nil
	}

	result := Location{}

	if err := Db().Where("id = ?", m.ID).First(&result).Error; err == nil {
		return &result
	} else if err := m.Create(); err != nil {
		log.Errorf("location: %s", err)
		return nil
	}

	return m
}

// Keywords computes keyword based on a Location
func (m *Location) Keywords() (result []string) {
	if m.Place == nil {
		log.Errorf("location: place for %s is nil - you might have found a bug", m.ID)
		return result
	}

	result = append(result, txt.Keywords(txt.ReplaceSpaces(m.City(), "-"))...)
	result = append(result, txt.Keywords(txt.ReplaceSpaces(m.State(), "-"))...)
	result = append(result, txt.Keywords(txt.ReplaceSpaces(m.CountryName(), "-"))...)
	result = append(result, txt.Keywords(m.Category())...)
	result = append(result, txt.Keywords(m.Name())...)
	result = append(result, txt.Keywords(m.Place.LocKeywords)...)

	result = txt.UniqueWords(result)

	return result
}

// Unknown checks if the location has no id
func (m *Location) Unknown() bool {
	return m.ID == "" || m.ID == UnknownLocation.ID
}

// Name returns name of location
func (m *Location) Name() string {
	return m.LocName
}

// NoName checks if the location has no name
func (m *Location) NoName() bool {
	return m.LocName == ""
}

// Category returns the location category
func (m *Location) Category() string {
	return m.LocCategory
}

// NoCategory checks id the location has no category
func (m *Location) NoCategory() bool {
	return m.LocCategory == ""
}

// Label returns the location place label
func (m *Location) Label() string {
	return m.Place.Label()
}

// City returns the location place city
func (m *Location) City() string {
	return m.Place.City()
}

// LongCity checks if the city name is more than 16 char
func (m *Location) LongCity() bool {
	return len(m.City()) > 16
}

// NoCity checks if the location has no city
func (m *Location) NoCity() bool {
	return m.City() == ""
}

// CityContains checks if the location city contains the text string
func (m *Location) CityContains(text string) bool {
	return strings.Contains(text, m.City())
}

// State returns the location place state
func (m *Location) State() string {
	return m.Place.State()
}

// NoState checks if the location place has no state
func (m *Location) NoState() bool {
	return m.Place.State() == ""
}

// CountryCode returns the location place country code
func (m *Location) CountryCode() string {
	return m.Place.CountryCode()
}

// CountryName returns the location place country name
func (m *Location) CountryName() string {
	return m.Place.CountryName()
}

// Notes returns the locations place notes
func (m *Location) Notes() string {
	return m.Place.Notes()
}

// Source returns the source of location information
func (m *Location) Source() string {
	return m.LocSource
}
