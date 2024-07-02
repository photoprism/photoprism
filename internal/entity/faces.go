package entity

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/tensorflow/face"
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
