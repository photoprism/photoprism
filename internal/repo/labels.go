package repo

import (
	"fmt"
	"strings"
	"time"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/models"
	"github.com/photoprism/photoprism/internal/util"
)

// LabelResult contains found labels
type LabelResult struct {
	// Label
	ID               uint
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        time.Time
	LabelSlug        string
	LabelName        string
	LabelPriority    int
	LabelCount       int
	LabelFavorite    bool
	LabelDescription string
	LabelNotes       string
}

// FindLabelBySlug returns a Label based on the slug name.
func (s *Repo) FindLabelBySlug(labelSlug string) (label models.Label, err error) {
	if err := s.db.Where("label_slug = ?", labelSlug).First(&label).Error; err != nil {
		return label, err
	}

	return label, nil
}

// FindLabelThumbBySlug returns a label preview file based on the slug name.
func (s *Repo) FindLabelThumbBySlug(labelSlug string) (file models.File, err error) {
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

// Labels searches labels based on their name.
func (s *Repo) Labels(f form.LabelSearch) (results []LabelResult, err error) {
	if err := f.ParseQueryString(); err != nil {
		return results, err
	}

	defer util.ProfileTime(time.Now(), fmt.Sprintf("search: %+v", f))

	q := s.db.NewScope(nil).DB()

	// q.LogMode(true)

	q = q.Table("labels").
		Select(`labels.*, COUNT(photos_labels.label_id) AS label_count`).
		Joins("JOIN photos_labels ON photos_labels.label_id = labels.id").
		Where("labels.deleted_at IS NULL").
		Group("labels.id")

	if f.Query != "" {
		var labelIds []uint
		var categories []models.Category
		var label models.Label

		likeString := "%" + strings.ToLower(f.Query) + "%"

		if result := s.db.First(&label, "LOWER(label_name) LIKE LOWER(?)", f.Query); result.Error != nil {
			log.Infof("search: label \"%s\" not found", f.Query)

			q = q.Where("LOWER(labels.label_name) LIKE ?", likeString)
		} else {
			labelIds = append(labelIds, label.ID)

			s.db.Where("category_id = ?", label.ID).Find(&categories)

			for _, category := range categories {
				labelIds = append(labelIds, category.LabelID)
			}

			log.Infof("search: label \"%s\" includes %d categories", label.LabelName, len(labelIds))

			q = q.Where("labels.id IN (?) OR LOWER(labels.label_name) LIKE ?", labelIds, likeString)
		}
	}

	if f.Favorites {
		q = q.Where("labels.label_favorite = 1")
	}

	if f.Priority != 0 {
		q = q.Where("labels.label_priority > ?", f.Priority)
	} else {
		q = q.Where("labels.label_priority >= -2")
	}

	switch f.Order {
	case "slug":
		q = q.Order("labels.label_favorite DESC, label_slug ASC")
	default:
		q = q.Order("labels.label_favorite DESC, labels.label_priority DESC, label_count DESC, labels.created_at DESC")
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
