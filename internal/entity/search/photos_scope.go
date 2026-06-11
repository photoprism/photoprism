package search

import (
	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
)

// sessionGrantsPhotos reports whether the session is granted perm on photos. For a client session
// the permission must be granted to both the client role and (when a user is present) the user
// role, since a client acting on behalf of a user is limited by both; an ordinary session is
// evaluated against its user role. A nil session means internal or CLI use, which is not restricted.
func sessionGrantsPhotos(sess *entity.Session, perm acl.Permission) bool {
	if sess == nil {
		return true
	}

	// For client sessions the client role must permit the action as well, so a restricted client
	// cannot inherit a privileged user's access.
	if sess.IsClient() {
		if !acl.Rules.Allow(acl.ResourcePhotos, sess.GetClientRole(), perm) {
			return false
		} else if sess.NoUser() {
			return true
		}
	}

	return acl.Rules.Allow(acl.ResourcePhotos, sess.GetUserRole(), perm)
}

// sessionGrantsAnyPhotos reports whether the session is granted at least one of the perms on photos,
// using the same client and user role intersection as sessionGrantsPhotos.
func sessionGrantsAnyPhotos(sess *entity.Session, perms acl.Permissions) bool {
	for i := range perms {
		if sessionGrantsPhotos(sess, perms[i]) {
			return true
		}
	}

	return false
}

// PhotoSessionSeesEverything reports whether a session may access every picture, including
// private and archived ones, so that no row scoping is required. It performs no database
// query, so full-access users (admins, regular users, full-access API clients) incur no cost.
func PhotoSessionSeesEverything(sess *entity.Session) bool {
	return sessionGrantsPhotos(sess, acl.AccessPrivate) &&
		sessionGrantsPhotos(sess, acl.ActionDelete) &&
		sessionGrantsAnyPhotos(sess, acl.Permissions{acl.AccessAll, acl.AccessLibrary})
}

// ScopePhotosForSession restricts a photos query to the shared content a session may access:
// shared albums, owned pictures, published links, and the user's base path. Sessions with full
// library or admin access, as well as nil sessions, are returned unchanged. This builds the same
// predicate searchPhotos applies, so single-item access and search stay aligned.
func ScopePhotosForSession(stmt *gorm.DB, sess *entity.Session) *gorm.DB {
	// Library and admin access (or no session) requires no shared-scope limitation. For client
	// sessions this considers the client role too, so a restricted client cannot inherit a
	// privileged user's whole-library scope.
	if sess == nil || sessionGrantsAnyPhotos(sess, acl.Permissions{acl.AccessAll, acl.AccessLibrary}) {
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
	if PhotoSessionSeesEverything(sess) {
		return stmt
	}

	// Exclude private pictures unless the session may access them (client and user role intersection).
	if !sessionGrantsPhotos(sess, acl.AccessPrivate) {
		stmt = stmt.Where("photos.photo_private = 0")
	}

	// Exclude archived (soft-deleted) pictures unless the session may access them.
	if !sessionGrantsPhotos(sess, acl.ActionDelete) {
		stmt = stmt.Where("photos.deleted_at IS NULL")
	}

	return ScopePhotosForSession(stmt, sess)
}

// PhotoVisibleToSession reports whether the session may access the photo with the given UID,
// applying the same scope, private, and archived rules as search.
func PhotoVisibleToSession(photoUID string, sess *entity.Session) (bool, error) {
	if photoUID == "" {
		return false, nil
	} else if PhotoSessionSeesEverything(sess) {
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
	} else if PhotoSessionSeesEverything(sess) {
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
