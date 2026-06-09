package search

import (
	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
)

// photoSessionSeesEverything reports whether a session may access every picture, including
// private and archived ones, so that no row scoping is required. It performs no database
// query, so full-access users (admins, regular users, full-access API clients) incur no cost.
func photoSessionSeesEverything(sess *entity.Session) bool {
	// No session means internal or CLI use, which is not restricted.
	if sess == nil {
		return true
	}

	// seesAll reports whether a role may access private and archived content across the whole library.
	seesAll := func(role acl.Role) bool {
		return acl.Rules.Allow(acl.ResourcePhotos, role, acl.AccessPrivate) &&
			acl.Rules.Allow(acl.ResourcePhotos, role, acl.ActionDelete) &&
			acl.Rules.AllowAny(acl.ResourcePhotos, role, acl.Permissions{acl.AccessAll, acl.AccessLibrary})
	}

	// For client sessions, the client role must permit full access as well; a client acting
	// on behalf of a restricted user is still limited by that user's role.
	if sess.IsClient() {
		if !seesAll(sess.GetClientRole()) {
			return false
		} else if sess.NoUser() {
			return true
		}
	}

	return seesAll(sess.GetUserRole())
}

// ScopePhotosForSession restricts a photos query to the shared content a session may access:
// shared albums, owned pictures, published links, and the user's base path. Sessions with full
// library or admin access, as well as nil sessions, are returned unchanged. This builds the same
// predicate searchPhotos applies, so single-item access and search stay aligned.
func ScopePhotosForSession(stmt *gorm.DB, sess *entity.Session) *gorm.DB {
	// Library and admin access (or no session) requires no shared-scope limitation.
	if sess == nil || acl.Rules.AllowAny(acl.ResourcePhotos, sess.GetUserRole(), acl.Permissions{acl.AccessAll, acl.AccessLibrary}) {
		return stmt
	}

	user := sess.GetUser()
	sharedAlbums := "photos.photo_uid IN (SELECT photo_uid FROM photos_albums WHERE hidden = 0 AND missing = 0 AND album_uid IN (?)) OR "

	if sess.IsVisitor() || sess.NotRegistered() {
		return stmt.Where(sharedAlbums+"photos.published_at > ?", sess.SharedUIDs(), entity.Now())
	} else if basePath := user.GetBasePath(); basePath == "" {
		return stmt.Where(sharedAlbums+"photos.created_by = ? OR photos.published_at > ?", sess.SharedUIDs(), user.UserUID, entity.Now())
	} else {
		return stmt.Where(sharedAlbums+"photos.created_by = ? OR photos.published_at > ? OR photos.photo_path = ? OR photos.photo_path LIKE ?",
			sess.SharedUIDs(), user.UserUID, entity.Now(), basePath, basePath+"/%")
	}
}

// ScopeVisiblePhotos restricts a photos or files query to the pictures a session may view: it
// excludes private and archived pictures for roles lacking those permissions and then applies
// ScopePhotosForSession. Full-access sessions are returned unchanged. Callers must ensure the
// photos table is part of the query so the photos.* conditions resolve.
func ScopeVisiblePhotos(stmt *gorm.DB, sess *entity.Session) *gorm.DB {
	if photoSessionSeesEverything(sess) {
		return stmt
	}

	aclRole := sess.GetUserRole()

	// Exclude private pictures unless the role may access them.
	if acl.Rules.Deny(acl.ResourcePhotos, aclRole, acl.AccessPrivate) {
		stmt = stmt.Where("photos.photo_private = 0")
	}

	// Exclude archived (soft-deleted) pictures unless the role may access them.
	if acl.Rules.Deny(acl.ResourcePhotos, aclRole, acl.ActionDelete) {
		stmt = stmt.Where("photos.deleted_at IS NULL")
	}

	return ScopePhotosForSession(stmt, sess)
}

// PhotoVisibleToSession reports whether the session may access the photo with the given UID,
// applying the same scope, private, and archived rules as search.
func PhotoVisibleToSession(photoUID string, sess *entity.Session) (bool, error) {
	if photoUID == "" {
		return false, nil
	} else if photoSessionSeesEverything(sess) {
		return true, nil
	}

	// UnscopedDb is deliberate: the archived (soft-deleted) decision is deferred to
	// ScopeVisiblePhotos, which adds "photos.deleted_at IS NULL" for roles without the archive
	// permission. Full-access roles short-circuited above and may resolve archived pictures by UID,
	// matching GetPhoto / searchPhotos.
	stmt := ScopeVisiblePhotos(UnscopedDb().Table("photos").Where("photos.photo_uid = ?", photoUID), sess)

	var count int
	if err := stmt.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// FileVisibleToSession reports whether the session may access the file with the given hash,
// based on the visibility of the photo it belongs to.
func FileVisibleToSession(fileHash string, sess *entity.Session) (bool, error) {
	if fileHash == "" {
		return false, nil
	} else if photoSessionSeesEverything(sess) {
		return true, nil
	}

	// A soft-deleted file is never accessible, so exclude it explicitly (UnscopedDb does not apply
	// the soft-delete scope). ScopeVisiblePhotos handles the photo's archived/private/scope state.
	stmt := ScopeVisiblePhotos(
		UnscopedDb().Table("files").
			Joins("JOIN photos ON photos.id = files.photo_id").
			Where("files.file_hash = ? AND files.deleted_at IS NULL", fileHash),
		sess,
	)

	var count int
	if err := stmt.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}
