package entity

import (
	"strings"
	"time"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/maps"
	"github.com/photoprism/photoprism/pkg/s2"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Geo represents a S2 cell with location data.
type Geo struct {
	ID          string    `gorm:"type:varbinary(42);primary_key;auto_increment:false;" json:"ID" yaml:"ID"`
	PlaceID     string    `gorm:"type:varbinary(42);default:'zz'" json:"-" yaml:"PlaceID"`
	Place       *Place    `gorm:"PRELOAD:true" json:"Place" yaml:"-"`
	GeoName     string    `gorm:"type:varchar(255);" json:"Name" yaml:"Name,omitempty"`
	GeoCategory string    `gorm:"type:varchar(64);" json:"Category" yaml:"Category,omitempty"`
	GeoSource   string    `gorm:"type:varbinary(16);" json:"Source" yaml:"Source,omitempty"`
	CreatedAt   time.Time `json:"CreatedAt" yaml:"-"`
	UpdatedAt   time.Time `json:"UpdatedAt" yaml:"-"`
}

// TableName return the database table name.
func (Geo) TableName() string {
	return "geo"
}

// UnknownLocation is PhotoPrism's default location.
var UnknownLocation = Geo{
	ID:          "zz",
	Place:       &UnknownPlace,
	PlaceID:     "zz",
	GeoName:     "",
	GeoCategory: "",
	GeoSource:   SrcAuto,
}

// CreateUnknownLocation creates the default location if not exists.
func CreateUnknownLocation() {
	FirstOrCreateGeo(&UnknownLocation)
}

// NewGeo creates a location using a token extracted from coordinate
func NewGeo(lat, lng float32) *Geo {
	result := &Geo{}

	result.ID = s2.PrefixedToken(float64(lat), float64(lng))

	return result
}

// Find retrieves location data from the database or an external api if not known already.
func (m *Geo) Find(api string) error {
	start := time.Now()
	db := Db()

	if err := db.Preload("Place").First(m, "id = ?", m.ID).Error; err == nil {
		log.Infof("geo: found %s (%+v)", m.ID, m)
		return nil
	}

	l := &maps.Location{
		ID: s2.NormalizeToken(m.ID),
	}

	if err := l.QueryApi(api); err != nil {
		log.Errorf("geo: %s failed %s", m.ID, err)
		return err
	}

	if found := FindPlace(l.PrefixedToken(), l.Label()); found != nil {
		m.Place = found
	} else {
		place := &Place{
			ID:          l.PrefixedToken(),
			GeoLabel:    l.Label(),
			GeoCity:     l.City(),
			GeoState:    l.State(),
			GeoCountry:  l.CountryCode(),
			GeoKeywords: l.KeywordString(),
			PhotoCount:  1,
		}

		if err := place.Create(); err == nil {
			event.Publish("count.places", event.Data{
				"count": 1,
			})

			log.Infof("place: added %s [%s]", place.ID, time.Since(start))

			m.Place = place
		} else if found := FindPlace(l.PrefixedToken(), l.Label()); found != nil {
			m.Place = found
		} else {
			log.Errorf("place: %s (add place %s for location %s)", err, place.ID, l.ID)
			m.Place = &UnknownPlace
		}
	}

	m.PlaceID = m.Place.ID
	m.GeoName = l.Name()
	m.GeoCategory = l.Category()
	m.GeoSource = l.Source()

	if err := db.Create(m).Error; err == nil {
		log.Infof("geo: added %s [%s]", m.ID, time.Since(start))
		return nil
	} else if err := db.Preload("Place").First(m, "id = ?", m.ID).Error; err != nil {
		log.Errorf("geo: failed adding %s %s [%s]", m.ID, err.Error(), time.Since(start))
		return err
	} else {
		log.Infof("geo: found %s after second try [%s]", m.ID, time.Since(start))
	}

	return nil
}

// Create inserts a new row to the database.
func (m *Geo) Create() error {
	return Db().Create(m).Error
}

// FirstOrCreateGeo fetches an existing row, inserts a new row or nil in case of errors.
func FirstOrCreateGeo(m *Geo) *Geo {
	if m.ID == "" {
		log.Errorf("geo: id must not be empty")
		return nil
	}

	if m.PlaceID == "" {
		log.Errorf("geo: place_id must not be empty (first or create %s)", m.ID)
		return nil
	}

	result := Geo{}

	if findErr := Db().Where("id = ?", m.ID).First(&result).Error; findErr == nil {
		return &result
	} else if createErr := m.Create(); createErr == nil {
		return m
	} else if err := Db().Where("id = ?", m.ID).First(&result).Error; err == nil {
		return &result
	} else {
		log.Errorf("geo: %s (first or create %s)", createErr, m.ID)
	}

	return nil
}

// Keywords returns search keywords for a location.
func (m *Geo) Keywords() (result []string) {
	if m.Place == nil {
		log.Errorf("geo: place for %s is nil - you might have found a bug", m.ID)
		return result
	}

	result = append(result, txt.Keywords(txt.ReplaceSpaces(m.City(), "-"))...)
	result = append(result, txt.Keywords(txt.ReplaceSpaces(m.State(), "-"))...)
	result = append(result, txt.Keywords(txt.ReplaceSpaces(m.CountryName(), "-"))...)
	result = append(result, txt.Keywords(m.Category())...)
	result = append(result, txt.Keywords(m.Name())...)
	result = append(result, txt.Keywords(m.Place.GeoKeywords)...)

	result = txt.UniqueWords(result)

	return result
}

// Unknown checks if the location has no id
func (m *Geo) Unknown() bool {
	return m.ID == "" || m.ID == UnknownLocation.ID
}

// Name returns name of location
func (m *Geo) Name() string {
	return m.GeoName
}

// NoName checks if the location has no name
func (m *Geo) NoName() bool {
	return m.GeoName == ""
}

// Category returns the location category
func (m *Geo) Category() string {
	return m.GeoCategory
}

// NoCategory checks id the location has no category
func (m *Geo) NoCategory() bool {
	return m.GeoCategory == ""
}

// Label returns the location place label
func (m *Geo) Label() string {
	return m.Place.Label()
}

// City returns the location place city
func (m *Geo) City() string {
	return m.Place.City()
}

// LongCity checks if the city name is more than 16 char
func (m *Geo) LongCity() bool {
	return len(m.City()) > 16
}

// NoCity checks if the location has no city
func (m *Geo) NoCity() bool {
	return m.City() == ""
}

// CityContains checks if the location city contains the text string
func (m *Geo) CityContains(text string) bool {
	return strings.Contains(text, m.City())
}

// State returns the location place state
func (m *Geo) State() string {
	return m.Place.State()
}

// NoState checks if the location place has no state
func (m *Geo) NoState() bool {
	return m.Place.State() == ""
}

// CountryCode returns the location place country code
func (m *Geo) CountryCode() string {
	return m.Place.CountryCode()
}

// CountryName returns the location place country name
func (m *Geo) CountryName() string {
	return m.Place.CountryName()
}

// Notes returns the locations place notes
func (m *Geo) Notes() string {
	return m.Place.Notes()
}

// Source returns the source of location information
func (m *Geo) Source() string {
	return m.GeoSource
}
