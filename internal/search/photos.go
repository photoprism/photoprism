package search

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Photos searches for photos based on a Form and returns PhotoResults ([]Photo).
func Photos(f form.SearchPhotos) (results PhotoResults, count int, err error) {
	start := time.Now()

	if err := f.ParseQueryString(); err != nil {
		return PhotoResults{}, 0, err
	}

	s := UnscopedDb()
	// s = s.LogMode(true)

	// Base query.
	s = s.Table("photos").
		Select(`photos.*, photos.id AS composite_id,
		files.id AS file_id, files.file_uid, files.instance_id, files.file_primary, files.file_sidecar, 
		files.file_portrait,files.file_video, files.file_missing, files.file_name, files.file_root, files.file_hash, 
		files.file_codec, files.file_type, files.file_mime, files.file_width, files.file_height, 
		files.file_aspect_ratio, files.file_orientation, files.file_main_color, files.file_colors, files.file_luminance, 
		files.file_chroma, files.file_projection, files.file_diff, files.file_duration, files.file_size,
		cameras.camera_make, cameras.camera_model,
		lenses.lens_make, lenses.lens_model,
		places.place_label, places.place_city, places.place_state, places.place_country`).
		Joins("JOIN files ON photos.id = files.photo_id AND files.file_missing = 0 AND files.deleted_at IS NULL").
		Joins("LEFT JOIN cameras ON photos.camera_id = cameras.id").
		Joins("LEFT JOIN lenses ON photos.lens_id = lenses.id").
		Joins("LEFT JOIN places ON photos.place_id = places.id")

	// Limit result count.
	if f.Count > 0 && f.Count <= MaxResults {
		s = s.Limit(f.Count).Offset(f.Offset)
	} else {
		s = s.Limit(MaxResults).Offset(f.Offset)
	}

	// Set sort order.
	switch f.Order {
	case entity.SortOrderEdited:
		s = s.Where("edited_at IS NOT NULL").Order("edited_at DESC, photos.photo_uid, files.file_primary DESC")
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
	case entity.SortOrderAdded:
		s = s.Order("photos.id DESC, files.file_primary DESC")
	case entity.SortOrderSimilar:
		s = s.Where("files.file_diff > 0")
		s = s.Order("photos.photo_color, photos.cell_id, files.file_diff, taken_at DESC, files.file_primary DESC")
	case entity.SortOrderName:
		s = s.Order("photos.photo_path, photos.photo_name, files.file_primary DESC")
	default:
		s = s.Order("taken_at DESC, photos.photo_uid, files.file_primary DESC")
	}

	// Include hidden files?
	if !f.Hidden {
		s = s.Where("files.file_type = 'jpg' OR files.file_video = 1")

		if f.Error {
			s = s.Where("files.file_error <> ''")
		} else {
			s = s.Where("files.file_error = ''")
		}
	}

	// Return primary files only.
	if f.Primary {
		s = s.Where("files.file_primary = 1")
	}

	if f.UID != "" {
		s = s.Where("photos.photo_uid IN (?)", strings.Split(strings.ToLower(f.UID), txt.Or))

		// Take shortcut?
		if f.Album == "" && f.Query == "" {
			s = s.Order("files.file_primary DESC")

			if result := s.Scan(&results); result.Error != nil {
				return results, 0, result.Error
			}

			log.Debugf("photos: found %s for %s [%s]", english.Plural(len(results), "result", "results"), f.SerializeAll(), time.Since(start))

			if f.Merged {
				return results.Merged()
			}

			return results, len(results), nil
		}
	}

	// Filter by label, label category and keywords.
	var categories []entity.Category
	var labels []entity.Label
	var labelIds []uint

	if f.Label != "" {
		if err := Db().Where(AnySlug("label_slug", f.Label, txt.Or)).Or(AnySlug("custom_slug", f.Label, txt.Or)).Find(&labels).Error; len(labels) == 0 || err != nil {
			log.Debugf("search: label %s not found", txt.LogParamLower(f.Label))
			return PhotoResults{}, 0, nil
		} else {
			for _, l := range labels {
				labelIds = append(labelIds, l.ID)

				Db().Where("category_id = ?", l.ID).Find(&categories)

				log.Infof("search: label %s includes %d categories", txt.LogParamLower(l.LabelName), len(categories))

				for _, category := range categories {
					labelIds = append(labelIds, category.LabelID)
				}
			}

			s = s.Joins("JOIN photos_labels ON photos_labels.photo_id = photos.id AND photos_labels.uncertainty < 100 AND photos_labels.label_id IN (?)", labelIds).
				Group("photos.id, files.id")
		}
	}

	// Set search filters based on search terms.
	if terms := txt.SearchTerms(f.Query); f.Query != "" && len(terms) == 0 {
		if f.Title == "" {
			f.Title = fmt.Sprintf("%s*", strings.Trim(f.Query, "%*"))
			f.Query = ""
		}
	} else if len(terms) > 0 {
		switch {
		case terms["faces"]:
			f.Query = strings.ReplaceAll(f.Query, "faces", "")
			f.Faces = "true"
		case terms["people"]:
			f.Query = strings.ReplaceAll(f.Query, "people", "")
			f.Faces = "true"
		case terms["videos"]:
			f.Query = strings.ReplaceAll(f.Query, "videos", "")
			f.Video = true
		case terms["video"]:
			f.Query = strings.ReplaceAll(f.Query, "video", "")
			f.Video = true
		case terms["live"]:
			f.Query = strings.ReplaceAll(f.Query, "live", "")
			f.Live = true
		case terms["raws"]:
			f.Query = strings.ReplaceAll(f.Query, "raws", "")
			f.Raw = true
		case terms["favorites"]:
			f.Query = strings.ReplaceAll(f.Query, "favorites", "")
			f.Favorite = true
		case terms["stacks"]:
			f.Query = strings.ReplaceAll(f.Query, "stacks", "")
			f.Stack = true
		case terms["panoramas"]:
			f.Query = strings.ReplaceAll(f.Query, "panoramas", "")
			f.Panorama = true
		case terms["scans"]:
			f.Query = strings.ReplaceAll(f.Query, "scans", "")
			f.Scan = true
		case terms["monochrome"]:
			f.Query = strings.ReplaceAll(f.Query, "monochrome", "")
			f.Mono = true
		case terms["mono"]:
			f.Query = strings.ReplaceAll(f.Query, "mono", "")
			f.Mono = true
		}
	}

	// Filter by location?
	if f.Geo == true {
		s = s.Where("photos.cell_id <> 'zz'")

		for _, where := range LikeAnyKeyword("k.keyword", f.Query) {
			s = s.Where("photos.id IN (SELECT pk.photo_id FROM keywords k JOIN photos_keywords pk ON k.id = pk.keyword_id WHERE (?))", gorm.Expr(where))
		}
	} else if f.Query != "" {
		if err := Db().Where(AnySlug("custom_slug", f.Query, " ")).Find(&labels).Error; len(labels) == 0 || err != nil {
			log.Debugf("search: label %s not found, using fuzzy search", txt.LogParamLower(f.Query))

			for _, where := range LikeAnyKeyword("k.keyword", f.Query) {
				s = s.Where("photos.id IN (SELECT pk.photo_id FROM keywords k JOIN photos_keywords pk ON k.id = pk.keyword_id WHERE (?))", gorm.Expr(where))
			}
		} else {
			for _, l := range labels {
				labelIds = append(labelIds, l.ID)

				Db().Where("category_id = ?", l.ID).Find(&categories)

				log.Debugf("search: label %s includes %d categories", txt.LogParamLower(l.LabelName), len(categories))

				for _, category := range categories {
					labelIds = append(labelIds, category.LabelID)
				}
			}

			if wheres := LikeAnyKeyword("k.keyword", f.Query); len(wheres) > 0 {
				for _, where := range wheres {
					s = s.Where("photos.id IN (SELECT pk.photo_id FROM keywords k JOIN photos_keywords pk ON k.id = pk.keyword_id WHERE (?)) OR "+
						"photos.id IN (SELECT pl.photo_id FROM photos_labels pl WHERE pl.uncertainty < 100 AND pl.label_id IN (?))", gorm.Expr(where), labelIds)
				}
			} else {
				s = s.Where("photos.id IN (SELECT pl.photo_id FROM photos_labels pl WHERE pl.uncertainty < 100 AND pl.label_id IN (?))", labelIds)
			}
		}
	}

	// Search for one or more keywords?
	if f.Keywords != "" {
		for _, where := range LikeAnyWord("k.keyword", f.Keywords) {
			s = s.Where("photos.id IN (SELECT pk.photo_id FROM keywords k JOIN photos_keywords pk ON k.id = pk.keyword_id WHERE (?))", gorm.Expr(where))
		}
	}

	// Filter by number of faces?
	if txt.IsUInt(f.Faces) {
		s = s.Where("photos.photo_faces >= ?", txt.Int(f.Faces))
	} else if txt.New(f.Faces) && f.Face == "" {
		f.Face = f.Faces
		f.Faces = ""
	} else if txt.Yes(f.Faces) {
		s = s.Where("photos.photo_faces > 0")
	} else if txt.No(f.Faces) {
		s = s.Where("photos.photo_faces = 0")
	}

	// Filter for specific face clusters? Example: PLJ7A3G4MBGZJRMVDIUCBLC46IAP4N7O
	if len(f.Face) >= 32 {
		for _, f := range strings.Split(strings.ToUpper(f.Face), txt.And) {
			s = s.Where(fmt.Sprintf("photos.id IN (SELECT photo_id FROM files f JOIN %s m ON f.file_uid = m.file_uid AND m.marker_invalid = 0 WHERE face_id IN (?))",
				entity.Marker{}.TableName()), strings.Split(f, txt.Or))
		}
	} else if txt.New(f.Face) {
		s = s.Where(fmt.Sprintf("photos.id IN (SELECT photo_id FROM files f JOIN %s m ON f.file_uid = m.file_uid AND m.marker_invalid = 0 AND m.marker_type = ? WHERE subj_uid IS NULL OR subj_uid = '')",
			entity.Marker{}.TableName()), entity.MarkerFace)
	} else if txt.No(f.Face) {
		s = s.Where(fmt.Sprintf("photos.id IN (SELECT photo_id FROM files f JOIN %s m ON f.file_uid = m.file_uid AND m.marker_invalid = 0 AND m.marker_type = ? WHERE face_id IS NULL OR face_id = '')",
			entity.Marker{}.TableName()), entity.MarkerFace)
	} else if txt.Yes(f.Face) {
		s = s.Where(fmt.Sprintf("photos.id IN (SELECT photo_id FROM files f JOIN %s m ON f.file_uid = m.file_uid AND m.marker_invalid = 0 AND m.marker_type = ? WHERE face_id IS NOT NULL AND face_id <> '')",
			entity.Marker{}.TableName()), entity.MarkerFace)
	}

	// Filter for one or more subjects?
	if f.Subject != "" {
		for _, subj := range strings.Split(strings.ToLower(f.Subject), txt.And) {
			if subjects := strings.Split(subj, txt.Or); rnd.ContainsUIDs(subjects, 'j') {
				s = s.Where(fmt.Sprintf("photos.id IN (SELECT photo_id FROM files f JOIN %s m ON f.file_uid = m.file_uid AND m.marker_invalid = 0 WHERE subj_uid IN (?))",
					entity.Marker{}.TableName()), subjects)
			} else {
				s = s.Where(fmt.Sprintf("photos.id IN (SELECT photo_id FROM files f JOIN %s m ON f.file_uid = m.file_uid AND m.marker_invalid = 0 JOIN %s s ON s.subj_uid = m.subj_uid WHERE (?))",
					entity.Marker{}.TableName(), entity.Subject{}.TableName()), gorm.Expr(AnySlug("s.subj_slug", subj, txt.Or)))
			}
		}
	} else if f.Subjects != "" {
		for _, where := range LikeAllNames(Cols{"subj_name", "subj_alias"}, f.Subjects) {
			s = s.Where(fmt.Sprintf("photos.id IN (SELECT photo_id FROM files f JOIN %s m ON f.file_uid = m.file_uid AND m.marker_invalid = 0 JOIN %s s ON s.subj_uid = m.subj_uid WHERE (?))",
				entity.Marker{}.TableName(), entity.Subject{}.TableName()), gorm.Expr(where))
		}
	}

	// Filter by status?
	if f.Hidden {
		s = s.Where("photos.photo_quality = -1")
		s = s.Where("photos.deleted_at IS NULL")
	} else if f.Archived {
		s = s.Where("photos.photo_quality > -1")
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

	// Filter by camera?
	if f.Camera > 0 {
		s = s.Where("photos.camera_id = ?", f.Camera)
	}

	// Filter by camera lens?
	if f.Lens > 0 {
		s = s.Where("photos.lens_id = ?", f.Lens)
	}

	// Filter by year?
	if f.Year != "" {
		s = s.Where(AnyInt("photos.photo_year", f.Year, txt.Or, entity.UnknownYear, txt.YearMax))
	}

	// Filter by month?
	if f.Month != "" {
		s = s.Where(AnyInt("photos.photo_month", f.Month, txt.Or, entity.UnknownMonth, txt.MonthMax))
	}

	// Filter by day?
	if f.Day != "" {
		s = s.Where(AnyInt("photos.photo_day", f.Day, txt.Or, entity.UnknownDay, txt.DayMax))
	}

	// Filter by main color?
	if f.Color != "" {
		s = s.Where("files.file_main_color IN (?)", strings.Split(strings.ToLower(f.Color), txt.Or))
	}

	// Find favorites only?
	if f.Favorite {
		s = s.Where("photos.photo_favorite = 1")
	}

	// Find scans only?
	if f.Scan {
		s = s.Where("photos.photo_scan = 1")
	}

	// Find panoramas only?
	if f.Panorama {
		s = s.Where("photos.photo_panorama = 1")
	}

	// Find portraits only?
	if f.Portrait {
		s = s.Where("files.file_portrait = 1")
	}

	if f.Stackable {
		s = s.Where("photos.photo_stack > -1")
	} else if f.Unstacked {
		s = s.Where("photos.photo_stack = -1")
	}

	// Filter by location country?
	if f.Country != "" {
		s = s.Where("photos.photo_country IN (?)", strings.Split(strings.ToLower(f.Country), txt.Or))
	}

	// Filter by location state?
	if f.State != "" {
		s = s.Where("places.place_state IN (?)", strings.Split(f.State, txt.Or))
	}

	// Filter by location category?
	if f.Category != "" {
		s = s.Joins("JOIN cells ON photos.cell_id = cells.id").
			Where("cells.cell_category IN (?)", strings.Split(strings.ToLower(f.Category), txt.Or))
	}

	// Filter by media type?
	if f.Type != "" {
		s = s.Where("photos.photo_type IN (?)", strings.Split(strings.ToLower(f.Type), txt.Or))
	} else if f.Video {
		s = s.Where("photos.photo_type = 'video'")
	} else if f.Photo {
		s = s.Where("photos.photo_type IN ('image','raw','live')")
	} else if f.Raw {
		s = s.Where("photos.photo_type = 'raw'")
	} else if f.Live {
		s = s.Where("photos.photo_type = 'live'")
	}

	// Filter by storage path?
	if f.Path != "" {
		p := f.Path

		if strings.HasPrefix(p, "/") {
			p = p[1:]
		}

		if strings.HasSuffix(p, "/") {
			s = s.Where("photos.photo_path = ?", p[:len(p)-1])
		} else {
			where, values := OrLike("photos.photo_path", p)
			s = s.Where(where, values...)
		}
	}

	// Filter by primary file name without path and extension.
	if f.Name != "" {
		where, names := OrLike("photos.photo_name", f.Name)

		// Omit file path and known extensions.
		for i := range names {
			names[i] = fs.StripKnownExt(path.Base(names[i].(string)))
		}

		s = s.Where(where, names...)
	}

	// Filter by complete file names?
	if f.Filename != "" {
		where, values := OrLike("files.file_name", f.Filename)
		s = s.Where(where, values...)
	}

	// Filter by original file name?
	if f.Original != "" {
		where, values := OrLike("photos.original_name", f.Original)
		s = s.Where(where, values...)
	}

	// Filter by photo title?
	if f.Title != "" {
		where, values := OrLike("photos.photo_title", f.Title)
		s = s.Where(where, values...)
	}

	// Filter by file hash?
	if f.Hash != "" {
		s = s.Where("files.file_hash IN (?)", strings.Split(strings.ToLower(f.Hash), txt.Or))
	}

	if f.Mono {
		s = s.Where("files.file_chroma = 0 OR file_colors = '111111111'")
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

	// Filter by approx distance to coordinates:
	if f.Lat != 0 {
		latMin := f.Lat - Radius*float32(f.Dist)
		latMax := f.Lat + Radius*float32(f.Dist)
		s = s.Where("photos.photo_lat BETWEEN ? AND ?", latMin, latMax)
	}
	if f.Lng != 0 {
		lngMin := f.Lng - Radius*float32(f.Dist)
		lngMax := f.Lng + Radius*float32(f.Dist)
		s = s.Where("photos.photo_lng BETWEEN ? AND ?", lngMin, lngMax)
	}

	if !f.Before.IsZero() {
		s = s.Where("photos.taken_at <= ?", f.Before.Format("2006-01-02"))
	}

	if !f.After.IsZero() {
		s = s.Where("photos.taken_at >= ?", f.After.Format("2006-01-02"))
	}

	// Find stacks only?
	if f.Stack {
		s = s.Where("photos.id IN (SELECT a.photo_id FROM files a JOIN files b ON a.id != b.id AND a.photo_id = b.photo_id AND a.file_type = b.file_type WHERE a.file_type='jpg')")
	}

	// Filter by album?
	if rnd.IsPPID(f.Album, 'a') {
		if f.Filter != "" {
			s = s.Where("photos.photo_uid NOT IN (SELECT photo_uid FROM photos_albums pa WHERE pa.hidden = 1 AND pa.album_uid = ?)", f.Album)
		} else {
			s = s.Joins("JOIN photos_albums ON photos_albums.photo_uid = photos.photo_uid").
				Where("photos_albums.hidden = 0 AND photos_albums.album_uid = ?", f.Album)
		}
	} else if f.Unsorted && f.Filter == "" {
		s = s.Where("photos.photo_uid NOT IN (SELECT photo_uid FROM photos_albums pa WHERE pa.hidden = 0)")
	} else if f.Albums != "" || f.Album != "" {
		if f.Albums == "" {
			f.Albums = f.Album
		}

		for _, where := range LikeAnyWord("a.album_title", f.Albums) {
			s = s.Where("photos.photo_uid IN (SELECT pa.photo_uid FROM photos_albums pa JOIN albums a ON a.album_uid = pa.album_uid AND pa.hidden = 0 WHERE (?))", gorm.Expr(where))
		}
	}

	if err := s.Scan(&results).Error; err != nil {
		return results, 0, err
	}

	log.Debugf("photos: found %s for %s [%s]", english.Plural(len(results), "result", "results"), f.SerializeAll(), time.Since(start))

	if f.Merged {
		return results.Merged()
	}

	return results, len(results), nil
}
