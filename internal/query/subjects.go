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

// RemoveDanglingMarkerSubjects permanently deletes dangling marker subjects from the index.
func RemoveDanglingMarkerSubjects() (removed int64, err error) {
	res := UnscopedDb().
		Where("subject_src = ?", entity.SrcMarker).
		Where(fmt.Sprintf("subject_uid NOT IN (SELECT subject_uid FROM %s)", entity.Face{}.TableName())).
		Where(fmt.Sprintf("subject_uid NOT IN (SELECT subject_uid FROM %s)", entity.Marker{}.TableName())).
		Delete(&entity.Subject{})

	return res.RowsAffected, res.Error
}
