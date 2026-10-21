package search

import (
	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

// ApplyPhotoFeatures applies the feature settings to a photo search form, as the photos endpoint does: the public
// filter is ignored while the Private feature is disabled, and while the Review feature is enabled, sessions that
// may not manage pictures find only reviewed pictures unless the search is scoped to an album.
func ApplyPhotoFeatures(frm *form.SearchPhotos, sess *entity.Session, private, review bool) {
	if frm == nil {
		return
	}

	if !private {
		frm.Public = false
	}

	if frm.Scope == "" && review && sess.Denies(acl.ResourcePhotos, acl.ActionManage) {
		frm.Quality = 3
	}
}

// ApplyAlbumFeatures applies the feature settings to an album search form: the public filter is ignored while
// the Private feature is disabled.
func ApplyAlbumFeatures(frm *form.SearchAlbums, private bool) {
	if frm != nil && !private {
		frm.Public = false
	}
}
