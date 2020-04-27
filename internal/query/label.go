package query

import (
	"fmt"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/capture"
)

// LabelResult contains found labels
type LabelResult struct {
	// Label
	ID               uint
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        time.Time
	LabelUUID        string
	LabelSlug        string
	CustomSlug       string
	LabelName        string
	LabelPriority    int
	LabelCount       int
	LabelFavorite    bool
	LabelDescription string
	LabelNotes       string
}

// PhotoLabel returns a photo label entity if exists.
func (q *Query) PhotoLabel(photoID, labelID uint) (label entity.PhotoLabel, err error) {
	if err := q.db.Where("photo_id = ? AND label_id = ?", photoID, labelID).Preload("Photo").Preload("Label").First(&label).Error; err != nil {
		return label, err
	}

	return label, nil
}

// LabelBySlug returns a Label based on the slug name.
func (q *Query) LabelBySlug(labelSlug string) (label entity.Label, err error) {
	if err := q.db.Where("label_slug = ? OR custom_slug = ?", labelSlug, labelSlug).Preload("Links").First(&label).Error; err != nil {
		return label, err
	}

	return label, nil
}

// LabelByUUID returns a Label based on the label UUID.
func (q *Query) LabelByUUID(labelUUID string) (label entity.Label, err error) {
	if err := q.db.Where("label_uuid = ?", labelUUID).Preload("Links").First(&label).Error; err != nil {
		return label, err
	}

	return label, nil
}

// LabelThumbBySlug returns a label preview file based on the slug name.
func (q *Query) LabelThumbBySlug(labelSlug string) (file entity.File, err error) {
	if err := q.db.Where("files.file_primary AND files.deleted_at IS NULL").
		Joins("JOIN labels ON labels.label_slug = ?", labelSlug).
		Joins("JOIN photos_labels ON photos_labels.label_id = labels.id AND photos_labels.photo_id = files.photo_id").
		Joins("JOIN photos ON photos.id = files.photo_id AND photos.deleted_at IS NULL").
		Order("photos.photo_quality DESC, photos_labels.uncertainty ASC").
		First(&file).Error; err != nil {
		return file, err
	}

	return file, nil
}

// LabelThumbByUUID returns a label preview file based on the label UUID.
func (q *Query) LabelThumbByUUID(labelUUID string) (file entity.File, err error) {
	// Search matching label
	err = q.db.Where("files.file_primary AND files.deleted_at IS NULL").
		Joins("JOIN labels ON labels.label_uuid = ?", labelUUID).
		Joins("JOIN photos_labels ON photos_labels.label_id = labels.id AND photos_labels.photo_id = files.photo_id").
		Joins("JOIN photos ON photos.id = files.photo_id AND photos.deleted_at IS NULL").
		Order("photos.photo_quality DESC, photos_labels.uncertainty ASC").
		First(&file).Error

	if err == nil {
		return file, nil
	}

	// If failed, search for category instead
	err = q.db.Where("files.file_primary AND files.deleted_at IS NULL").
		Joins("JOIN photos_labels ON photos_labels.photo_id = files.photo_id").
		Joins("JOIN categories c ON photos_labels.label_id = c.label_id").
		Joins("JOIN labels ON c.category_id = labels.id AND labels.label_uuid= ?", labelUUID).
		Joins("JOIN photos ON photos.id = files.photo_id AND photos.deleted_at IS NULL").
		Order("photos.photo_quality DESC, photos_labels.uncertainty ASC").
		First(&file).Error

	return file, err
}

// Labels searches labels based on their name.
func (q *Query) Labels(f form.LabelSearch) (results []LabelResult, err error) {
	if err := f.ParseQueryString(); err != nil {
		return results, err
	}

	defer log.Debug(capture.Time(time.Now(), fmt.Sprintf("labels: %+v", f)))

	s := q.db.NewScope(nil).DB()

	// s.LogMode(true)

	s = s.Table("labels").
		Select(`labels.*`).
		Where("labels.deleted_at IS NULL").
		Group("labels.id")

	if f.ID != "" {
		s = s.Where("labels.label_uuid = ?", f.ID)

		if result := s.Scan(&results); result.Error != nil {
			return results, result.Error
		}

		return results, nil
	}

	if f.Query != "" {
		var labelIds []uint
		var categories []entity.Category
		var label entity.Label

		slugString := slug.Make(f.Query)
		likeString := "%" + strings.ToLower(f.Query) + "%"

		if result := q.db.First(&label, "label_slug = ? OR custom_slug = ?", slugString, slugString); result.Error != nil {
			log.Infof("search: label \"%s\" not found", f.Query)

			s = s.Where("LOWER(labels.label_name) LIKE ?", likeString)
		} else {
			labelIds = append(labelIds, label.ID)

			q.db.Where("category_id = ?", label.ID).Find(&categories)

			for _, category := range categories {
				labelIds = append(labelIds, category.LabelID)
			}

			log.Infof("search: label \"%s\" includes %d categories", label.LabelName, len(labelIds))

			s = s.Where("labels.id IN (?)", labelIds)
		}
	}

	if f.Favorites {
		s = s.Where("labels.label_favorite = 1")
	}

	if !f.All {
		s = s.Where("labels.label_priority >= 0 OR labels.label_favorite = 1")
	}

	switch f.Order {
	case "slug":
		s = s.Order("labels.label_favorite DESC, custom_slug ASC")
	default:
		s = s.Order("labels.label_favorite DESC, custom_slug ASC")
	}

	if f.Count > 0 && f.Count <= 1000 {
		s = s.Limit(f.Count).Offset(f.Offset)
	} else {
		s = s.Limit(100).Offset(0)
	}

	if result := s.Scan(&results); result.Error != nil {
		return results, result.Error
	}

	return results, nil
}
