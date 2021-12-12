package entity

import (
	"crypto/sha1"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/photoprism/photoprism/internal/face"
	"github.com/photoprism/photoprism/pkg/rnd"
)

var faceMutex = sync.Mutex{}

// Face represents the face of a Subject.
type Face struct {
	ID              string          `gorm:"type:VARBINARY(42);primary_key;auto_increment:false;" json:"ID" yaml:"ID"`
	FaceSrc         string          `gorm:"type:VARBINARY(8);" json:"Src" yaml:"Src,omitempty"`
	FaceHidden      bool            `json:"Hidden" yaml:"Hidden,omitempty"`
	SubjUID         string          `gorm:"type:VARBINARY(42);index;default:'';" json:"SubjUID" yaml:"SubjUID,omitempty"`
	Samples         int             `json:"Samples" yaml:"Samples,omitempty"`
	SampleRadius    float64         `json:"SampleRadius" yaml:"SampleRadius,omitempty"`
	Collisions      int             `json:"Collisions" yaml:"Collisions,omitempty"`
	CollisionRadius float64         `json:"CollisionRadius" yaml:"CollisionRadius,omitempty"`
	EmbeddingJSON   json.RawMessage `gorm:"type:MEDIUMBLOB;" json:"-" yaml:"EmbeddingJSON,omitempty"`
	embedding       face.Embedding  `gorm:"-"`
	MatchedAt       *time.Time      `json:"MatchedAt" yaml:"MatchedAt,omitempty"`
	CreatedAt       time.Time       `json:"CreatedAt" yaml:"CreatedAt,omitempty"`
	UpdatedAt       time.Time       `json:"UpdatedAt" yaml:"UpdatedAt,omitempty"`
}

// Faceless can be used as argument to match unmatched face markers.
var Faceless = []string{""}

// TableName returns the entity database table name.
func (Face) TableName() string {
	return "faces"
}

// NewFace returns a new face.
func NewFace(subjUID, faceSrc string, embeddings face.Embeddings) *Face {
	result := &Face{
		SubjUID: subjUID,
		FaceSrc: faceSrc,
	}

	if err := result.SetEmbeddings(embeddings); err != nil {
		log.Errorf("face: failed setting embeddings (%s)", err)
	}

	return result
}

// Unsuitable tests if the face is unsuitable for clustering and matching.
func (m *Face) Unsuitable() bool {
	return m.Embedding().Unsuitable()
}

// SetEmbeddings assigns face embeddings.
func (m *Face) SetEmbeddings(embeddings face.Embeddings) (err error) {
	m.embedding, m.SampleRadius, m.Samples = face.EmbeddingsMidpoint(embeddings)

	// Limit sample radius to reduce false positives.
	if m.SampleRadius > 0.35 {
		m.SampleRadius = 0.35
	}

	m.EmbeddingJSON, err = json.Marshal(m.embedding)

	if err != nil {
		return err
	}

	s := sha1.Sum(m.EmbeddingJSON)
	m.ID = base32.StdEncoding.EncodeToString(s[:])
	m.UpdatedAt = TimeStamp()

	// Reset match timestamp.
	m.MatchedAt = nil

	if m.CreatedAt.IsZero() {
		m.CreatedAt = m.UpdatedAt
	}

	return nil
}

// Matched updates the match timestamp.
func (m *Face) Matched() error {
	m.MatchedAt = TimePointer()
	return UnscopedDb().Model(m).UpdateColumns(Values{"MatchedAt": m.MatchedAt}).Error
}

// Embedding returns parsed face embedding.
func (m *Face) Embedding() face.Embedding {
	if len(m.EmbeddingJSON) == 0 {
		return face.Embedding{}
	} else if len(m.embedding) > 0 {
		return m.embedding
	} else if err := json.Unmarshal(m.EmbeddingJSON, &m.embedding); err != nil {
		log.Errorf("failed parsing face embedding json: %s", err)
	}

	return m.embedding
}

// Match tests if embeddings match this face.
func (m *Face) Match(embeddings face.Embeddings) (match bool, dist float64) {
	dist = -1

	if embeddings.Empty() {
		// Np embeddings, no match.
		return false, dist
	}

	faceEmbedding := m.Embedding()

	if len(faceEmbedding) == 0 {
		// Should never happen.
		return false, dist
	}

	// Calculate the smallest distance to embeddings.
	for _, e := range embeddings {
		if d := e.Distance(faceEmbedding); d < dist || dist < 0 {
			dist = d
		}
	}

	// Any reasons embeddings do not match this face?
	switch {
	case dist < 0:
		// Should never happen.
		return false, dist
	case dist > (m.SampleRadius + face.MatchDist):
		// Too far.
		return false, dist
	case m.CollisionRadius > 0.1 && dist > m.CollisionRadius:
		// Within radius of reported collisions.
		return false, dist
	}

	// If not, at least one of the embeddings match!
	return true, dist
}

// ResolveCollision resolves a collision with a different subject's face.
func (m *Face) ResolveCollision(embeddings face.Embeddings) (resolved bool, err error) {
	if m.SubjUID == "" {
		// Ignore reports for anonymous faces.
		return false, nil
	} else if m.ID == "" {
		return false, fmt.Errorf("invalid face id")
	} else if len(m.EmbeddingJSON) == 0 {
		return false, fmt.Errorf("embedding must not be empty")
	}

	if match, dist := m.Match(embeddings); !match {
		// Embeddings don't match this face. Ignore.
		return false, nil
	} else if dist < 0 {
		// Should never happen.
		return false, fmt.Errorf("collision distance must be positive")
	} else if dist < 0.02 {
		// Ignore if distance is very small as faces may belong to the same person.
		log.Warnf("face %s: clearing ambiguous subject %s, similar face at dist %f with source %s", m.ID, m.SubjUID, dist, SrcString(m.FaceSrc))

		// Reset subject UID just in case.
		m.SubjUID = ""

		return false, m.Updates(Values{"SubjUID": m.SubjUID})
	} else {
		m.MatchedAt = nil
		m.Collisions++
		m.CollisionRadius = dist - 0.01
	}

	err = m.Updates(Values{"Collisions": m.Collisions, "CollisionRadius": m.CollisionRadius, "MatchedAt": m.MatchedAt})

	if err != nil {
		return true, err
	}

	if revised, err := m.ReviseMatches(); err != nil {
		return true, err
	} else if r := len(revised); r > 0 {
		log.Infof("faces: revised %d matches after conflict", r)
	}

	return true, nil
}

// ReviseMatches updates marker matches after face parameters have been changed.
func (m *Face) ReviseMatches() (revised Markers, err error) {
	if m.ID == "" {
		return revised, fmt.Errorf("empty face id")
	}

	var matches Markers

	if err := Db().Where("face_id = ?", m.ID).Where("marker_type = ?", MarkerFace).
		Find(&matches).Error; err != nil {
		log.Debugf("faces: %s (revise matches)", err)
		return revised, err
	} else {
		for _, marker := range matches {
			if ok, _ := m.Match(marker.Embeddings()); !ok {
				if updated, err := marker.ClearFace(); err != nil {
					log.Debugf("faces: %s (revise matches)", err)
					return revised, err
				} else if updated {
					revised = append(revised, marker)
				}
			}
		}
	}

	return revised, nil
}

// MatchMarkers finds and references matching markers.
func (m *Face) MatchMarkers(faceIds []string) error {
	var markers Markers

	err := Db().
		Where("marker_invalid = 0 AND marker_type = ? AND face_id IN (?)", MarkerFace, faceIds).
		Find(&markers).Error

	if err != nil {
		log.Debugf("faces: %s (match markers)", err)
		return err
	}

	for _, marker := range markers {
		if ok, dist := m.Match(marker.Embeddings()); !ok {
			// Ignore.
		} else if _, err = marker.SetFace(m, dist); err != nil {
			return err
		}
	}

	return nil
}

// SetSubjectUID updates the face's subject uid and related markers.
func (m *Face) SetSubjectUID(subjUID string) (err error) {
	// Update face.
	if err = m.Update("SubjUID", subjUID); err != nil {
		return err
	} else {
		m.SubjUID = subjUID
	}

	// Update related markers.
	if err = Db().Model(&Marker{}).
		Where("face_id = ?", m.ID).
		Where("subj_src = ?", SrcAuto).
		Where("subj_uid <> ?", m.SubjUID).
		Where("marker_invalid = 0").
		UpdateColumns(Values{"subj_uid": m.SubjUID, "marker_review": false}).Error; err != nil {
		return err
	}

	return m.RefreshPhotos()
}

// RefreshPhotos flags related photos for metadata maintenance.
func (m *Face) RefreshPhotos() error {
	if m.ID == "" {
		return fmt.Errorf("empty face id")
	}

	var err error
	switch DbDialect() {
	case MySQL:
		update := fmt.Sprintf(`UPDATE photos p JOIN files f ON f.photo_id = p.id JOIN %s m ON m.file_uid = f.file_uid
			SET p.checked_at = NULL WHERE m.face_id = ?`, Marker{}.TableName())
		err = UnscopedDb().Exec(update, m.ID).Error
	default:
		update := fmt.Sprintf(`UPDATE photos SET checked_at = NULL WHERE id IN (SELECT f.photo_id FROM files f
			JOIN %s m ON m.file_uid = f.file_uid WHERE m.face_id = ?)`, Marker{}.TableName())
		err = UnscopedDb().Exec(update, m.ID).Error
	}

	return err
}

// Hide hides the face by default.
func (m *Face) Hide() (err error) {
	return m.Update("FaceHidden", true)
}

// Show shows the face by default.
func (m *Face) Show() (err error) {
	return m.Update("FaceHidden", false)
}

// Save updates the existing or inserts a new face.
func (m *Face) Save() error {
	faceMutex.Lock()
	defer faceMutex.Unlock()

	return Save(m, "ID")
}

// Create inserts the face to the database.
func (m *Face) Create() error {
	faceMutex.Lock()
	defer faceMutex.Unlock()

	return Db().Create(m).Error
}

// Delete removes the face from the database.
func (m *Face) Delete() error {
	// Remove face id from markers before deleting.
	if err := Db().Model(&Marker{}).
		Where("face_id = ?", m.ID).
		UpdateColumns(Values{"face_id": "", "face_dist": -1}).Error; err != nil {
		return err
	}

	return Db().Delete(m).Error
}

// Update a face property in the database.
func (m *Face) Update(attr string, value interface{}) error {
	return UnscopedDb().Model(m).Update(attr, value).Error
}

// Updates face properties in the database.
func (m *Face) Updates(values interface{}) error {
	return UnscopedDb().Model(m).Updates(values).Error
}

// FirstOrCreateFace returns the existing entity, inserts a new entity or nil in case of errors.
func FirstOrCreateFace(m *Face) *Face {
	result := Face{}

	if err := UnscopedDb().Where("id = ?", m.ID).First(&result).Error; err == nil {
		log.Warnf("faces: %s has ambiguous subject %s", m.ID, m.SubjUID)
		return &result
	} else if createErr := m.Create(); createErr == nil {
		return m
	} else if err := UnscopedDb().Where("id = ?", m.ID).First(&result).Error; err == nil {
		log.Warnf("faces: %s has ambiguous subject %s", m.ID, m.SubjUID)
		return &result
	} else {
		log.Errorf("faces: %s when trying to create %s", createErr, m.ID)
	}

	return nil
}

// FindFace returns an existing entity if exists.
func FindFace(id string) *Face {
	if id == "" {
		return nil
	}

	f := Face{}

	if err := Db().Where("id = ?", strings.ToUpper(id)).First(&f).Error; err != nil {
		return nil
	}

	return &f
}

// ValidFaceCount counts the number of valid face markers for a file uid.
func ValidFaceCount(fileUID string) (c int) {
	if !rnd.IsPPID(fileUID, 'f') {
		return
	}

	if err := Db().Model(Marker{}).
		Where("file_uid = ? AND marker_type = ?", fileUID, MarkerFace).
		Where("marker_invalid = 0").
		Count(&c).Error; err != nil {
		log.Errorf("file: %s (count faces)", err)
		return 0
	} else {
		return c
	}
}
