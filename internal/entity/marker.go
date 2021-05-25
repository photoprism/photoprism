package entity

import (
	"time"
)

const (
	MarkerUnknown = ""
	MarkerFace    = "Face"
	MarkerLabel   = "Label"
)

// Marker represents an image marker point.
type Marker struct {
	ID          uint    `gorm:"primary_key" json:"ID" yaml:"-"`
	FileUID     string  `gorm:"type:VARBINARY(42);index;" json:"FileUID" yaml:"FileUID"`
	RefUID      string  `gorm:"type:VARBINARY(42);index;" json:"UID" yaml:"UID,omitempty"`
	MarkerSrc   string  `gorm:"type:VARBINARY(8);default:'';" json:"Src" yaml:"Src,omitempty"`
	MarkerType  string  `gorm:"type:VARBINARY(8);default:'';" json:"Type" yaml:"Type"`
	MarkerLabel string  `gorm:"type:VARCHAR(255);" json:"Label" yaml:"Label,omitempty"`
	MarkerMeta  string  `gorm:"type:TEXT;" json:"Meta" yaml:"Meta,omitempty"`
	Uncertainty int     `gorm:"type:SMALLINT"`
	X           float32 `gorm:"type:FLOAT;" json:"X" yaml:"X,omitempty"`
	Y           float32 `gorm:"type:FLOAT;" json:"Y" yaml:"Y,omitempty"`
	W           float32 `gorm:"type:FLOAT;" json:"W" yaml:"W,omitempty"`
	H           float32 `gorm:"type:FLOAT;" json:"H" yaml:"H,omitempty"`
	File        *File
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TableName returns the entity database table name.
func (Marker) TableName() string {
	return "markers_dev"
}

// NewMarker creates a new entity.
func NewMarker(fileUID, refUID, markerSrc, markerType string) *Marker {
	result := &Marker{
		FileUID:    fileUID,
		RefUID:     refUID,
		MarkerSrc:  markerSrc,
		MarkerType: markerType,
	}

	return result
}

// Updates multiple columns in the database.
func (m *Marker) Updates(values interface{}) error {
	return UnscopedDb().Model(m).UpdateColumns(values).Error
}

// Update updates a column in the database.
func (m *Marker) Update(attr string, value interface{}) error {
	return UnscopedDb().Model(m).UpdateColumn(attr, value).Error
}

// Save updates the existing or inserts a new row.
func (m *Marker) Save() error {
	return Db().Save(m).Error
}

// Create inserts a new row to the database.
func (m *Marker) Create() error {
	return Db().Create(m).Error
}

// FirstOrCreateMarker returns the existing row, inserts a new row or nil in case of errors.
func FirstOrCreateMarker(m *Marker) *Marker {
	result := Marker{}

	if err := Db().Where("file_uid = ? AND ref_uid = ?", m.FileUID, m.RefUID).First(&result).Error; err == nil {
		return &result
	} else if err := m.Create(); err != nil {
		log.Errorf("marker: %s", err)
		return nil
	}

	return m
}
