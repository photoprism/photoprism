package query

import (
	"github.com/photoprism/photoprism/internal/entity"
)

// Faces returns (known) faces from the index.
func Faces(knownOnly bool) (result entity.Faces, err error) {
	stmt := Db().
		Where("id <> ?", entity.UnknownFace.ID).
		Order("id")

	if knownOnly {
		stmt = stmt.Where("subject_uid <> ''")
	}

	err = stmt.Find(&result).Error

	return result, err
}

// MatchKnownFaces matches known faces with file markers.
func MatchKnownFaces() (affected int64, err error) {
	faces, err := Faces(true)

	if err != nil {
		return affected, err
	}

	for _, match := range faces {
		if res := Db().Model(&entity.Marker{}).
			Where("face_id = ?", match.ID).
			Updates(entity.Values{"SubjectUID": match.SubjectUID, "SubjectSrc": entity.SrcAuto, "FaceID": ""}); res.Error != nil {
			return affected, err
		} else if res.RowsAffected > 0 {
			affected += res.RowsAffected
		}
	}

	return affected, nil
}

// PurgeAnonymousFaces removes anonymous faces from the index.
func PurgeAnonymousFaces() error {
	return UnscopedDb().Delete(
		entity.Face{},
		"id <> ? AND subject_uid = '' AND updated_at < ?", entity.UnknownFace.ID, entity.Yesterday()).Error
}

// ResetFaces removes all face clusters from the index.
func ResetFaces() error {
	return UnscopedDb().
		Delete(entity.Face{}, "id <> ? AND face_src = ''", entity.UnknownFace.ID).
		Error
}

// CountNewFaceMarkers returns the number of new face markers in the index.
func CountNewFaceMarkers() (n int) {
	var f entity.Face

	if err := Db().Where("id <> ?", entity.UnknownFace.ID).Order("created_at DESC").Take(&f).Error; err != nil {
		log.Debugf("faces: no existing clusters")
	}

	q := Db().Model(&entity.Markers{}).Where("marker_type = ? AND embeddings_json <> ''", entity.MarkerFace)

	if !f.CreatedAt.IsZero() {
		q = q.Where("created_at > ?", f.CreatedAt)
	}

	if err := q.Order("created_at DESC").Count(&n).Error; err != nil {
		log.Errorf("faces: %s (count new markers)", err)
	}

	return n
}
