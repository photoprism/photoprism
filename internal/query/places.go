package query

import (
	"github.com/photoprism/photoprism/internal/entity"
)

type Cell struct {
	ID      string
	PlaceID string
}

type Cells []Cell

// CellIDs returns all known S2 cell ids as Cell slice.
func CellIDs() (result Cells, err error) {
	tableName := entity.Cell{}.TableName()

	var count int64

	if err = UnscopedDb().Table(tableName).Where("id <> 'zz'").Count(&count).Error; err != nil {
		return result, err
	}

	result = make(Cells, 0, count)

	err = UnscopedDb().Table(tableName).Select("id, place_id").Where("id <> 'zz'").Order("id").Scan(&result).Error

	return result, err
}

// PurgePlaces removes unused entries from the places table.
func PurgePlaces() error {
	query := `DELETE FROM places 
       WHERE id NOT IN (SELECT DISTINCT place_id FROM cells)
       AND id NOT IN (SELECT DISTINCT place_id FROM photos)`

	return Db().Exec(query).Error
}
