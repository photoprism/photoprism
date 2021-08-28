package query

import (
	"fmt"
	"time"

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
func Markers(limit, offset int, markerType string, embeddings, subjects bool, matchedBefore time.Time) (result entity.Markers, err error) {
	db := Db()

	if markerType != "" {
		db = db.Where("marker_type = ?", markerType)
	}

	if embeddings {
		db = db.Where("embeddings_json <> ''")
	}

	if subjects {
		db = db.Where("subject_uid <> ''")
	}

	if !matchedBefore.IsZero() {
		db = db.Where("matched_at IS NULL OR matched_at < ?", matchedBefore)
	}

	db = db.Order("matched_at, id").Limit(limit).Offset(offset)

	err = db.Find(&result).Error

	return result, err
}

// Embeddings returns existing face embeddings.
func Embeddings(single, unclustered bool, score int) (result entity.Embeddings, err error) {
	var col []string

	stmt := Db().
		Model(&entity.Marker{}).
		Where("marker_type = ?", entity.MarkerFace).
		Where("marker_invalid = 0").
		Where("embeddings_json <> ''").
		Order("id")

	if score > 0 {
		stmt = stmt.Where("score >= ?", score)
	}

	if unclustered {
		stmt = stmt.Where("face_id = ''")
	}

	if err := stmt.Pluck("embeddings_json", &col).Error; err != nil {
		return result, err
	}

	for _, embeddingsJson := range col {
		if embeddings := entity.UnmarshalEmbeddings(embeddingsJson); len(embeddings) > 0 {
			if single {
				// Single embedding per face detected.
				result = append(result, embeddings[0])
			} else {
				// Return all embedding otherwise.
				result = append(result, embeddings...)
			}
		}
	}

	return result, nil
}

// RemoveInvalidMarkerReferences deletes invalid reference IDs from the markers table.
func RemoveInvalidMarkerReferences() (removed int64, err error) {
	// Remove subject and face relationships for invalid markers.
	if res := Db().
		Model(&entity.Marker{}).
		Where("marker_invalid = 1 AND (subject_uid <> '' OR face_id <> '')").
		UpdateColumns(entity.Values{"subject_uid": "", "face_id": ""}); res.Error != nil {
		return removed, res.Error
	} else {
		removed += res.RowsAffected
	}

	// Remove invalid face IDs.
	if res := Db().
		Model(&entity.Marker{}).
		Where("marker_type = ?", entity.MarkerFace).
		Where(fmt.Sprintf("face_id <> '' AND face_id NOT IN (SELECT id FROM %s)", entity.Face{}.TableName())).
		UpdateColumns(entity.Values{"face_id": ""}); res.Error != nil {
		return removed, res.Error
	} else {
		removed += res.RowsAffected
	}

	// Remove invalid subject UIDs.
	if res := Db().
		Model(&entity.Marker{}).
		Where(fmt.Sprintf("subject_uid <> '' AND subject_uid NOT IN (SELECT subject_uid FROM %s)", entity.Subject{}.TableName())).
		UpdateColumns(entity.Values{"subject_uid": ""}); res.Error != nil {
		return removed, res.Error
	} else {
		removed += res.RowsAffected
	}

	return removed, nil
}

// MarkersWithInvalidReferences finds markers with invalid references.
func MarkersWithInvalidReferences() (faces entity.Markers, subjects entity.Markers, err error) {
	// Find markers with invalid face IDs.
	if res := Db().
		Where("marker_type = ?", entity.MarkerFace).
		Where(fmt.Sprintf("face_id <> '' AND face_id NOT IN (SELECT id FROM %s)", entity.Face{}.TableName())).
		Find(&faces); res.Error != nil {
		err = res.Error
	}

	// Find markers with invalid subject UIDs.
	if res := Db().
		Where(fmt.Sprintf("subject_uid <> '' AND subject_uid NOT IN (SELECT subject_uid FROM %s)", entity.Subject{}.TableName())).
		Find(&subjects); res.Error != nil {
		err = res.Error
	}

	return faces, subjects, err
}

// ResetFaceMarkerMatches removes automatically added subject and face references from the markers table.
func ResetFaceMarkerMatches() (removed int64, err error) {
	res := Db().Model(&entity.Marker{}).
		Where("subject_src <> ? AND marker_type = ?", entity.SrcManual, entity.MarkerFace).
		UpdateColumns(entity.Values{"subject_uid": "", "subject_src": "", "face_id": "", "matched_at": nil})

	return res.RowsAffected, res.Error
}

// CountUnmatchedFaceMarkers counts the number of unmatched face markers in the index.
func CountUnmatchedFaceMarkers() (n int, matchedBefore time.Time) {
	var f entity.Face

	if err := Db().Where("face_src <> ?", entity.SrcDefault).
		Order("updated_at DESC").Limit(1).Take(&f).Error; err != nil || f.UpdatedAt.IsZero() {
		return 0, matchedBefore
	}

	matchedBefore = time.Now().UTC().Round(time.Second).Add(-2 * time.Hour)

	if f.UpdatedAt.Before(matchedBefore) {
		matchedBefore = f.UpdatedAt.Add(time.Second)
	}

	q := Db().Model(&entity.Markers{}).
		Where("marker_type = ?", entity.MarkerFace).
		Where("face_id = '' AND subject_src = '' AND marker_invalid = 0 AND embeddings_json <> ''").
		Where("matched_at IS NULL OR matched_at < ?", matchedBefore)

	if err := q.Count(&n).Error; err != nil {
		log.Errorf("faces: %s (count unmatched markers)", err)
	}

	return n, matchedBefore
}

// CountMarkers counts the number of face markers in the index.
func CountMarkers(markerType string) (n int) {
	q := Db().Model(&entity.Markers{})

	if markerType != "" {
		q = q.Where("marker_type = ?", markerType)
	}

	if err := q.Count(&n).Error; err != nil {
		log.Errorf("faces: %s (count markers)", err)
	}

	return n
}
