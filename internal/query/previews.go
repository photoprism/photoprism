package query

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/entity"
)

// UpdateAlbumDefaultPreviews updates default album preview images.
func UpdateAlbumDefaultPreviews() (err error) {
	start := time.Now()

	err = Db().Table(entity.Album{}.TableName()).
		UpdateColumn("thumb", gorm.Expr(`(
			SELECT f.file_hash FROM files f 
			JOIN photos_albums pa ON pa.album_uid = albums.album_uid AND pa.photo_uid = f.photo_uid AND pa.hidden = 0
			JOIN photos p ON p.id = f.photo_id AND p.photo_private = 0 AND p.deleted_at IS NULL AND p.photo_quality > -1
			WHERE f.deleted_at IS NULL AND f.file_missing = 0  AND f.file_hash <> '' AND f.file_primary = 1 AND f.file_type = 'jpg' 
			ORDER BY p.taken_at DESC LIMIT 1
		) WHERE thumb_src='' AND album_type = 'album' AND deleted_at IS NULL`)).Error

	log.Debugf("previews: updated albums [%s]", time.Since(start))

	return err
}

// UpdateAlbumFolderPreviews updates folder album preview images.
func UpdateAlbumFolderPreviews() (err error) {
	start := time.Now()

	err = Db().Table(entity.Album{}.TableName()).
		UpdateColumn("thumb", gorm.Expr(`(
			SELECT f.file_hash FROM files f 
			JOIN photos p ON p.id = f.photo_id AND p.photo_path = albums.album_path AND p.photo_private = 0 AND p.deleted_at IS NULL AND p.photo_quality > -1
			WHERE f.deleted_at IS NULL AND f.file_hash <> '' AND f.file_missing = 0 AND f.file_primary = 1 AND f.file_type = 'jpg' 
			ORDER BY p.taken_at DESC LIMIT 1
		) WHERE thumb_src = '' AND album_type = 'folder' AND deleted_at IS NULL`)).
		Error

	log.Debugf("previews: updated folders [%s]", time.Since(start))

	return err
}

// UpdateAlbumMonthPreviews updates month album preview images.
func UpdateAlbumMonthPreviews() (err error) {
	start := time.Now()

	err = Db().Table(entity.Album{}.TableName()).
		Where("album_type = ?", entity.AlbumMonth).
		Where("thumb IS NOT NULL AND thumb_src = ?", entity.SrcAuto).
		UpdateColumns(entity.Values{"thumb": nil}).Error

	/* TODO: Slow with many photos due to missing index.

	switch DbDialect() {
	case MySQL:
		err = Db().Table(entity.Album{}.TableName()).
			UpdateColumn("thumb", gorm.Expr(`(
			SELECT f.file_hash FROM files f JOIN photos p ON p.id = f.photo_id
			WHERE YEAR(p.taken_at) = albums.album_year AND MONTH(p.taken_at) = albums.album_month
			AND p.photo_private = 0 AND p.deleted_at IS NULL AND p.photo_quality > -1 AND f.deleted_at IS NULL
			AND f.file_hash <> '' AND f.file_missing = 0 AND f.file_primary = 1 AND f.file_type = 'jpg'
			ORDER BY p.taken_at DESC LIMIT 1
		) WHERE thumb IS NULL AND thumb_src = '' AND album_type = 'month' AND deleted_at IS NULL`)).
			Error
	case SQLite:
		err = Db().Table(entity.Album{}.TableName()).
			UpdateColumn("thumb", gorm.Expr(`(
			SELECT f.file_hash FROM files f JOIN photos p ON p.id = f.photo_id
			WHERE strftime('%Y%m', p.taken_at) = (albums.album_year || printf('%02d', albums.album_month))
			AND p.photo_private = 0 AND p.deleted_at IS NULL AND p.photo_quality > -1 AND f.deleted_at IS NULL
			AND f.file_hash <> '' AND f.file_missing = 0 AND f.file_primary = 1 AND f.file_type = 'jpg'
			ORDER BY p.taken_at DESC LIMIT 1
		) WHERE thumb IS NULL AND thumb_src = '' AND album_type = 'month' AND deleted_at IS NULL`)).
			Error
	default:
		return nil
	}
	*/
	log.Debugf("previews: updated calendar [%s]", time.Since(start))

	return err
}

// UpdateAlbumPreviews updates album preview images.
func UpdateAlbumPreviews() (err error) {
	// Update Default Albums.
	if err = UpdateAlbumDefaultPreviews(); err != nil {
		return err
	}

	// Update Folder Albums.
	if err = UpdateAlbumFolderPreviews(); err != nil {
		return err
	}

	// Update Monthly Albums.
	if err = UpdateAlbumMonthPreviews(); err != nil {
		return err
	}

	return nil
}

// UpdateLabelPreviews updates label preview images.
func UpdateLabelPreviews() (err error) {
	start := time.Now()

	// Labels.
	if err = Db().Table(entity.Label{}.TableName()).
		UpdateColumn("thumb", gorm.Expr(`(
			SELECT f.file_hash FROM files f 
			JOIN photos_labels pl ON pl.label_id = labels.id AND pl.photo_id = f.photo_id AND pl.uncertainty < 100
			JOIN photos p ON p.id = f.photo_id AND p.photo_private = 0 AND p.deleted_at IS NULL AND p.photo_quality > -1
			WHERE f.deleted_at IS NULL AND f.file_hash <> '' AND f.file_missing = 0 AND f.file_primary = 1 AND f.file_type = 'jpg' 
			ORDER BY p.photo_quality DESC, pl.uncertainty ASC, p.taken_at DESC LIMIT 1
		) WHERE thumb_src = '' AND deleted_at IS NULL`)).
		Error; err != nil {
		return err
	}

	log.Debugf("previews: updated labels [%s]", time.Since(start))

	return nil
}

// UpdateCategoryPreviews updates category preview images.
func UpdateCategoryPreviews() (err error) {
	start := time.Now()

	// Categories.
	if err = Db().Table(entity.Label{}.TableName()).
		UpdateColumn("thumb", gorm.Expr(`(
			SELECT f.file_hash FROM files f 
			JOIN photos_labels pl ON pl.photo_id = f.photo_id AND pl.uncertainty < 100
			JOIN categories c ON c.label_id = pl.label_id AND c.category_id = labels.id
			JOIN photos p ON p.id = f.photo_id AND p.photo_private = 0 AND p.deleted_at IS NULL AND p.photo_quality > -1
			WHERE f.deleted_at IS NULL AND f.file_hash <> '' AND f.file_missing = 0 AND f.file_primary = 1 AND f.file_type = 'jpg' 
			ORDER BY p.photo_quality DESC, pl.uncertainty ASC, p.taken_at DESC LIMIT 1
		) WHERE thumb IS NULL AND thumb_src = '' AND deleted_at IS NULL`)).
		Error; err != nil {
		return err
	}

	log.Debugf("previews: updated categories [%s]", time.Since(start))

	return nil
}

// UpdateSubjectPreviews updates subject preview images.
func UpdateSubjectPreviews() (err error) {
	start := time.Now()

	/* Previous implementation for reference:

	return Db().Table(entity.Subject{}.TableName()).
	UpdateColumn("thumb", gorm.Expr("(SELECT f.file_hash FROM files f "+
		fmt.Sprintf(
			"JOIN %s m ON f.file_uid = m.file_uid AND m.subj_uid = %s.subj_uid",
			entity.Marker{}.TableName(),
			entity.Subject{}.TableName())+
		` JOIN photos p ON f.photo_id = p.id
		WHERE m.marker_invalid = 0 AND f.deleted_at IS NULL AND f.file_hash <> '' AND p.deleted_at IS NULL
		AND f.file_primary = 1 AND f.file_missing = 0 AND p.photo_private = 0 AND p.photo_quality > -1
		ORDER BY p.taken_at DESC LIMIT 1)
		WHERE thumb_src='' AND deleted_at IS NULL`)).
	Error */

	err = Db().Table(entity.Subject{}.TableName()).
		UpdateColumn("marker_uid", gorm.Expr("(SELECT m.marker_uid FROM "+
			fmt.Sprintf(
				"%s m WHERE m.subj_uid = %s.subj_uid AND m.subj_src = 'manual' ",
				entity.Marker{}.TableName(),
				entity.Subject{}.TableName())+
			` AND m.file_hash <> '' ORDER BY m.q DESC LIMIT 1) 
			WHERE marker_src = '' AND deleted_at IS NULL`)).
		Error

	/** err = Db().Table(entity.Subject{}.TableName()).
	UpdateColumn("thumb", gorm.Expr("(SELECT m.file_hash FROM "+
		fmt.Sprintf(
			"%s m WHERE m.subj_uid = %s.subj_uid AND m.subj_src = 'manual' ",
			entity.Marker{}.TableName(),
			entity.Subject{}.TableName())+
		` AND m.file_hash <> '' ORDER BY m.w DESC LIMIT 1)
		WHERE thumb_src = '' AND deleted_at IS NULL`)).
	Error

	*/

	log.Debugf("previews: updated subjects [%s]", time.Since(start))

	return err
}

// UpdatePreviews updates album, labels, and subject preview images.
func UpdatePreviews() (err error) {
	// Update Albums.
	if err = UpdateAlbumPreviews(); err != nil {
		return err
	}

	// Update Labels.
	if err = UpdateLabelPreviews(); err != nil {
		return err
	}

	// Update Categories.
	if err = UpdateCategoryPreviews(); err != nil {
		return err
	}

	// Update Subjects.
	if err = UpdateSubjectPreviews(); err != nil {
		return err
	}

	return nil
}
