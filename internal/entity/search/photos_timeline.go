package search

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

const (
	PhotoTimelineBucketMonth = "month"
	PhotoTimelineBucketDay   = "day"
	PhotoTimelineOrderDesc   = "desc"
)

// PhotoTimeline contains local-date buckets for photo search results.
type PhotoTimeline struct {
	Bucket           string                `json:"bucket"`
	Order            string                `json:"order"`
	PhotoCount       int                   `json:"photoCount"`
	UnknownDateCount int                   `json:"unknownDateCount"`
	Buckets          []PhotoTimelineBucket `json:"buckets"`
}

// PhotoTimelineBucket contains a single local-date photo count bucket.
type PhotoTimelineBucket struct {
	Key        string `json:"key"`
	Label      string `json:"label"`
	Year       int    `json:"year"`
	Month      int    `json:"month"`
	Day        *int   `json:"day,omitempty"`
	From       string `json:"from"`
	Until      string `json:"until"`
	PhotoCount int    `json:"photoCount"`
}

type photoTimelineBucketRow struct {
	Year       int `gorm:"column:year"`
	Month      int `gorm:"column:month"`
	Day        int `gorm:"column:day"`
	PhotoCount int `gorm:"column:photo_count"`
}

type photoTimelineCountRow struct {
	PhotoCount int `gorm:"column:photo_count"`
}

// UserPhotoTimeline returns local-date timeline buckets for matching photos.
func UserPhotoTimeline(frm form.SearchPhotos, sess *entity.Session, bucket string) (timeline PhotoTimeline, err error) {
	if bucket == "" {
		bucket = PhotoTimelineBucketMonth
	}

	if bucket != PhotoTimelineBucketMonth && bucket != PhotoTimelineBucketDay {
		return timeline, ErrBadRequest
	}

	frm.Count = 0
	frm.Offset = 0
	frm.Reverse = false
	frm.Merged = false

	query, err := newSearchPhotosQueryWithOptions(frm, sess, searchPhotosQueryOptions{
		applyOrder:        false,
		applyOrderFilters: true,
		applyPagination:   false,
		applyLabelGroup:   false,
		allowUIDFast:      false,
	})

	if err != nil {
		return timeline, err
	} else if query.db == nil {
		return PhotoTimeline{Bucket: bucket, Order: PhotoTimelineOrderDesc, Buckets: []PhotoTimelineBucket{}}, nil
	}

	total, err := photoTimelineCount(query.db, "")

	if err != nil {
		return timeline, err
	}

	validWhere := "photos.photo_year > 0 AND photos.photo_month > 0"

	if bucket == PhotoTimelineBucketDay {
		validWhere += " AND photos.photo_day > 0"
	}

	buckets, err := photoTimelineBuckets(query.db, bucket, validWhere)

	if err != nil {
		return timeline, err
	}

	valid := 0

	for i := range buckets {
		valid += buckets[i].PhotoCount
	}

	return PhotoTimeline{
		Bucket:           bucket,
		Order:            PhotoTimelineOrderDesc,
		PhotoCount:       total,
		UnknownDateCount: total - valid,
		Buckets:          buckets,
	}, nil
}

// photoTimelineCount returns the number of distinct photos matching the query.
func photoTimelineCount(s *gorm.DB, where string) (count int, err error) {
	row := photoTimelineCountRow{}
	q := s.Select("COUNT(DISTINCT photos.id) AS photo_count")

	if where != "" {
		q = q.Where(where)
	}

	if err = q.Scan(&row).Error; err != nil {
		return 0, err
	}

	return row.PhotoCount, nil
}

// photoTimelineBuckets returns grouped local-date photo buckets.
func photoTimelineBuckets(s *gorm.DB, bucket, validWhere string) (buckets []PhotoTimelineBucket, err error) {
	rows := make([]photoTimelineBucketRow, 0)
	selectCols := "photos.photo_year AS year, photos.photo_month AS month, COUNT(DISTINCT photos.id) AS photo_count"
	groupCols := "photos.photo_year, photos.photo_month"
	orderCols := "photos.photo_year DESC, photos.photo_month DESC"

	if bucket == PhotoTimelineBucketDay {
		selectCols = "photos.photo_year AS year, photos.photo_month AS month, photos.photo_day AS day, COUNT(DISTINCT photos.id) AS photo_count"
		groupCols = "photos.photo_year, photos.photo_month, photos.photo_day"
		orderCols = "photos.photo_year DESC, photos.photo_month DESC, photos.photo_day DESC"
	}

	if err = s.Select(selectCols).Where(validWhere).Group(groupCols).Order(orderCols).Scan(&rows).Error; err != nil {
		return buckets, err
	}

	buckets = make([]PhotoTimelineBucket, 0, len(rows))

	for _, row := range rows {
		buckets = append(buckets, newPhotoTimelineBucket(row, bucket))
	}

	return buckets, nil
}

// newPhotoTimelineBucket formats a local-date bucket row for API responses.
func newPhotoTimelineBucket(row photoTimelineBucketRow, bucket string) PhotoTimelineBucket {
	if bucket == PhotoTimelineBucketDay {
		day := row.Day
		date := time.Date(row.Year, time.Month(row.Month), row.Day, 0, 0, 0, 0, time.UTC)
		next := date.AddDate(0, 0, 1)

		return PhotoTimelineBucket{
			Key:        date.Format("2006-01-02"),
			Label:      date.Format("Jan 2, 2006"),
			Year:       row.Year,
			Month:      row.Month,
			Day:        &day,
			From:       date.Format("2006-01-02"),
			Until:      next.Format("2006-01-02"),
			PhotoCount: row.PhotoCount,
		}
	}

	date := time.Date(row.Year, time.Month(row.Month), 1, 0, 0, 0, 0, time.UTC)
	next := date.AddDate(0, 1, 0)

	return PhotoTimelineBucket{
		Key:        date.Format("2006-01"),
		Label:      fmt.Sprintf("%s %d", date.Month(), row.Year),
		Year:       row.Year,
		Month:      row.Month,
		Day:        nil,
		From:       date.Format("2006-01-02"),
		Until:      next.Format("2006-01-02"),
		PhotoCount: row.PhotoCount,
	}
}
