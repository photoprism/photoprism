package mcp

import (
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search"
)

// coverUIDs maps cover thumbs, which are file hashes, optionally followed by a crop area, to the UIDs of their
// pictures, leaving out the pictures the session may not see.
func coverUIDs(thumbs []string, s *entity.Session) map[string]string {
	var unique []string

	for _, thumb := range thumbs {
		if h, _, _ := strings.Cut(thumb, "-"); h != "" {
			unique = appendUnique(unique, h)
		}
	}

	covers := make(map[string]string, len(thumbs))

	if len(unique) == 0 {
		return covers
	}

	var rows []struct {
		Hash     string
		PhotoUID string
	}

	stmt := entity.UnscopedDb().Table("files").Select("files.file_hash AS hash, photos.photo_uid").
		Joins("JOIN photos ON photos.id = files.photo_id").
		Where("files.file_hash IN (?) AND files.deleted_at IS NULL", unique)

	entity.Log("mcp", "find covers", search.ScopeVisiblePhotos(stmt, s).Scan(&rows).Error)

	byHash := make(map[string]string, len(rows))

	for _, r := range rows {
		byHash[r.Hash] = r.PhotoUID
	}

	for _, thumb := range thumbs {
		if h, _, _ := strings.Cut(thumb, "-"); byHash[h] != "" {
			covers[thumb] = byHash[h]
		}
	}

	return covers
}
