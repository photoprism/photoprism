package search

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/enum"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/geo"
	"github.com/photoprism/photoprism/pkg/geo/pluscode"
	"github.com/photoprism/photoprism/pkg/geo/s2"
	"github.com/photoprism/photoprism/pkg/log/status"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

// GeoCols specifies the UserPhotosGeo result column names.
var GeoCols = SelectString(GeoResult{}, []string{"*"})

// PhotosGeo finds GeoResults based on the search form without checking rights or permissions.
func PhotosGeo(frm form.SearchPhotosGeo) (results GeoResults, err error) {
	return UserPhotosGeo(frm, nil)
}

// UserPhotosGeo finds photos based on the search form and user session then returns them as GeoResults.
func UserPhotosGeo(frm form.SearchPhotosGeo, sess *entity.Session) (results GeoResults, err error) {
	start := time.Now()

	// Parse query string and filter.
	if err = frm.ParseQueryString(); err != nil {
		log.Debugf("search: %s", err)
		return GeoResults{}, ErrBadRequest
	}

	// Find photos near another?
	if txt.NotEmpty(frm.Near) {
		photo := Photo{}

		// Find a nearby picture using the UID or return an empty result otherwise.
		if err = Db().First(&photo, "photo_uid = ?", frm.Near).Error; err != nil {
			log.Debugf("search: %s (find nearby)", err)
			return GeoResults{}, ErrNotFound
		}

		// Set the S2 Cell ID to search for.
		frm.S2 = photo.CellID

		// Set the search distance if unspecified.
		if frm.Dist <= 0 {
			frm.Dist = geo.DefaultDist
		}
	}

	// Set default search distance.
	if frm.Dist <= 0 {
		frm.Dist = geo.DefaultDist
	} else if frm.Dist > geo.DistLimit {
		frm.Dist = geo.DistLimit
	}

	// Specify table names and joins.
	s := UnscopedDb().Table(entity.Photo{}.TableName()).Select(GeoCols).
		Joins(`JOIN files ON files.photo_id = photos.id AND files.file_primary = 1 AND files.media_id IS NOT NULL`).
		Joins("LEFT JOIN places ON photos.place_id = places.id").
		Where("photos.deleted_at IS NULL").
		Where("photos.photo_lat <> 0")

	// Accept the album UID as scope for backward compatibility.
	if rnd.IsUID(frm.Album, entity.AlbumUID) {
		if txt.Empty(frm.Scope) {
			frm.Scope = frm.Album
		}

		frm.Album = ""
	}

	var album entity.Album
	var albumErr error

	// Limit search results to a specific UID scope, e.g. when sharing.
	if txt.NotEmpty(frm.Scope) {
		frm.Scope = strings.ToLower(frm.Scope)

		if idType, idPrefix := rnd.IdType(frm.Scope); idType != rnd.TypeUID || idPrefix != entity.AlbumUID {
			return GeoResults{}, ErrInvalidId
		} else if album, albumErr = entity.CachedAlbumByUID(frm.Scope); albumErr != nil || album.AlbumUID == "" {
			return GeoResults{}, ErrInvalidId
		} else if album.AlbumFilter == "" {
			s = s.Joins("JOIN photos_albums ON photos_albums.photo_uid = files.photo_uid").
				Where("photos_albums.hidden = 0 AND photos_albums.album_uid = ?", album.AlbumUID)
		} else if formErr := form.Unserialize(&frm, album.AlbumFilter); formErr != nil {
			log.Debugf("search: %s (%s)", clean.Error(formErr), clean.Log(album.AlbumFilter))
			return GeoResults{}, ErrBadFilter
		} else {
			frm.Filter = album.AlbumFilter
			s = s.Where("files.photo_uid NOT IN (SELECT photo_uid FROM photos_albums pa WHERE pa.hidden = 1 AND pa.album_uid = ?)", album.AlbumUID)
		}

		// Enforce search distance range (km).
		if frm.Dist <= 0 {
			frm.Dist = geo.DefaultDist
		} else if frm.Dist > geo.ScopeDistLimit {
			frm.Dist = geo.ScopeDistLimit
		}
	} else {
		frm.Scope = ""
	}

	// Check session permissions and apply as needed.
	if sess != nil {
		user := sess.GetUser()
		aclRole := user.AclRole()

		// Exclude private content.
		if acl.Rules.Deny(acl.ResourcePlaces, aclRole, acl.AccessPrivate) {
			frm.Public = true
			frm.Private = false
		}

		// Exclude archived content.
		if acl.Rules.Deny(acl.ResourcePlaces, aclRole, acl.ActionDelete) {
			frm.Archived = false
			frm.Review = false
		}

		// Visitors and other restricted users can only access shared content.
		if frm.Scope != "" && album.CreatedBy != user.UserUID && !sess.HasShare(frm.Scope) && (sess.GetUser().HasSharedAccessOnly(acl.ResourcePlaces) || sess.NotRegistered()) ||
			frm.Scope == "" && acl.Rules.Deny(acl.ResourcePlaces, aclRole, acl.ActionSearch) {
			event.AuditErr([]string{sess.IP(), "session %s", "%s %s as %s", status.Denied}, sess.RefID, acl.ActionSearch.String(), string(acl.ResourcePlaces), aclRole)
			return GeoResults{}, ErrForbidden
		}

		// Limit results for external users.
		if frm.Scope == "" && acl.Rules.DenyAll(acl.ResourcePlaces, aclRole, acl.Permissions{acl.AccessAll, acl.AccessLibrary}) {
			sharedAlbums := "photos.photo_uid IN (SELECT photo_uid FROM photos_albums WHERE hidden = 0 AND missing = 0 AND album_uid IN (?)) OR "

			if sess.IsVisitor() || sess.NotRegistered() {
				s = s.Where(sharedAlbums+"photos.published_at > ?", sess.SharedUIDs(), entity.Now())
			} else if basePath := user.GetBasePath(); basePath == "" {
				s = s.Where(sharedAlbums+"photos.created_by = ? OR photos.published_at > ?", sess.SharedUIDs(), user.UserUID, entity.Now())
			} else {
				s = s.Where(sharedAlbums+"photos.created_by = ? OR photos.published_at > ? OR photos.photo_path = ? OR photos.photo_path LIKE ?",
					sess.SharedUIDs(), user.UserUID, entity.Now(), basePath, basePath+"/%")
			}
		}
	}

	// Set sort order.
	if frm.Near == "" {
		s = s.Order("taken_at, photos.photo_uid")
	} else {
		// Sort by distance to UID.
		s = s.Order(gorm.Expr("(photos.photo_uid = ?) DESC, ABS(? - photos.photo_lat)+ABS(? - photos.photo_lng)", frm.Near, frm.Lat, frm.Lng))
	}

	// Find specific UIDs only.
	if txt.NotEmpty(frm.UID) {
		ids := SplitOr(strings.ToLower(frm.UID))
		idType, prefix := rnd.ContainsType(ids)

		if idType == rnd.TypeUnknown {
			return GeoResults{}, fmt.Errorf("%s ids specified", idType)
		} else if idType.SHA() {
			s = s.Where("files.file_hash IN (?)", ids)
		} else if idType == rnd.TypeUID {
			switch prefix {
			case entity.PhotoUID:
				s = s.Where("photos.photo_uid IN (?)", ids)
			case entity.FileUID:
				s = s.Where("files.file_uid IN (?)", ids)
			default:
				return GeoResults{}, fmt.Errorf("invalid ids specified")
			}
		}

		// Find UIDs only to improve performance.
		if sess == nil && frm.FindUidOnly() {
			// Fetch results.
			if result := s.Scan(&results); result.Error != nil {
				return results, result.Error
			}

			log.Debugf("places: found %s for %s [%s]", english.Plural(len(results), "result", "results"), frm.SerializeAll(), time.Since(start))

			return results, nil
		}
	}

	// Find Unique Image ID (Exif), Document ID, or Instance ID (XMP).
	if txt.NotEmpty(frm.ID) {
		for _, id := range SplitAnd(strings.ToLower(frm.ID)) {
			if ids := SplitOr(id); len(ids) > 0 {
				s = s.Where("files.instance_id IN (?) OR photos.uuid IN (?)", ids, ids)
			}
		}
	}

	// Filter by label, label category and keywords.
	if txt.NotEmpty(frm.Label) {
		var categories []entity.Category
		var labels []entity.Label
		var labelIds []uint

		if labelErr := Db().Where(AnySlug("label_slug", frm.Label, txt.Or)).Or(AnySlug("custom_slug", frm.Label, txt.Or)).Find(&labels).Error; len(labels) == 0 || labelErr != nil {
			log.Debugf("search: label %s not found", txt.LogParamLower(frm.Label))
			return GeoResults{}, nil
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
	if terms := txt.SearchTerms(frm.Query); frm.Query != "" && len(terms) == 0 {
		if frm.Title == "" && frm.Caption == "" && frm.Description == "" {
			frm.Description = fmt.Sprintf("%s*", strings.Trim(frm.Query, "%*"))
			frm.Query = ""
		}
	} else if len(terms) > 0 {
		switch {
		case terms["faces"]:
			frm.Query = strings.ReplaceAll(frm.Query, "faces", "")
			frm.Faces = "true"
		case terms["people"]:
			frm.Query = strings.ReplaceAll(frm.Query, "people", "")
			frm.Faces = "true"
		case terms["videos"]:
			frm.Query = strings.ReplaceAll(frm.Query, "videos", "")
			frm.Video = true
		case terms["video"]:
			frm.Query = strings.ReplaceAll(frm.Query, "video", "")
			frm.Video = true
		case terms["audio"]:
			frm.Query = strings.ReplaceAll(frm.Query, "audio", "")
			frm.Audio = true
		case terms["sounds"]:
			frm.Query = strings.ReplaceAll(frm.Query, "sounds", "")
			frm.Audio = true
		case terms["documents"]:
			frm.Query = strings.ReplaceAll(frm.Query, "documents", "")
			frm.Document = true
		case terms["document"]:
			frm.Query = strings.ReplaceAll(frm.Query, "document", "")
			frm.Document = true
		case terms["vectors"]:
			frm.Query = strings.ReplaceAll(frm.Query, "vectors", "")
			frm.Vector = true
		case terms["vector"]:
			frm.Query = strings.ReplaceAll(frm.Query, "vector", "")
			frm.Vector = true
		case terms["animated"]:
			frm.Query = strings.ReplaceAll(frm.Query, "animated", "")
			frm.Animated = true
		case terms["gifs"]:
			frm.Query = strings.ReplaceAll(frm.Query, "gifs", "")
			frm.Animated = true
		case terms["gif"]:
			frm.Query = strings.ReplaceAll(frm.Query, "gif", "")
			frm.Animated = true
		case terms["live"]:
			frm.Query = strings.ReplaceAll(frm.Query, "live", "")
			frm.Live = true
		case terms["raws"]:
			frm.Query = strings.ReplaceAll(frm.Query, "raws", "")
			frm.Raw = true
		case terms["raw"]:
			frm.Query = strings.ReplaceAll(frm.Query, "raw", "")
			frm.Raw = true
		case terms["favorites"]:
			frm.Query = strings.ReplaceAll(frm.Query, "favorites", "")
			frm.Favorite = "true"
		case terms["panoramas"]:
			frm.Query = strings.ReplaceAll(frm.Query, "panoramas", "")
			frm.Panorama = true
		case terms["scans"]:
			frm.Query = strings.ReplaceAll(frm.Query, "scans", "")
			frm.Scan = "true"
		case terms["monochrome"]:
			frm.Query = strings.ReplaceAll(frm.Query, "monochrome", "")
			frm.Mono = true
		case terms["mono"]:
			frm.Query = strings.ReplaceAll(frm.Query, "mono", "")
			frm.Mono = true
		}
	}

	// Filter by label, label category, and keywords.
	if frm.Query != "" {
		var categories []entity.Category
		var labels []entity.Label
		var labelIds []uint

		if labelsErr := Db().Where(AnySlug("custom_slug", frm.Query, " ")).Find(&labels).Error; len(labels) == 0 || labelsErr != nil {
			log.Tracef("search: label %s not found, using fuzzy search", txt.LogParamLower(frm.Query))

			for _, where := range LikeAnyKeyword("k.keyword", frm.Query) {
				s = s.Where("photos.id IN (SELECT pk.photo_id FROM keywords k JOIN photos_keywords pk ON k.id = pk.keyword_id WHERE (?))", gorm.Expr(where))
			}
		} else {
			for _, l := range labels {
				labelIds = append(labelIds, l.ID)

				Log("find categories", Db().Where("category_id = ?", l.ID).Find(&categories).Error)
				log.Tracef("search: label %s includes %d categories", txt.LogParamLower(l.LabelName), len(categories))

				for _, category := range categories {
					labelIds = append(labelIds, category.LabelID)
				}
			}

			if wheres := LikeAnyKeyword("k.keyword", frm.Query); len(wheres) > 0 {
				for _, where := range wheres {
					s = s.Where("photos.id IN (SELECT pk.photo_id FROM keywords k JOIN photos_keywords pk ON k.id = pk.keyword_id WHERE (?)) OR "+
						"photos.id IN (SELECT pl.photo_id FROM photos_labels pl WHERE pl.uncertainty < 100 AND pl.label_id IN (?))", gorm.Expr(where), labelIds)
				}
			} else {
				s = s.Where("photos.id IN (SELECT pl.photo_id FROM photos_labels pl WHERE pl.uncertainty < 100 AND pl.label_id IN (?))", labelIds)
			}
		}
	}

	// Search for one or more keywords.
	if frm.Keywords != "" {
		for _, where := range LikeAnyWord("k.keyword", frm.Keywords) {
			s = s.Where("photos.id IN (SELECT pk.photo_id FROM keywords k JOIN photos_keywords pk ON k.id = pk.keyword_id WHERE (?))", gorm.Expr(where))
		}
	}

	// Filter by number of faces.
	if frm.Faces == "" {
		// Do nothing.
	} else if txt.IsUInt(frm.Faces) {
		s = s.Where("photos.photo_faces >= ?", txt.Int(frm.Faces))
	} else if txt.New(frm.Faces) && frm.Face == "" {
		frm.Face = frm.Faces
		frm.Faces = ""
	} else if txt.Yes(frm.Faces) {
		s = s.Where("photos.photo_faces > 0")
	} else if txt.No(frm.Faces) {
		s = s.Where("photos.photo_faces = 0")
	}

	// Filter for specific face clusters? Example: PLJ7A3G4MBGZJRMVDIUCBLC46IAP4N7O
	if frm.Face == "" {
		// Do nothing.
	} else if len(frm.Face) >= 32 {
		for _, f := range SplitAnd(strings.ToUpper(frm.Face)) {
			s = s.Where(fmt.Sprintf("photos.id IN (SELECT photo_id FROM files f JOIN %s m ON f.file_uid = m.file_uid AND m.marker_invalid = 0 WHERE face_id IN (?))",
				entity.Marker{}.TableName()), SplitOr(f))
		}
	} else if txt.New(frm.Face) {
		s = s.Where(fmt.Sprintf("photos.id IN (SELECT photo_id FROM files f JOIN %s m ON f.file_uid = m.file_uid AND m.marker_invalid = 0 AND m.marker_type = ? WHERE subj_uid IS NULL OR subj_uid = '')",
			entity.Marker{}.TableName()), entity.MarkerFace)
	} else if txt.No(frm.Face) {
		s = s.Where(fmt.Sprintf("photos.id IN (SELECT photo_id FROM files f JOIN %s m ON f.file_uid = m.file_uid AND m.marker_invalid = 0 AND m.marker_type = ? WHERE face_id IS NULL OR face_id = '')",
			entity.Marker{}.TableName()), entity.MarkerFace)
	} else if txt.Yes(frm.Face) {
		s = s.Where(fmt.Sprintf("photos.id IN (SELECT photo_id FROM files f JOIN %s m ON f.file_uid = m.file_uid AND m.marker_invalid = 0 AND m.marker_type = ? WHERE face_id IS NOT NULL AND face_id <> '')",
			entity.Marker{}.TableName()), entity.MarkerFace)
	} else if txt.IsUInt(frm.Face) {
		s = s.Where("files.photo_id IN (SELECT photo_id FROM files f JOIN markers m ON f.file_uid = m.file_uid AND m.marker_invalid = 0 AND m.marker_type = ? JOIN faces ON faces.id = m.face_id WHERE m.face_id IS NOT NULL AND m.face_id <> '' AND faces.face_kind = ?)",
			entity.MarkerFace, txt.Int(frm.Face))
	}

	// Filter for one or more subjects.
	if frm.Subject != "" {
		for _, subj := range SplitAnd(strings.ToLower(frm.Subject)) {
			if subjects := SplitOr(subj); rnd.ContainsUID(subjects, 'j') {
				s = s.Where(fmt.Sprintf("photos.id IN (SELECT photo_id FROM files f JOIN %s m ON f.file_uid = m.file_uid AND m.marker_invalid = 0 WHERE subj_uid IN (?))",
					entity.Marker{}.TableName()), subjects)
			} else {
				s = s.Where(fmt.Sprintf("photos.id IN (SELECT photo_id FROM files f JOIN %s m ON f.file_uid = m.file_uid AND m.marker_invalid = 0 JOIN %s s ON s.subj_uid = m.subj_uid WHERE (?))",
					entity.Marker{}.TableName(), entity.Subject{}.TableName()), gorm.Expr(AnySlug("s.subj_slug", subj, txt.Or)))
			}
		}
	} else if frm.Subjects != "" {
		for _, where := range LikeAllNames(Cols{"subj_name", "subj_alias"}, frm.Subjects) {
			s = s.Where(fmt.Sprintf("photos.id IN (SELECT photo_id FROM files f JOIN %s m ON f.file_uid = m.file_uid AND m.marker_invalid = 0 JOIN %s s ON s.subj_uid = m.subj_uid WHERE (?))",
				entity.Marker{}.TableName(), entity.Subject{}.TableName()), gorm.Expr(where))
		}
	}

	// Find photos in albums or not in an album, unless search results are limited to a scope.
	if frm.Scope == "" {
		if frm.Unsorted {
			s = s.Where("photos.photo_uid NOT IN (SELECT photo_uid FROM photos_albums pa JOIN albums a ON a.album_uid = pa.album_uid WHERE pa.hidden = 0 AND a.deleted_at IS NULL)")
		} else if txt.NotEmpty(frm.Album) {
			v := strings.Trim(frm.Album, "*%") + "%"
			s = s.Where("photos.photo_uid IN (SELECT pa.photo_uid FROM photos_albums pa JOIN albums a ON a.album_uid = pa.album_uid AND pa.hidden = 0 WHERE (a.album_title LIKE ? OR a.album_slug LIKE ?))", v, v)
		} else if txt.NotEmpty(frm.Albums) {
			for _, where := range LikeAnyWord("a.album_title", frm.Albums) {
				s = s.Where("photos.photo_uid IN (SELECT pa.photo_uid FROM photos_albums pa JOIN albums a ON a.album_uid = pa.album_uid AND pa.hidden = 0 WHERE (?))", gorm.Expr(where))
			}
		}
	}

	// Filter by camera.
	if frm.Camera > 0 {
		s = s.Where("photos.camera_id = ?", frm.Camera)
	}

	// Filter by camera lens.
	if frm.Lens > 0 {
		s = s.Where("photos.lens_id = ?", frm.Lens)
	}

	// Filter by ISO Number (light sensitivity) range.
	if rangeStart, rangeEnd, rangeErr := txt.IntRange(frm.Iso, 0, 10000000); rangeErr == nil {
		s = s.Where("photos.photo_iso >= ? AND photos.photo_iso <= ?", rangeStart, rangeEnd)
	}

	// Filter by Focal Length (35mm equivalent) range.
	if rangeStart, rangeEnd, rangeErr := txt.IntRange(frm.Mm, 0, 10000000); rangeErr == nil {
		s = s.Where("photos.photo_focal_length >= ? AND photos.photo_focal_length <= ?", rangeStart, rangeEnd)
	}

	// Filter by Aperture (f-number) range.
	if rangeStart, rangeEnd, rangeErr := txt.FloatRange(frm.F, 0, 10000000); rangeErr == nil {
		s = s.Where("photos.photo_f_number >= ? AND photos.photo_f_number <= ?", rangeStart-0.01, rangeEnd+0.01)
	}

	// Filter by year.
	if frm.Year != "" {
		s = s.Where(AnyInt("photos.photo_year", frm.Year, txt.Or, entity.UnknownYear, txt.YearMax))
	}

	// Filter by month.
	if frm.Month != "" {
		s = s.Where(AnyInt("photos.photo_month", frm.Month, txt.Or, entity.UnknownMonth, txt.MonthMax))
	}

	// Filter by day.
	if frm.Day != "" {
		s = s.Where(AnyInt("photos.photo_day", frm.Day, txt.Or, entity.UnknownDay, txt.DayMax))
	}

	// Filter by Resolution in Megapixels (MP).
	if rangeStart, rangeEnd, rangeErr := txt.IntRange(frm.Mp, 0, 32000); rangeErr == nil {
		s = s.Where("photos.photo_resolution >= ? AND photos.photo_resolution <= ?", rangeStart, rangeEnd)
	}

	// Find panoramic pictures only.
	if frm.Panorama {
		s = s.Where("photos.photo_panorama = 1")
	}

	// Find portrait/landscape/square pictures only.
	if frm.Portrait {
		s = s.Where("files.file_portrait = 1")
	} else if frm.Landscape {
		s = s.Where("files.file_aspect_ratio > 1.25")
	} else if frm.Square {
		s = s.Where("files.file_aspect_ratio = 1")
	}

	// Filter by file main color.
	if txt.NotEmpty(frm.Color) {
		s = s.Where("files.file_main_color IN (?)", SplitOr(strings.ToLower(frm.Color)))
	}

	// Filter by file codec.
	if txt.NotEmpty(frm.Codec) {
		s = s.Where("files.file_codec IN (?)", SplitOr(strings.ToLower(frm.Codec)))
	}

	// Filter by chroma.
	if frm.Mono {
		s = s.Where("files.file_chroma = 0")
	} else if frm.Chroma > 9 {
		s = s.Where("files.file_chroma > ?", frm.Chroma)
	} else if frm.Chroma > 0 {
		s = s.Where("files.file_chroma > 0 AND files.file_chroma <= ?", frm.Chroma)
	}

	// Filter by favorite flag.
	if txt.No(frm.Favorite) {
		s = s.Where("photos.photo_favorite = 0")
	} else if txt.NotEmpty(frm.Favorite) {
		s = s.Where("photos.photo_favorite = 1")
	}

	// Filter by scan flag.
	if txt.No(frm.Scan) {
		s = s.Where("photos.photo_scan = 0")
	} else if txt.NotEmpty(frm.Scan) {
		s = s.Where("photos.photo_scan = 1")
	}

	// Filter by location country.
	if frm.Country != "" {
		s = s.Where("photos.photo_country IN (?)", SplitOr(strings.ToLower(frm.Country)))
	}

	// Filter by location state.
	if txt.NotEmpty(frm.State) {
		s = s.Where("places.place_state IN (?)", SplitOr(frm.State))
	}

	// Filter by location city.
	if txt.NotEmpty(frm.City) {
		s = s.Where("places.place_city IN (?)", SplitOr(frm.City))
	}

	// Filter by location category.
	if txt.NotEmpty(frm.Category) {
		s = s.Joins("JOIN cells ON photos.cell_id = cells.id").
			Where("cells.cell_category IN (?)", SplitOr(strings.ToLower(frm.Category)))
	}

	// Filter by media type.
	if txt.NotEmpty(frm.Type) {
		s = s.Where("photos.photo_type IN (?)", SplitOr(strings.ToLower(frm.Type)))
	} else if frm.Document {
		s = s.Where("photos.photo_type = ?", media.Document)
	} else if frm.Image {
		s = s.Where("photos.photo_type = ?", media.Image)
	} else if frm.Raw {
		s = s.Where("photos.photo_type = ?", media.Raw)
	} else if frm.Vector {
		s = s.Where("photos.photo_type = ?", media.Vector)
	} else if frm.Animated {
		s = s.Where("photos.photo_type = ?", media.Animated)
	} else if frm.Audio {
		s = s.Where("photos.photo_type = ?", media.Audio)
	} else if frm.Video {
		s = s.Where("photos.photo_type = ?", media.Video)
	} else if frm.Live {
		s = s.Where("photos.photo_type = ?", media.Live)
	} else if frm.Media {
		s = s.Where("photos.photo_type IN ('live','video','audio','animated')")
	} else if frm.Photo {
		s = s.Where("photos.photo_type IN ('image','raw','live')")
	}

	// Filter by storage path.
	if frm.Path != "" {
		p := frm.Path

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
	if frm.Name != "" {
		where, names := OrLike("photos.photo_name", frm.Name)

		// Omit file path and known extensions.
		for i := range names {
			names[i] = fs.StripKnownExt(path.Base(names[i].(string)))
		}

		s = s.Where(where, names...)
	}

	// Filter by title.
	if txt.NotEmpty(frm.Title) {
		if frm.Title == enum.False {
			s = s.Where("photos.photo_title = ''")
		} else if frm.Title == enum.True {
			s = s.Where("photos.photo_title <> ''")
		} else {
			where, values := OrLike("photos.photo_title", frm.Title)
			s = s.Where(where, values...)
		}
	}

	// Filter by caption.
	if txt.NotEmpty(frm.Caption) {
		if frm.Caption == enum.False {
			s = s.Where("photos.photo_caption = ''")
		} else if frm.Caption == enum.True {
			s = s.Where("photos.photo_caption <> ''")
		} else {
			where, values := OrLike("photos.photo_caption", frm.Caption)
			s = s.Where(where, values...)
		}
	}

	// Filter by description.
	if txt.NotEmpty(frm.Description) {
		if frm.Description == enum.False {
			s = s.Where("photos.photo_title = '' AND photos.photo_caption = ''")
		} else if frm.Description == enum.True {
			s = s.Where("photos.photo_title <> '' OR photos.photo_caption <> ''")
		} else {
			where, values := OrLikeCols([]string{"photos.photo_title", "photos.photo_caption"}, frm.Description)
			s = s.Where(where, values...)
		}
	}

	// Filter by status.
	if frm.Archived {
		s = s.Where("photos.photo_quality > -1")
		s = s.Where("photos.deleted_at IS NOT NULL")
	} else {
		s = s.Where("photos.deleted_at IS NULL")

		if frm.Review {
			s = s.Where("photos.photo_quality < 3")
		} else if frm.Quality != 0 && frm.Private == false {
			s = s.Where("photos.photo_quality >= ?", frm.Quality)
		}
	}

	// Filter private pictures.
	if frm.Public {
		s = s.Where("photos.photo_private = 0")
	} else if frm.Private {
		s = s.Where("photos.photo_private = 1")
	}

	// Filter by location code.
	if txt.NotEmpty(frm.S2) {
		// S2 Cell ID.
		s2Min, s2Max := s2.PrefixedRange(frm.S2, s2.Level(frm.Dist))
		s = s.Where("photos.cell_id BETWEEN ? AND ?", s2Min, s2Max)
	} else if txt.NotEmpty(frm.Olc) {
		// Open Location Code (OLC).
		s2Min, s2Max := s2.PrefixedRange(pluscode.S2(frm.Olc), s2.Level(frm.Dist))
		s = s.Where("photos.cell_id BETWEEN ? AND ?", s2Min, s2Max)
	}

	// Filter by GPS Bounds (Lat N, Lng E, Lat S, Lng W).
	if latN, lngE, latS, lngW, boundsErr := clean.GPSBounds(frm.Latlng); boundsErr == nil {
		s = s.Where("photos.photo_lat BETWEEN ? AND ?", latS, latN)
		s = s.Where("photos.photo_lng BETWEEN ? AND ?", lngW, lngE)
	}

	// Filter by GPS Latitude range (from +90 to -90 degrees).
	if latN, latS, latErr := clean.GPSLatRange(frm.Lat, frm.Dist); latErr == nil {
		s = s.Where("photos.photo_lat BETWEEN ? AND ?", latS, latN)
	}

	// Filter by GPS Longitude range (from -180 to +180 degrees).
	if lngE, lngW, lngErr := clean.GPSLngRange(frm.Lat, frm.Lng, frm.Dist); lngErr == nil {
		s = s.Where("photos.photo_lng BETWEEN ? AND ?", lngW, lngE)
	}

	// Filter by GPS Altitude (m) range.
	if rangeStart, rangeEnd, rangeErr := txt.IntRange(frm.Alt, -6378000, 1000000000); rangeErr == nil {
		s = s.Where("photos.photo_altitude BETWEEN ? AND ?", rangeStart, rangeEnd)
	}

	// Find pictures added at or after this time (UTC).
	if !frm.Added.IsZero() {
		s = s.Where("photos.created_at >= ?", frm.Added.UTC().Format("2006-01-02 15:04:05"))
	}

	// Find pictures updated at or after this time (UTC).
	if !frm.Updated.IsZero() {
		s = s.Where("photos.updated_at >= ?", frm.Updated.UTC().Format("2006-01-02 15:04:05"))
	}

	// Find pictures edited at or after this time (UTC).
	if !frm.Edited.IsZero() {
		s = s.Where("photos.edited_at >= ?", frm.Edited.UTC().Format("2006-01-02 15:04:05"))
	}

	// Find pictures taken on the specified date.
	if !frm.Taken.IsZero() {
		s = s.Where("DATE(photos.taken_at) = DATE(?)", frm.Taken.UTC().Format("2006-01-02"))
	}

	// Finds pictures taken on or before this date.
	if !frm.Before.IsZero() {
		s = s.Where("photos.taken_at <= ?", frm.Before.UTC().Format("2006-01-02"))
	}

	// Finds pictures taken on or after this date.
	if !frm.After.IsZero() {
		s = s.Where("photos.taken_at >= ?", frm.After.UTC().Format("2006-01-02"))
	}

	// Limit offset and count.
	if frm.Count > 0 {
		s = s.Limit(frm.Count).Offset(frm.Offset)
	} else {
		s = s.Limit(1000000).Offset(frm.Offset)
	}

	// Fetch results.
	if result := s.Scan(&results); result.Error != nil {
		return results, result.Error
	}

	log.Debugf("places: found %s for %s [%s]", english.Plural(len(results), "result", "results"), frm.SerializeAll(), time.Since(start))

	return results, nil
}
