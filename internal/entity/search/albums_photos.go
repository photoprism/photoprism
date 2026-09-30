package search

import (
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

// AlbumPhotos returns up to count photos from an album. When shared is set, the selection is
// bounded to what a link may expose, which SharedPhotos applies after the album's stored filter.
func AlbumPhotos(a entity.Album, count int, shared bool) (results PhotoResults, err error) {
	frm := form.SearchPhotos{
		Album:  a.AlbumUID,
		Filter: a.AlbumFilter,
		Count:  count,
		Offset: 0,
	}

	// Parse query string and filter.
	if err = frm.ParseQueryString(); err != nil {
		return results, err
	}

	if shared {
		results, _, err = SharedPhotos(frm)
	} else {
		results, _, err = Photos(frm)
	}

	return results, err
}

// folderAlbumPath returns the path filter of a folder album, which the path filter compares exactly
// because of its trailing slash, or an empty string for other albums.
func folderAlbumPath(a entity.Album) string {
	if a.AlbumType != entity.AlbumFolder {
		return ""
	} else if p := strings.Trim(a.AlbumPath, "/"); p != "" {
		return p + "/"
	}

	return ""
}
