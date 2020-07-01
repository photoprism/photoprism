package query

import (
	"github.com/photoprism/photoprism/internal/entity"
)

// Errors returns the error log.
func Errors(limit, offset int) (results entity.Errors, err error) {
	stmt := Db()

	err = stmt.Order("error_time DESC").Limit(limit).Offset(offset).Find(&results).Error

	return results, err
}
