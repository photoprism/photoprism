package search

import (
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Cameras searches cameras based on their name.
func Cameras(frm form.SearchCameras) (results []Camera, err error) {
	if err = frm.ParseQueryString(); err != nil {
		return results, err
	}

	results = make([]Camera, 0)
	s := Db()

	// Base query.
	s = s.Table("cameras").
		Select(`cameras.*`)

	// Filter cameras without make.
	if frm.NoMake {
		s = s.Where("cameras.camera_make = ''")
	}

	// Limit result count.
	if frm.Count > 0 && frm.Count <= MaxResults {
		s = s.Limit(frm.Count).Offset(frm.Offset)
	} else {
		s = s.Limit(MaxResults).Offset(frm.Offset)
	}

	// Set sort order.
	s = s.Order(OrderExpr("cameras.camera_make ASC, cameras.camera_model ASC, cameras.camera_slug ASC", frm.Reverse))

	if frm.ID != "" {
		s = s.Where("cameras.id IN (?)", strings.Split(frm.ID, txt.Or))

		if result := s.Scan(&results); result.Error != nil {
			return results, result.Error
		}

		return results, nil
	}

	if frm.Query != "" {
		likeString := SqlParam(frm.Query, "%", "%")
		switch entity.DbDialect() {
		case dsn.DialectPostgreSQL:
			where, values := OrLikeCols([]string{"lower(cameras.camera_name)", "lower(cameras.camera_make)", "lower(cameras.camera_model)"}, strings.ToLower(likeString))
			s = s.Where(where, values...)
		default:
			where, values := OrLikeCols([]string{"cameras.camera_name", "cameras.camera_make", "cameras.camera_model"}, likeString)
			s = s.Where(where, values...)
		}
	}

	if result := s.Scan(&results); result.Error != nil {
		return results, result.Error
	}

	return results, nil
}
