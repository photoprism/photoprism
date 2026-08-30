package entity

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/ai/face"
)

// Faces represents a Face slice.
type Faces []Face

// Embeddings returns all face embeddings in this slice.
func (f Faces) Embeddings() (embeddings face.Embeddings) {
	for _, m := range f {
		embeddings = append(embeddings, m.Embedding())
	}

	return embeddings
}

// EmbedModel returns the embedding model shared by all faces in this slice, and reports
// whether they belong to one embedding space. Legacy rows without a recorded model are
// FaceNet, so they resolve to the name their siblings carry.
func (f Faces) EmbedModel() (model face.ModelName, ok bool) {
	for _, m := range f {
		if m.EmbedModel != "" {
			model = m.EmbedModel
			break
		}
	}

	for _, m := range f {
		if !face.ModelsComparable(m.EmbedModel, model) {
			return model, false
		}
	}

	return model, true
}

// CollisionBound returns the tightest active collision radius in this slice, with the collision
// count of the cluster carrying it.
//
// Dropped when it falls inside extent: a midpoint reaches further than its sources, so a tighter
// bound would refuse the members the merge was made of.
func (f Faces) CollisionBound(extent float64) (radius float64, collisions int) {
	for _, m := range f {
		if m.CollisionRadius > face.CollisionDist && (radius == 0 || m.CollisionRadius < radius) {
			radius, collisions = m.CollisionRadius, m.Collisions
		}
	}

	if radius < extent {
		return 0, 0
	}

	return radius, collisions
}

// IDs returns all face IDs in this slice.
func (f Faces) IDs() (ids []string) {
	for _, m := range f {
		ids = append(ids, m.ID)
	}

	return ids
}

// Delete (soft) deletes all subjects.
func (f Faces) Delete() error {
	for _, m := range f {
		if err := m.Delete(); err != nil {
			return err
		}
	}

	return nil
}

// OrphanFaces returns unused faces.
func OrphanFaces() (Faces, error) {
	orphans := Faces{}

	err := Db().
		Where(fmt.Sprintf("id NOT IN (SELECT DISTINCT face_id FROM %s)", Marker{}.TableName())).
		Find(&orphans).Error

	return orphans, err
}

// DeleteOrphanFaces finds and (soft) deletes all unused face clusters.
func DeleteOrphanFaces() (count int, err error) {
	orphans, err := OrphanFaces()

	if err != nil {
		return 0, err
	}

	return len(orphans), orphans.Delete()
}
