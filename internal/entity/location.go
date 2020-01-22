package entity

import (
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/maps"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/s2"
	"github.com/photoprism/photoprism/pkg/txt"
)

var locationMutex = sync.Mutex{}

// Photo location
type Location struct {
	ID          string `gorm:"type:varbinary(16);primary_key;auto_increment:false;"`
	PlaceID     string `gorm:"type:varbinary(16);"`
	Place       *Place
	LocName     string `gorm:"type:varchar(128);"`
	LocCategory string `gorm:"type:varchar(64);"`
	LocSource   string `gorm:"type:varbinary(16);"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Locks locations for updates
func (Location) Lock() {
	locationMutex.Lock()
}

// Unlock locations for updates
func (Location) Unlock() {
	locationMutex.Unlock()
}

func NewLocation(lat, lng float64) *Location {
	result := &Location{}

	result.ID = s2.Token(lat, lng)

	return result
}

func (m *Location) Find(db *gorm.DB, api string) error {
	mutex.Db.Lock()
	defer mutex.Db.Unlock()

	if err := db.First(m, "id = ?", m.ID).Error; err == nil {
		m.Place = FindPlace(m.PlaceID, db)
		return nil
	}

	l := &maps.Location{
		ID: m.ID,
	}

	if err := l.QueryApi(api); err != nil {
		return err
	}

	m.Place = FindPlaceByLabel(l.ID, l.LocLabel, db)

	if m.Place.NoID() {
		m.Place.ID = l.ID
		m.Place.LocLabel = l.LocLabel
		m.Place.LocCity = l.LocCity
		m.Place.LocState = l.LocState
		m.Place.LocCountry = l.LocCountry
	}

	m.LocName = l.LocName
	m.LocCategory = l.LocCategory
	m.LocSource = l.LocSource

	if err := db.Create(m).Error; err != nil {
		log.Errorf("location: %s", err)
		return err
	}

	return nil
}

func (m *Location) Keywords() []string {
	result := []string{
		strings.ToLower(m.City()),
		strings.ToLower(m.State()),
		strings.ToLower(m.CountryName()),
		strings.ToLower(m.Category()),
	}

	result = append(result, txt.Keywords(m.Name())...)
	result = append(result, txt.Keywords(m.Label())...)
	result = append(result, txt.Keywords(m.Notes())...)

	return result
}

func (m *Location) Unknown() bool {
	return m.ID == ""
}

func (m *Location) Name() string {
	return m.LocName
}

func (m *Location) NoName() bool {
	return m.LocName == ""
}

func (m *Location) Category() string {
	return m.LocCategory
}

func (m *Location) NoCategory() bool {
	return m.LocCategory == ""
}

func (m *Location) Label() string {
	return m.Place.Label()
}

func (m *Location) City() string {
	return m.Place.City()
}

func (m *Location) LongCity() bool {
	return len(m.City()) > 16
}

func (m *Location) NoCity() bool {
	return m.City() == ""
}

func (m *Location) CityContains(text string) bool {
	return strings.Contains(text, m.City())
}

func (m *Location) State() string {
	return m.Place.State()
}

func (m *Location) NoState() bool {
	return m.Place.State() == ""
}

func (m *Location) CountryCode() string {
	return m.Place.CountryCode()
}

func (m *Location) CountryName() string {
	return m.Place.CountryName()
}

func (m *Location) Notes() string {
	return m.Place.Notes()
}

func (m *Location) Source() string {
	return m.LocSource
}
