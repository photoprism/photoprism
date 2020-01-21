package query

import (
	"fmt"
	"strings"
	"time"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/photoprism/photoprism/pkg/pluscode"
	"github.com/photoprism/photoprism/pkg/s2"
)

// GeoResult represents a photo for displaying it on a map.
type GeoResult struct {
	ID         string    `json:"ID"`
	PhotoLat   float64   `json:"Lat"`
	PhotoLng   float64   `json:"Lng"`
	PhotoUUID  string    `json:"PhotoUUID"`
	PhotoTitle string    `json:"PhotoTitle"`
	FileHash   string    `json:"FileHash"`
	FileWidth  int       `json:"FileWidth"`
	FileHeight int       `json:"FileHeight"`
	TakenAt    time.Time `json:"TakenAt"`
}

// Geo searches for photos based on a Form and returns a PhotoResult slice.
func (s *Repo) Geo(f form.GeoSearch) (results []GeoResult, err error) {
	if err := f.ParseQueryString(); err != nil {
		return results, err
	}

	defer log.Debug(capture.Time(time.Now(), fmt.Sprintf("search: %+v", f)))

	q := s.db.NewScope(nil).DB()

	q = q.Table("photos").
		Select(`photos.id, photos.photo_uuid, photos.photo_lat, photos.photo_lng, photos.photo_title, photos.taken_at, 
		files.file_hash, files.file_width, files.file_height`).
		Joins(`JOIN files ON files.photo_id = photos.id 
		AND files.file_missing = 0 AND files.file_primary AND files.deleted_at IS NULL`).
		Where("photos.deleted_at IS NULL").
		Where("photos.photo_lat <> 0").
		Group("photos.id, files.id")

	if f.Query != "" {
		q = q.Joins("LEFT JOIN photos_keywords ON photos_keywords.photo_id = photos.id").
			Joins("LEFT JOIN keywords ON photos_keywords.keyword_id = keywords.id").
			Where("keywords.keyword LIKE ?", strings.ToLower(f.Query)+"%")
	}

	if f.S2 != "" {
		s2Min, s2Max := s2.Range(f.S2, 7)
		q = q.Where("photos.location_id BETWEEN ? AND ?", s2Min, s2Max)
	} else if f.Olc != "" {
		s2Min, s2Max := s2.Range(pluscode.S2(f.Olc), 7)
		q = q.Where("photos.location_id BETWEEN ? AND ?", s2Min, s2Max)
	} else {
		// Inaccurate distance search, but probably 'good enough' for now
		if f.Lat > 0 {
			latMin := f.Lat - SearchRadius*float64(f.Dist)
			latMax := f.Lat + SearchRadius*float64(f.Dist)
			q = q.Where("photos.photo_lat BETWEEN ? AND ?", latMin, latMax)
		}

		if f.Lng > 0 {
			lngMin := f.Lng - SearchRadius*float64(f.Dist)
			lngMax := f.Lng + SearchRadius*float64(f.Dist)
			q = q.Where("photos.photo_lng BETWEEN ? AND ?", lngMin, lngMax)
		}
	}

	if !f.Before.IsZero() {
		q = q.Where("photos.taken_at <= ?", f.Before.Format("2006-01-02"))
	}

	if !f.After.IsZero() {
		q = q.Where("photos.taken_at >= ?", f.After.Format("2006-01-02"))
	}

	q = q.Order("taken_at, photos.photo_uuid")

	if result := q.Scan(&results); result.Error != nil {
		return results, result.Error
	}

	return results, nil
}
