package entity

import (
	"github.com/photoprism/photoprism/internal/auth/acl"
)

// SharedWithSession reports whether the session reaches this album without whole-library access to
// albums: through a share, or because its own account created it.
func (m *Album) SharedWithSession(sess *Session) bool {
	switch {
	case m == nil || sess == nil:
		return false
	case sess.HasShare(m.AlbumUID):
		return true
	}

	// An album with no owner is owned by nobody, whatever a session without an account carries.
	return sess.IsRegistered() && m.CreatedBy != "" && m.CreatedBy == sess.UserUID
}

// VisibleToSession reports whether the session may read this album as a record, from its role and
// the row alone. Callers MUST already be scope-checked on albums, or be authorizing an action on
// another resource that includes the album: an album download is admitted on pictures, which a
// scope arm here would refuse.
func (m *Album) VisibleToSession(sess *Session) bool {
	if m == nil || sess == nil {
		return false
	} else if sess.GrantsAny(acl.ResourceAlbums, acl.Permissions{acl.AccessAll, acl.AccessLibrary}) {
		return true
	}

	return m.SharedWithSession(sess)
}

// SharedAlbums returns the albums of the given list the session reaches through a share or its own
// ownership, in a new slice, or nil when it reaches none.
func SharedAlbums(albums []Album, sess *Session) []Album {
	var kept []Album

	for i := range albums {
		if albums[i].SharedWithSession(sess) {
			kept = append(kept, albums[i])
		}
	}

	return kept
}
