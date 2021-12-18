package search

import (
	"fmt"
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Faces searches faces and returns them.
func Faces(f form.SearchFaces) (results FaceResults, err error) {
	if err := f.ParseQueryString(); err != nil {
		return results, err
	}

	facesTable := entity.Face{}.TableName()

	// Base query.
	s := UnscopedDb().Table(facesTable)

	if f.Markers {
		s = s.Select(fmt.Sprintf(`%s.*, m.marker_uid, m.file_uid, m.marker_name, m.subj_src, m.marker_src, 
			m.marker_type, m.marker_review, m.marker_invalid, m.size, m.score, m.thumb, m.face_dist`, facesTable))

		if txt.Yes(f.Unknown) {
			s = s.Joins(`JOIN (
	        SELECT face_id, MIN(marker_uid) AS marker_uid FROM markers
	        WHERE face_id <> '' AND subj_uid = '' AND marker_name = '' AND marker_type = 'face' AND marker_src = 'image'
	          AND marker_invalid = 0 AND face_dist <= 0.64 AND size >= 80 AND score >= 15
	        GROUP BY face_id) fm
	        ON faces.id = fm.face_id`)
		} else if txt.No(f.Unknown) {
			s = s.Joins(`JOIN (
	        SELECT face_id, MIN(marker_uid) AS marker_uid FROM markers
	        WHERE face_id <> '' AND subj_uid <> '' AND marker_name <> '' AND marker_type = 'face' AND marker_src = 'image'
	          AND marker_invalid = 0 AND face_dist <= 0.64 AND size >= 80 AND score >= 15
	        GROUP BY face_id) fm
	        ON faces.id = fm.face_id`)
		} else {
			s = s.Joins(`JOIN (
	        SELECT face_id, MIN(marker_uid) AS marker_uid FROM markers
	        WHERE face_id <> '' AND marker_type = 'face' AND marker_src = 'image'
	          AND marker_invalid = 0 AND face_dist <= 0.64 AND size >= 80 AND score >= 15
	        GROUP BY face_id) fm
	        ON faces.id = fm.face_id`)
		}

		s = s.Joins("JOIN markers m ON m.marker_uid = fm.marker_uid")
	} else {
		s = s.Select(fmt.Sprintf(`%s.*`, facesTable))
	}

	// Limit result count.
	if f.Count > 0 && f.Count <= MaxResults {
		s = s.Limit(f.Count).Offset(f.Offset)
	} else {
		s = s.Limit(MaxResults).Offset(f.Offset)
	}

	// Set sort order.
	switch f.Order {
	case "subject":
		s = s.Order(fmt.Sprintf("%s.subj_uid", facesTable))
	case "added":
		s = s.Order(fmt.Sprintf("%s.created_at DESC", facesTable))
	case "samples":
		s = s.Order(fmt.Sprintf("%s.samples DESC, %s.id", facesTable, facesTable))
	default:
		s = s.Order(fmt.Sprintf("%s.samples DESC, %s.id", facesTable, facesTable))
	}

	// Find specific IDs?
	if f.UID != "" {
		s = s.Where(fmt.Sprintf("%s.id IN (?)", facesTable), strings.Split(strings.ToUpper(f.UID), txt.Or))

		if result := s.Scan(&results); result.Error != nil {
			return results, result.Error
		}

		return results, nil
	}

	// Exclude unknown faces?
	if txt.Yes(f.Unknown) {
		s = s.Where(fmt.Sprintf("%s.subj_uid = '' OR %s.subj_uid IS NULL", facesTable, facesTable))
	} else if txt.No(f.Unknown) {
		s = s.Where(fmt.Sprintf("%s.subj_uid <> '' AND %s.subj_uid IS NOT NULL", facesTable, facesTable))
	}

	// Show hidden faces?
	if !txt.Yes(f.Hidden) {
		s = s.Where(fmt.Sprintf("%s.face_hidden = 0", facesTable))
	}

	// Perform query.
	if res := s.Scan(&results); res.Error != nil {
		return results, res.Error
	}

	return results, nil
}
