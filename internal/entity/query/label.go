package query

import (
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/media"
)

// PhotoLabel returns a photo label entity if exists.
func PhotoLabel(photoID, labelID uint) (label entity.PhotoLabel, err error) {
	if err := Db().Where("photo_id = ? AND label_id = ?", photoID, labelID).Preload("Photo").Preload("Label").First(&label).Error; err != nil {
		return label, err
	}

	return label, nil
}

// LabelBySlug returns a Label based on the slug name.
func LabelBySlug(labelSlug string) (label entity.Label, err error) {
	if err := Db().Where("label_slug = ? OR custom_slug = ?", labelSlug, labelSlug).First(&label).Error; err != nil {
		return label, err
	}

	return label, nil
}

// LabelByUID returns a Label based on the label UID.
func LabelByUID(labelUID string) (label entity.Label, err error) {
	if err := Db().Where("label_uid = ?", labelUID).First(&label).Error; err != nil {
		return label, err
	}

	return label, nil
}

// LabelThumbBySlug returns a label cover file based on the slug name.
func LabelThumbBySlug(labelSlug string) (file entity.File, err error) {
	if err := Db().Where("files.file_primary AND files.file_type IN (?) AND files.deleted_at IS NULL", media.PreviewExpr).
		Joins("JOIN labels ON labels.label_slug = ?", labelSlug).
		Joins("JOIN photos_labels ON photos_labels.label_id = labels.id AND photos_labels.photo_id = files.photo_id AND photos_labels.uncertainty < 100").
		Joins("JOIN photos ON photos.id = files.photo_id AND photos.photo_private = 0 AND photos.deleted_at IS NULL").
		Order("photos.photo_quality DESC, photos_labels.uncertainty ASC").
		First(&file).Error; err != nil {
		return file, err
	}

	return file, nil
}

// LabelThumbByUID returns a label cover file based on the label UID.
func LabelThumbByUID(labelUID string) (file entity.File, err error) {
	// Search matching label
	err = Db().Where("files.file_primary AND files.deleted_at IS NULL").
		Joins("JOIN labels ON labels.label_uid = ?", labelUID).
		Joins("JOIN photos_labels ON photos_labels.label_id = labels.id AND photos_labels.photo_id = files.photo_id AND photos_labels.uncertainty < 100").
		Joins("JOIN photos ON photos.id = files.photo_id AND photos.photo_private = 0 AND photos.deleted_at IS NULL").
		Order("photos.photo_quality DESC, photos_labels.uncertainty ASC").
		First(&file).Error

	if err == nil {
		return file, nil
	}

	// If failed, search for category instead
	err = Db().Where("files.file_primary AND files.deleted_at IS NULL").
		Joins("JOIN photos_labels ON photos_labels.photo_id = files.photo_id AND photos_labels.uncertainty < 100").
		Joins("JOIN categories c ON photos_labels.label_id = c.label_id").
		Joins("JOIN labels ON c.category_id = labels.id AND labels.label_uid= ?", labelUID).
		Joins("JOIN photos ON photos.id = files.photo_id AND photos.photo_private = 0 AND photos.deleted_at IS NULL").
		Order("photos.photo_quality DESC, photos_labels.uncertainty ASC").
		First(&file).Error

	return file, err
}
