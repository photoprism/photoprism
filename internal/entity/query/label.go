package query

import (
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/media"
)

// PhotoLabel returns a photo label entity if exists.
func PhotoLabel(photoID, labelID uint) (*entity.PhotoLabel, error) {
	result := &entity.PhotoLabel{}

	if err := Db().Where("photo_id = ? AND label_id = ?", photoID, labelID).Preload("Photo").Preload("Label").First(result).Error; err != nil {
		return result, err
	}

	return result, nil
}

// LabelBySlug returns a Label based on the slug name.
func LabelBySlug(labelSlug string) (*entity.Label, error) {
	result := &entity.Label{}

	if err := Db().Where("label_slug = ? OR custom_slug = ?", labelSlug, labelSlug).First(result).Error; err != nil {
		return result, err
	}

	return result, nil
}

// LabelByUID returns a Label based on the label UID.
func LabelByUID(labelUID string) (*entity.Label, error) {
	result := &entity.Label{}

	if err := Db().Where("label_uid = ?", labelUID).First(result).Error; err != nil {
		return result, err
	}

	return result, nil
}

// LabelThumbBySlug returns a label cover file based on the slug name.
func LabelThumbBySlug(labelSlug string) (*entity.File, error) {
	result := &entity.File{}

	if err := Db().Where("files.file_primary AND files.file_type IN (?) AND files.deleted_at IS NULL", media.PreviewExpr).
		Joins("JOIN labels ON labels.label_slug = ?", labelSlug).
		Joins("JOIN photos_labels ON photos_labels.label_id = labels.id AND photos_labels.photo_id = files.photo_id AND photos_labels.uncertainty < 100").
		Joins("JOIN photos ON photos.id = files.photo_id AND photos.photo_private = 0 AND photos.deleted_at IS NULL").
		Order("photos.photo_quality DESC, photos_labels.uncertainty ASC").
		First(result).Error; err != nil {
		return result, err
	}

	return result, nil
}

// LabelThumbByUID returns a label cover file based on the label UID.
func LabelThumbByUID(labelUID string) (*entity.File, error) {
	result := &entity.File{}

	// Search matching label
	err := Db().Where("files.file_primary AND files.deleted_at IS NULL").
		Joins("JOIN labels ON labels.label_uid = ?", labelUID).
		Joins("JOIN photos_labels ON photos_labels.label_id = labels.id AND photos_labels.photo_id = files.photo_id AND photos_labels.uncertainty < 100").
		Joins("JOIN photos ON photos.id = files.photo_id AND photos.photo_private = 0 AND photos.deleted_at IS NULL").
		Order("photos.photo_quality DESC, photos_labels.uncertainty ASC").
		First(result).Error

	if err == nil {
		return result, nil
	}

	// If failed, search for category instead
	err = Db().Where("files.file_primary AND files.deleted_at IS NULL").
		Joins("JOIN photos_labels ON photos_labels.photo_id = files.photo_id AND photos_labels.uncertainty < 100").
		Joins("JOIN categories c ON photos_labels.label_id = c.label_id").
		Joins("JOIN labels ON c.category_id = labels.id AND labels.label_uid= ?", labelUID).
		Joins("JOIN photos ON photos.id = files.photo_id AND photos.photo_private = 0 AND photos.deleted_at IS NULL").
		Order("photos.photo_quality DESC, photos_labels.uncertainty ASC").
		First(result).Error

	return result, err
}
