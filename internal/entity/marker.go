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
	MarkerFace    = "face"
	MarkerLabel   = "label"
)

// Marker represents an image marker point.
type Marker struct {
	ID             uint    `gorm:"primary_key" json:"ID" yaml:"-"`
	FileID         uint    `gorm:"index;" json:"-" yaml:"-"`
	MarkerType     string  `gorm:"type:VARBINARY(8);index:idx_markers_subject;default:'';" json:"Type" yaml:"Type"`
	MarkerSrc      string  `gorm:"type:VARBINARY(8);default:'';" json:"Src" yaml:"Src,omitempty"`
	MarkerName     string  `gorm:"type:VARCHAR(255);" json:"Name" yaml:"Name,omitempty"`
	SubjectUID     string  `gorm:"type:VARBINARY(42);index:idx_markers_subject;" json:"SubjectUID" yaml:"SubjectUID,omitempty"`
	SubjectSrc     string  `gorm:"type:VARBINARY(8);default:'';" json:"SubjectSrc" yaml:"SubjectSrc,omitempty"`
	FaceID         string  `gorm:"type:VARBINARY(42);index;" json:"FaceID" yaml:"FaceID,omitempty"`
	EmbeddingsJSON []byte  `gorm:"type:MEDIUMBLOB;" json:"EmbeddingsJSON" yaml:"EmbeddingsJSON,omitempty"`
	MarkerScore    int     `gorm:"type:SMALLINT" json:"Score" yaml:"Score,omitempty"`
	MarkerInvalid  bool    `json:"Invalid" yaml:"Invalid,omitempty"`
	MarkerJSON     []byte  `gorm:"type:MEDIUMBLOB;" json:"MarkerJSON" yaml:"MarkerJSON,omitempty"`
	X              float32 `gorm:"type:FLOAT;" json:"X" yaml:"X,omitempty"`
	Y              float32 `gorm:"type:FLOAT;" json:"Y" yaml:"Y,omitempty"`
	W              float32 `gorm:"type:FLOAT;" json:"W" yaml:"W,omitempty"`
	H              float32 `gorm:"type:FLOAT;" json:"H" yaml:"H,omitempty"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Subject        *Subject   `gorm:"foreignkey:SubjectUID;association_foreignkey:SubjectUID;association_autoupdate:false;association_autocreate:false;association_save_reference:false" json:"Subject,omitempty" yaml:"-"`
	embeddings     Embeddings `gorm:"-"`
}

// UnknownMarker can be used as a default for unknown markers.
var UnknownMarker = NewMarker(0, "", SrcDefault, MarkerUnknown, 0, 0, 0, 0)

// TableName returns the entity database table name.
func (Marker) TableName() string {
	return "markers_dev3"
}

// NewMarker creates a new entity.
func NewMarker(fileUID uint, refUID, markerSrc, markerType string, x, y, w, h float32) *Marker {
	m := &Marker{
		FileID:     fileUID,
		SubjectUID: refUID,
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
	m.MarkerJSON = f.RelativeLandmarksJSON()
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

	if f.MarkerName != "" {
		m.MarkerName = txt.Title(txt.Clip(f.MarkerName, txt.ClipKeyword))
	}

	if err := m.Save(); err != nil {
		return err
	}

	faceId := m.FaceID

	if faceId != "" && m.MarkerName != "" && m.SubjectUID == "" && m.MarkerType == MarkerFace {
		if subj := NewSubject(m.MarkerName, SubjectPerson, SrcMarker); subj == nil {
			return fmt.Errorf("marker: subject should not be nil (save form)")
		} else if subj = FirstOrCreateSubject(subj); subj == nil {
			return fmt.Errorf("marker: failed adding subject %s for marker %d (save form)", txt.Quote(m.MarkerName), m.ID)
		} else if err := m.Updates(Values{"SubjectUID": subj.SubjectUID, "SubjectSrc": SrcManual, "FaceID": ""}); err != nil {
			return fmt.Errorf("marker: %s (save form)", err)
		} else if err := Db().Model(&Face{}).Where("id = ? AND subject_uid = ''", faceId).Update("SubjectUID", subj.SubjectUID).Error; err != nil {
			return fmt.Errorf("marker: %s (update face)", err)
		} else if err := Db().Model(&Marker{}).
			Where("face_id = ?", faceId).
			Updates(Values{"SubjectUID": subj.SubjectUID, "SubjectSrc": SrcManual, "FaceID": ""}).Error; err != nil {
			return fmt.Errorf("marker: %s (update related markers)", err)
		} else {
			log.Infof("marker: matched subject %s with %s", subj.SubjectUID, txt.Quote(m.MarkerName))
		}
	} else if err := m.UpdateSubject(); err != nil {
		log.Error(err)
	}

	return nil
}

// UpdateSubject changes and saves the related subject's name in the index.
func (m *Marker) UpdateSubject() error {
	if m.MarkerName == "" || m.SubjectUID == "" || m.MarkerType == MarkerFace {
		return nil
	}

	subj := FindSubject(m.SubjectUID)

	if subj == nil {
		return fmt.Errorf("marker: subject %s not found", m.SubjectUID)
	}

	return subj.UpdateName(m.MarkerName)
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
			"MarkerJSON":     m.MarkerJSON,
			"EmbeddingsJSON": m.EmbeddingsJSON,
			"SubjectUID":     m.SubjectUID,
		})

		log.Debugf("faces: updated existing marker %d for file %d", result.ID, result.FileID)

		return &result, err
	} else if err := m.Create(); err != nil {
		log.Debugf("faces: added marker %d for file %d", m.ID, m.FileID)
		return m, err
	}

	return m, nil
}
