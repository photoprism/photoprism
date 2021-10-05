package search

import (
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Faces searches faces and returns them.
func Faces(f form.FaceSearch) (results FaceResults, err error) {
	if err := f.ParseQueryString(); err != nil {
		return results, err
	}

	// Base query.
	s := UnscopedDb().Table(entity.Face{}.TableName())

	// Limit result count.
	if f.Count > 0 && f.Count <= MaxResults {
		s = s.Limit(f.Count).Offset(f.Offset)
	} else {
		s = s.Limit(MaxResults).Offset(f.Offset)
	}

	// Set sort order.
	switch f.Order {
	case "subject":
		s = s.Order("subj_uid")
	case "added":
		s = s.Order(fmt.Sprintf("%s.created_at DESC", entity.Face{}.TableName()))
	case "samples":
		s = s.Order("samples DESC")
	default:
		s = s.Order("samples DESC")
	}

	// Make sure at least one marker exists.
	if f.Markers || txt.Yes(f.Unknown) {
		s = s.Where("id IN (SELECT face_id FROM ? WHERE "+
			"face_id IS NOT NULL AND face_id <> '' AND marker_type = ? AND  marker_src = ? AND marker_invalid = 0)",
			gorm.Expr(entity.Marker{}.TableName()), entity.MarkerFace, entity.SrcImage)
	}

	// Adds markers to search results if requested.
	addMarkers := func(results FaceResults) FaceResults {
		r := make(FaceResults, 0, len(results))

		// Add markers to results.
		for i := range results {
			if marker := entity.FindFaceMarker(results[i].ID); marker != nil {
				m := results[i]
				m.Marker = marker
				r = append(r, m)
			}
		}

		return r
	}

	// Find specific IDs?
	if f.ID != "" {
		s = s.Where(fmt.Sprintf("%s.id IN (?)", entity.Face{}.TableName()), strings.Split(strings.ToUpper(f.ID), txt.Or))

		if result := s.Scan(&results); result.Error != nil {
			return results, result.Error
		} else if f.Markers {
			return addMarkers(results), nil
		}

		return results, nil
	}

	// Exclude unknown faces?
	if txt.Yes(f.Unknown) {
		s = s.Where("subj_uid = '' OR subj_uid IS NULL")
	} else if txt.No(f.Unknown) {
		s = s.Where("subj_uid <> '' AND subj_uid IS NOT NULL")
	}

	// Exclude hidden faces?
	if f.Hidden == "" || txt.No(f.Hidden) {
		s = s.Where("face_hidden = 0 OR face_hidden IS NULL")
	} else if txt.Yes(f.Hidden) {
		s = s.Where("face_hidden = 1")
	}

	// Perform query.
	if res := s.Scan(&results); res.Error != nil {
		return results, res.Error
	} else if f.Markers {
		return addMarkers(results), nil
	}

	return results, nil
}
