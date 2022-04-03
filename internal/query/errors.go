package query

import (
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
)

// Errors returns the error log filtered with an optional search string.
func Errors(limit, offset int, search string) (results entity.Errors, err error) {
	stmt := Db()

	search = strings.TrimSpace(search)

	if search == "error" || search == "errors" {
		stmt = stmt.Where("error_level = 'error'")
	} else if search == "warning" || search == "warnings" {
		stmt = stmt.Where("error_level = 'warning'")
	} else if len(search) >= 3 {
		stmt = stmt.Where("error_message LIKE ?", "%"+search+"%")
	}

	err = stmt.Order("error_time DESC").Limit(limit).Offset(offset).Find(&results).Error

	return results, err
}

// DeleteErrors removes all entries from the errors table.
func DeleteErrors() (err error) {
	return UnscopedDb().Delete(entity.Error{}).Error
}
