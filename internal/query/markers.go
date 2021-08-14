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
		db = db.Where("embeddings <> ''")
	}

	if unmatched {
		db = db.Where("ref_uid = ''")
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
		Where("embeddings <> ''").
		Order("id")

	if err := stmt.Pluck("embeddings", &col).Error; err != nil {
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

func MatchMarkersWithPeople() (affected int, err error) {
	var markers entity.Markers

	if err := Db().
		Where("face_id <> '' AND ref_uid = '' AND ref_src = ''").
		Where("marker_invalid = 0 AND marker_type = ?", entity.MarkerFace).
		Where("marker_label <> ''").
		Order("marker_label").
		Find(&markers).Error; err != nil {
		return affected, err
	} else if len(markers) == 0 {
		return affected, nil
	}

	for _, m := range markers {
		if p := entity.NewPerson(m.MarkerLabel, entity.SrcMarker, 1); p == nil {
			log.Errorf("faces: person should not be nil - bug?")
		} else if p = entity.FirstOrCreatePerson(p); p == nil {
			log.Errorf("faces: failed adding person %s for marker %d", txt.Quote(m.MarkerLabel), m.ID)
		} else if err := m.Updates(entity.Val{"RefUID": p.PersonUID, "RefSrc": entity.SrcPeople, "FaceID": ""}); err != nil {
			return affected, err
		} else if err := Db().Model(&entity.Face{}).Where("id = ? AND person_uid = ''", m.FaceID).Update("PersonUID", p.PersonUID).Error; err != nil {
			return affected, err
		} else {
			affected++
		}
	}

	return affected, nil
}
