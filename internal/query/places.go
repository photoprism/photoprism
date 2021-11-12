package query

import (
	"github.com/photoprism/photoprism/internal/entity"
)

// CellIDs returns all known S2 cell ids as string slice.
func CellIDs() (ids []string, err error) {
	tableName := entity.Cell{}.TableName()

	var count int64

	if err = UnscopedDb().Table(tableName).Where("id <> 'zz'").Count(&count).Error; err != nil {
		return []string{}, err
	}

	ids = make([]string, 0, count)

	err = UnscopedDb().Table(tableName).Select("id").Where("id <> 'zz'").Pluck("id", &ids).Error

	return ids, err
}

// PlaceIDs returns all known S2 place ids as string slice.
func PlaceIDs() (ids []string, err error) {
	tableName := entity.Place{}.TableName()

	var count int64

	if err = UnscopedDb().Table(tableName).Where("id <> 'zz'").Count(&count).Error; err != nil {
		return []string{}, err
	}

	ids = make([]string, 0, count)

	err = UnscopedDb().Table(tableName).Select("id").Where("id <> 'zz'").Pluck("id", &ids).Error

	return ids, err
}
