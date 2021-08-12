package query

import (
	"github.com/photoprism/photoprism/internal/entity"
)

// People finds a list of people.
func People(limit, offset int, embeddings bool) (result entity.People, err error) {
	stmt := Db()

	if embeddings {
		stmt = stmt.Where("embeddings <> ''")
	}

	stmt = stmt.Order("id").Limit(limit).Offset(offset)
	err = stmt.Find(&result).Error

	return result, err
}

// PeopleFaces finds a list of faces.
func PeopleFaces() (result entity.PeopleFaces, err error) {
	stmt := Db().
		Order("id")

	err = stmt.Find(&result).Error

	return result, err
}

// PurgeUnknownFaces removes unknown faces from the index.
func PurgeUnknownFaces() error {
	return UnscopedDb().Delete(
		entity.PersonFace{},
		"person_uid = '' AND updated_at < ?", entity.Yesterday()).Error
}
