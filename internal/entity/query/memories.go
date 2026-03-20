package query

import (
	"time"
)

// MemoryYear represents a year that has photos taken on the same month and day.
type MemoryYear struct {
	Year       int `json:"Year"`
	Month      int `json:"Month"`
	Day        int `json:"Day"`
	PhotoCount int `json:"PhotoCount"`
}

// MemoriesOnThisDay returns years that have photos taken on today's month and day,
// ordered from most recent to oldest, so the frontend can render "X years ago today" cards.
func MemoriesOnThisDay(public bool) (results []MemoryYear, err error) {
	now := time.Now()
	month := int(now.Month())
	day := now.Day()
	currentYear := now.Year()

	stmt := UnscopedDb().Table("photos").
		Select("photos.photo_year AS year, photos.photo_month AS month, photos.photo_day AS day, COUNT(*) AS photo_count").
		Where("photos.deleted_at IS NULL AND photos.photo_month = ? AND photos.photo_day = ? AND photos.photo_year > 0 AND photos.photo_year < ?", month, day, currentYear)

	// Ignore private pictures?
	if public {
		stmt = stmt.Where("photo_private = 0")
	}

	stmt = stmt.Group("photos.photo_year").
		Order("photos.photo_year DESC")

	if err = stmt.Scan(&results).Error; err != nil {
		return results, err
	}

	return results, nil
}
