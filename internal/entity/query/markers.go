package query

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
)

// MarkerByUID returns a Marker based on the UID.
func MarkerByUID(uid string) (*entity.Marker, error) {
	result := entity.Marker{}

	err := UnscopedDb().Where("marker_uid = ?", uid).First(&result).Error

	return &result, err
}

// Markers finds a list of file markers filtered by type, embeddings, and sorted by id.
func Markers(limit, offset int, markerType string, embeddings, subjects bool, matchedBefore time.Time) (result entity.Markers, err error) {
	db := Db()

	if markerType != "" {
		db = db.Where("marker_type = ?", markerType)
	}

	if embeddings {
		db = db.Where("LENGTH(embeddings_json) > 0")
	}

	if subjects {
		db = db.Where("subj_uid <> ''")
	}

	if !matchedBefore.IsZero() {
		db = db.Where("matched_at IS NULL OR matched_at < ?", matchedBefore)
	}

	db = db.Order("matched_at, marker_uid").Limit(limit).Offset(offset)

	err = db.Find(&result).Error

	return result, err
}

// UnmatchedFaceMarkers returns the next page of markers that still need matching, after the given uid.
//
// Paged by cursor, not offset: a run stamps what it matches, so rows leave this set as it reads, and
// an offset would skip whatever shifted into its window. A fixed offset also returns the markers a
// run visits without stamping until a batch holds nothing new and the run stops early.
func UnmatchedFaceMarkers(limit int, after string, matchedBefore *time.Time) (result entity.Markers, err error) {
	db := whereEmbeddingModel(Db().
		Where("marker_type = ?", entity.MarkerFace).
		Where("marker_invalid = 0").
		Where("LENGTH(embeddings_json) > 0"), face.EmbeddingModelName())

	if matchedBefore == nil {
		db = db.Where("matched_at IS NULL")
	} else if !matchedBefore.IsZero() {
		db = db.Where("matched_at IS NULL OR matched_at < ?", matchedBefore)
	}

	if after != "" {
		db = db.Where("marker_uid > ?", after)
	}

	// Ordered by the cursor column, so a page cannot shift under the run that is reading it.
	db = db.Order("marker_uid").Limit(limit)

	err = db.Find(&result).Error

	return result, err
}

// FaceMarkers returns all face markers sorted by id.
func FaceMarkers(limit, offset int) (result entity.Markers, err error) {
	err = whereEmbeddingModel(Db().
		Where("marker_type = ?", entity.MarkerFace), face.EmbeddingModelName()).
		Order("marker_uid").Limit(limit).Offset(offset).
		Find(&result).Error

	return result, err
}

// Embeddings returns existing face embeddings.
func Embeddings(single, unclustered bool, size, score int, model string) (result face.Embeddings, err error) {
	var col []string

	stmt := Db().
		Model(&entity.Marker{}).
		Where("marker_type = ?", entity.MarkerFace).
		Where("marker_invalid = 0").
		Where("LENGTH(embeddings_json) > 0").
		Order("marker_uid")

	stmt = whereEmbeddingModel(stmt, model)

	if size > 0 {
		sizeCond, sizeArgs := entity.ClusterSizeCond("", size)
		stmt = stmt.Where(sizeCond, sizeArgs...)
	}

	stmt = whereClusterScore(stmt, score)

	if unclustered {
		stmt = stmt.Where("face_id = ''")
	}

	if err := stmt.Pluck("embeddings_json", &col).Error; err != nil {
		return result, err
	}

	for _, embeddingsJson := range col {
		if embeddingsJson == "" {
			continue
		} else if embeddings, err := face.UnmarshalEmbeddings(embeddingsJson); err != nil {
			log.Warnf("faces: %s", err)
		} else if !embeddings.Empty() {
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

// MarkerCountsByFaceIDs returns a map of marker counts for the provided face IDs.
func MarkerCountsByFaceIDs(faceIDs []string) (map[string]int, error) {
	counts := make(map[string]int, len(faceIDs))

	if len(faceIDs) == 0 {
		return counts, nil
	}

	type row struct {
		FaceID string
		Count  int
	}

	var rows []row

	if err := Db().
		Model(&entity.Marker{}).
		Select("face_id, COUNT(*) AS count").
		Where("marker_invalid = 0").
		Where("marker_type = ?", entity.MarkerFace).
		Where("face_id IN (?)", faceIDs).
		Group("face_id").
		Scan(&rows).Error; err != nil {
		return counts, err
	}

	for _, r := range rows {
		counts[r.FaceID] = r.Count
	}

	return counts, nil
}

// RemoveInvalidMarkerReferences removes face and subject references from invalid markers.
func RemoveInvalidMarkerReferences() (removed int64, err error) {
	result := Db().
		Model(&entity.Marker{}).
		Where("marker_invalid = 1 AND (subj_uid <> '' OR face_id <> '')").
		UpdateColumns(entity.Values{"subj_uid": "", "face_id": "", "face_dist": -1.0, "matched_at": nil})

	return result.RowsAffected, result.Error
}

// RemoveNonExistentMarkerFaces removes non-existent face IDs from the markers table.
func RemoveNonExistentMarkerFaces() (removed int64, err error) {
	result := Db().
		Model(&entity.Marker{}).
		Where("marker_type = ?", entity.MarkerFace).
		Where(fmt.Sprintf("face_id <> '' AND face_id NOT IN (SELECT id FROM %s)", entity.Face{}.TableName())).
		UpdateColumns(entity.Values{"face_id": "", "face_dist": -1.0, "matched_at": nil})

	return result.RowsAffected, result.Error
}

// RemoveNonExistentMarkerSubjects removes non-existent subject UIDs from the markers table.
func RemoveNonExistentMarkerSubjects() (removed int64, err error) {
	result := Db().
		Model(&entity.Marker{}).
		Where(fmt.Sprintf("subj_uid <> '' AND subj_uid NOT IN (SELECT subj_uid FROM %s)", entity.Subject{}.TableName())).
		UpdateColumns(entity.Values{"subj_uid": "", "matched_at": nil})

	return result.RowsAffected, result.Error
}

// FixMarkerReferences repairs invalid or non-existent references in the markers table.
func FixMarkerReferences() (removed int64, err error) {
	if r, err := RemoveInvalidMarkerReferences(); err != nil {
		return removed, err
	} else {
		removed += r
	}

	if r, err := RemoveNonExistentMarkerFaces(); err != nil {
		return removed, err
	} else {
		removed += r
	}

	if r, err := RemoveNonExistentMarkerSubjects(); err != nil {
		return removed, err
	} else {
		removed += r
	}

	return removed, nil
}

// MarkersWithNonExistentReferences finds markers with non-existent face or subject references.
func MarkersWithNonExistentReferences() (faces entity.Markers, subjects entity.Markers, err error) {
	// Find markers with invalid face IDs.
	if res := Db().
		Where("marker_type = ?", entity.MarkerFace).
		Where(fmt.Sprintf("face_id <> '' AND face_id NOT IN (SELECT id FROM %s)", entity.Face{}.TableName())).
		Find(&faces); res.Error != nil {
		err = res.Error
	}

	// Find markers with invalid subject UIDs.
	if res := Db().
		Where(fmt.Sprintf("subj_uid <> '' AND subj_uid NOT IN (SELECT subj_uid FROM %s)", entity.Subject{}.TableName())).
		Find(&subjects); res.Error != nil {
		err = res.Error
	}

	return faces, subjects, err
}

// MarkersWithSubjectConflict finds markers with conflicting subjects.
func MarkersWithSubjectConflict() (results entity.Markers, err error) {
	faces := gorm.Expr(entity.Face{}.TableName())
	markers := gorm.Expr(entity.Marker{}.TableName())
	hasSubj := gorm.Expr("(?.subj_uid <> '' OR ?.subj_src <> '')", markers, markers)

	err = Db().
		Joins("JOIN ? f ON f.id = face_id AND f.subj_uid <> ?.subj_uid AND ?", faces, markers, hasSubj).
		Order("face_id").
		Find(&results).Error

	return results, err
}

// ResetFaceMarkerMatches removes automatically added subject and face references from the markers table.
func ResetFaceMarkerMatches() (removed int64, err error) {
	return resetFaceMarkerMatches(Db().Where("subj_src = ?", entity.SrcAuto))
}

// ResetAllFaceMarkerMatches clears the references of every face marker, including the ones a person
// or an XMP sidecar named. Those columns are the only record of a hand-verified identity, so a
// caller that measures cluster purity against them has to export them first.
func ResetAllFaceMarkerMatches() (removed int64, err error) {
	return resetFaceMarkerMatches(Db())
}

// resetFaceMarkerMatches clears the subject and face references of the face markers a scope selects.
// Geometry, embeddings, size and score are left alone, which lets a later run re-cluster without
// decoding a file. The columns are listed explicitly rather than written through the struct, and a
// test instance relies on that: a column this package does not name survives a reset.
func resetFaceMarkerMatches(scope *gorm.DB) (removed int64, err error) {
	res := scope.Model(&entity.Marker{}).
		Where("marker_type = ?", entity.MarkerFace).
		UpdateColumns(entity.Values{"marker_name": "", "subj_uid": "", "subj_src": "", "face_id": "", "face_dist": -1.0, "matched_at": nil})

	return res.RowsAffected, res.Error
}

// CountUnmatchedFaceMarkers counts the number of unmatched face markers in the index.
func CountUnmatchedFaceMarkers() (n int) {
	q := whereEmbeddingModel(Db().Model(&entity.Markers{}).
		Where("matched_at IS NULL AND marker_invalid = 0 AND LENGTH(embeddings_json) > 0").
		Where("marker_type = ?", entity.MarkerFace), face.EmbeddingModelName())

	if err := q.Count(&n).Error; err != nil {
		log.Errorf("faces: %s (count unmatched markers)", err)
	}

	return n
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

// RemoveOrphanMarkers removes markers without an existing file.
func RemoveOrphanMarkers() (removed int64, err error) {
	where := fmt.Sprintf("file_uid NOT IN (SELECT file_uid FROM %s)", entity.File{}.TableName())

	if res := UnscopedDb().
		Delete(&entity.Marker{}, where); res.Error != nil {
		return removed, fmt.Errorf("markers: %s (purge orphans)", res.Error)
	} else {
		removed += res.RowsAffected
	}

	return removed, nil
}
