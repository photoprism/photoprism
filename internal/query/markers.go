package query

import (
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/txt"
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
func Markers(limit, offset int, markerType string, embeddings, unmatched bool) (result entity.Markers, err error) {
	db := Db()

	if markerType != "" {
		db = db.Where("marker_type = ?", markerType)
	}

	if embeddings {
		db = db.Where("embeddings_json <> ''")
	}

	if unmatched {
		db = db.Where("subject_uid = ''")
	}

	db = db.Order("id").Limit(limit).Offset(offset)

	err = db.Find(&result).Error

	return result, err
}

// Embeddings returns existing face embeddings.
func Embeddings(single bool) (result entity.Embeddings, err error) {
	var col []string

	stmt := Db().
		Model(&entity.Marker{}).
		Where("marker_type = ?", entity.MarkerFace).
		Where("embeddings_json <> ''").
		Order("id")

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

// MatchMarkersWithSubjects automatically creates and assigns subjects to markers.
func MatchMarkersWithSubjects() (affected int, err error) {
	var markers entity.Markers

	if err := Db().
		Where("face_id <> '' AND subject_uid = '' AND subject_src = ''").
		Where("marker_invalid = 0 AND marker_type = ?", entity.MarkerFace).
		Where("marker_name <> ''").
		Order("marker_name").
		Find(&markers).Error; err != nil {
		return affected, err
	} else if len(markers) == 0 {
		return affected, nil
	}

	for _, m := range markers {
		faceId := m.FaceID

		if subj := entity.NewSubject(m.MarkerName, entity.SubjectPerson, entity.SrcMarker); subj == nil {
			log.Errorf("faces: subject should not be nil - bug?")
		} else if subj = entity.FirstOrCreateSubject(subj); subj == nil {
			log.Errorf("faces: failed adding subject %s for marker %d", txt.Quote(m.MarkerName), m.ID)
		} else if err := m.Updates(entity.Values{"SubjectUID": subj.SubjectUID, "SubjectSrc": entity.SrcAuto, "FaceID": ""}); err != nil {
			return affected, err
		} else if err := Db().Model(&entity.Face{}).Where("id = ? AND subject_uid = ''", faceId).Update("SubjectUID", subj.SubjectUID).Error; err != nil {
			return affected, err
		} else {
			affected++
		}
	}

	return affected, nil
}

// ResetFaceMarkerMatches removes people and face matches from face markers.
func ResetFaceMarkerMatches() error {
	v := entity.Values{"subject_uid": "", "subject_src": "", "face_id": ""}

	return Db().Model(&entity.Marker{}).Where("marker_type = ?", entity.MarkerFace).UpdateColumns(v).Error
}
