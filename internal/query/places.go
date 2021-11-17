package query

import (
	"github.com/jinzhu/gorm"

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

// UpdatePlaceIDs finds and replaces invalid place references.
func UpdatePlaceIDs() (fixed int64, err error) {
	photosTable := entity.Photo{}.TableName()
	placesTable := entity.Place{}.TableName()

	res := Db().Table(photosTable).Where("place_id NOT IN (SELECT place_id FROM ?)", gorm.Expr(placesTable)).
		UpdateColumn("place_id", "zz")

	if res.Error != nil {
		return res.RowsAffected, res.Error
	}

	res = Db().Table(photosTable).Where("cell_id IS NOT NULL AND cell_id <> 'zz'").
		UpdateColumn("place_id", gorm.Expr("(SELECT id FROM cells WHERE id = cell_id)"))

	return res.RowsAffected, res.Error
}
