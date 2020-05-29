package query

import (
	"fmt"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/txt"
)

// PhotoSearch searches for photos based on a Form and returns PhotoResults ([]PhotoResult).
func PhotoSearch(f form.PhotoSearch) (results PhotoResults, count int, err error) {
	start := time.Now()

	if err := f.ParseQueryString(); err != nil {
		return results, 0, err
	}

	s := UnscopedDb()

	// s.LogMode(true)

	// Main search query, avoids (slow) left joins.
	s = s.Table("photos").
		Select(`photos.*,
		files.id AS file_id, files.file_uid, files.file_primary, files.file_missing, files.file_name,
		files.file_root, files.file_hash, files.file_codec, files.file_type, files.file_mime, files.file_width, 
		files.file_height, files.file_aspect_ratio, files.file_orientation, files.file_main_color, 
		files.file_colors, files.file_luminance, files.file_chroma,
		files.file_diff, files.file_video, files.file_duration, files.file_size,
		cameras.camera_make, cameras.camera_model,
		lenses.lens_make, lenses.lens_model,
		places.loc_label, places.loc_city, places.loc_state, places.loc_country`).
		Joins("JOIN files ON photos.id = files.photo_id AND files.file_missing = 0 AND files.deleted_at IS NULL").
		Joins("JOIN cameras ON photos.camera_id = cameras.id").
		Joins("JOIN lenses ON photos.lens_id = lenses.id").
		Joins("JOIN places ON photos.place_id = places.id").
		Where("files.file_type = 'jpg' OR files.file_video = 1")

	// Shortcut for known photo ids.
	if f.ID != "" {
		s = s.Where("photos.photo_uid IN (?)", strings.Split(f.ID, ","))
		s = s.Order("files.file_primary DESC")

		if result := s.Scan(&results); result.Error != nil {
			return results, 0, result.Error
		}

		log.Infof("photos: found %d results for %s [%s]", len(results), f.SerializeAll(), time.Since(start))

		if f.Merged {
			return results.Merged()
		}

		return results, len(results), nil
	}

	// Filter by label, label category and keywords.
	var categories []entity.Category
	var labels []entity.Label
	var labelIds []uint

	if f.Label != "" {
		if err := Db().Where(AnySlug("label_slug", f.Label, ",")).Or(AnySlug("custom_slug", f.Label, ",")).Find(&labels).Error; len(labels) == 0 || err != nil {
			log.Errorf("search: labels %s not found", txt.Quote(f.Label))
			return results, 0, fmt.Errorf("%s not found", txt.Quote(f.Label))
		} else {
			for _, l := range labels {
				labelIds = append(labelIds, l.ID)

				Db().Where("category_id = ?", l.ID).Find(&categories)

				log.Infof("search: label %s includes %d categories", txt.Quote(l.LabelName), len(categories))

				for _, category := range categories {
					labelIds = append(labelIds, category.LabelID)
				}
			}

			s = s.Joins("JOIN photos_labels ON photos_labels.photo_id = photos.id AND photos_labels.uncertainty < 100 AND photos_labels.label_id IN (?)", labelIds).
				Group("photos.id, files.id")
		}
	}

	// Filter by location.
	if f.Location == true {
		s = s.Where("location_id <> ''")

		if likeAny := LikeAny("k.keyword", f.Query); likeAny != "" {
			s = s.Where("photos.id IN (SELECT pk.photo_id FROM keywords k JOIN photos_keywords pk ON k.id = pk.keyword_id WHERE (?))", gorm.Expr(likeAny))
		}
	} else if f.Query != "" {
		if len(f.Query) < 2 {
			return results, 0, fmt.Errorf("query too short")
		}

		if err := Db().Where(AnySlug("custom_slug", f.Query, " ")).Find(&labels).Error; len(labels) == 0 || err != nil {
			log.Infof("search: label %s not found, using fuzzy search", txt.Quote(f.Query))

			if likeAny := LikeAny("k.keyword", f.Query); likeAny != "" {
				s = s.Where("photos.id IN (SELECT pk.photo_id FROM keywords k JOIN photos_keywords pk ON k.id = pk.keyword_id WHERE (?))", gorm.Expr(likeAny))
			}
		} else {
			for _, l := range labels {
				labelIds = append(labelIds, l.ID)

				Db().Where("category_id = ?", l.ID).Find(&categories)

				log.Infof("search: label %s includes %d categories", txt.Quote(l.LabelName), len(categories))

				for _, category := range categories {
					labelIds = append(labelIds, category.LabelID)
				}
			}

			if likeAny := LikeAny("k.keyword", f.Query); likeAny != "" {
				s = s.Where("photos.id IN (SELECT pk.photo_id FROM keywords k JOIN photos_keywords pk ON k.id = pk.keyword_id WHERE (?)) OR "+
					"photos.id IN (SELECT pl.photo_id FROM photos_labels pl WHERE pl.uncertainty < 100 AND pl.label_id IN (?))", gorm.Expr(likeAny), labelIds)
			} else {
				s = s.Where("photos.id IN (SELECT pl.photo_id FROM photos_labels pl WHERE pl.uncertainty < 100 AND pl.label_id IN (?))", labelIds)
			}
		}
	}

	// Filter by status.
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

	// Filter by additional flags and metadata.
	if f.Error {
		s = s.Where("files.file_error <> ''")
	} else {
		s = s.Where("files.file_error = ''")
	}

	if f.Camera > 0 {
		s = s.Where("photos.camera_id = ?", f.Camera)
	}

	if f.Lens > 0 {
		s = s.Where("photos.lens_id = ?", f.Lens)
	}

	if (f.Year > 0 && f.Year <= txt.YearMax) || f.Year == entity.YearUnknown {
		s = s.Where("photos.photo_year = ?", f.Year)
	}

	if (f.Month >= txt.MonthMin && f.Month <= txt.MonthMax) || f.Month == entity.MonthUnknown {
		s = s.Where("photos.photo_month = ?", f.Month)
	}

	if f.Color != "" {
		s = s.Where("files.file_main_color IN (?)", strings.Split(strings.ToLower(f.Color), ","))
	}

	if f.Favorite {
		s = s.Where("photos.photo_favorite = 1")
	}

	if f.Country != "" {
		s = s.Where("photos.photo_country IN (?)", strings.Split(strings.ToLower(f.Country), ","))
	}

	if f.State != "" {
		s = s.Where("places.loc_state IN (?)", strings.Split(txt.Title(f.State), ","))
	}

	if f.Category != "" {
		s = s.Joins("JOIN locations ON photos.location_id = locations.id").
			Where("locations.loc_category IN (?)", strings.Split(strings.ToLower(f.Category), ","))
	}

	// Filter by media type.
	if f.Type != "" {
		s = s.Where("photos.photo_type IN (?)", strings.Split(strings.ToLower(f.Type), ","))
	}

	if f.Video {
		s = s.Where("photos.photo_type = 'video'")
	} else if f.Photo {
		s = s.Where("photos.photo_type IN ('image','raw','live')")
	}

	if f.Path != "" {
		p := f.Path

		if strings.HasPrefix(p, "/") {
			p = p[1:]
		}

		if strings.HasSuffix(p, "/") {
			s = s.Where("photos.photo_path = ?", p[:len(p)-1])
		} else if strings.Contains(p, ",") {
			s = s.Where("photos.photo_path IN (?)", strings.Split(p, ","))
		} else {
			s = s.Where("photos.photo_path LIKE ?", strings.ReplaceAll(p, "*", "%"))
		}
	}

	if f.Name != "" {
		s = s.Where("photos.photo_name LIKE ?", strings.ReplaceAll(f.Name, "*", "%"))
	}

	if f.Title != "" {
		s = s.Where("LOWER(photos.photo_title) LIKE ?", strings.ReplaceAll(strings.ToLower(f.Title), "*", "%"))
	}

	if f.Hash != "" {
		s = s.Where("files.file_hash IN (?)", strings.Split(strings.ToLower(f.Hash), ","))
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

	// Filter by distance (approximation).
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

	if f.Album != "" {
		if f.Filter != "" {
			s = s.Where("photos.photo_uid NOT IN (SELECT photo_uid FROM photos_albums pa WHERE pa.hidden = 1 AND pa.album_uid = ?)", f.Album)
		} else {
			s = s.Joins("JOIN photos_albums ON photos_albums.photo_uid = photos.photo_uid").Where("photos_albums.hidden = 0 AND photos_albums.album_uid = ?", f.Album)
		}
	}

	// Set sort order for results.
	switch f.Order {
	case entity.SortOrderRelevance:
		if f.Label != "" {
			s = s.Order("photo_quality DESC, photos_labels.uncertainty ASC, taken_at DESC, files.file_primary DESC")
		} else {
			s = s.Order("photo_quality DESC, taken_at DESC, files.file_primary DESC")
		}
	case entity.SortOrderNewest:
		s = s.Order("taken_at DESC, photos.photo_uid, files.file_primary DESC")
	case entity.SortOrderOldest:
		s = s.Order("taken_at, photos.photo_uid, files.file_primary DESC")
	case entity.SortOrderImported:
		s = s.Order("photos.id DESC, files.file_primary DESC")
	case entity.SortOrderSimilar:
		s = s.Where("files.file_diff > 0")
		s = s.Order("files.file_main_color, photos.location_id, files.file_diff, taken_at DESC, files.file_primary DESC")
	case entity.SortOrderName:
		s = s.Order("photos.photo_path, photos.photo_name, files.file_primary DESC")
	default:
		s = s.Order("taken_at DESC, photos.photo_uid, files.file_primary DESC")
	}

	if f.Count > 0 && f.Count <= 1000 {
		s = s.Limit(f.Count).Offset(f.Offset)
	} else {
		s = s.Limit(100).Offset(0)
	}

	if result := s.Scan(&results); result.Error != nil {
		return results, 0, result.Error
	}

	log.Infof("photos: found %d results for %s [%s]", len(results), f.SerializeAll(), time.Since(start))

	if f.Merged {
		return results.Merged()
	}

	return results, len(results), nil
}
