package query

import (
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
)

// Errors returns the error log filtered with an optional search string.
func Errors(limit, offset int, s string) (results entity.Errors, err error) {
	stmt := Db()

	s = strings.TrimSpace(s)

	if len(s) >= 3 {
		stmt = stmt.Where("error_message LIKE ?", "%"+s+"%")
	}

	err = stmt.Order("error_time DESC").Limit(limit).Offset(offset).Find(&results).Error

	return results, err
}
