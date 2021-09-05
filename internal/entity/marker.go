package entity

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/pkg/clusters"
	"github.com/photoprism/photoprism/pkg/rnd"

	"github.com/photoprism/photoprism/internal/crop"
	"github.com/photoprism/photoprism/internal/face"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/txt"
)

const (
	MarkerUnknown = ""
	MarkerFace    = "face"
	MarkerLabel   = "label"
)

// Marker represents an image marker point.
type Marker struct {
	MarkerUID      string          `gorm:"type:VARBINARY(42);primary_key;auto_increment:false;" json:"UID" yaml:"UID"`
	FileUID        string          `gorm:"type:VARBINARY(42);index;" json:"FileUID" yaml:"FileUID"`
	FileHash       string          `gorm:"type:VARBINARY(128);index" json:"FileHash" yaml:"FileHash,omitempty"`
	FileArea       string          `gorm:"type:VARBINARY(16);default:''" json:"FileArea" yaml:"FileArea,omitempty"`
	MarkerType     string          `gorm:"type:VARBINARY(8);default:'';" json:"Type" yaml:"Type"`
	MarkerSrc      string          `gorm:"type:VARBINARY(8);default:'';" json:"Src" yaml:"Src,omitempty"`
	MarkerName     string          `gorm:"type:VARCHAR(255);" json:"Name" yaml:"Name,omitempty"`
	MarkerInvalid  bool            `json:"Invalid" yaml:"Invalid,omitempty"`
	SubjectUID     string          `gorm:"type:VARBINARY(42);index:idx_markers_subject_uid_src;" json:"SubjectUID" yaml:"SubjectUID,omitempty"`
	SubjectSrc     string          `gorm:"type:VARBINARY(8);index:idx_markers_subject_uid_src;default:'';" json:"SubjectSrc" yaml:"SubjectSrc,omitempty"`
	subject        *Subject        `gorm:"foreignkey:SubjectUID;association_foreignkey:SubjectUID;association_autoupdate:false;association_autocreate:false;association_save_reference:false"`
	FaceID         string          `gorm:"type:VARBINARY(42);index;" json:"FaceID" yaml:"FaceID,omitempty"`
	FaceDist       float64         `gorm:"default:-1" json:"FaceDist" yaml:"FaceDist,omitempty"`
	face           *Face           `gorm:"foreignkey:FaceID;association_foreignkey:ID;association_autoupdate:false;association_autocreate:false;association_save_reference:false"`
	EmbeddingsJSON json.RawMessage `gorm:"type:MEDIUMBLOB;" json:"-" yaml:"EmbeddingsJSON,omitempty"`
	embeddings     Embeddings      `gorm:"-"`
	LandmarksJSON  json.RawMessage `gorm:"type:MEDIUMBLOB;" json:"-" yaml:"LandmarksJSON,omitempty"`
	X              float32         `gorm:"type:FLOAT;" json:"X" yaml:"X,omitempty"`
	Y              float32         `gorm:"type:FLOAT;" json:"Y" yaml:"Y,omitempty"`
	W              float32         `gorm:"type:FLOAT;" json:"W" yaml:"W,omitempty"`
	H              float32         `gorm:"type:FLOAT;" json:"H" yaml:"H,omitempty"`
	Size           int             `gorm:"default:-1" json:"Size" yaml:"Size,omitempty"`
	Score          int             `gorm:"type:SMALLINT" json:"Score" yaml:"Score,omitempty"`
	Review         bool            `json:"Review" yaml:"Review,omitempty"`
	MatchedAt      *time.Time      `sql:"index" json:"MatchedAt" yaml:"MatchedAt,omitempty"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// TableName returns the entity database table name.
func (Marker) TableName() string {
	return "markers_dev7"
}

// BeforeCreate creates a random UID if needed before inserting a new row to the database.
func (m *Marker) BeforeCreate(scope *gorm.Scope) error {
	if rnd.IsUID(m.MarkerUID, 'm') {
		return nil
	}

	return scope.SetColumn("MarkerUID", rnd.PPID('m'))
}

// NewMarker creates a new entity.
func NewMarker(file File, area crop.Area, subjectUID, markerSrc, markerType string) *Marker {
	m := &Marker{
		FileUID:    file.FileUID,
		FileHash:   file.FileHash,
		FileArea:   area.String(),
		MarkerSrc:  markerSrc,
		MarkerType: markerType,
		SubjectUID: subjectUID,
		X:          area.X,
		Y:          area.Y,
		W:          area.W,
		H:          area.H,
		MatchedAt:  nil,
	}

	return m
}

// NewFaceMarker creates a new entity.
func NewFaceMarker(f face.Face, file File, subjectUID string) *Marker {
	m := NewMarker(file, f.CropArea(), subjectUID, SrcImage, MarkerFace)

	m.Size = f.Size()
	m.Score = f.Score
	m.Review = f.Score < 30
	m.FaceDist = -1
	m.EmbeddingsJSON = f.EmbeddingsJSON()
	m.LandmarksJSON = f.RelativeLandmarksJSON()

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
	changed := false

	if m.MarkerInvalid != f.MarkerInvalid {
		m.MarkerInvalid = f.MarkerInvalid
		changed = true
	}

	if m.Review != f.Review {
		m.Review = f.Review
		changed = true
	}

	if f.SubjectSrc == SrcManual && strings.TrimSpace(f.MarkerName) != "" {
		m.SubjectSrc = SrcManual
		m.MarkerName = txt.Title(txt.Clip(f.MarkerName, txt.ClipDefault))

		if err := m.SyncSubject(true); err != nil {
			return err
		}

		changed = true
	}

	if changed {
		return m.Save()
	}

	return nil
}

// HasFace tests if the marker already has the best matching face.
func (m *Marker) HasFace(f *Face, dist float64) bool {
	if m.FaceID == "" {
		return false
	} else if f == nil {
		return m.FaceID != ""
	} else if m.FaceID == f.ID {
		return m.FaceID != ""
	} else if m.FaceDist < 0 {
		return false
	} else if dist < 0 {
		return true
	}

	return m.FaceDist <= dist
}

// SetFace sets a new face for this marker.
func (m *Marker) SetFace(f *Face, dist float64) (updated bool, err error) {
	if f == nil {
		return false, fmt.Errorf("face is nil")
	}

	if m.MarkerType != MarkerFace {
		return false, fmt.Errorf("not a face marker")
	}

	// Any reason we don't want to set a new face for this marker?
	if m.SubjectSrc == SrcAuto || f.SubjectUID == "" || m.SubjectUID == "" || f.SubjectUID == m.SubjectUID {
		// Don't skip if subject wasn't set manually, or subjects match.
	} else if reported, err := f.ResolveCollision(m.Embeddings()); err != nil {
		return false, err
	} else if reported {
		log.Infof("faces: collision of marker %s, subject %s, face %s, subject %s, source %s", m.MarkerUID, m.SubjectUID, f.ID, f.SubjectUID, m.SubjectSrc)
		return false, nil
	} else {
		return false, nil
	}

	// Update face with known subject from marker?
	if m.SubjectSrc == SrcAuto || m.SubjectUID == "" || f.SubjectUID != "" {
		// Don't update if face has a known subject, or marker subject is unknown.
	} else if err = f.SetSubjectUID(m.SubjectUID); err != nil {
		return false, err
	}

	// Set face.
	m.face = f

	// Skip update if the same face is already set.
	if m.SubjectUID == f.SubjectUID && m.FaceID == f.ID {
		// Update matching timestamp.
		m.MatchedAt = TimePointer()
		return false, m.Updates(Values{"MatchedAt": m.MatchedAt})
	}

	// Remember current values for comparison.
	faceID := m.FaceID
	subjectUID := m.SubjectUID
	SubjectSrc := m.SubjectSrc

	m.FaceID = f.ID
	m.FaceDist = dist

	if m.FaceDist < 0 {
		faceEmbedding := f.Embedding()

		// Calculate smallest distance to embeddings.
		for _, e := range m.Embeddings() {
			if len(e) != len(faceEmbedding) {
				continue
			}

			if d := clusters.EuclideanDistance(e, faceEmbedding); d < m.FaceDist || m.FaceDist < 0 {
				m.FaceDist = d
			}
		}
	}

	if f.SubjectUID != "" {
		m.SubjectUID = f.SubjectUID
	}

	if err = m.SyncSubject(false); err != nil {
		return false, err
	}

	// Update face subject?
	if m.SubjectSrc == SrcAuto || m.SubjectUID == "" || f.SubjectUID == m.SubjectUID {
		// Not needed.
	} else if err = f.SetSubjectUID(m.SubjectUID); err != nil {
		return false, err
	}

	updated = m.FaceID != faceID || m.SubjectUID != subjectUID || m.SubjectSrc != SubjectSrc

	// Update matching timestamp.
	m.MatchedAt = TimePointer()

	return updated, m.Updates(Values{"FaceID": m.FaceID, "FaceDist": m.FaceDist, "SubjectUID": m.SubjectUID, "SubjectSrc": m.SubjectSrc, "MatchedAt": m.MatchedAt})
}

// SyncSubject maintains the marker subject relationship.
func (m *Marker) SyncSubject(updateRelated bool) (err error) {
	// Face marker? If not, return.
	if m.MarkerType != MarkerFace {
		return nil
	}

	subj := m.Subject()

	if subj == nil || m.SubjectSrc == SrcAuto {
		return nil
	}

	// Update subject with marker name?
	if m.MarkerName == "" || subj.SubjectName == m.MarkerName {
		// Do nothing.
	} else if subj, err = subj.UpdateName(m.MarkerName); err != nil {
		return err
	} else if subj != nil {
		// Update subject fields in case it was merged.
		m.subject = subj
		m.SubjectUID = subj.SubjectUID
		m.MarkerName = subj.SubjectName
	}

	// Create known face for subject?
	if m.FaceID != "" {
		// Do nothing.
	} else if f := m.Face(); f != nil {
		m.FaceID = f.ID
	}

	// Update related markers?
	if m.FaceID == "" || m.SubjectUID == "" {
		// Do nothing.
	} else if err := Db().Model(&Face{}).Where("id = ? AND subject_uid = ''", m.FaceID).Update("SubjectUID", m.SubjectUID).Error; err != nil {
		return fmt.Errorf("%s (update known face)", err)
	} else if !updateRelated {
		return nil
	} else if err := Db().Model(&Marker{}).
		Where("marker_uid <> ?", m.MarkerUID).
		Where("face_id = ?", m.FaceID).
		Where("subject_src = ?", SrcAuto).
		Where("subject_uid <> ?", m.SubjectUID).
		Updates(Values{"SubjectUID": m.SubjectUID, "SubjectSrc": SrcAuto}).Error; err != nil {
		return fmt.Errorf("%s (update related markers)", err)
	} else {
		log.Debugf("marker: matched %s with %s", subj.SubjectName, m.FaceID)
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

// SubjectName returns the matching subject's name.
func (m *Marker) SubjectName() string {
	if m.MarkerName != "" {
		return m.MarkerName
	} else if s := m.Subject(); s != nil {
		return s.SubjectName
	}

	return ""
}

// Subject returns the matching subject or nil.
func (m *Marker) Subject() (subj *Subject) {
	if m.subject != nil {
		if m.SubjectUID == m.subject.SubjectUID {
			return m.subject
		}
	}

	// Create subject?
	if m.SubjectSrc != SrcAuto && m.MarkerName != "" && m.SubjectUID == "" {
		if subj = NewSubject(m.MarkerName, SubjectPerson, m.SubjectSrc); subj == nil {
			return nil
		} else if subj = FirstOrCreateSubject(subj); subj == nil {
			log.Debugf("marker: invalid subject %s", txt.Quote(m.MarkerName))
			return nil
		} else {
			m.subject = subj
			m.SubjectUID = subj.SubjectUID
		}

		return m.subject
	}

	m.subject = FindSubject(m.SubjectUID)

	return m.subject
}

// ClearSubject removes an existing subject association, and reports a collision.
func (m *Marker) ClearSubject(src string) error {
	// Find the matching face.
	if m.face == nil {
		m.face = FindFace(m.FaceID)
	}

	// Update index & resolve collisions.
	if err := m.Updates(Values{"MarkerName": "", "FaceID": "", "FaceDist": -1.0, "SubjectUID": "", "SubjectSrc": src}); err != nil {
		return err
	} else if m.face == nil {
		m.subject = nil
		return nil
	} else if resolved, err := m.face.ResolveCollision(m.Embeddings()); err != nil {
		return err
	} else if resolved {
		log.Debugf("faces: resolved collision with %s", m.face.ID)
	}

	// Clear references.
	m.face = nil
	m.subject = nil

	return nil
}

// Face returns a matching face entity if possible.
func (m *Marker) Face() (f *Face) {
	if m.face != nil {
		if m.FaceID == m.face.ID {
			return m.face
		}
	}

	// Add face if size
	if m.SubjectSrc != SrcAuto && m.FaceID == "" {
		if m.Size < face.ClusterMinSize || m.Score < face.ClusterMinScore {
			log.Debugf("faces: skipped adding face for low-quality marker %s, size %d, score %d", m.MarkerUID, m.Size, m.Score)
			return nil
		} else if emb := m.Embeddings(); len(emb) == 0 {
			log.Warnf("marker: %s has no embeddings", m.MarkerUID)
			return nil
		} else if f = NewFace(m.SubjectUID, m.SubjectSrc, emb); f == nil {
			log.Warnf("marker: failed adding face for id %s", m.MarkerUID)
			return nil
		} else if f = FirstOrCreateFace(f); f == nil {
			log.Warnf("marker: failed adding face for id %s", m.MarkerUID)
			return nil
		} else if err := f.MatchMarkers(Faceless); err != nil {
			log.Errorf("faces: %s (match markers)", err)
		}

		m.face = f
		m.FaceID = f.ID
		m.FaceDist = 0
	} else {
		m.face = FindFace(m.FaceID)
	}

	return m.face
}

// ClearFace removes an existing face association.
func (m *Marker) ClearFace() (updated bool, err error) {
	if m.FaceID == "" {
		return false, m.Matched()
	}

	updated = true

	// Remove face references.
	m.face = nil
	m.FaceID = ""
	m.MatchedAt = TimePointer()

	// Remove subject if set automatically.
	if m.SubjectSrc == SrcAuto {
		m.SubjectUID = ""
		err = m.Updates(Values{"FaceID": "", "FaceDist": -1.0, "SubjectUID": "", "MatchedAt": m.MatchedAt})
	} else {
		err = m.Updates(Values{"FaceID": "", "FaceDist": -1.0, "MatchedAt": m.MatchedAt})
	}

	return updated, err
}

// Matched updates the match timestamp.
func (m *Marker) Matched() error {
	m.MatchedAt = TimePointer()
	return UnscopedDb().Model(m).UpdateColumns(Values{"MatchedAt": m.MatchedAt}).Error
}

// FindMarker returns an existing row if exists.
func FindMarker(uid string) *Marker {
	var result Marker

	if err := Db().Where("marker_uid = ?", uid).First(&result).Error; err != nil {
		return nil
	}

	return &result
}

// UpdateOrCreateMarker updates a marker in the database or creates a new one if needed.
func UpdateOrCreateMarker(m *Marker) (*Marker, error) {
	const d = 0.07

	result := Marker{}

	if m.MarkerUID != "" {
		err := m.Save()
		log.Debugf("faces: saved marker %s for file %s", m.MarkerUID, m.FileUID)
		return m, err
	} else if err := Db().Where(`file_uid = ? AND x > ? AND x < ? AND y > ? AND y < ?`,
		m.FileUID, m.X-d, m.X+d, m.Y-d, m.Y+d).First(&result).Error; err == nil {

		if SrcPriority[m.MarkerSrc] < SrcPriority[result.MarkerSrc] {
			// Ignore.
			return &result, nil
		}

		err := result.Updates(map[string]interface{}{
			"MarkerType":     m.MarkerType,
			"MarkerSrc":      m.MarkerSrc,
			"FileArea":       m.FileArea,
			"X":              m.X,
			"Y":              m.Y,
			"W":              m.W,
			"H":              m.H,
			"Score":          m.Score,
			"Size":           m.Size,
			"LandmarksJSON":  m.LandmarksJSON,
			"EmbeddingsJSON": m.EmbeddingsJSON,
		})

		log.Debugf("faces: updated existing marker %s for file %s", result.MarkerUID, result.FileUID)

		return &result, err
	} else if err := m.Create(); err != nil {
		log.Debugf("faces: added marker %s for file %s", m.MarkerUID, m.FileUID)
		return m, err
	}

	return m, nil
}
