package query

import (
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
)

// Duplicates finds duplicate files in the range of limit and offset sorted by file name.
func Duplicates(limit, offset int, dir string) (files entity.Duplicates, err error) {
	if strings.HasPrefix(dir, "/") {
		dir = dir[1:]
	}

	stmt := Db()

	if dir != "" {
		stmt = stmt.Where("file_name LIKE ?", dir+"/%")
	}

	err = stmt.Order("file_name").Limit(limit).Offset(offset).Find(&files).Error

	return files, err
}
