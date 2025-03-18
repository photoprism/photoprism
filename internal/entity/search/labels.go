package search

import (
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Labels searches labels based on their name.
func Labels(frm form.SearchLabels) (results []Label, err error) {
	if err = frm.ParseQueryString(); err != nil {
		return results, err
	}

	s := UnscopedDb()
	// s.LogMode(true)

	// Base query.
	s = s.Table("labels").
		Select(`labels.*`).
		Where("labels.deleted_at IS NULL").
		Where("labels.photo_count > 0").
		Group("labels.id")

	// Limit result count.
	if frm.Count > 0 && frm.Count <= MaxResults {
		s = s.Limit(frm.Count).Offset(frm.Offset)
	} else {
		s = s.Limit(MaxResults).Offset(frm.Offset)
	}

	// Set sort order.
	switch frm.Order {
	case "slug":
		s = s.Order("labels.label_favorite DESC, custom_slug ASC")
	default:
		s = s.Order("labels.label_favorite DESC, custom_slug ASC")
	}

	if frm.UID != "" {
		s = s.Where("labels.label_uid IN (?)", strings.Split(strings.ToLower(frm.UID), txt.Or))

		if result := s.Scan(&results); result.Error != nil {
			return results, result.Error
		}

		return results, nil
	}

	if frm.Query != "" {
		var labelIds []uint
		var categories []entity.Category
		var label entity.Label

		slugString := txt.Slug(frm.Query)
		likeString := "%" + frm.Query + "%"

		if result := Db().First(&label, "label_slug = ? OR custom_slug = ?", slugString, slugString); result.Error != nil {
			log.Infof("search: label %s not found", clean.Log(frm.Query))

			s = s.Where("labels.label_name LIKE ?", likeString)
		} else {
			labelIds = append(labelIds, label.ID)

			Db().Where("category_id = ?", label.ID).Find(&categories)

			for _, category := range categories {
				labelIds = append(labelIds, category.LabelID)
			}

			log.Infof("search: label %s includes %d categories", clean.Log(label.LabelName), len(labelIds))

			s = s.Where("labels.id IN (?)", labelIds)
		}
	}

	if frm.Favorite {
		s = s.Where("labels.label_favorite = 1")
	}

	if !frm.All {
		s = s.Where("labels.label_priority >= 0 OR labels.label_favorite = 1")
	}

	if result := s.Scan(&results); result.Error != nil {
		return results, result.Error
	}

	return results, nil
}
