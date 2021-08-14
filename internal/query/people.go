package query

import (
	"github.com/photoprism/photoprism/internal/entity"
)

// People returns people (with face embeddings) from the index.
func People(limit, offset int, withEmbeddings bool) (result entity.People, err error) {
	stmt := Db()

	if withEmbeddings {
		stmt = stmt.Where("embeddings <> ''")
	}

	stmt = stmt.Order("id").Limit(limit).Offset(offset)
	err = stmt.Find(&result).Error

	return result, err
}
