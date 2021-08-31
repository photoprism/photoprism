package entity

import (
	"crypto/sha1"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/photoprism/photoprism/internal/face"

	"github.com/photoprism/photoprism/pkg/clusters"
)

var faceMutex = sync.Mutex{}

// Face represents the face of a Subject.
type Face struct {
	ID              string          `gorm:"type:VARBINARY(42);primary_key;auto_increment:false;" json:"ID" yaml:"ID"`
	FaceSrc         string          `gorm:"type:VARBINARY(8);" json:"Src" yaml:"Src,omitempty"`
	SubjectUID      string          `gorm:"type:VARBINARY(42);index;" json:"SubjectUID" yaml:"SubjectUID,omitempty"`
	Samples         int             `json:"Samples" yaml:"Samples,omitempty"`
	SampleRadius    float64         `json:"SampleRadius" yaml:"SampleRadius,omitempty"`
	Collisions      int             `json:"Collisions" yaml:"Collisions,omitempty"`
	CollisionRadius float64         `json:"CollisionRadius" yaml:"CollisionRadius,omitempty"`
	EmbeddingJSON   json.RawMessage `gorm:"type:MEDIUMBLOB;" json:"-" yaml:"EmbeddingJSON,omitempty"`
	embedding       Embedding       `gorm:"-"`
	MatchedAt       *time.Time      `json:"MatchedAt" yaml:"MatchedAt,omitempty"`
	CreatedAt       time.Time       `json:"CreatedAt" yaml:"CreatedAt,omitempty"`
	UpdatedAt       time.Time       `json:"UpdatedAt" yaml:"UpdatedAt,omitempty"`
}

// UnknownFace can be used as a placeholder for unknown faces.
var UnknownFace = Face{
	ID:            UnknownID,
	FaceSrc:       SrcDefault,
	MatchedAt:     TimePointer(),
	SubjectUID:    UnknownPerson.SubjectUID,
	EmbeddingJSON: []byte{},
}

// Faceless can be used as argument to match unmatched face markers.
var Faceless = []string{""}

// CreateUnknownFace initializes the database with a placeholder for unknown faces.
func CreateUnknownFace() {
	_ = UnknownFace.Create()
}

// TableName returns the entity database table name.
func (Face) TableName() string {
	return "faces_dev5"
}

// NewFace returns a new face.
func NewFace(subjectUID, faceSrc string, embeddings Embeddings) *Face {
	result := &Face{
		SubjectUID: subjectUID,
		FaceSrc:    faceSrc,
	}

	if err := result.SetEmbeddings(embeddings); err != nil {
		log.Errorf("face: failed setting embeddings (%s)", err)
	}

	return result
}

// SetEmbeddings assigns face embeddings.
func (m *Face) SetEmbeddings(embeddings Embeddings) (err error) {
	m.embedding, m.SampleRadius, m.Samples = EmbeddingsMidpoint(embeddings)
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
func (m *Face) Embedding() Embedding {
	if len(m.EmbeddingJSON) == 0 {
		return Embedding{}
	} else if len(m.embedding) > 0 {
		return m.embedding
	} else if err := json.Unmarshal(m.EmbeddingJSON, &m.embedding); err != nil {
		log.Errorf("failed parsing face embedding json: %s", err)
	}

	return m.embedding
}

// Match tests if embeddings match this face.
func (m *Face) Match(embeddings Embeddings) (match bool, dist float64) {
	dist = -1

	if len(embeddings) == 0 {
		// Np embeddings, no match.
		return false, dist
	}

	faceEmbedding := m.Embedding()

	if len(faceEmbedding) == 0 {
		// Should never happen.
		return false, dist
	}

	// Calculate smallest distance to embeddings.
	for _, e := range embeddings {
		if d := clusters.EuclideanDistance(e, faceEmbedding); d < dist || dist < 0 {
			dist = d
		}
	}

	// Any reasons embeddings do not match this face?
	switch {
	case dist < 0:
		// Should never happen.
		return false, dist
	case dist > (m.SampleRadius + face.ClusterRadius):
		// Too far.
		return false, dist
	case m.CollisionRadius > 0.1 && dist > m.CollisionRadius:
		// Within radius of reported collisions.
		return false, dist
	}

	// If not, at least one of the embeddings match!
	return true, dist
}

// ReportCollision reports a collision with a different subject's face.
func (m *Face) ReportCollision(embeddings Embeddings) (reported bool, err error) {
	if m.SubjectUID == "" {
		// Ignore reports for anonymous faces.
		return false, nil
	} else if m.ID == "" {
		return false, fmt.Errorf("invalid face id")
	} else if len(m.EmbeddingJSON) == 0 {
		return false, fmt.Errorf("embedding must not be empty")
	}

	revise := false

	if match, dist := m.Match(embeddings); !match {
		// Embeddings don't match this face. Ignore.
		return false, nil
	} else if dist < 0 {
		// Should never happen.
		return false, fmt.Errorf("collision distance must be positive")
	} else if dist > 0.2 {
		m.MatchedAt = nil
		m.Collisions++
		m.CollisionRadius = dist - 0.1
		revise = true
	} else {
		// Don't set a radius yet if distance is very small.
		m.Collisions++
	}

	err = m.Updates(Values{"Collisions": m.Collisions, "CollisionRadius": m.CollisionRadius, "MatchedAt": m.MatchedAt})

	if err == nil && revise {
		var revised Markers
		revised, err = m.ReviseMatches()

		if n := len(revised); n > 0 {
			log.Infof("faces: revised %d matches after collision", n)
		}
	}

	return true, err
}

// ReviseMatches updates marker matches after face parameters have been changed.
func (m *Face) ReviseMatches() (revised Markers, err error) {
	var matches Markers

	if err := Db().Where("face_id = ?", m.ID).Where("marker_type = ?", MarkerFace).
		Find(&matches).Error; err != nil {
		log.Debugf("faces: %s (find matching markers)", err)
		return revised, err
	} else {
		for _, marker := range matches {
			if ok, _ := m.Match(marker.Embeddings()); !ok {
				if updated, err := marker.ClearFace(); err != nil {
					log.Debugf("faces: %s (revise match)", err)
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
		log.Warnf("faces: found %s with conflicting subject %s", m.ID, m.SubjectUID)
		return &result
	} else if createErr := m.Create(); createErr == nil {
		return m
	} else if err := UnscopedDb().Where("id = ?", m.ID).First(&result).Error; err == nil {
		log.Warnf("faces: found %s with conflicting subject %s", m.ID, m.SubjectUID)
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

	result := Face{}

	db := Db()
	db = db.Where("id = ?", id)

	if err := db.First(&result).Error; err != nil {
		return nil
	}

	return &result
}
