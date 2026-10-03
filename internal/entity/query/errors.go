package query

import (
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
)

// Errors returns the error log filtered with an optional search string.
func Errors(limit, offset int, search string) (results entity.Errors, err error) {
	stmt := Db()

	search = strings.TrimSpace(search)

	switch {
	case search == "error" || search == "errors":
		stmt = stmt.Where("error_level = 'error'")
	case search == "warning" || search == "warnings":
		stmt = stmt.Where("error_level = 'warning'")
	case len(search) >= 3:
		stmt = stmt.Where(LikeCond("error_message"), "%"+clean.SqlLike(search)+"%")
	}

	err = stmt.Order("error_time DESC").Limit(limit).Offset(offset).Find(&results).Error

	return results, err
}

// DeleteErrors removes all entries from the errors table.
func DeleteErrors() (err error) {
	return UnscopedDb().Delete(entity.Error{}).Error
}
