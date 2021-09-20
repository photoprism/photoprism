package search

import (
	"fmt"
	"strings"

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
	default:
		s = s.Order("samples DESC")
	}

	// Find specific IDs?
	if f.ID != "" {
		s = s.Where(fmt.Sprintf("%s.id IN (?)", entity.Face{}.TableName()), strings.Split(strings.ToUpper(f.ID), txt.Or))

		if result := s.Scan(&results); result.Error != nil {
			return results, result.Error
		} else if f.Markers {
			// Add markers to results.
			for i := range results {
				results[i].Marker = entity.FindFaceMarker(results[i].ID)
			}
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
		// Add markers to results.
		for i := range results {
			results[i].Marker = entity.FindFaceMarker(results[i].ID)
		}
	}

	return results, nil
}
