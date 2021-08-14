package query

import (
	"github.com/photoprism/photoprism/internal/entity"
)

// Faces returns (known) faces from the index.
func Faces(knownOnly bool) (result entity.Faces, err error) {
	stmt := Db().
		Order("id")

	if knownOnly {
		stmt = stmt.Where("person_uid <> ''")
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
			Updates(entity.Val{"RefUID": match.PersonUID, "RefSrc": entity.SrcPeople, "FaceID": ""}); res.Error != nil {
			return affected, err
		} else if res.RowsAffected > 0 {
			affected += res.RowsAffected
		}
	}

	return affected, nil
}

// PurgeUnknownFaces removes unknown faces from the index.
func PurgeUnknownFaces() error {
	return UnscopedDb().Delete(
		entity.Face{},
		"person_uid = '' AND updated_at < ?", entity.Yesterday()).Error
}
