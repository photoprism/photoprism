package query

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/entity"
)

// Faces returns all (known) faces from the index.
func Faces(knownOnly bool, src string) (result entity.Faces, err error) {
	stmt := Db()

	if src == "" {
		stmt = stmt.Where("face_src <> ?", entity.SrcDefault)
	} else {
		stmt = stmt.Where("face_src = ?", src)
	}

	if knownOnly {
		stmt = stmt.Where("subject_uid <> ''").Order("subject_uid, samples DESC")
	} else {
		stmt = stmt.Order("samples DESC")
	}

	err = stmt.Find(&result).Error

	return result, err
}

// MatchFaceMarkers matches markers with known faces.
func MatchFaceMarkers() (affected int64, err error) {
	faces, err := Faces(true, "")

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

		if err := f.UpdateMatchTime(); err != nil {
			return affected, err
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
func CountNewFaceMarkers() (n int) {
	var f entity.Face

	if err := Db().Where("face_src = ?", entity.SrcAuto).
		Order("created_at DESC").Limit(1).Take(&f).Error; err != nil {
		log.Debugf("faces: no existing clusters")
	}

	q := Db().Model(&entity.Markers{}).
		Where("marker_type = ?", entity.MarkerFace).
		Where("face_id = '' AND marker_invalid = 0 AND embeddings_json <> ''")

	if !f.CreatedAt.IsZero() {
		q = q.Where("created_at > ?", f.CreatedAt)
	}

	if err := q.Count(&n).Error; err != nil {
		log.Errorf("faces: %s (count new markers)", err)
	}

	return n
}

// CountUnmatchedFaceMarkers counts the number of unmatched face markers in the index.
func CountUnmatchedFaceMarkers() (n int) {
	var f entity.Face

	if err := Db().Where("face_src = ?", entity.SrcAuto).Order("matched_at ASC").Limit(1).Take(&f).Error; err != nil {
		log.Debugf("faces: no unmatched clusters")
	}

	q := Db().Model(&entity.Markers{}).
		Where("marker_type = ?", entity.MarkerFace).
		Where("face_id = '' AND subject_src = '' AND marker_invalid = 0 AND embeddings_json <> ''")

	if f.MatchedAt != nil {
		q = q.Where("updated_at > ?", f.MatchedAt)
	}

	if err := q.Count(&n).Error; err != nil {
		log.Errorf("faces: %s (count unmatched markers)", err)
	}

	return n
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
	}

	// Update marker matches.
	if err := Db().Model(&entity.Marker{}).Where("face_id IN (?)", merge.IDs()).
		Updates(entity.Values{"face_id": merged.ID, "subject_uid": merged.SubjectUID}).Error; err != nil {
		return merged, err
	}

	// Delete merged faces.
	if err := Db().Where("id IN (?) AND id <> ?", merge.IDs(), merged.ID).Delete(&entity.Face{}).Error; err != nil {
		return merged, err
	}

	// Find and reference additional matching markers.
	if err := merged.MatchMarkers(); err != nil {
		return merged, err
	}

	return merged, err
}
