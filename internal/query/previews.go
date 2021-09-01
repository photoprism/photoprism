package query

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/entity"
)

// UpdateAlbumDefaultPreviews updates default album preview images.
func UpdateAlbumDefaultPreviews() error {
	return Db().Table(entity.Album{}.TableName()).
		UpdateColumn("thumb", gorm.Expr(`(
			SELECT file_hash FROM files f 
			JOIN photos_albums pa ON pa.album_uid = albums.album_uid AND pa.photo_uid = f.photo_uid AND pa.hidden = 0
			JOIN photos p ON p.id = f.photo_id AND p.photo_private = 0 AND p.deleted_at IS NULL AND p.photo_quality > -1
			WHERE f.deleted_at IS NULL AND f.file_missing = 0  AND f.file_hash <> '' AND f.file_primary = 1 AND f.file_type = 'jpg' 
			ORDER BY p.taken_at DESC LIMIT 1
		) WHERE thumb_src='' AND album_type = 'album' AND deleted_at IS NULL`)).Error
}

// UpdateAlbumFolderPreviews updates folder album preview images.
func UpdateAlbumFolderPreviews() error {
	return Db().Table(entity.Album{}.TableName()).
		UpdateColumn("thumb", gorm.Expr(`(
			SELECT file_hash FROM files f 
			JOIN photos p ON p.id = f.photo_id AND p.photo_path = albums.album_path AND p.photo_private = 0 AND p.deleted_at IS NULL AND p.photo_quality > -1
			WHERE f.deleted_at IS NULL AND f.file_hash <> '' AND f.file_missing = 0 AND f.file_primary = 1 AND f.file_type = 'jpg' 
			ORDER BY p.taken_at DESC LIMIT 1
		) WHERE thumb_src = '' AND album_type = 'folder' AND deleted_at IS NULL`)).
		Error
}

// UpdateAlbumMonthPreviews updates month album preview images.
func UpdateAlbumMonthPreviews() error {
	return Db().Table(entity.Album{}.TableName()).
		UpdateColumn("thumb", gorm.Expr(`(
			SELECT file_hash FROM files f 
			JOIN photos p ON p.id = f.photo_id AND p.photo_private = 0 AND p.deleted_at IS NULL AND p.photo_quality > -1
			AND p.photo_year = albums.album_year AND p.photo_month = albums.album_month AND p.photo_month = albums.album_month 	
			WHERE f.deleted_at IS NULL AND f.file_hash <> '' AND f.file_missing = 0 AND f.file_primary = 1 AND f.file_type = 'jpg' 
			ORDER BY p.taken_at DESC LIMIT 1
		) WHERE thumb_src = '' AND album_type = 'month' AND deleted_at IS NULL`)).
		Error
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
	// Labels.
	if err = Db().Table(entity.Label{}.TableName()).
		UpdateColumn("thumb", gorm.Expr(`(
			SELECT file_hash FROM files f 
			JOIN photos_labels pl ON pl.label_id = labels.id AND pl.photo_id = f.photo_id AND pl.uncertainty < 100
			JOIN photos p ON p.id = f.photo_id AND p.photo_private = 0 AND p.deleted_at IS NULL AND p.photo_quality > -1
			WHERE f.deleted_at IS NULL AND f.file_hash <> '' AND f.file_missing = 0 AND f.file_primary = 1 AND f.file_type = 'jpg' 
			ORDER BY p.photo_quality DESC, pl.uncertainty ASC, p.taken_at DESC LIMIT 1
		) WHERE thumb_src = '' AND deleted_at IS NULL`)).
		Error; err != nil {
		return err
	}

	// Categories.
	if err = Db().Table(entity.Label{}.TableName()).
		UpdateColumn("thumb", gorm.Expr(`(
			SELECT file_hash FROM files f 
			JOIN photos_labels pl ON pl.photo_id = f.photo_id AND pl.uncertainty < 100
			JOIN categories c ON c.label_id = pl.label_id AND c.category_id = labels.id
			JOIN photos p ON p.id = f.photo_id AND p.photo_private = 0 AND p.deleted_at IS NULL AND p.photo_quality > -1
			WHERE f.deleted_at IS NULL AND f.file_hash <> '' AND f.file_missing = 0 AND f.file_primary = 1 AND f.file_type = 'jpg' 
			ORDER BY p.photo_quality DESC, pl.uncertainty ASC, p.taken_at DESC LIMIT 1
		) WHERE thumb IS NULL AND thumb_src = '' AND deleted_at IS NULL`)).
		Error; err != nil {
		return err
	}

	return nil
}

// UpdateSubjectPreviews updates subject preview images.
func UpdateSubjectPreviews() error {
	return Db().Table(entity.Subject{}.TableName()).
		UpdateColumn("thumb", gorm.Expr("(SELECT file_hash FROM files f "+
			fmt.Sprintf(
				"JOIN %s m ON f.file_uid = m.file_uid AND m.subject_uid = %s.subject_uid",
				entity.Marker{}.TableName(),
				entity.Subject{}.TableName())+
			` JOIN photos p ON f.photo_id = p.id 
			WHERE m.marker_invalid = 0 AND f.deleted_at IS NULL AND f.file_hash <> '' AND p.deleted_at IS NULL 
			AND f.file_primary = 1 AND f.file_missing = 0 AND p.photo_private = 0 AND p.photo_quality > -1 
			ORDER BY p.taken_at DESC LIMIT 1) 
			WHERE thumb_src='' AND deleted_at IS NULL AND subject_src <> 'default'`)).
		Error
}

// UpdatePreviews updates album, labels, and subject preview images.
func UpdatePreviews() (err error) {
	// Update Albums.
	if err = UpdateAlbumPreviews(); err != nil {
		return err
	}

	// Update Labels, and Categories.
	if err = UpdateLabelPreviews(); err != nil {
		return err
	}

	// Update Subjects.
	if err = UpdateSubjectPreviews(); err != nil {
		return err
	}

	return nil
}
