package entity

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/txt"
	"github.com/ulule/deepcopier"

	"github.com/photoprism/photoprism/internal/face"
)

const (
	MarkerUnknown = ""
	MarkerFace    = "Face"
	MarkerLabel   = "Label"
)

// Marker represents an image marker point.
type Marker struct {
	ID             uint    `gorm:"primary_key" json:"ID" yaml:"-"`
	FileID         uint    `gorm:"index;" json:"-" yaml:"-"`
	FaceID         string  `gorm:"type:VARBINARY(42);index;" json:"FaceID" yaml:"FaceID,omitempty"`
	RefUID         string  `gorm:"type:VARBINARY(42);index:idx_markers_uid_type;" json:"RefUID" yaml:"RefUID,omitempty"`
	RefSrc         string  `gorm:"type:VARBINARY(8);default:'';" json:"RefSrc" yaml:"RefSrc,omitempty"`
	MarkerType     string  `gorm:"type:VARBINARY(8);index:idx_markers_uid_type;default:'';" json:"Type" yaml:"Type"`
	MarkerSrc      string  `gorm:"type:VARBINARY(8);default:'';" json:"Src" yaml:"Src,omitempty"`
	MarkerScore    int     `gorm:"type:SMALLINT" json:"Score" yaml:"Score,omitempty"`
	MarkerInvalid  bool    `json:"Invalid" yaml:"Invalid,omitempty"`
	MarkerLabel    string  `gorm:"type:VARCHAR(255);" json:"Label" yaml:"Label,omitempty"`
	MetaJSON       []byte  `gorm:"type:MEDIUMBLOB;" json:"MetaJSON" yaml:"MetaJSON,omitempty"`
	EmbeddingsJSON []byte  `gorm:"type:MEDIUMBLOB;" json:"EmbeddingsJSON" yaml:"EmbeddingsJSON,omitempty"`
	X              float32 `gorm:"type:FLOAT;" json:"X" yaml:"X,omitempty"`
	Y              float32 `gorm:"type:FLOAT;" json:"Y" yaml:"Y,omitempty"`
	W              float32 `gorm:"type:FLOAT;" json:"W" yaml:"W,omitempty"`
	H              float32 `gorm:"type:FLOAT;" json:"H" yaml:"H,omitempty"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Person         *Person    `gorm:"foreignkey:RefUID;association_foreignkey:PersonUID;association_autoupdate:false;association_autocreate:false;association_save_reference:false" json:"Person" yaml:"-"`
	embeddings     Embeddings `gorm:"-"`
}

// UnknownMarker can be used as a default for unknown markers.
var UnknownMarker = NewMarker(0, "", SrcDefault, MarkerUnknown, 0, 0, 0, 0)

// TableName returns the entity database table name.
func (Marker) TableName() string {
	return "markers_dev2"
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
	pos := f.Marker()

	m := NewMarker(fileID, refUID, SrcImage, MarkerFace, pos.X, pos.Y, pos.W, pos.H)

	m.MarkerScore = f.Score
	m.MetaJSON = f.RelativeLandmarksJSON()
	m.EmbeddingsJSON = f.EmbeddingsJSON()

	return m
}

// Updates multiple columns in the database.
func (m *Marker) Updates(values interface{}) error {
	return UnscopedDb().Model(m).Updates(values).Error
}

// Update updates a column in the database.
func (m *Marker) Update(attr string, value interface{}) error {
	return UnscopedDb().Model(m).Update(attr, value).Error
}

// SaveForm updates the entity using form data and stores it in the database.
func (m *Marker) SaveForm(f form.Marker) error {
	if err := deepcopier.Copy(m).From(f); err != nil {
		return err
	}

	if f.MarkerLabel != "" {
		m.MarkerLabel = txt.Title(txt.Clip(f.MarkerLabel, txt.ClipKeyword))
	}

	if err := m.Save(); err != nil {
		return err
	}

	faceId := m.FaceID

	if faceId != "" && m.MarkerLabel != "" && m.RefUID == "" && m.MarkerType == MarkerFace {
		if p := NewPerson(m.MarkerLabel, SrcMarker, 1); p == nil {
			return fmt.Errorf("marker: person should not be nil (save form)")
		} else if p = FirstOrCreatePerson(p); p == nil {
			return fmt.Errorf("marker: failed adding person %s for marker %d (save form)", txt.Quote(m.MarkerLabel), m.ID)
		} else if err := m.Updates(Val{"RefUID": p.PersonUID, "RefSrc": SrcManual, "FaceID": ""}); err != nil {
			return fmt.Errorf("marker: %s (save form)", err)
		} else if err := Db().Model(&Face{}).Where("id = ? AND person_uid = ''", faceId).Update("PersonUID", p.PersonUID).Error; err != nil {
			return fmt.Errorf("marker: %s (update face)", err)
		} else if err := Db().Model(&Marker{}).
			Where("face_id = ?", faceId).
			Updates(Val{"RefUID": p.PersonUID, "RefSrc": SrcManual, "FaceID": ""}).Error; err != nil {
			return fmt.Errorf("marker: %s (update related markers)", err)
		} else {
			log.Infof("marker: matched person %s with label %s", p.PersonUID, txt.Quote(m.MarkerLabel))
		}
	} else if m.MarkerLabel != "" && m.RefUID != "" && m.MarkerType == MarkerFace {
		if p := FindPerson(m.RefUID); p != nil {
			p.SetName(m.MarkerLabel)

			if err := p.Save(); err != nil {
				return fmt.Errorf("marker: %s (update person)", err)
			}
		}
	}

	return nil
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

// Embeddings returns parsed marker embeddings.
func (m *Marker) Embeddings() Embeddings {
	if len(m.EmbeddingsJSON) == 0 {
		return Embeddings{}
	} else if len(m.embeddings) > 0 {
		return m.embeddings
	} else if err := json.Unmarshal(m.EmbeddingsJSON, &m.embeddings); err != nil {
		log.Errorf("failed parsing marker embeddings json: %s", err)
	}

	return m.embeddings
}

// FindMarker returns an existing row if exists.
func FindMarker(id uint) *Marker {
	result := Marker{}

	if err := Db().Where("id = ?", id).First(&result).Error; err == nil {
		return &result
	}

	return nil
}

// UpdateOrCreateMarker updates a marker in the database or creates a new one if needed.
func UpdateOrCreateMarker(m *Marker) (*Marker, error) {
	const d = 0.07

	result := Marker{}

	if m.ID > 0 {
		err := m.Save()
		log.Debugf("faces: saved marker %d for file %d", m.ID, m.FileID)
		return m, err
	} else if err := Db().Where(`file_id = ? AND x > ? AND x < ? AND y > ? AND y < ?`,
		m.FileID, m.X-d, m.X+d, m.Y-d, m.Y+d).First(&result).Error; err == nil {

		if SrcPriority[m.MarkerSrc] < SrcPriority[result.MarkerSrc] {
			// Ignore.
			return &result, nil
		}

		err := result.Updates(map[string]interface{}{
			"X":              m.X,
			"Y":              m.Y,
			"W":              m.W,
			"H":              m.H,
			"MarkerScore":    m.MarkerScore,
			"MetaJSON":       m.MetaJSON,
			"EmbeddingsJSON": m.EmbeddingsJSON,
			"RefUID":         m.RefUID,
		})

		log.Debugf("faces: updated existing marker %d for file %d", result.ID, result.FileID)

		return &result, err
	} else if err := m.Create(); err != nil {
		log.Debugf("faces: added marker %d for file %d", m.ID, m.FileID)
		return m, err
	}

	return m, nil
}
