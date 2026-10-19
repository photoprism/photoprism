package search

import (
	"errors"
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/clean"
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
	if err = frm.ParseQueryString(); errors.Is(err, form.ErrNoMatch) {
		return PhotoResults{}, nil
	} else if err != nil {
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
// because of its trailing slash, or an empty string for other albums. A folder album without a path uses
// the path of its stored filter if that is the filter the folder sync writes, quoted or not, so a filter
// that was cut off, edited, or had NUL bytes or invalid UTF-8 names no folder.
func folderAlbumPath(a entity.Album) string {
	if a.AlbumType != entity.AlbumFolder {
		return ""
	}

	p := strings.Trim(a.AlbumPath, "/")

	if f := a.AlbumFilter; p == "" && clean.SearchBindable(f) == f {
		if v := form.FilterValue(f, "path"); f == "path:"+v+" public:true" || f == "path:\""+v+"\" public:true" {
			p = strings.Trim(v, "/")
		}
	}

	if p != "" {
		return p + "/"
	}

	return ""
}
