package query

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/entity"
)

// Subjects returns subjects from the index.
func Subjects(limit, offset int) (result entity.Subjects, err error) {
	stmt := Db()

	stmt = stmt.Order("subject_slug").Limit(limit).Offset(offset)
	err = stmt.Find(&result).Error

	return result, err
}

// ResetSubjects removes all unused subjects from the index.
func ResetSubjects() error {
	return UnscopedDb().
		Where("subject_src = ?", entity.SrcMarker).
		Where(fmt.Sprintf("subject_uid NOT IN (SELECT subject_uid FROM %s)", entity.Face{}.TableName())).
		Delete(&entity.Subject{}).
		Error
}
