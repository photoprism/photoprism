package search

import (
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/sortby"
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
	case sortby.Slug:
		s = s.Order("custom_slug ASC, labels.photo_count DESC")
	case sortby.Count:
		s = s.Order("labels.photo_count DESC, custom_slug ASC")
	default:
		s = s.Order("labels.label_favorite DESC, labels.label_priority DESC, labels.photo_count DESC, custom_slug ASC")
	}

	if frm.UID != "" {
		s = s.Where("labels.label_uid IN (?)", strings.Split(strings.ToLower(frm.UID), txt.Or))

		if result := s.Scan(&results); result.Error != nil {
			return results, result.Error
		}

		return results, nil
	}

	if frm.Query != "" {
		var labelIDs []uint
		var categories []entity.Category
		// var label entity.Label
		var labels []entity.Label

		slugString := txt.Slug(frm.Query)
		likeString := "%" + frm.Query + "%"
		slugLike := slugString + "-c-_"
		cleanName := clean.LabelName(frm.Query)

		if result := Db().Where("label_slug = ? OR label_slug like ? OR custom_slug = ?", slugString, slugLike, slugString).Find(&labels); result.Error != nil {
			log.Errorf("search: label %s not found with error %s", clean.Log(frm.Query), result.Error)
			return results, result.Error
		}
		if len(labels) == 0 {
			log.Infof("search: label %s not found", clean.Log(frm.Query))

			s = s.Where("labels.label_name LIKE ?", likeString)
		} else {
			labelName := ""
			for _, label := range labels {
				if strings.EqualFold(clean.LabelName(label.LabelName), cleanName) {
					labelIDs = append(labelIDs, label.ID)
					labelName = label.LabelName
				}
			}
			labelCount := len(labelIDs)

			Db().Where("category_id in (?)", labelIDs).Find(&categories)

			for _, category := range categories {
				labelIDs = append(labelIDs, category.LabelID)
			}

			log.Infof("search: label %s includes %d categories", clean.Log(labelName), len(labelIDs)-labelCount)

			s = s.Where("labels.id IN (?)", labelIDs)
		}
	}

	if frm.Favorite {
		s = s.Where("labels.label_favorite = 1")
	}

	if frm.Query == "" && !frm.All {
		s = s.Where("labels.label_priority >= 0 AND labels.photo_count > 1 OR labels.label_favorite = 1")
	}

	if result := s.Scan(&results); result.Error != nil {
		return results, result.Error
	}

	return results, nil
}
