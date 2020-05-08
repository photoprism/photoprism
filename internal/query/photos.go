package query

import (
	"fmt"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Photos searches for photos based on a Form and returns a PhotoResult slice.
func Photos(f form.PhotoSearch) (results PhotoResults, count int, err error) {
	if err := f.ParseQueryString(); err != nil {
		return results, 0, err
	}

	defer log.Debug(capture.Time(time.Now(), fmt.Sprintf("photos: %+v", f)))

	s := UnscopedDb()

	// s.LogMode(true)

	s = s.Table("photos").
		Select(`photos.*,
		files.id AS file_id, files.file_uuid, files.file_primary, files.file_missing, files.file_name, files.file_hash, 
		files.file_type, files.file_mime, files.file_width, files.file_height, files.file_aspect_ratio, 
		files.file_orientation, files.file_main_color, files.file_colors, files.file_luminance, files.file_chroma,
		files.file_diff,
		cameras.camera_make, cameras.camera_model,
		lenses.lens_make, lenses.lens_model,
		places.loc_label, places.loc_city, places.loc_state, places.loc_country
		`).
		Joins("JOIN files ON files.photo_id = photos.id AND files.file_type = 'jpg' AND files.file_missing = 0 AND files.deleted_at IS NULL").
		Joins("JOIN cameras ON cameras.id = photos.camera_id").
		Joins("JOIN lenses ON lenses.id = photos.lens_id").
		Joins("JOIN places ON photos.place_id = places.id").
		Joins("LEFT JOIN photos_labels ON photos_labels.photo_id = photos.id AND photos_labels.uncertainty < 100").
		Group("photos.id, files.id")

	if f.ID != "" {
		s = s.Where("photos.photo_uuid = ?", f.ID)
		s = s.Order("files.file_primary DESC")

		if result := s.Scan(&results); result.Error != nil {
			return results, 0, result.Error
		}

		if f.Merged {
			return results.Merged()
		}

		return results, len(results), nil
	}

	var categories []entity.Category
	var label entity.Label
	var labelIds []uint

	if f.Label != "" {
		slugString := strings.ToLower(f.Label)
		if result := Db().First(&label, "label_slug =? OR custom_slug = ?", slugString, slugString); result.Error != nil {
			log.Errorf("search: label %s not found", txt.Quote(f.Label))
			return results, 0, fmt.Errorf("label %s not found", txt.Quote(f.Label))
		} else {
			labelIds = append(labelIds, label.ID)

			Db().Where("category_id = ?", label.ID).Find(&categories)

			for _, category := range categories {
				labelIds = append(labelIds, category.LabelID)
			}

			s = s.Where("photos_labels.label_id IN (?)", labelIds)
		}
	}

	if f.Location == true {
		s = s.Where("location_id > 0")

		if f.Query != "" {
			s = s.Joins("LEFT JOIN photos_keywords ON photos_keywords.photo_id = photos.id").
				Joins("LEFT JOIN keywords ON photos_keywords.keyword_id = keywords.id").
				Where("keywords.keyword LIKE ?", strings.ToLower(txt.Clip(f.Query, txt.ClipKeyword))+"%")
		}
	} else if f.Query != "" {
		if len(f.Query) < 2 {
			return results, 0, fmt.Errorf("query too short")
		}

		slugString := slug.Make(f.Query)
		lowerString := strings.ToLower(f.Query)
		likeString := txt.Clip(lowerString, txt.ClipKeyword) + "%"

		s = s.Joins("LEFT JOIN photos_keywords ON photos_keywords.photo_id = photos.id").
			Joins("LEFT JOIN keywords ON photos_keywords.keyword_id = keywords.id")

		if result := Db().First(&label, "label_slug = ? OR custom_slug = ?", slugString, slugString); result.Error != nil {
			log.Infof("search: label %s not found, using fuzzy search", txt.Quote(f.Query))

			s = s.Where("keywords.keyword LIKE ?", likeString)
		} else {
			labelIds = append(labelIds, label.ID)

			Db().Where("category_id = ?", label.ID).Find(&categories)

			for _, category := range categories {
				labelIds = append(labelIds, category.LabelID)
			}

			log.Infof("search: label %s includes %d categories", txt.Quote(label.LabelName), len(labelIds))

			s = s.Where("photos_labels.label_id IN (?) OR keywords.keyword LIKE ?", labelIds, likeString)
		}
	}

	if f.Archived {
		s = s.Where("photos.deleted_at IS NOT NULL")
	} else {
		s = s.Where("photos.deleted_at IS NULL")

		if f.Private {
			s = s.Where("photos.photo_private = 1")
		} else if f.Public {
			s = s.Where("photos.photo_private = 0")
		}

		if f.Review {
			s = s.Where("photos.photo_quality < 3")
		} else if f.Quality != 0 && f.Private == false {
			s = s.Where("photos.photo_quality >= ?", f.Quality)
		}
	}

	if f.Error {
		s = s.Where("files.file_error <> ''")
	}

	if f.Album != "" {
		s = s.Joins("JOIN photos_albums ON photos_albums.photo_uuid = photos.photo_uuid").Where("photos_albums.album_uuid = ?", f.Album)
	}

	if f.Camera > 0 {
		s = s.Where("photos.camera_id = ?", f.Camera)
	}

	if f.Lens > 0 {
		s = s.Where("photos.lens_id = ?", f.Lens)
	}

	if f.Year > 0 {
		s = s.Where("photos.photo_year = ?", f.Year)
	}

	if f.Month > 0 {
		s = s.Where("photos.photo_month = ?", f.Month)
	}

	if f.Color != "" {
		s = s.Where("files.file_main_color = ?", strings.ToLower(f.Color))
	}

	if f.Favorites {
		s = s.Where("photos.photo_favorite = 1")
	}

	if f.Story {
		s = s.Where("photos.photo_story = 1")
	}

	if f.Country != "" {
		s = s.Where("photos.photo_country = ?", f.Country)
	}

	if f.Title != "" {
		s = s.Where("LOWER(photos.photo_title) LIKE ?", fmt.Sprintf("%%%s%%", strings.ToLower(f.Title)))
	}

	if f.Hash != "" {
		s = s.Where("files.file_hash = ?", f.Hash)
	}

	if f.Duplicate {
		s = s.Where("files.file_duplicate = 1")
	}

	if f.Portrait {
		s = s.Where("files.file_portrait = 1")
	}

	if f.Mono {
		s = s.Where("files.file_chroma = 0")
	} else if f.Chroma > 9 {
		s = s.Where("files.file_chroma > ?", f.Chroma)
	} else if f.Chroma > 0 {
		s = s.Where("files.file_chroma > 0 AND files.file_chroma <= ?", f.Chroma)
	}

	if f.Diff != 0 {
		s = s.Where("files.file_diff = ?", f.Diff)
	}

	if f.Fmin > 0 {
		s = s.Where("photos.photo_f_number >= ?", f.Fmin)
	}

	if f.Fmax > 0 {
		s = s.Where("photos.photo_f_number <= ?", f.Fmax)
	}

	if f.Dist == 0 {
		f.Dist = 20
	} else if f.Dist > 5000 {
		f.Dist = 5000
	}

	// Inaccurate distance search, but probably 'good enough' for now
	if f.Lat > 0 {
		latMin := f.Lat - SearchRadius*float32(f.Dist)
		latMax := f.Lat + SearchRadius*float32(f.Dist)
		s = s.Where("photos.photo_lat BETWEEN ? AND ?", latMin, latMax)
	}

	if f.Lng > 0 {
		lngMin := f.Lng - SearchRadius*float32(f.Dist)
		lngMax := f.Lng + SearchRadius*float32(f.Dist)
		s = s.Where("photos.photo_lng BETWEEN ? AND ?", lngMin, lngMax)
	}

	if !f.Before.IsZero() {
		s = s.Where("photos.taken_at <= ?", f.Before.Format("2006-01-02"))
	}

	if !f.After.IsZero() {
		s = s.Where("photos.taken_at >= ?", f.After.Format("2006-01-02"))
	}

	switch f.Order {
	case entity.SortOrderRelevance:
		if f.Label != "" {
			s = s.Order("photo_quality DESC, photos_labels.uncertainty ASC, taken_at DESC, files.file_primary DESC")
		} else {
			s = s.Order("photo_quality DESC, taken_at DESC, files.file_primary DESC")
		}
	case entity.SortOrderNewest:
		s = s.Order("taken_at DESC, photos.photo_uuid, files.file_primary DESC")
	case entity.SortOrderOldest:
		s = s.Order("taken_at, photos.photo_uuid, files.file_primary DESC")
	case entity.SortOrderImported:
		s = s.Order("photos.id DESC, files.file_primary DESC")
	case entity.SortOrderSimilar:
		s = s.Order("files.file_main_color, photos.location_id, files.file_diff, taken_at DESC, files.file_primary DESC")
	default:
		s = s.Order("taken_at DESC, photos.photo_uuid, files.file_primary DESC")
	}

	if f.Count > 0 && f.Count <= 1000 {
		s = s.Limit(f.Count).Offset(f.Offset)
	} else {
		s = s.Limit(100).Offset(0)
	}

	if result := s.Scan(&results); result.Error != nil {
		return results, 0, result.Error
	}

	if f.Merged {
		return results.Merged()
	}

	return results, len(results), nil
}
