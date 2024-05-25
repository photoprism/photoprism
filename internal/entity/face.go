package entity

import (
	"crypto/sha1"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/photoprism/photoprism/internal/face"
	"github.com/photoprism/photoprism/pkg/rnd"
)

var faceMutex = sync.Mutex{}
var UpdateFaces = atomic.Bool{}

// Face represents the face of a Subject.
type Face struct {
	ID              string          `gorm:"type:VARBINARY(64);primary_key;auto_increment:false;" json:"ID" yaml:"ID"`
	FaceSrc         string          `gorm:"type:VARBINARY(8);" json:"Src" yaml:"Src,omitempty"`
	FaceKind        int             `json:"Kind" yaml:"Kind,omitempty"`
	FaceHidden      bool            `json:"Hidden" yaml:"Hidden,omitempty"`
	SubjUID         string          `gorm:"type:VARBINARY(42);index;default:'';" json:"SubjUID" yaml:"SubjUID,omitempty"`
	Samples         int             `json:"Samples" yaml:"Samples,omitempty"`
	SampleRadius    float64         `json:"SampleRadius" yaml:"SampleRadius,omitempty"`
	Collisions      int             `json:"Collisions" yaml:"Collisions,omitempty"`
	CollisionRadius float64         `json:"CollisionRadius" yaml:"CollisionRadius,omitempty"`
	EmbeddingJSON   json.RawMessage `gorm:"type:MEDIUMBLOB;" json:"-" yaml:"EmbeddingJSON,omitempty"`
	embedding       face.Embedding  `gorm:"-" yaml:"-"`
	MatchedAt       *time.Time      `json:"MatchedAt" yaml:"MatchedAt,omitempty"`
	CreatedAt       time.Time       `json:"CreatedAt" yaml:"CreatedAt,omitempty"`
	UpdatedAt       time.Time       `json:"UpdatedAt" yaml:"UpdatedAt,omitempty"`
}

// Faceless can be used as argument to match unmatched face markers.
var Faceless = []string{""}

// TableName returns the entity table name.
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

// MatchId returns a compound id for matching.
func (m *Face) MatchId(f Face) string {
	if m.ID == "" || f.ID == "" {
		return ""
	}

	if m.ID < f.ID {
		return fmt.Sprintf("%s-%s", m.ID, f.ID)
	} else {
		return fmt.Sprintf("%s-%s", f.ID, m.ID)
	}
}

// SkipMatching checks whether the face should be skipped when matching.
func (m *Face) SkipMatching() bool {
	return m.FaceKind > 1 || m.Embedding().SkipMatching()
}

// SetEmbeddings assigns face embeddings.
func (m *Face) SetEmbeddings(embeddings face.Embeddings) (err error) {
	if len(embeddings) == 0 {
		return fmt.Errorf("invalid embedding")
	}

	m.embedding, m.SampleRadius, m.Samples = face.EmbeddingsMidpoint(embeddings)

	if len(m.embedding) != len(face.NullEmbedding) {
		return fmt.Errorf("embedding has invalid number of values")
	}

	// Limit sample radius to reduce false positives.
	if m.SampleRadius > 0.35 {
		m.SampleRadius = 0.35
	}

	m.EmbeddingJSON, err = json.Marshal(m.embedding)

	if err != nil {
		return err
	}

	s := sha1.Sum(m.EmbeddingJSON)

	// Update Face ID, Kind, and reset match timestamp,
	m.ID = base32.StdEncoding.EncodeToString(s[:])

	if k := int(m.embedding.Kind()); k > m.FaceKind {
		m.FaceKind = k
	}

	m.MatchedAt = nil

	return nil
}

// Matched updates the match timestamp.
func (m *Face) Matched() error {
	m.MatchedAt = TimePointer()
	return UnscopedDb().Model(m).UpdateColumns(Map{"MatchedAt": m.MatchedAt}).Error
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
		if d := e.Dist(faceEmbedding); d < dist || dist < 0 {
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
		log.Warnf("faces: %s has ambiguous subject %s with a similar face at dist %f with source %s", m.ID, SubjNames.Log(m.SubjUID), dist, SrcString(m.FaceSrc))

		m.FaceKind = int(face.AmbiguousFace)
		m.UpdatedAt = TimeStamp()
		m.MatchedAt = &m.UpdatedAt
		m.Collisions++
		m.CollisionRadius = dist
		UpdateFaces.Store(true)
		return true, m.Updates(Map{"Collisions": m.Collisions, "CollisionRadius": m.CollisionRadius, "FaceKind": m.FaceKind, "UpdatedAt": m.UpdatedAt, "MatchedAt": m.MatchedAt})
	} else {
		m.MatchedAt = nil
		m.Collisions++
		m.CollisionRadius = dist - 0.01
		UpdateFaces.Store(true)
	}

	err = m.Updates(Map{"Collisions": m.Collisions, "CollisionRadius": m.CollisionRadius, "MatchedAt": m.MatchedAt})

	if err != nil {
		return true, err
	}

	if revised, err := m.ReviseMatches(); err != nil {
		return true, err
	} else if r := len(revised); r > 0 {
		log.Infof("faces: resolved %d conflicts", r)
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
		log.Debugf("faces: found no matching markers for conflict resolution (%s)", err)
		return revised, err
	} else {
		for _, marker := range matches {
			if ok, _ := m.Match(marker.Embeddings()); !ok {
				if updated, err := marker.ClearFace(); err != nil {
					log.Debugf("faces: failed to remove match with marker (%s)", err) // Conflict resolution
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
		log.Debugf("faces: failed fetching markers matching face id %s (%s)", strings.Join(faceIds, ", "), err)
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
func (m *Face) SetSubjectUID(subjUid string) (err error) {
	// Update face.
	if err = m.Update("SubjUID", subjUid); err != nil {
		return err
	} else {
		m.SubjUID = subjUid
	}

	UpdateFaces.Store(true)

	// Update related markers.
	if err = Db().Model(&Marker{}).
		Where("face_id = ?", m.ID).
		Where("subj_src = ?", SrcAuto).
		Where("subj_uid <> ?", m.SubjUID).
		Where("marker_invalid = 0").
		UpdateColumns(Map{"subj_uid": m.SubjUID, "marker_review": false}).Error; err != nil {
		return err
	}

	return m.RefreshPhotos()
}

// RefreshPhotos flags related photos for metadata maintenance.
func (m *Face) RefreshPhotos() error {
	if m.ID == "" {
		return fmt.Errorf("empty face id")
	}

	UpdateFaces.Store(true)

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

// Create inserts the face to the database.
func (m *Face) Create() error {
	if m.ID == "" {
		return fmt.Errorf("empty id")
	}

	faceMutex.Lock()
	defer faceMutex.Unlock()

	UpdateFaces.Store(true)

	return Db().Create(m).Error
}

// Delete removes the face from the database.
func (m *Face) Delete() error {
	if m.ID == "" {
		return fmt.Errorf("empty id")
	}

	UpdateFaces.Store(true)

	// Remove face id from markers before deleting.
	if err := Db().Model(&Marker{}).
		Where("face_id = ?", m.ID).
		UpdateColumns(Map{"face_id": "", "face_dist": -1}).Error; err != nil {
		return err
	}

	return Db().Delete(m).Error
}

// Update a face property in the database.
func (m *Face) Update(attr string, value interface{}) error {
	if m.ID == "" {
		return fmt.Errorf("empty id")
	}

	UpdateFaces.Store(true)

	return UnscopedDb().Model(m).Update(attr, value).Error
}

// Updates face properties in the database.
func (m *Face) Updates(values interface{}) error {
	if m.ID == "" {
		return fmt.Errorf("empty id")
	}

	UpdateFaces.Store(true)

	return UnscopedDb().Model(m).Updates(values).Error
}

// FirstOrCreateFace returns the existing entity, inserts a new entity or nil in case of errors.
func FirstOrCreateFace(m *Face) *Face {
	if m == nil {
		return nil
	}

	if m.ID == "" {
		return nil
	}

	result := Face{}

	// Search existing face with the same ID. Report if found and it belongs to another person.
	if findErr := UnscopedDb().Where("id = ?", m.ID).First(&result).Error; findErr == nil && result.ID != "" {
		if m.SubjUID != result.SubjUID {
			log.Warnf("faces: %s has ambiguous subjects %s and %s", m.ID, SubjNames.Log(m.SubjUID), SubjNames.Log(result.SubjUID))
		}
		return &result
	} else if err := m.Create(); err == nil {
		UpdateFaces.Store(true)
		return m
	} else if findErr = UnscopedDb().Where("id = ?", m.ID).First(&result).Error; findErr == nil && result.ID != "" {
		if m.SubjUID != result.SubjUID {
			log.Warnf("faces: %s has ambiguous subjects %s and %s", m.ID, SubjNames.Log(m.SubjUID), SubjNames.Log(result.SubjUID))
		}
		return &result
	} else {
		log.Errorf("faces: failed to add %s (%s)", m.ID, err)
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
func ValidFaceCount(fileUid string) (c int) {
	if !rnd.IsUID(fileUid, FileUID) {
		return
	}

	if err := Db().Model(Marker{}).
		Where("file_uid = ? AND marker_type = ?", fileUid, MarkerFace).
		Where("marker_invalid = 0").
		Count(&c).Error; err != nil {
		log.Errorf("file: %s (count faces)", err)
		return 0
	} else {
		return c
	}
}
