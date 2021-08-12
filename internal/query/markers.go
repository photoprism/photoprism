package query

import (
	"github.com/photoprism/photoprism/internal/entity"
)

// MarkerByID returns a Marker based on the ID.
func MarkerByID(id uint) (marker entity.Marker, err error) {
	if err := UnscopedDb().Where("id = ?", id).
		First(&marker).Error; err != nil {
		return marker, err
	}

	return marker, nil
}

// Markers finds a list of file markers filtered by type, embeddings, and sorted by id.
func Markers(limit, offset int, markerType string, embeddings, noRef bool) (result entity.Markers, err error) {
	stmt := Db()

	if markerType != "" {
		stmt = stmt.Where("marker_type = ?", markerType)
	}

	if embeddings {
		stmt = stmt.Where("embeddings <> ''")
	}

	if noRef {
		stmt = stmt.Where("ref = ''")
	}

	stmt = stmt.Order("id").Limit(limit).Offset(offset)
	err = stmt.Find(&result).Error

	return result, err
}

// Embeddings finds all face embeddings.
func Embeddings() (result entity.Embeddings, err error) {
	var col []string

	stmt := Db().
		Model(&entity.Marker{}).
		Where("marker_type = ?", entity.MarkerFace).
		Where("embeddings <> ''").
		Order("id")

	if err := stmt.Pluck("embeddings", &col).Error; err != nil {
		return result, err
	}

	for _, embeddingsJson := range col {
		if embeddings := entity.UnmarshalEmbeddings(embeddingsJson); len(embeddings) > 0 {
			result = append(result, embeddings...)
		}
	}

	return result, nil
}
