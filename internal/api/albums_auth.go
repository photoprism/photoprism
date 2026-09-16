package api

import (
	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
)

// albumViewableBySession reports whether the session may view or download the given album. Sessions
// without whole-library reach on albums need a share unless they created the album; a nil session is
// never permitted. The row policy lives on the entity, so the picture that names its albums answers
// the same question about the same session.
func albumViewableBySession(s *entity.Session, album entity.Album) bool {
	return album.VisibleToSession(s)
}

// albumShareRequired reports whether the session lacks a share required to mutate the album, refusing
// restricted users (visitors, unregistered, shared-access-only). Unlike albumViewableBySession there is
// no self-owned exception — creating an album grants no right to modify it without a share.
func albumShareRequired(s *entity.Session, uid string) bool {
	return (s.HasSharedAccessOnly(acl.ResourceAlbums) || s.NotRegistered()) && !s.HasShare(uid)
}
