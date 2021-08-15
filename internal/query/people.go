package query

import (
	"github.com/photoprism/photoprism/internal/entity"
)

// People returns people from the index.
func People(limit, offset int) (result entity.People, err error) {
	stmt := Db()

	stmt = stmt.Order("id").Limit(limit).Offset(offset)
	err = stmt.Find(&result).Error

	return result, err
}
