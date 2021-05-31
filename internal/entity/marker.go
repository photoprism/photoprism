package entity

import (
	"fmt"
	"time"

	"github.com/photoprism/photoprism/internal/face"
)

const (
	MarkerUnknown = ""
	MarkerFace    = "Face"
	MarkerLabel   = "Label"
)

// Marker represents an image marker point.
type Marker struct {
	ID            uint    `gorm:"primary_key" json:"ID" yaml:"-"`
	FileID        uint    `gorm:"index;" json:"-" yaml:"-"`
	RefUID        string  `gorm:"type:VARBINARY(42);index;" json:"RefUID" yaml:"RefUID,omitempty"`
	MarkerSrc     string  `gorm:"type:VARBINARY(8);default:'';" json:"Src" yaml:"Src,omitempty"`
	MarkerType    string  `gorm:"type:VARBINARY(8);default:'';" json:"Type" yaml:"Type"`
	MarkerScore   int     `gorm:"type:SMALLINT" json:"Score" yaml:"Score"`
	MarkerInvalid bool    `json:"Invalid" yaml:"Invalid,omitempty"`
	MarkerLabel   string  `gorm:"type:VARCHAR(255);" json:"Label" yaml:"Label,omitempty"`
	MarkerMeta    string  `gorm:"type:TEXT;" json:"Meta" yaml:"Meta,omitempty"`
	X             float32 `gorm:"type:FLOAT;" json:"X" yaml:"X,omitempty"`
	Y             float32 `gorm:"type:FLOAT;" json:"Y" yaml:"Y,omitempty"`
	W             float32 `gorm:"type:FLOAT;" json:"W" yaml:"W,omitempty"`
	H             float32 `gorm:"type:FLOAT;" json:"H" yaml:"H,omitempty"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// TableName returns the entity database table name.
func (Marker) TableName() string {
	return "markers_dev"
}

// NewMarker creates a new entity.
func NewMarker(fileUID uint, refUID, markerSrc, markerType string, x, y, w, h float32) *Marker {
	m := &Marker{
		FileID:     fileUID,
		RefUID:     refUID,
		MarkerSrc:  markerSrc,
		MarkerType: markerType,
		X:          x,
		Y:          y,
		W:          w,
		H:          h,
	}

	return m
}

// NewFaceMarker creates a new entity.
func NewFaceMarker(f face.Face, fileID uint, refUID string) *Marker {
	fm := f.Marker()
	m := NewMarker(fileID, refUID, SrcImage, MarkerFace, fm.X, fm.Y, fm.W, fm.H)

	m.MarkerScore = f.Score
	m.MarkerMeta = string(f.RelativeLandmarksJSON())

	return m
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
	if m.X == 0 || m.Y == 0 || m.X > 1 || m.Y > 1 || m.X < -1 || m.Y < -1 {
		return fmt.Errorf("marker: invalid position")
	}

	return Db().Save(m).Error
}

// Create inserts a new row to the database.
func (m *Marker) Create() error {
	if m.X == 0 || m.Y == 0 || m.X > 1 || m.Y > 1 || m.X < -1 || m.Y < -1 {
		return fmt.Errorf("marker: invalid position")
	}

	return Db().Create(m).Error
}

// UpdateOrCreateMarker updates a marker in the database or creates a new one if needed.
func UpdateOrCreateMarker(m *Marker) (*Marker, error) {
	result := Marker{}

	if m.ID > 0 {
		err := m.Save()
		log.Debugf("faces: saved marker %d for file %d", m.ID, m.FileID)
		return m, err
	} else if err := Db().Where(`file_id = ? AND x > ? AND x < ? AND y > ? AND y < ?`,
		m.FileID, m.X-m.W, m.X+m.W, m.Y-m.H, m.Y+m.H).First(&result).Error; err == nil {

		if SrcPriority[m.MarkerSrc] < SrcPriority[result.MarkerSrc] {
			// Ignore.
			return &result, nil
		}

		err := result.Updates(map[string]interface{}{
			"X":           m.X,
			"Y":           m.Y,
			"W":           m.W,
			"H":           m.H,
			"MarkerScore": m.MarkerScore,
			"MarkerMeta":  m.MarkerMeta,
			"RefUID":      m.RefUID,
		})

		log.Debugf("faces: updated existing marker %d for file %d", result.ID, result.FileID)

		return &result, err
	} else if err := m.Create(); err != nil {
		log.Debugf("faces: added marker %d for file %d", m.ID, m.FileID)
		return m, err
	}

	return m, nil
}
