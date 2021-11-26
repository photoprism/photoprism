package search

import (
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

// AlbumPhotos returns up to count photos from an album.
func AlbumPhotos(a entity.Album, count int) (results PhotoResults, err error) {
	results, _, err = Photos(form.SearchPhotos{
		Album:  a.AlbumUID,
		Filter: a.AlbumFilter,
		Count:  count,
		Offset: 0,
	})

	return results, err
}
