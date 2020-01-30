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
	LabelName        string
	LabelPriority    int
	LabelCount       int
	LabelFavorite    bool
	LabelDescription string
	LabelNotes       string
}

// FindLabelBySlug returns a Label based on the slug name.
func (s *Repo) FindLabelBySlug(labelSlug string) (label entity.Label, err error) {
	if err := s.db.Where("label_slug = ?", labelSlug).First(&label).Error; err != nil {
		return label, err
	}

	return label, nil
}

// FindLabelByUUID returns a Label based on the label UUID.
func (s *Repo) FindLabelByUUID(labelUUID string) (label entity.Label, err error) {
	if err := s.db.Where("label_uuid = ?", labelUUID).First(&label).Error; err != nil {
		return label, err
	}

	return label, nil
}

// FindLabelThumbBySlug returns a label preview file based on the slug name.
func (s *Repo) FindLabelThumbBySlug(labelSlug string) (file entity.File, err error) {
	// s.db.LogMode(true)

	if err := s.db.Where("files.file_primary AND files.deleted_at IS NULL").
		Joins("JOIN labels ON labels.label_slug = ?", labelSlug).
		Joins("JOIN photos_labels ON photos_labels.label_id = labels.id AND photos_labels.photo_id = files.photo_id").
		Order("photos_labels.label_uncertainty ASC").
		First(&file).Error; err != nil {
		return file, err
	}

	return file, nil
}

// FindLabelThumbByUUID returns a label preview file based on the label UUID.
func (s *Repo) FindLabelThumbByUUID(labelUUID string) (file entity.File, err error) {
	// Search matching label
	err = s.db.Where("files.file_primary AND files.deleted_at IS NULL").
		Joins("JOIN labels ON labels.label_uuid = ?", labelUUID).
		Joins("JOIN photos_labels ON photos_labels.label_id = labels.id AND photos_labels.photo_id = files.photo_id").
		Order("photos_labels.label_uncertainty ASC").
		First(&file).Error

	if err == nil {
		return file, nil
	}

	// If failed, search for category instead
	err = s.db.Where("files.file_primary AND files.deleted_at IS NULL").
		Joins("JOIN photos_labels ON photos_labels.photo_id = files.photo_id").
		Joins("JOIN categories c ON photos_labels.label_id = c.label_id").
		Joins("JOIN labels ON c.category_id = labels.id AND labels.label_uuid= ?", labelUUID).
		Order("photos_labels.label_uncertainty ASC").
		First(&file).Error

	return file, err
}

// Labels searches labels based on their name.
func (s *Repo) Labels(f form.LabelSearch) (results []LabelResult, err error) {
	if err := f.ParseQueryString(); err != nil {
		return results, err
	}

	defer log.Debug(capture.Time(time.Now(), fmt.Sprintf("labels: %+v", f)))

	q := s.db.NewScope(nil).DB()

	// q.LogMode(true)

	q = q.Table("labels").
		Select(`labels.*`).
		Where("labels.deleted_at IS NULL").
		Group("labels.id")

	if f.ID != "" {
		q = q.Where("labels.label_uuid = ?", f.ID)

		if result := q.Scan(&results); result.Error != nil {
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

		if result := s.db.First(&label, "label_slug = ?", slugString); result.Error != nil {
			log.Infof("search: label \"%s\" not found", f.Query)

			q = q.Where("LOWER(labels.label_name) LIKE ?", likeString)
		} else {
			labelIds = append(labelIds, label.ID)

			s.db.Where("category_id = ?", label.ID).Find(&categories)

			for _, category := range categories {
				labelIds = append(labelIds, category.LabelID)
			}

			log.Infof("search: label \"%s\" includes %d categories", label.LabelName, len(labelIds))

			q = q.Where("labels.id IN (?)", labelIds)
		}
	}

	if f.Favorites {
		q = q.Where("labels.label_favorite = 1")
	}

	if !f.All {
		q = q.Where("labels.label_priority >= 0 OR labels.label_favorite = 1")
	}

	switch f.Order {
	case "slug":
		q = q.Order("labels.label_favorite DESC, label_slug ASC")
	default:
		q = q.Order("labels.label_favorite DESC, label_slug ASC")
	}

	if f.Count > 0 && f.Count <= 1000 {
		q = q.Limit(f.Count).Offset(f.Offset)
	} else {
		q = q.Limit(100).Offset(0)
	}

	if result := q.Scan(&results); result.Error != nil {
		return results, result.Error
	}

	return results, nil
}
