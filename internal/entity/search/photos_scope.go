package search

import (
	"strings"

	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/clean"
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

// PhotoSessionSeesPrivate reports whether the session may view private pictures (the AccessPrivate
// permission on photos, evaluated with the same client/user role intersection as the scope predicates).
// A nil session (no identified requester, e.g. a coarse download token) is treated as not permitted.
func PhotoSessionSeesPrivate(sess *entity.Session) bool {
	if sess == nil {
		return false
	}

	return sessionGrantsPhotos(sess, acl.AccessPrivate)
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
	return scopePhotosForSession(stmt, sess, nil)
}

// scopePhotosForSession builds the shared-scope predicate for ScopePhotosForSession, additionally
// allowing the photo UIDs in allowUIDs regardless of the predicate. The allow-list carries pictures
// shared only through filter-based smart albums (folder, moment, calendar, region), whose membership
// the photos_albums predicate cannot resolve; callers pass those UIDs via sharedSmartAlbumPhotoUIDs.
func scopePhotosForSession(stmt *gorm.DB, sess *entity.Session, allowUIDs []string) *gorm.DB {
	// Library and admin access (or no session) requires no shared-scope limitation. For client
	// sessions this considers the client role too, so a restricted client cannot inherit a
	// privileged user's whole-library scope.
	if sess == nil || sessionGrantsAnyPhotos(sess, acl.Permissions{acl.AccessAll, acl.AccessLibrary}) {
		return stmt
	}

	user := sess.GetUser()

	// Pictures shared through a regular album resolve via their photos_albums rows.
	conds := []string{"photos.photo_uid IN (SELECT photo_uid FROM photos_albums WHERE hidden = FALSE AND missing = FALSE AND album_uid IN (?))"}
	args := []interface{}{sess.SharedUIDs()}

	// Explicitly-allowed photo UIDs (e.g. members of shared smart albums resolved by their filter).
	if len(allowUIDs) > 0 {
		conds = append(conds, "photos.photo_uid IN (?)")
		args = append(args, allowUIDs)
	}

	if sess.IsVisitor() || sess.NotRegistered() {
		conds = append(conds, "photos.published_at > ?")
		args = append(args, entity.Now())
	} else if basePath := user.GetBasePath(); basePath == "" {
		conds = append(conds, "photos.created_by = ?", "photos.published_at > ?")
		args = append(args, user.UserUID, entity.Now())
	} else {
		conds = append(conds, "photos.created_by = ?", "photos.published_at > ?", "photos.photo_path = ?", "photos.photo_path LIKE ?")
		args = append(args, user.UserUID, entity.Now(), basePath, basePath+"/%")
	}

	return stmt.Where(strings.Join(conds, " OR "), args...)
}

// excludeRestrictedPhotos removes private and archived (soft-deleted) pictures from a photos or files
// query unless the session may access them (client and user role intersection).
func excludeRestrictedPhotos(stmt *gorm.DB, sess *entity.Session) *gorm.DB {
	if !sessionGrantsPhotos(sess, acl.AccessPrivate) {
		stmt = stmt.Where("photos.photo_private = FALSE")
	}

	if !sessionGrantsPhotos(sess, acl.ActionDelete) {
		stmt = stmt.Where("photos.deleted_at IS NULL")
	}

	return stmt
}

// ScopeVisiblePhotos restricts a photos or files query to the pictures a session may view: it
// excludes private and archived pictures for roles lacking those permissions and then applies
// ScopePhotosForSession. Full-access sessions are returned unchanged. Callers must ensure the
// photos table is part of the query so the photos.* conditions resolve.
func ScopeVisiblePhotos(stmt *gorm.DB, sess *entity.Session) *gorm.DB {
	if PhotoSessionSeesEverything(sess) {
		return stmt
	}

	return ScopePhotosForSession(excludeRestrictedPhotos(stmt, sess), sess)
}

// ScopeVisibleSelection restricts a photos or files query to what a session may view for a bulk
// selection (e.g. a multi-file download), like ScopeVisiblePhotos but additionally allowing the
// selected photo UIDs shared through filter-based smart albums (folder, moment, calendar, region).
// Those albums have no photos_albums rows, so ScopePhotosForSession alone would drop them and the
// download would return no files; this keeps multi-select downloads aligned with single-item access
// (FileVisibleToSession) and album browsing. Full-access sessions are returned unchanged.
func ScopeVisibleSelection(stmt *gorm.DB, sess *entity.Session, photoUIDs []string) *gorm.DB {
	if PhotoSessionSeesEverything(sess) {
		return stmt
	}

	return scopePhotosForSession(excludeRestrictedPhotos(stmt, sess), sess, sharedSmartAlbumPhotoUIDs(photoUIDs, sess))
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

	var count int64
	if err := stmt.Count(&count).Error; err != nil {
		return false, err
	} else if count > 0 {
		return true, nil
	}

	// Fall back to smart albums shared with the session (folder, moment, calendar, region), whose
	// members derive from a filter rather than photos_albums rows and are thus invisible to
	// ScopePhotosForSession above.
	return sharedSmartAlbumContains(photoUID, sess)
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

	var count int64
	if err := stmt.Count(&count).Error; err != nil {
		return false, err
	} else if count > 0 {
		return true, nil
	}

	// Fall back to smart albums shared with the session; the file hash resolves to its photo, so a
	// picture shared only through a folder, moment, calendar, or region link stays downloadable.
	var f entity.File
	if err := UnscopedDb().Table("files").Select("photo_uid").
		Where("file_hash = ? AND deleted_at IS NULL", fileHash).
		Limit(1).Scan(&f).Error; err != nil {
		return false, err
	}

	return sharedSmartAlbumContains(f.PhotoUID, sess)
}

// FileVisibleToPublic reports whether the file's photo may be downloaded with only a coarse,
// unidentified download token (public mode, a configured static token, or the instance default): the
// photo must be public (not private), not archived, and not hidden. Unlike FileVisibleToSession it
// applies no shared-scope predicate, so a static token keeps its broad — but never private — access.
// The public predicate is duplicated verbatim in PhotoVisibleToPublic; keep the two in sync.
func FileVisibleToPublic(fileHash string) (bool, error) {
	if fileHash == "" {
		return false, nil
	}

	var count int64
	err := UnscopedDb().Table("files").
		Joins("JOIN photos ON photos.id = files.photo_id").
		Where("files.file_hash = ? AND files.deleted_at IS NULL", fileHash).
		Where("photos.photo_private = 0 AND photos.deleted_at IS NULL AND photos.photo_quality > -1").
		Count(&count).Error

	return count > 0, err
}

// PhotoVisibleToPublic reports whether the photo may be downloaded with only a coarse, unidentified
// download token: it must be public (not private), not archived, and not hidden. The public predicate
// is duplicated verbatim in FileVisibleToPublic; keep the two in sync.
func PhotoVisibleToPublic(photoUID string) (bool, error) {
	if photoUID == "" {
		return false, nil
	}

	var count int64
	err := UnscopedDb().Table("photos").
		Where("photo_uid = ? AND photo_private = 0 AND deleted_at IS NULL AND photo_quality > -1", photoUID).
		Count(&count).Error

	return count > 0, err
}

// FileDownloadable reports whether the file with the given hash may be downloaded: an identified session
// is limited to what it may see (FileVisibleToSession), a nil session (coarse download token) to public
// content (FileVisibleToPublic). Centralizes the "coarse token sees only public" rule for all endpoints.
func FileDownloadable(fileHash string, sess *entity.Session) (bool, error) {
	if sess != nil {
		return FileVisibleToSession(fileHash, sess)
	}

	return FileVisibleToPublic(fileHash)
}

// PhotoDownloadable reports whether the photo with the given UID may be downloaded, picking the
// session-scoped or public visibility path like FileDownloadable.
func PhotoDownloadable(photoUID string, sess *entity.Session) (bool, error) {
	if sess != nil {
		return PhotoVisibleToSession(photoUID, sess)
	}

	return PhotoVisibleToPublic(photoUID)
}

// sharedSmartAlbumMembers returns the subset of the given photo UIDs belonging to a smart album (folder,
// moment, calendar, region) shared with the session. Such albums derive members from a filter, not
// photos_albums rows, so ScopePhotosForSession misses them; this evaluates each shared album's filter as
// the album view does. Regular albums (empty filter) are skipped, and it stops once every requested UID
// is matched so the single-UID caller short-circuits.
func sharedSmartAlbumMembers(photoUIDs []string, sess *entity.Session) []string {
	if len(photoUIDs) == 0 || sess == nil {
		return nil
	}

	seen := make(map[string]struct{})
	var visible []string

	for _, albumUID := range sess.SharedUIDs() {
		if len(visible) == len(photoUIDs) {
			break // every requested UID already matched
		}

		album, err := entity.CachedAlbumByUID(albumUID)

		// Regular albums (empty filter) are already covered by the photos_albums predicate; skip them.
		if err != nil || album.AlbumUID == "" || album.AlbumFilter == "" {
			continue
		}

		// Match membership as the album view does. Primary:true makes Count bound photos, not file rows —
		// without it a multi-file picture (RAW+JPEG, video, sidecars) exhausts the limit and later selected
		// photos silently drop from a multi-file download.
		results, _, searchErr := UserPhotos(form.SearchPhotos{Scope: albumUID, UID: strings.Join(photoUIDs, "|"), Count: len(photoUIDs), Primary: true}, sess)

		if searchErr != nil {
			// A single unusable share (e.g. an invalid filter) must not mask the others.
			log.Debugf("scope: %s while checking shared album %s", clean.Error(searchErr), clean.Log(albumUID))
			continue
		}

		for _, r := range results {
			if _, ok := seen[r.PhotoUID]; ok {
				continue
			}

			seen[r.PhotoUID] = struct{}{}
			visible = append(visible, r.PhotoUID)
		}
	}

	return visible
}

// sharedSmartAlbumContains reports whether the given photo UID belongs to a smart album shared with the
// session — the single-item case of sharedSmartAlbumMembers, which short-circuits on the first match.
func sharedSmartAlbumContains(photoUID string, sess *entity.Session) (bool, error) {
	if photoUID == "" {
		return false, nil
	}

	return len(sharedSmartAlbumMembers([]string{photoUID}, sess)) > 0, nil
}

// sharedSmartAlbumPhotoUIDs returns the subset of the given photo UIDs that belong to a smart album
// shared with the session — the bulk case of sharedSmartAlbumMembers used by the download/selection path.
func sharedSmartAlbumPhotoUIDs(photoUIDs []string, sess *entity.Session) []string {
	return sharedSmartAlbumMembers(photoUIDs, sess)
}
