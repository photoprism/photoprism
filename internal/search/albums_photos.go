package search

import (
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

// AlbumPhotos returns up to count photos from an album.
func AlbumPhotos(a entity.Album, count int, shared bool) (results PhotoResults, err error) {
	frm := form.SearchPhotos{
		Album:  a.AlbumUID,
		Filter: a.AlbumFilter,
		Count:  count,
		Offset: 0,
	}

	if shared {
		frm.Public = true
		frm.Private = false
		frm.Hidden = false
		frm.Archived = false
		frm.Review = false
	}

	results, _, err = Photos(frm)

	return results, err
}
