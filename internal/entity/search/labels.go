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

	// Filter private labels.
	if frm.Public {
		s = s.Where("labels.label_nsfw = 0")
	} else if frm.NSFW {
		s = s.Where("labels.label_nsfw = 1")
	}

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
		likeString := "%" + frm.Query + "%"

		if labelIds, findErr := entity.FindLabelIDs(frm.Query, " ", true); findErr != nil || len(labelIds) == 0 {
			log.Infof("search: label %s not found", clean.Log(frm.Query))

			s = s.Where("labels.label_name LIKE ?", likeString)
		} else {
			log.Infof("search: label %s resolves to %d labels", clean.Log(frm.Query), len(labelIds))

			s = s.Where("labels.id IN (?)", labelIds)
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
