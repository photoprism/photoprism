package query

import (
	"github.com/photoprism/photoprism/internal/entity"
)

// Faces returns all (known) faces from the index.
func Faces(knownOnly bool) (result entity.Faces, err error) {
	stmt := Db().
		Where("face_src <> ?", entity.SrcDefault).
		Order("id")

	if knownOnly {
		stmt = stmt.Where("subject_uid <> ''")
	}

	err = stmt.Find(&result).Error

	return result, err
}

// MatchFaceMarkers matches markers with known faces.
func MatchFaceMarkers() (affected int64, err error) {
	faces, err := Faces(true)

	if err != nil {
		return affected, err
	}

	for _, match := range faces {
		if res := Db().Model(&entity.Marker{}).
			Where("face_id = ?", match.ID).
			Where("subject_src = ?", entity.SrcAuto).
			Where("subject_uid <> ?", match.SubjectUID).
			Updates(entity.Values{"SubjectUID": match.SubjectUID}); res.Error != nil {
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

// CountNewFaceMarkers returns the number of new face markers in the index.
func CountNewFaceMarkers() (n int) {
	var f entity.Face

	if err := Db().Where("face_src = ?", entity.SrcAuto).Order("created_at DESC").Take(&f).Error; err != nil {
		log.Debugf("faces: no existing clusters")
	}

	q := Db().Model(&entity.Markers{}).Where("marker_type = ? AND marker_invalid = 0 AND embeddings_json <> ''", entity.MarkerFace)

	if !f.CreatedAt.IsZero() {
		q = q.Where("created_at > ?", f.CreatedAt)
	}

	if err := q.Order("created_at DESC").Count(&n).Error; err != nil {
		log.Errorf("faces: %s (count new markers)", err)
	}

	return n
}
