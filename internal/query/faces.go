package query

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/entity"
)

// Faces returns all (known / unmatched) faces from the index.
func Faces(knownOnly, unmatched bool) (result entity.Faces, err error) {
	stmt := Db().Where("face_src <> ?", entity.SrcDefault)

	if unmatched {
		stmt = stmt.Where("matched_at IS NULL")
	}

	if knownOnly {
		stmt = stmt.Where("subject_uid <> ''")
	}

	err = stmt.Order("subject_uid, samples DESC").Find(&result).Error

	return result, err
}

// ManuallyAddedFaces returns all manually added face clusters.
func ManuallyAddedFaces() (result entity.Faces, err error) {
	err = Db().
		Where("face_src = ?", entity.SrcManual).
		Where("subject_uid <> ''").Order("subject_uid, samples DESC").
		Find(&result).Error

	return result, err
}

// MatchFaceMarkers matches markers with known faces.
func MatchFaceMarkers() (affected int64, err error) {
	faces, err := Faces(true, false)

	if err != nil {
		return affected, err
	}

	for _, f := range faces {
		if res := Db().Model(&entity.Marker{}).
			Where("face_id = ?", f.ID).
			Where("subject_src = ?", entity.SrcAuto).
			Where("subject_uid <> ?", f.SubjectUID).
			Updates(entity.Values{"SubjectUID": f.SubjectUID}); res.Error != nil {
			return affected, err
		} else if res.RowsAffected > 0 {
			affected += res.RowsAffected
		}
	}

	return affected, nil
}

// RemoveAnonymousFaceClusters removes anonymous faces from the index.
func RemoveAnonymousFaceClusters() (removed int64, err error) {
	res := UnscopedDb().Delete(
		entity.Face{},
		"face_src = ? AND subject_uid = ''", entity.SrcAuto)

	return res.RowsAffected, res.Error
}

// RemoveAutoFaceClusters removes automatically added face clusters from the index.
func RemoveAutoFaceClusters() (removed int64, err error) {
	res := UnscopedDb().
		Delete(entity.Face{}, "id <> ? AND face_src = ?", entity.UnknownFace.ID, entity.SrcAuto)

	return res.RowsAffected, res.Error
}

// CountNewFaceMarkers counts the number of new face markers in the index.
func CountNewFaceMarkers(size, score int) (n int) {
	var f entity.Face

	if err := Db().Where("face_src = ?", entity.SrcAuto).
		Order("created_at DESC").Limit(1).Take(&f).Error; err != nil {
		log.Debugf("faces: no existing clusters")
	}

	q := Db().Model(&entity.Markers{}).
		Where("marker_type = ?", entity.MarkerFace).
		Where("face_id = '' AND marker_invalid = 0 AND embeddings_json <> ''")

	if size > 0 {
		q = q.Where("size >= ?", size)
	}

	if score > 0 {
		q = q.Where("score >= ?", score)
	}

	if !f.CreatedAt.IsZero() {
		q = q.Where("created_at > ?", f.CreatedAt)
	}

	if err := q.Count(&n).Error; err != nil {
		log.Errorf("faces: %s (count new markers)", err)
	}

	return n
}

// RemoveUnusedFaces removes unused faces from the index.
func RemoveUnusedFaces(faceIds []string) (removed int64, err error) {
	// Remove invalid face IDs.
	if res := Db().
		Where("id IN (?)", faceIds).
		Where(fmt.Sprintf("id NOT IN (SELECT face_id FROM %s)", entity.Marker{}.TableName())).
		Delete(&entity.Face{}); res.Error != nil {
		return removed, res.Error
	} else {
		removed += res.RowsAffected
	}

	return removed, nil
}

// MergeFaces returns a new face that replaces multiple others.
func MergeFaces(merge entity.Faces) (merged *entity.Face, err error) {
	if len(merge) < 2 {
		// Nothing to merge.
		return merged, fmt.Errorf("at least two faces required for merging")
	}

	// Create merged face cluster.
	if merged = entity.NewFace(merge[0].SubjectUID, merge[0].FaceSrc, merge.Embeddings()); merged == nil {
		return merged, fmt.Errorf("merged face must not be nil")
	} else if err := merged.Create(); err != nil {
		return merged, err
	} else if err := merged.MatchMarkers(append(merge.IDs(), "")); err != nil {
		return merged, err
	}

	// RemoveUnusedFaces removes unused faces from the index.
	if removed, err := RemoveUnusedFaces(merge.IDs()); err != nil {
		log.Errorf("faces: %s", err)
	} else {
		log.Debugf("faces: removed %d unused faces", removed)
	}

	return merged, err
}
