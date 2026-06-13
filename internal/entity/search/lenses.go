package search

import (
	"strings"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Lenses searches lenses based on their name.
func Lenses(frm form.SearchLenses) (results []Lens, err error) {
	if err = frm.ParseQueryString(); err != nil {
		return results, err
	}

	s := Db()
	// s.LogMode(true)

	// Base query.
	s = s.Table("lenses").
		Select(`lenses.*`)

	// Filter private lenses.
	if frm.NoMake {
		s = s.Where("lenses.lens_make = ''")
	}

	// Limit result count.
	if frm.Count > 0 && frm.Count <= MaxResults {
		s = s.Limit(frm.Count).Offset(frm.Offset)
	} else {
		s = s.Limit(MaxResults).Offset(frm.Offset)
	}

	// Set sort order.
	s = s.Order(OrderExpr("lenses.lens_make ASC, lenses.lens_model ASC, lenses.lens_slug ASC", frm.Reverse))

	if frm.ID != "" {
		s = s.Where("lenses.id IN (?)", strings.Split(frm.ID, txt.Or))

		if result := s.Scan(&results); result.Error != nil {
			return results, result.Error
		}

		return results, nil
	}

	if frm.Query != "" {
		likeString := SqlParam(frm.Query, "%", "%")
		s = s.Where("lenses.lens_name LIKE ? OR lenses.lens_make LIKE ? OR lenses.lens_model LIKE ?", likeString, likeString, likeString)
	}

	if result := s.Scan(&results); result.Error != nil {
		return results, result.Error
	}

	return results, nil
}
