package query

import (
	"errors"
	"fmt"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/dsn"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/form"
)

// SelectedPhotoUIDsForSession returns the subset of the given photo UIDs that the session is
// allowed to access, applying the same shared-scope, private, and archived rules as photo search,
// so restricted callers can only act on pictures within their shared scope. Full library or admin
// sessions (client and user role intersection) see every picture, so the input is returned without
// a database query; it is therefore safe and cheap to call for any session.
func SelectedPhotoUIDsForSession(photoUIDs []string, sess *entity.Session) (scoped []string, err error) {
	if len(photoUIDs) == 0 || search.PhotoSessionSeesEverything(sess) {
		return photoUIDs, nil
	}

	stmt := search.ScopeVisiblePhotos(
		UnscopedDb().Table("photos").Where("photos.photo_uid IN (?)", photoUIDs),
		sess,
	)

	if err = stmt.Pluck("photos.photo_uid", &scoped).Error; err != nil {
		return nil, err
	}

	return scoped, nil
}

// subfolderCond returns the join condition that matches the subfolders b of the folders a.
func subfolderCond(dialect string) (string, error) {
	switch dialect {
	case dsn.DriverMySQL:
		return fmt.Sprintf("b.path LIKE CONCAT(%s, '/%%') ESCAPE '%s' AND SUBSTR(b.path, 1, LENGTH(a.path) + 1) = CONCAT(a.path, '/')",
			clean.SqlLikeExpr("a.path"), clean.SqlLikeEscape), nil
	case dsn.DriverSQLite3:
		return fmt.Sprintf("b.path LIKE %s || '/%%' ESCAPE '%s' AND SUBSTR(b.path, 1, LENGTH(a.path) + 1) = a.path || '/'",
			clean.SqlLikeExpr("a.path"), clean.SqlLikeEscape), nil
	default:
		return "", fmt.Errorf("unknown sql dialect: %s", dialect)
	}
}

// SelectedPhotos finds photos based on the given selection form, e.g. for adding them to an album,
// without applying a session's scope. Handlers serving a request use SelectedPhotosForSession instead.
func SelectedPhotos(frm form.Selection) (results entity.Photos, err error) {
	return selectedPhotos(frm, nil)
}

// SelectedPhotosForSession finds the photos of a selection form that the session may access, applying
// its scope to the photos reached through every selection field.
func SelectedPhotosForSession(frm form.Selection, sess *entity.Session) (results entity.Photos, err error) {
	return selectedPhotos(frm, sess)
}

// selectedPhotos finds photos based on the given selection form, optionally limited to the content
// the session may access when sess is not nil.
func selectedPhotos(frm form.Selection, sess *entity.Session) (results entity.Photos, err error) {
	if frm.Empty() {
		return results, errors.New("no items selected")
	}

	// Resolve photos in smart albums.
	if photoIds, err := AlbumsPhotoUIDs(frm.Albums, false, false); err != nil {
		log.Warnf("query: %s", err.Error())
	} else if len(photoIds) > 0 {
		frm.Photos = append(frm.Photos, photoIds...)
	}

	subfolders, err := subfolderCond(DbDialect())

	if err != nil {
		return results, err
	}

	where := fmt.Sprintf(`photos.photo_uid IN (?) 
		OR photos.place_id IN (?) 
		OR photos.photo_uid IN (SELECT photo_uid FROM files WHERE file_uid IN (?))
		OR photos.photo_path IN (
			SELECT a.path FROM folders a WHERE a.folder_uid IN (?) UNION
			SELECT b.path FROM folders a JOIN folders b ON %s WHERE a.folder_uid IN (?))
		OR photos.photo_uid IN (SELECT photo_uid FROM photos_albums WHERE hidden = 0 AND album_uid IN (?))
		OR photos.id IN (SELECT f.photo_id FROM files f JOIN %s m ON f.file_uid = m.file_uid WHERE f.deleted_at IS NULL AND m.subj_uid IN (?))
		OR photos.id IN (SELECT pl.photo_id FROM photos_labels pl JOIN labels l ON pl.label_id = l.id AND pl.uncertainty < 100 AND l.deleted_at IS NULL WHERE l.label_uid IN (?))
		OR photos.id IN (SELECT pl.photo_id FROM photos_labels pl JOIN categories c ON c.label_id = pl.label_id AND pl.uncertainty < 100 JOIN labels lc ON lc.id = c.category_id AND lc.deleted_at IS NULL WHERE lc.label_uid IN (?))`,
		subfolders, entity.Marker{}.TableName())

	s := UnscopedDb().Table("photos").
		Select("photos.*").
		Where(where, frm.Photos, frm.Places, frm.Files, frm.Files, frm.Files, frm.Albums, frm.Subjects, frm.Labels, frm.Labels)

	// Limit the selection to the session's shared scope (no-op for full-access sessions).
	if sess != nil {
		s = search.ScopeVisibleSelection(s, sess, frm.Photos)
	}

	if result := s.Scan(&results); result.Error != nil {
		return results, result.Error
	}

	return results, nil
}
