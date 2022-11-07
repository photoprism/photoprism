package search

import (
	"strings"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/photoprism/photoprism/pkg/rnd"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Albums finds AlbumResults based on the search form without checking rights or permissions.
func Albums(f form.SearchAlbums) (results AlbumResults, err error) {
	return UserAlbums(f, nil)
}

// UserAlbums finds AlbumResults based on the search form and user session.
func UserAlbums(f form.SearchAlbums, sess *entity.Session) (results AlbumResults, err error) {
	start := time.Now()

	if err = f.ParseQueryString(); err != nil {
		log.Debugf("albums: %s", err)
		return AlbumResults{}, err
	}

	// Base query.
	s := UnscopedDb().Table("albums").
		Select("albums.*, cp.photo_count, cl.link_count, CASE WHEN albums.album_year = 0 THEN 0 ELSE 1 END AS has_year").
		Joins("LEFT JOIN (SELECT album_uid, count(photo_uid) AS photo_count FROM photos_albums WHERE hidden = 0 AND missing = 0 GROUP BY album_uid) AS cp ON cp.album_uid = albums.album_uid").
		Joins("LEFT JOIN (SELECT share_uid, count(share_uid) AS link_count FROM links GROUP BY share_uid) AS cl ON cl.share_uid = albums.album_uid").
		Where("albums.deleted_at IS NULL")

	// Check session permissions and apply as needed.
	if sess != nil {
		user := sess.User()
		aclRole := user.AclRole()

		// Determine resource to check.
		var aclResource acl.Resource
		switch f.Type {
		case entity.AlbumDefault:
			aclResource = acl.ResourceAlbums
		case entity.AlbumFolder:
			aclResource = acl.ResourceFolders
		case entity.AlbumMoment:
			aclResource = acl.ResourceMoments
		case entity.AlbumMonth:
			aclResource = acl.ResourceCalendar
		case entity.AlbumState:
			aclResource = acl.ResourcePlaces
		}

		// Check user permissions.
		if acl.Resources.DenyAll(aclResource, aclRole, acl.Permissions{acl.AccessAll, acl.AccessLibrary, acl.AccessShared, acl.AccessOwn}) {
			return AlbumResults{}, ErrForbidden
		}

		// Limit results by UID, owner and path.
		if sess.IsVisitor() || sess.NotRegistered() {
			s = s.Where("albums.album_uid IN (?) OR albums.published_at > ?", sess.SharedUIDs(), entity.TimeStamp())
		} else if acl.Resources.DenyAll(aclResource, aclRole, acl.Permissions{acl.AccessAll, acl.AccessLibrary}) {
			if user.BasePath == "" {
				s = s.Where("albums.album_uid IN (?) OR albums.created_by = ? OR albums.published_at > ?", sess.SharedUIDs(), user.UserUID, entity.TimeStamp())
			} else {
				s = s.Where("albums.album_uid IN (?) OR albums.created_by = ? OR albums.published_at > ? OR albums.album_type = ? AND (albums.album_path = ? OR albums.album_path LIKE ?)",
					sess.SharedUIDs(), user.UserUID, entity.TimeStamp(), entity.AlbumFolder, user.BasePath, user.BasePath+"/%")
			}
		}

		// Exclude private content?
		if acl.Resources.Deny(acl.ResourcePhotos, aclRole, acl.AccessPrivate) || acl.Resources.Deny(aclResource, aclRole, acl.AccessPrivate) {
			f.Public = true
			f.Private = false
		}
	}

	// Set sort order.
	switch f.Order {
	case entity.SortOrderCount:
		s = s.Order("photo_count DESC, albums.album_title, albums.album_uid DESC")
	case entity.SortOrderRelevance:
		s = s.Order("albums.album_favorite DESC, albums.updated_at DESC, albums.album_uid DESC")
	case entity.SortOrderNewest:
		s = s.Order("albums.album_favorite DESC, albums.album_year DESC, albums.album_month DESC, albums.album_day DESC, albums.album_title, albums.album_uid DESC")
	case entity.SortOrderOldest:
		s = s.Order("albums.album_favorite DESC, albums.album_year ASC, albums.album_month ASC, albums.album_day ASC, albums.album_title, albums.album_uid ASC")
	case entity.SortOrderAdded:
		s = s.Order("albums.album_uid DESC")
	case entity.SortOrderMoment:
		s = s.Order("albums.album_favorite DESC, has_year, albums.album_year DESC, albums.album_month DESC, albums.album_title ASC, albums.album_uid DESC")
	case entity.SortOrderPlace:
		s = s.Order("albums.album_favorite DESC, albums.album_location, albums.album_title, albums.album_year DESC, albums.album_month ASC, albums.album_day ASC, albums.album_uid DESC")
	case entity.SortOrderPath:
		s = s.Order("albums.album_path, albums.album_uid DESC")
	case entity.SortOrderCategory:
		s = s.Order("albums.album_category, albums.album_title, albums.album_uid DESC")
	case entity.SortOrderSlug:
		s = s.Order("albums.album_favorite DESC, albums.album_slug ASC, albums.album_uid DESC")
	case entity.SortOrderName:
		s = s.Order("albums.album_favorite DESC, albums.album_title ASC, albums.album_uid DESC")
	default:
		s = s.Order("albums.album_favorite DESC, albums.album_title ASC, albums.album_uid DESC")
	}

	// Find specific UIDs only?
	if txt.NotEmpty(f.UID) {
		ids := SplitOr(strings.ToLower(f.UID))

		if rnd.ContainsUID(ids, entity.AlbumUID) {
			s = s.Where("albums.album_uid IN (?)", ids)
		}
	}

	// Filter by title or path?
	if txt.NotEmpty(f.Query) {
		if f.Type != entity.AlbumFolder {
			likeString := "%" + f.Query + "%"
			s = s.Where("albums.album_title LIKE ? OR albums.album_location LIKE ?", likeString, likeString)
		} else {
			f.Order = entity.SortOrderPath

			p := f.Query

			if strings.HasPrefix(p, "/") {
				p = p[1:]
			}

			if strings.HasSuffix(p, "/") {
				s = s.Where("albums.album_path = ?", p[:len(p)-1])
			} else {
				p = p + "*"

				where, values := OrLike("albums.album_path", p)

				if w, v := OrLike("albums.album_title", p); len(v) > 0 {
					where = where + " OR " + w
					values = append(values, v...)
				}

				s = s.Where(where, values...)
			}
		}
	}

	// Albums with public pictures only?
	if f.Public {
		s = s.Where("albums.album_type <> 'folder' OR albums.album_path IN (SELECT photo_path FROM photos WHERE photo_private = 0 AND photo_quality > -1 AND deleted_at IS NULL)")
	} else {
		s = s.Where("albums.album_type <> 'folder' OR albums.album_path IN (SELECT photo_path FROM photos WHERE photo_quality > -1 AND deleted_at IS NULL)")
	}

	if txt.NotEmpty(f.Type) {
		s = s.Where("albums.album_type IN (?)", strings.Split(f.Type, txt.Or))
	}

	if txt.NotEmpty(f.Category) {
		s = s.Where("albums.album_category IN (?)", strings.Split(f.Category, txt.Or))
	}

	if txt.NotEmpty(f.Location) {
		s = s.Where("albums.album_location IN (?)", strings.Split(f.Location, txt.Or))
	}

	if txt.NotEmpty(f.Country) {
		s = s.Where("albums.album_country IN (?)", strings.Split(f.Country, txt.Or))
	}

	if f.Favorite {
		s = s.Where("albums.album_favorite = 1")
	}

	if (f.Year > 0 && f.Year <= txt.YearMax) || f.Year == entity.UnknownYear {
		s = s.Where("albums.album_year = ?", f.Year)
	}

	if (f.Month >= txt.MonthMin && f.Month <= txt.MonthMax) || f.Month == entity.UnknownMonth {
		s = s.Where("albums.album_month = ?", f.Month)
	}

	if (f.Day >= txt.DayMin && f.Month <= txt.DayMax) || f.Day == entity.UnknownDay {
		s = s.Where("albums.album_day = ?", f.Day)
	}

	// Limit result count.
	if f.Count > 0 && f.Count <= MaxResults {
		s = s.Limit(f.Count).Offset(f.Offset)
	} else {
		s = s.Limit(MaxResults).Offset(f.Offset)
	}

	// Query database.
	if result := s.Scan(&results); result.Error != nil {
		return results, result.Error
	}

	// Log number of results.
	log.Debugf("albums: found %s [%s]", english.Plural(len(results), "result", "results"), time.Since(start))

	return results, nil
}
