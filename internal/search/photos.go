package search

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/sortby"
	"github.com/photoprism/photoprism/pkg/txt"
)

// PhotosColsAll contains all supported result column names.
var PhotosColsAll = SelectString(Photo{}, []string{"*"})

// PhotosColsView contains the result column names necessary for the photo viewer.
var PhotosColsView = SelectString(Photo{}, SelectCols(GeoResult{}, []string{"*"}))

// FileTypes contains a list of browser-compatible file formats returned by search queries.
var FileTypes = []string{fs.ImageJPEG.String(), fs.ImagePNG.String(), fs.ImageGIF.String(), fs.ImageAVIF.String(), fs.ImageAVIFS.String(), fs.ImageWebP.String(), fs.VectorSVG.String()}

// Photos finds PhotoResults based on the search form without checking rights or permissions.
func Photos(f form.SearchPhotos) (results PhotoResults, count int, err error) {
	return searchPhotos(f, nil, PhotosColsAll)
}

// UserPhotos finds PhotoResults based on the search form and user session.
func UserPhotos(f form.SearchPhotos, sess *entity.Session) (results PhotoResults, count int, err error) {
	return searchPhotos(f, sess, PhotosColsAll)
}

// PhotoIds finds photo and file ids based on the search form provided and returns them as PhotoResults.
func PhotoIds(f form.SearchPhotos) (files PhotoResults, count int, err error) {
	f.Merged = false
	f.Primary = true
	return searchPhotos(f, nil, "photos.id, photos.photo_uid, files.file_uid")
}

// searchPhotos finds photos based on the search form and user session then returns them as PhotoResults.
func searchPhotos(f form.SearchPhotos, sess *entity.Session, resultCols string) (results PhotoResults, count int, err error) {
	start := time.Now()

	// Parse query string and filter.
	if err = f.ParseQueryString(); err != nil {
		log.Debugf("search: %s", err)
		return PhotoResults{}, 0, ErrBadRequest
	}

	// Specify table names and joins.
	s := UnscopedDb().Table(entity.File{}.TableName()).Select(resultCols).
		Joins("JOIN photos ON files.photo_id = photos.id AND files.media_id IS NOT NULL").
		Joins("LEFT JOIN cameras ON photos.camera_id = cameras.id").
		Joins("LEFT JOIN lenses ON photos.lens_id = lenses.id").
		Joins("LEFT JOIN places ON photos.place_id = places.id")

	// Accept the album UID as scope for backward compatibility.
	if rnd.IsUID(f.Album, entity.AlbumUID) {
		if txt.Empty(f.Scope) {
			f.Scope = f.Album
		}

		f.Album = ""
	}

	// Limit search results to a specific UID scope, e.g. when sharing.
	if txt.NotEmpty(f.Scope) {
		f.Scope = strings.ToLower(f.Scope)

		if idType, idPrefix := rnd.IdType(f.Scope); idType != rnd.TypeUID || idPrefix != entity.AlbumUID {
			return PhotoResults{}, 0, ErrInvalidId
		} else if a, err := entity.CachedAlbumByUID(f.Scope); err != nil || a.AlbumUID == "" {
			return PhotoResults{}, 0, ErrInvalidId
		} else if a.AlbumFilter == "" {
			s = s.Joins("JOIN photos_albums ON photos_albums.photo_uid = files.photo_uid").
				Where("photos_albums.hidden = 0 AND photos_albums.album_uid = ?", a.AlbumUID)
		} else if err = form.Unserialize(&f, a.AlbumFilter); err != nil {
			return PhotoResults{}, 0, ErrBadFilter
		} else {
			f.Filter = a.AlbumFilter
			s = s.Where("files.photo_uid NOT IN (SELECT photo_uid FROM photos_albums pa WHERE pa.hidden = 1 AND pa.album_uid = ?)", a.AlbumUID)
		}
	} else {
		f.Scope = ""
	}

	// Check session permissions and apply as needed.
	if sess != nil {
		user := sess.User()
		aclRole := user.AclRole()

		// Exclude private content.
		if acl.Resources.Deny(acl.ResourcePhotos, aclRole, acl.AccessPrivate) {
			f.Public = true
			f.Private = false
		}

		// Exclude archived content.
		if acl.Resources.Deny(acl.ResourcePhotos, aclRole, acl.ActionDelete) {
			f.Archived = false
			f.Review = false
		}

		// Exclude hidden files.
		if acl.Resources.Deny(acl.ResourceFiles, aclRole, acl.AccessAll) {
			f.Hidden = false
		}

		// Visitors and other restricted users can only access shared content.
		if f.Scope != "" && !sess.HasShare(f.Scope) && (sess.IsVisitor() || sess.NotRegistered()) ||
			f.Scope == "" && acl.Resources.Deny(acl.ResourcePhotos, aclRole, acl.ActionSearch) {
			event.AuditErr([]string{sess.IP(), "session %s", "%s %s as %s", "denied"}, sess.RefID, acl.ActionSearch.String(), string(acl.ResourcePhotos), aclRole)
			return PhotoResults{}, 0, ErrForbidden
		}

		// Limit results for external users.
		if f.Scope == "" && acl.Resources.DenyAll(acl.ResourcePhotos, aclRole, acl.Permissions{acl.AccessAll, acl.AccessLibrary}) {
			sharedAlbums := "photos.photo_uid IN (SELECT photo_uid FROM photos_albums WHERE hidden = 0 AND missing = 0 AND album_uid IN (?)) OR "

			if sess.IsVisitor() || sess.NotRegistered() {
				s = s.Where(sharedAlbums+"photos.published_at > ?", sess.SharedUIDs(), entity.TimeStamp())
			} else if basePath := user.GetBasePath(); basePath == "" {
				s = s.Where(sharedAlbums+"photos.created_by = ? OR photos.published_at > ?", sess.SharedUIDs(), user.UserUID, entity.TimeStamp())
			} else {
				s = s.Where(sharedAlbums+"photos.created_by = ? OR photos.published_at > ? OR photos.photo_path = ? OR photos.photo_path LIKE ?",
					sess.SharedUIDs(), user.UserUID, entity.TimeStamp(), basePath, basePath+"/%")
			}
		}
	}

	// Set sort order.
	switch f.Order {
	case sortby.Edited:
		s = s.Where("photos.edited_at IS NOT NULL").Order("photos.edited_at DESC, files.media_id")
	case sortby.Relevance:
		if f.Label != "" {
			s = s.Order("photos.photo_quality DESC, photos_labels.uncertainty ASC, files.time_index")
		} else {
			s = s.Order("photos.photo_quality DESC, files.time_index")
		}
	case sortby.Duration:
		s = s.Order("photos.photo_duration DESC, files.time_index")
	case sortby.Size:
		s = s.Order("files.file_size DESC, files.time_index")
	case sortby.Newest:
		s = s.Order("files.time_index")
	case sortby.Oldest:
		s = s.Order("files.photo_taken_at, files.media_id")
	case sortby.Similar:
		s = s.Where("files.file_diff > 0")
		s = s.Order("photos.photo_color, photos.cell_id, files.file_diff, files.time_index")
	case sortby.Name:
		s = s.Order("photos.photo_path, photos.photo_name, files.time_index")
	case sortby.Random:
		s = s.Order(sortby.RandomExpr(s.Dialect()))
	case sortby.Default, sortby.Imported, sortby.Added:
		s = s.Order("files.media_id")
	default:
		return PhotoResults{}, 0, ErrBadSortOrder
	}

	// Limit the result file types if hidden images/videos should not be found.
	if !f.Hidden {
		s = s.Where("files.file_type IN (?) OR files.media_type IN ('vector','video')", FileTypes)

		if f.Error {
			s = s.Where("files.file_error <> ''")
		} else {
			s = s.Where("files.file_error = ''")
		}
	}

	// Primary files only.
	if f.Primary {
		s = s.Where("files.file_primary = 1")
	}

	// Find specific UIDs only.
	if txt.NotEmpty(f.UID) {
		ids := SplitOr(strings.ToLower(f.UID))

		idType, prefix := rnd.ContainsType(ids)

		if idType == rnd.TypeUnknown {
			return PhotoResults{}, 0, fmt.Errorf("%s ids specified", idType)
		} else if idType.SHA() {
			s = s.Where("files.file_hash IN (?)", ids)
		} else if idType == rnd.TypeUID {
			switch prefix {
			case entity.PhotoUID:
				s = s.Where("photos.photo_uid IN (?)", ids)
			case entity.FileUID:
				s = s.Where("files.file_uid IN (?)", ids)
			default:
				return PhotoResults{}, 0, fmt.Errorf("invalid ids specified")
			}
		}

		// Find UIDs only to improve performance.
		if sess == nil && f.FindUidOnly() {
			if result := s.Scan(&results); result.Error != nil {
				return results, 0, result.Error
			}

			log.Debugf("photos: found %s for %s [%s]", english.Plural(len(results), "result", "results"), f.SerializeAll(), time.Since(start))

			if f.Merged {
				return results.Merge()
			}

			return results, len(results), nil
		}
	}

	// Find Unique Image ID (Exif), Document ID, or Instance ID (XMP).
	if txt.NotEmpty(f.ID) {
		for _, id := range SplitAnd(strings.ToLower(f.ID)) {
			if ids := SplitOr(id); len(ids) > 0 {
				s = s.Where("files.instance_id IN (?) OR photos.uuid IN (?)", ids, ids)
			}
		}
	}

	// Filter by label, label category and keywords.
	var categories []entity.Category
	var labels []entity.Label
	var labelIds []uint
	if txt.NotEmpty(f.Label) {
		if err := Db().Where(AnySlug("label_slug", f.Label, txt.Or)).Or(AnySlug("custom_slug", f.Label, txt.Or)).Find(&labels).Error; len(labels) == 0 || err != nil {
			log.Debugf("search: label %s not found", txt.LogParamLower(f.Label))
			return PhotoResults{}, 0, nil
		} else {
			for _, l := range labels {
				labelIds = append(labelIds, l.ID)

				Log("find categories", Db().Where("category_id = ?", l.ID).Find(&categories).Error)
				log.Debugf("search: label %s includes %d categories", txt.LogParamLower(l.LabelName), len(categories))

				for _, category := range categories {
					labelIds = append(labelIds, category.LabelID)
				}
			}

			s = s.Joins("JOIN photos_labels ON photos_labels.photo_id = files.photo_id AND photos_labels.uncertainty < 100 AND photos_labels.label_id IN (?)", labelIds).
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
		case terms["vectors"]:
			f.Query = strings.ReplaceAll(f.Query, "vectors", "")
			f.Vector = true
		case terms["vector"]:
			f.Query = strings.ReplaceAll(f.Query, "vector", "")
			f.Vector = true
		case terms["animated"]:
			f.Query = strings.ReplaceAll(f.Query, "animated", "")
			f.Animated = true
		case terms["gifs"]:
			f.Query = strings.ReplaceAll(f.Query, "gifs", "")
			f.Animated = true
		case terms["gif"]:
			f.Query = strings.ReplaceAll(f.Query, "gif", "")
			f.Animated = true
		case terms["live"]:
			f.Query = strings.ReplaceAll(f.Query, "live", "")
			f.Live = true
		case terms["raws"]:
			f.Query = strings.ReplaceAll(f.Query, "raws", "")
			f.Raw = true
		case terms["raw"]:
			f.Query = strings.ReplaceAll(f.Query, "raw", "")
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

	// Filter by location info.
	if txt.No(f.Geo) {
		s = s.Where("photos.cell_id = 'zz'")
	} else if txt.NotEmpty(f.Geo) {
		s = s.Where("photos.cell_id <> 'zz'")
	}

	// Filter by query string.
	if f.Query != "" {
		if err := Db().Where(AnySlug("custom_slug", f.Query, " ")).Find(&labels).Error; len(labels) == 0 || err != nil {
			log.Tracef("search: label %s not found, using fuzzy search", txt.LogParamLower(f.Query))

			for _, where := range LikeAnyKeyword("k.keyword", f.Query) {
				s = s.Where("files.photo_id IN (SELECT pk.photo_id FROM keywords k JOIN photos_keywords pk ON k.id = pk.keyword_id WHERE (?))", gorm.Expr(where))
			}
		} else {
			for _, l := range labels {
				labelIds = append(labelIds, l.ID)

				Db().Where("category_id = ?", l.ID).Find(&categories)

				log.Tracef("search: label %s includes %d categories", txt.LogParamLower(l.LabelName), len(categories))

				for _, category := range categories {
					labelIds = append(labelIds, category.LabelID)
				}
			}

			if wheres := LikeAnyKeyword("k.keyword", f.Query); len(wheres) > 0 {
				for _, where := range wheres {
					s = s.Where("files.photo_id IN (SELECT pk.photo_id FROM keywords k JOIN photos_keywords pk ON k.id = pk.keyword_id WHERE (?)) OR "+
						"files.photo_id IN (SELECT pl.photo_id FROM photos_labels pl WHERE pl.uncertainty < 100 AND pl.label_id IN (?))", gorm.Expr(where), labelIds)
				}
			} else {
				s = s.Where("files.photo_id IN (SELECT pl.photo_id FROM photos_labels pl WHERE pl.uncertainty < 100 AND pl.label_id IN (?))", labelIds)
			}
		}
	}

	// Search for one or more keywords.
	if txt.NotEmpty(f.Keywords) {
		for _, where := range LikeAnyWord("k.keyword", f.Keywords) {
			s = s.Where("files.photo_id IN (SELECT pk.photo_id FROM keywords k JOIN photos_keywords pk ON k.id = pk.keyword_id WHERE (?))", gorm.Expr(where))
		}
	}

	// Filter by number of faces.
	if f.Faces == "" {
		// Do nothing.
	} else if txt.IsUInt(f.Faces) {
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
	if f.Face == "" {
		// Do nothing.
	} else if len(f.Face) >= 32 {
		for _, f := range SplitAnd(strings.ToUpper(f.Face)) {
			s = s.Where(fmt.Sprintf("files.photo_id IN (SELECT photo_id FROM files f JOIN %s m ON f.file_uid = m.file_uid AND m.marker_invalid = 0 WHERE face_id IN (?))",
				entity.Marker{}.TableName()), SplitOr(f))
		}
	} else if txt.New(f.Face) {
		s = s.Where(fmt.Sprintf("files.photo_id IN (SELECT photo_id FROM files f JOIN %s m ON f.file_uid = m.file_uid AND m.marker_invalid = 0 AND m.marker_type = ? WHERE subj_uid IS NULL OR subj_uid = '')",
			entity.Marker{}.TableName()), entity.MarkerFace)
	} else if txt.No(f.Face) {
		s = s.Where(fmt.Sprintf("files.photo_id IN (SELECT photo_id FROM files f JOIN %s m ON f.file_uid = m.file_uid AND m.marker_invalid = 0 AND m.marker_type = ? WHERE face_id IS NULL OR face_id = '')",
			entity.Marker{}.TableName()), entity.MarkerFace)
	} else if txt.Yes(f.Face) {
		s = s.Where(fmt.Sprintf("files.photo_id IN (SELECT photo_id FROM files f JOIN %s m ON f.file_uid = m.file_uid AND m.marker_invalid = 0 AND m.marker_type = ? WHERE face_id IS NOT NULL AND face_id <> '')",
			entity.Marker{}.TableName()), entity.MarkerFace)
	} else if txt.IsUInt(f.Face) {
		s = s.Where("files.photo_id IN (SELECT photo_id FROM files f JOIN markers m ON f.file_uid = m.file_uid AND m.marker_invalid = 0 AND m.marker_type = ? JOIN faces ON faces.id = m.face_id WHERE m.face_id IS NOT NULL AND m.face_id <> '' AND faces.face_kind = ?)",
			entity.MarkerFace, txt.Int(f.Face))
	}

	// Filter for one or more subjects.
	if txt.NotEmpty(f.Subject) {
		for _, subj := range SplitAnd(strings.ToLower(f.Subject)) {
			if subjects := SplitOr(subj); rnd.ContainsUID(subjects, 'j') {
				s = s.Where(fmt.Sprintf("files.photo_id IN (SELECT photo_id FROM files f JOIN %s m ON f.file_uid = m.file_uid AND m.marker_invalid = 0 WHERE subj_uid IN (?))",
					entity.Marker{}.TableName()), subjects)
			} else {
				s = s.Where(fmt.Sprintf("files.photo_id IN (SELECT photo_id FROM files f JOIN %s m ON f.file_uid = m.file_uid AND m.marker_invalid = 0 JOIN %s s ON s.subj_uid = m.subj_uid WHERE (?))",
					entity.Marker{}.TableName(), entity.Subject{}.TableName()), gorm.Expr(AnySlug("s.subj_slug", subj, txt.Or)))
			}
		}
	} else if txt.NotEmpty(f.Subjects) {
		for _, where := range LikeAllNames(Cols{"subj_name", "subj_alias"}, f.Subjects) {
			s = s.Where(fmt.Sprintf("files.photo_id IN (SELECT photo_id FROM files f JOIN %s m ON f.file_uid = m.file_uid AND m.marker_invalid = 0 JOIN %s s ON s.subj_uid = m.subj_uid WHERE (?))",
				entity.Marker{}.TableName(), entity.Subject{}.TableName()), gorm.Expr(where))
		}
	}

	// Filter by status.
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

	// Filter by camera id or name.
	if txt.IsPosInt(f.Camera) {
		s = s.Where("photos.camera_id = ?", txt.UInt(f.Camera))
	} else if txt.NotEmpty(f.Camera) {
		v := strings.Trim(f.Camera, "*%") + "%"
		s = s.Where("cameras.camera_name LIKE ? OR cameras.camera_model LIKE ? OR cameras.camera_slug LIKE ?", v, v, v)
	}

	// Filter by lens id or name.
	if txt.IsPosInt(f.Lens) {
		s = s.Where("photos.lens_id = ?", txt.UInt(f.Lens))
	} else if txt.NotEmpty(f.Lens) {
		v := strings.Trim(f.Lens, "*%") + "%"
		s = s.Where("lenses.lens_name LIKE ? OR lenses.lens_model LIKE ? OR lenses.lens_slug LIKE ?", v, v, v)
	}

	// Filter by year.
	if f.Year != "" {
		s = s.Where(AnyInt("photos.photo_year", f.Year, txt.Or, entity.UnknownYear, txt.YearMax))
	}

	// Filter by month.
	if f.Month != "" {
		s = s.Where(AnyInt("photos.photo_month", f.Month, txt.Or, entity.UnknownMonth, txt.MonthMax))
	}

	// Filter by day.
	if f.Day != "" {
		s = s.Where(AnyInt("photos.photo_day", f.Day, txt.Or, entity.UnknownDay, txt.DayMax))
	}

	// Filter by main color.
	if f.Color != "" {
		s = s.Where("files.file_main_color IN (?)", SplitOr(strings.ToLower(f.Color)))
	}

	// Find favorites only.
	if f.Favorite {
		s = s.Where("photos.photo_favorite = 1")
	}

	// Find scans only.
	if f.Scan {
		s = s.Where("photos.photo_scan = 1")
	}

	// Find panoramas only.
	if f.Panorama {
		s = s.Where("photos.photo_panorama = 1")
	}

	// Find portrait/landscape/square pictures only.
	if f.Portrait {
		s = s.Where("files.file_portrait = 1")
	} else if f.Landscape {
		s = s.Where("files.file_aspect_ratio > 1.25")
	} else if f.Square {
		s = s.Where("files.file_aspect_ratio = 1")
	}

	if f.Stackable {
		s = s.Where("photos.photo_stack > -1")
	} else if f.Unstacked {
		s = s.Where("photos.photo_stack = -1")
	}

	// Filter by location country.
	if txt.NotEmpty(f.Country) {
		s = s.Where("photos.photo_country IN (?)", SplitOr(strings.ToLower(f.Country)))
	}

	// Filter by location state.
	if txt.NotEmpty(f.State) {
		s = s.Where("places.place_state IN (?)", SplitOr(f.State))
	}

	// Filter by location city.
	if txt.NotEmpty(f.City) {
		s = s.Where("places.place_city IN (?)", SplitOr(f.City))
	}

	// Filter by location category.
	if txt.NotEmpty(f.Category) {
		s = s.Joins("JOIN cells ON photos.cell_id = cells.id").
			Where("cells.cell_category IN (?)", SplitOr(strings.ToLower(f.Category)))
	}

	// Filter by media type.
	if txt.NotEmpty(f.Type) {
		s = s.Where("photos.photo_type IN (?)", SplitOr(strings.ToLower(f.Type)))
	} else if f.Video {
		s = s.Where("photos.photo_type = ?", entity.MediaVideo)
	} else if f.Vector {
		s = s.Where("photos.photo_type = ?", entity.MediaVector)
	} else if f.Animated {
		s = s.Where("photos.photo_type = ?", entity.MediaAnimated)
	} else if f.Raw {
		s = s.Where("photos.photo_type = ?", entity.MediaRaw)
	} else if f.Live {
		s = s.Where("photos.photo_type = ?", entity.MediaLive)
	} else if f.Photo {
		s = s.Where("photos.photo_type IN ('image','live','animated','vector','raw')")
	}

	// Filter by storage path.
	if txt.NotEmpty(f.Path) {
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
	if txt.NotEmpty(f.Name) {
		where, names := OrLike("photos.photo_name", f.Name)

		// Omit file path and known extensions.
		for i := range names {
			names[i] = fs.StripKnownExt(path.Base(names[i].(string)))
		}

		s = s.Where(where, names...)
	}

	// Filter by complete file names.
	if txt.NotEmpty(f.Filename) {
		where, values := OrLike("files.file_name", f.Filename)
		s = s.Where(where, values...)
	}

	// Filter by original file name.
	if txt.NotEmpty(f.Original) {
		where, values := OrLike("photos.original_name", f.Original)
		s = s.Where(where, values...)
	}

	// Filter by title.
	if txt.NotEmpty(f.Title) {
		where, values := OrLike("photos.photo_title", f.Title)
		s = s.Where(where, values...)
	}

	// Filter by hash.
	if txt.NotEmpty(f.Hash) {
		s = s.Where("files.file_hash IN (?)", SplitOr(strings.ToLower(f.Hash)))
	}

	// Filter by chroma.
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

	// Filter by approx distance to co-ordinates:
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

	// Find stacks only.
	if f.Stack {
		s = s.Where("photos.id IN (SELECT a.photo_id FROM files a JOIN files b ON a.id != b.id AND a.photo_id = b.photo_id AND a.file_type = b.file_type WHERE a.file_type='jpg')")
	}

	// Find photos in albums or not in an album, unless search results are limited to a scope.
	if f.Scope == "" {
		if f.Unsorted {
			s = s.Where("photos.photo_uid NOT IN (SELECT photo_uid FROM photos_albums pa JOIN albums a ON a.album_uid = pa.album_uid WHERE pa.hidden = 0 AND a.deleted_at IS NULL)")
		} else if txt.NotEmpty(f.Album) {
			v := strings.Trim(f.Album, "*%") + "%"
			s = s.Where("photos.photo_uid IN (SELECT pa.photo_uid FROM photos_albums pa JOIN albums a ON a.album_uid = pa.album_uid AND pa.hidden = 0 WHERE (a.album_title LIKE ? OR a.album_slug LIKE ?))", v, v)
		} else if txt.NotEmpty(f.Albums) {
			for _, where := range LikeAnyWord("a.album_title", f.Albums) {
				s = s.Where("photos.photo_uid IN (SELECT pa.photo_uid FROM photos_albums pa JOIN albums a ON a.album_uid = pa.album_uid AND pa.hidden = 0 WHERE (?))", gorm.Expr(where))
			}
		}
	}

	// Limit offset and count.
	if f.Count > 0 && f.Count <= MaxResults {
		s = s.Limit(f.Count).Offset(f.Offset)
	} else {
		s = s.Limit(MaxResults).Offset(f.Offset)
	}

	// Query database.
	if err = s.Scan(&results).Error; err != nil {
		return results, 0, err
	}

	// Log number of results.
	log.Debugf("photos: found %s for %s [%s]", english.Plural(len(results), "result", "results"), f.SerializeAll(), time.Since(start))

	// Merge files that belong to the same photo.
	if f.Merged {
		// Return merged files.
		return results.Merge()
	}

	// Return unmerged files.
	return results, len(results), nil
}
