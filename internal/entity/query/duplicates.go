package query

import (
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
)

// Duplicates finds duplicate files in the range of limit and offset sorted by file name.
func Duplicates(limit, offset int, dir string) (files entity.Duplicates, err error) {
	dir = strings.TrimPrefix(dir, "/")

	stmt := Db()

	if dir != "" {
		stmt = stmt.Where(clean.SqlPrefixCond("file_name"), clean.SqlPrefixArgs(dir+"/")...)
	}

	err = stmt.Order("file_name").Limit(limit).Offset(offset).Find(&files).Error

	return files, err
}
