package query

import (
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
)

// Duplicates finds duplicate files in the range of limit and offset sorted by file name.
func Duplicates(limit, offset int, pathName string) (files entity.Duplicates, err error) {
	if strings.HasPrefix(pathName, "/") {
		pathName = pathName[1:]
	}

	stmt := Db()

	if pathName != "" {
		stmt = stmt.Where("file_name LIKE ?", pathName+"/%")
	}

	err = stmt.Order("file_name").Limit(limit).Offset(offset).Find(&files).Error

	return files, err
}

// GetDuplicateByName finds a single duplicate file by file name.
func GetDuplicateByName(pathName string) (file entity.Duplicate, err error) {
	if strings.HasPrefix(pathName, "/") {
		pathName = pathName[1:]
	}

	stmt := Db()

	if pathName != "" {
		stmt = stmt.Where("file_name = ?", pathName)
	}

	err = stmt.First(&file).Error

	return file, err
}

// GetDuplicateByHash finds a single duplicate file by fileHash.
func GetDuplicatesByHash(fileHash string) (files entity.Duplicates, err error) {
	if err := Db().Where("file_hash = ?", fileHash).Find(&files).Error; err != nil {
		return files, err
	}

	return files, err
}
