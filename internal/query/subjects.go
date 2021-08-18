package query

import (
	"github.com/photoprism/photoprism/internal/entity"
)

// Subjects returns subjects from the index.
func Subjects(limit, offset int) (result entity.Subjects, err error) {
	stmt := Db()

	stmt = stmt.Order("subject_slug").Limit(limit).Offset(offset)
	err = stmt.Find(&result).Error

	return result, err
}
