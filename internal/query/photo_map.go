package query

import (
	"path"
	"strconv"
	"time"
)

type PhotoMap map[string]uint

// IndexedPhotos returns a map of already indexed files with their mod time unix timestamp as value.
func IndexedPhotos() (result PhotoMap, err error) {
	result = make(PhotoMap)

	type Photo struct {
		ID      uint
		TakenAt time.Time
		CellID  string
	}

	var rows []Photo

	if err := UnscopedDb().Raw("SELECT id, taken_at, cell_id FROM photos WHERE deleted_at IS NULL").Scan(&rows).Error; err != nil {
		return result, err
	}

	for _, row := range rows {
		result[path.Join(strconv.FormatInt(row.TakenAt.Unix(), 36), row.CellID)] = row.ID
	}

	return result, err
}
