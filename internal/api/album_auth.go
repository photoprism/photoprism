package api

import (
	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
)

// albumViewableBySession reports whether the session may view or download the given album. Unregistered
// visitors and shared-access-only users need a share unless they created the album; a nil session is
// never permitted.
func albumViewableBySession(s *entity.Session, album entity.Album) bool {
	if s == nil {
		return false
	}

	if s.NotRegistered() && !s.HasShare(album.AlbumUID) {
		return false
	}

	if s.GetUser().HasSharedAccessOnly(acl.ResourceAlbums) && album.CreatedBy != s.UserUID && !s.HasShare(album.AlbumUID) {
		return false
	}

	return true
}

// albumShareRequired reports whether the session lacks a share required to mutate the album, refusing
// restricted users (visitors, unregistered, shared-access-only). Unlike albumViewableBySession there is
// no self-owned exception — creating an album grants no right to modify it without a share.
func albumShareRequired(s *entity.Session, uid string) bool {
	return (s.GetUser().HasSharedAccessOnly(acl.ResourceAlbums) || s.NotRegistered()) && !s.HasShare(uid)
}
