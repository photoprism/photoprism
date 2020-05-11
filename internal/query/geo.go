package query

import (
	"fmt"
	"strings"
	"time"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/photoprism/photoprism/pkg/pluscode"
	"github.com/photoprism/photoprism/pkg/s2"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Geo searches for photos based on a Form and returns GeoResults ([]GeoResult).
func Geo(f form.GeoSearch) (results GeoResults, err error) {
	if err := f.ParseQueryString(); err != nil {
		return results, err
	}

	defer log.Debug(capture.Time(time.Now(), fmt.Sprintf("search: %+v", f)))

	s := UnscopedDb()

	s = s.Table("photos").
		Select(`photos.id, photos.photo_uuid, photos.photo_lat, photos.photo_lng, photos.photo_title, 
		photos.photo_favorite, photos.taken_at, files.file_hash, files.file_width, files.file_height`).
		Joins(`JOIN files ON files.photo_id = photos.id 
		AND files.file_missing = 0 AND files.file_primary AND files.deleted_at IS NULL`).
		Where("photos.deleted_at IS NULL").
		Where("photos.photo_lat <> 0").
		Group("photos.id, files.id")

	f.Query = txt.Clip(f.Query, txt.ClipKeyword)

	if f.Query != "" {
		s = s.Joins("LEFT JOIN photos_keywords ON photos_keywords.photo_id = photos.id").
			Joins("LEFT JOIN keywords ON photos_keywords.keyword_id = keywords.id").
			Where("keywords.keyword LIKE ?", strings.ToLower(f.Query)+"%")
	}

	if f.Review {
		s = s.Where("photos.photo_quality < 3")
	} else if f.Quality != 0 {
		s = s.Where("photos.photo_quality >= ?", f.Quality)
	}

	if f.Favorite {
		s = s.Where("photos.photo_favorite = 1")
	}

	if f.S2 != "" {
		s2Min, s2Max := s2.Range(f.S2, 7)
		s = s.Where("photos.location_id BETWEEN ? AND ?", s2Min, s2Max)
	} else if f.Olc != "" {
		s2Min, s2Max := s2.Range(pluscode.S2(f.Olc), 7)
		s = s.Where("photos.location_id BETWEEN ? AND ?", s2Min, s2Max)
	} else {
		// Inaccurate distance search, but probably 'good enough' for now
		if f.Lat > 0 {
			latMin := f.Lat - SearchRadius*float32(f.Dist)
			latMax := f.Lat + SearchRadius*float32(f.Dist)
			s = s.Where("photos.photo_lat BETWEEN ? AND ?", latMin, latMax)
		}

		if f.Lng > 0 {
			lngMin := f.Lng - SearchRadius*float32(f.Dist)
			lngMax := f.Lng + SearchRadius*float32(f.Dist)
			s = s.Where("photos.photo_lng BETWEEN ? AND ?", lngMin, lngMax)
		}
	}

	if !f.Before.IsZero() {
		s = s.Where("photos.taken_at <= ?", f.Before.Format("2006-01-02"))
	}

	if !f.After.IsZero() {
		s = s.Where("photos.taken_at >= ?", f.After.Format("2006-01-02"))
	}

	s = s.Order("taken_at, photos.photo_uuid")

	if result := s.Scan(&results); result.Error != nil {
		return results, result.Error
	}

	return results, nil
}
