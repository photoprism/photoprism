package query

import (
	"fmt"
	"strings"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/media"
)

// UpdateAlbumDefaultCovers updates default album cover thumbs.
func UpdateAlbumDefaultCovers() (err error) {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	start := time.Now()

	var res *gorm.DB

	condition := gorm.Expr("album_type = ? AND thumb_src = ?", entity.AlbumManual, entity.SrcAuto)

	switch DbDialect() {
	case MySQL:
		res = Db().Exec(`UPDATE albums LEFT JOIN (
    	SELECT p2.album_uid, f.file_hash FROM files f, (
        	SELECT pa.album_uid, max(p.id) AS photo_id FROM photos p
            JOIN photos_albums pa ON pa.photo_uid = p.photo_uid AND pa.hidden = 0 AND pa.missing = 0
        	WHERE p.photo_quality > 0 AND p.photo_private = 0 AND p.deleted_at IS NULL
        	GROUP BY pa.album_uid) p2 WHERE p2.photo_id = f.photo_id AND f.file_primary = 1 AND f.file_error = '' AND f.file_type IN (?)
			) b ON b.album_uid = albums.album_uid
		SET thumb = b.file_hash WHERE ?`, media.PreviewExpr, condition)
	case SQLite3:
		res = Db().Table(entity.Album{}.TableName()).
			UpdateColumn("thumb", gorm.Expr(`(
		SELECT f.file_hash FROM files f 
			JOIN photos_albums pa ON pa.album_uid = albums.album_uid AND pa.photo_uid = f.photo_uid AND pa.hidden = 0 AND pa.missing = 0
			JOIN photos p ON p.id = f.photo_id AND p.photo_private = 0 AND p.deleted_at IS NULL AND p.photo_quality > 0
			WHERE f.deleted_at IS NULL AND f.file_missing = 0 AND f.file_hash <> '' AND f.file_primary = 1 AND f.file_error = '' AND f.file_type IN (?)
			ORDER BY p.taken_at DESC LIMIT 1
		) WHERE ?`, media.PreviewExpr, condition))
	default:
		log.Warnf("sql: unsupported dialect %s", DbDialect())
		return nil
	}

	err = res.Error

	if err == nil {
		log.Debugf("covers: updated %s [%s]", english.Plural(int(res.RowsAffected), "album", "albums"), time.Since(start))
	} else if strings.Contains(err.Error(), "Error 1054") {
		log.Errorf("covers: failed updating albums, potentially incompatible database version")
		log.Errorf("%s see https://jira.mariadb.org/browse/MDEV-25362", err)
		return nil
	}

	return err
}

// UpdateAlbumFolderCovers updates folder album cover thumbs.
func UpdateAlbumFolderCovers() (err error) {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	start := time.Now()

	var res *gorm.DB

	condition := gorm.Expr("album_type = ? AND thumb_src = ?", entity.AlbumFolder, entity.SrcAuto)

	switch DbDialect() {
	case MySQL:
		res = Db().Exec(`UPDATE albums LEFT JOIN (
		SELECT p2.photo_path, f.file_hash FROM files f, (
			SELECT p.photo_path, max(p.id) AS photo_id FROM photos p
			WHERE p.photo_quality > 0 AND p.photo_private = 0 AND p.deleted_at IS NULL
			GROUP BY p.photo_path) p2 WHERE p2.photo_id = f.photo_id AND f.file_primary = 1 AND f.file_error = '' AND f.file_type IN (?)
			) b ON b.photo_path = albums.album_path
		SET thumb = b.file_hash WHERE ?`, media.PreviewExpr, condition)
	case SQLite3:
		res = Db().Table(entity.Album{}.TableName()).UpdateColumn("thumb", gorm.Expr(`(
		SELECT f.file_hash FROM files f,(
			SELECT p.photo_path, max(p.id) AS photo_id FROM photos p
			  WHERE p.photo_quality > 0 AND p.photo_private = 0 AND p.deleted_at IS NULL
			  GROUP BY p.photo_path
			) b
		WHERE f.photo_id = b.photo_id  AND f.file_primary = 1 AND f.file_error = '' AND f.file_type IN (?)
		AND b.photo_path = albums.album_path LIMIT 1)
		WHERE ?`, media.PreviewExpr, condition))
	default:
		log.Warnf("sql: unsupported dialect %s", DbDialect())
		return nil
	}

	err = res.Error

	if err == nil {
		log.Debugf("covers: updated %s [%s]", english.Plural(int(res.RowsAffected), "folder", "folders"), time.Since(start))
	} else if strings.Contains(err.Error(), "Error 1054") {
		log.Errorf("covers: failed updating folders, potentially incompatible database version")
		log.Errorf("%s see https://jira.mariadb.org/browse/MDEV-25362", err)
		return nil
	}

	return err
}

// UpdateAlbumMonthCovers updates month album cover thumbs.
func UpdateAlbumMonthCovers() (err error) {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	start := time.Now()

	var res *gorm.DB

	condition := gorm.Expr("album_type = ? AND thumb_src = ?", entity.AlbumMonth, entity.SrcAuto)

	switch DbDialect() {
	case MySQL:
		res = Db().Exec(`UPDATE albums LEFT JOIN (
		SELECT p2.photo_year, p2.photo_month, f.file_hash FROM files f, (
			SELECT p.photo_year, p.photo_month, max(p.id) AS photo_id FROM photos p
			WHERE p.photo_quality > 0 AND p.photo_private = 0 AND p.deleted_at IS NULL
			GROUP BY p.photo_year, p.photo_month) p2 WHERE p2.photo_id = f.photo_id AND f.file_primary = 1 AND f.file_error = '' AND f.file_type IN (?)
			) b ON b.photo_year = albums.album_year AND b.photo_month = albums.album_month
		SET thumb = b.file_hash WHERE ?`, media.PreviewExpr, condition)
	case SQLite3:
		res = Db().Table(entity.Album{}.TableName()).UpdateColumn("thumb", gorm.Expr(`(
		SELECT f.file_hash FROM files f,(
			SELECT p.photo_year, p.photo_month, max(p.id) AS photo_id FROM photos p
			  WHERE p.photo_quality > 0 AND p.photo_private = 0 AND p.deleted_at IS NULL
			  GROUP BY p.photo_year, p.photo_month
			) b
		WHERE f.photo_id = b.photo_id AND f.file_primary = 1 AND f.file_error = '' AND f.file_type IN (?)
		AND b.photo_year = albums.album_year AND b.photo_month = albums.album_month LIMIT 1)
		WHERE ?`, media.PreviewExpr, condition))
	default:
		log.Warnf("sql: unsupported dialect %s", DbDialect())
		return nil
	}

	err = res.Error

	if err == nil {
		log.Debugf("covers: updated %s [%s]", english.Plural(int(res.RowsAffected), "month", "months"), time.Since(start))
	} else if strings.Contains(err.Error(), "Error 1054") {
		log.Errorf("covers: failed updating calendar, potentially incompatible database version")
		log.Errorf("%s see https://jira.mariadb.org/browse/MDEV-25362", err)
		return nil
	}

	return err
}

// UpdateAlbumCovers updates album cover thumbs.
func UpdateAlbumCovers() (err error) {
	// Update Default Albums.
	if err = UpdateAlbumDefaultCovers(); err != nil {
		return err
	}

	// Update Folder Albums.
	if err = UpdateAlbumFolderCovers(); err != nil {
		return err
	}

	// Update Monthly Albums.
	if err = UpdateAlbumMonthCovers(); err != nil {
		return err
	}

	return nil
}

// UpdateLabelCovers updates label cover thumbs.
func UpdateLabelCovers() (err error) {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	start := time.Now()

	var res *gorm.DB

	condition := gorm.Expr("thumb_src = ?", entity.SrcAuto)

	switch DbDialect() {
	case MySQL:
		res = Db().Exec(`UPDATE labels LEFT JOIN (
		SELECT p2.label_id, f.file_hash FROM files f, (
			SELECT pl.label_id as label_id, max(p.id) AS photo_id FROM photos p
				JOIN photos_labels pl ON pl.photo_id = p.id AND pl.uncertainty < 100
			WHERE p.photo_quality > 0 AND p.photo_private = 0 AND p.deleted_at IS NULL
			GROUP BY pl.label_id
			UNION
			SELECT c.category_id as label_id, max(p.id) AS photo_id FROM photos p
				JOIN photos_labels pl ON pl.photo_id = p.id AND pl.uncertainty < 100
				JOIN categories c ON c.label_id = pl.label_id
			WHERE p.photo_quality > 0 AND p.photo_private = 0 AND p.deleted_at IS NULL
			GROUP BY c.category_id
			) p2 WHERE p2.photo_id = f.photo_id AND f.file_primary = 1 AND f.file_error = '' AND f.file_type IN (?) AND f.file_missing = 0
		) b ON b.label_id = labels.id
		SET thumb = b.file_hash WHERE ?`, media.PreviewExpr, condition)
	case SQLite3:
		res = Db().Table(entity.Label{}.TableName()).UpdateColumn("thumb", gorm.Expr(`(
		SELECT f.file_hash FROM files f 
			JOIN photos_labels pl ON pl.label_id = labels.id AND pl.photo_id = f.photo_id AND pl.uncertainty < 100
			JOIN photos p ON p.id = f.photo_id AND p.photo_private = 0 AND p.deleted_at IS NULL AND p.photo_quality > 0
			WHERE f.deleted_at IS NULL AND f.file_hash <> '' AND f.file_missing = 0 AND f.file_primary = 1 AND f.file_error = '' AND f.file_type IN (?)
			ORDER BY p.photo_quality DESC, pl.uncertainty ASC, p.taken_at DESC LIMIT 1
		) WHERE ?`, media.PreviewExpr, condition))

		if res.Error == nil {
			catRes := Db().Table(entity.Label{}.TableName()).UpdateColumn("thumb", gorm.Expr(`(
			SELECT f.file_hash FROM files f 
			JOIN photos_labels pl ON pl.photo_id = f.photo_id AND pl.uncertainty < 100
			JOIN categories c ON c.label_id = pl.label_id AND c.category_id = labels.id
			JOIN photos p ON p.id = f.photo_id AND p.photo_private = 0 AND p.deleted_at IS NULL AND p.photo_quality > 0
			WHERE f.deleted_at IS NULL AND f.file_hash <> '' AND f.file_missing = 0 AND f.file_primary = 1 AND f.file_error = '' AND f.file_type IN (?)
			ORDER BY p.photo_quality DESC, pl.uncertainty ASC, p.taken_at DESC LIMIT 1
			) WHERE thumb IS NULL`, media.PreviewExpr))

			res.RowsAffected += catRes.RowsAffected
		}
	default:
		log.Warnf("sql: unsupported dialect %s", DbDialect())
		return nil
	}

	err = res.Error

	if err == nil {
		log.Debugf("covers: updated %s [%s]", english.Plural(int(res.RowsAffected), "label", "labels"), time.Since(start))
	} else if strings.Contains(err.Error(), "Error 1054") {
		log.Errorf("covers: failed updating labels, potentially incompatible database version")
		log.Errorf("%s see https://jira.mariadb.org/browse/MDEV-25362", err)
		return nil
	}

	return err
}

// UpdateSubjectCovers updates subject cover thumbs.
func UpdateSubjectCovers() (err error) {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	start := time.Now()

	var res *gorm.DB

	subjTable := entity.Subject{}.TableName()
	markerTable := entity.Marker{}.TableName()

	condition := gorm.Expr(
		fmt.Sprintf("%s.subj_type = ? AND thumb_src = ?", subjTable),
		entity.SubjPerson, entity.SrcAuto)

	// TODO: Avoid using private photos as subject covers.
	// See https://github.com/photoprism/photoprism/issues/2570#issuecomment-1231690056
	switch DbDialect() {
	case MySQL:
		res = Db().Exec(`UPDATE ? LEFT JOIN (
    	SELECT m.subj_uid, m.q, MAX(m.thumb) AS marker_thumb FROM ? m
			WHERE m.subj_uid <> '' AND m.subj_uid IS NOT NULL
			  AND m.marker_invalid = 0 AND m.thumb IS NOT NULL AND m.thumb <> ''
			GROUP BY m.subj_uid, m.q
			) b ON b.subj_uid = subjects.subj_uid
		SET thumb = marker_thumb WHERE ?`, gorm.Expr(subjTable), gorm.Expr(markerTable), condition)
	case SQLite3:
		from := gorm.Expr(fmt.Sprintf("%s m WHERE m.subj_uid = %s.subj_uid ", markerTable, subjTable))
		res = Db().Table(entity.Subject{}.TableName()).UpdateColumn("thumb", gorm.Expr(`(
		SELECT m.thumb FROM ? AND m.thumb <> '' ORDER BY m.subj_src DESC, m.q DESC LIMIT 1
		) WHERE ?`, from, condition))
	default:
		log.Warnf("sql: unsupported dialect %s", DbDialect())
		return nil
	}

	err = res.Error

	if err == nil {
		log.Debugf("covers: updated %s [%s]", english.Plural(int(res.RowsAffected), "subject", "subjects"), time.Since(start))
	} else if strings.Contains(err.Error(), "Error 1054") {
		log.Errorf("covers: failed updating subjects, potentially incompatible database version")
		log.Errorf("%s see https://jira.mariadb.org/browse/MDEV-25362", err)
		return nil
	}

	return err
}

// UpdateCovers updates album, subject, and label cover thumbs.
func UpdateCovers() (err error) {
	log.Debugf("index: updating covers")

	// Update Albums.
	if err = UpdateAlbumCovers(); err != nil {
		return fmt.Errorf("%s while updating album covers", err)
	}

	// Update Labels.
	if err = UpdateLabelCovers(); err != nil {
		return fmt.Errorf("%s while updating label covers", err)
	}

	// Update Subjects.
	if err = UpdateSubjectCovers(); err != nil {
		return fmt.Errorf("%s while updating subject covers", err)
	}

	return nil
}
