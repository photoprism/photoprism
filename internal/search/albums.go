package search

import (
	"strings"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/sortby"
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
		Select("albums.*, cp.photo_count, cl.link_count, CASE WHEN albums.album_year = 0 THEN 0 ELSE 1 END AS has_year, CASE WHEN albums.album_location = '' THEN 1 ELSE 0 END AS no_location").
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
		case entity.AlbumManual:
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
			if basePath := user.GetBasePath(); basePath == "" {
				s = s.Where("albums.album_uid IN (?) OR albums.created_by = ? OR albums.published_at > ?", sess.SharedUIDs(), user.UserUID, entity.TimeStamp())
			} else {
				s = s.Where("albums.album_uid IN (?) OR albums.created_by = ? OR albums.published_at > ? OR albums.album_type = ? AND (albums.album_path = ? OR albums.album_path LIKE ?)",
					sess.SharedUIDs(), user.UserUID, entity.TimeStamp(), entity.AlbumFolder, basePath, basePath+"/%")
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
	case sortby.Count:
		s = s.Order("photo_count DESC, albums.album_title, albums.album_uid DESC")
	case sortby.Moment, sortby.Newest:
		if f.Type == entity.AlbumManual || f.Type == entity.AlbumState {
			s = s.Order("albums.album_uid DESC")
		} else if f.Type == entity.AlbumMoment {
			s = s.Order("has_year, albums.album_year DESC, albums.album_month DESC, albums.album_day DESC, albums.album_title, albums.album_uid DESC")
		} else {
			s = s.Order("albums.album_year DESC, albums.album_month DESC, albums.album_day DESC, albums.album_title, albums.album_uid DESC")
		}
	case sortby.Oldest:
		if f.Type == entity.AlbumManual || f.Type == entity.AlbumState {
			s = s.Order("albums.album_uid ASC")
		} else if f.Type == entity.AlbumMoment {
			s = s.Order("has_year, albums.album_year ASC, albums.album_month ASC, albums.album_day ASC, albums.album_title, albums.album_uid ASC")
		} else {
			s = s.Order("albums.album_year ASC, albums.album_month ASC, albums.album_day ASC, albums.album_title, albums.album_uid ASC")
		}
	case sortby.Added:
		s = s.Order("albums.album_uid DESC")
	case sortby.Edited:
		s = s.Order("albums.updated_at DESC, albums.album_uid DESC")
	case sortby.Place:
		s = s.Order("no_location, albums.album_location, has_year, albums.album_year DESC, albums.album_month ASC, albums.album_day ASC, albums.album_title, albums.album_uid DESC")
	case sortby.Path:
		s = s.Order("albums.album_path, albums.album_uid DESC")
	case sortby.Category:
		s = s.Order("albums.album_category, albums.album_title, albums.album_uid DESC")
	case sortby.Slug:
		s = s.Order("albums.album_slug ASC, albums.album_uid DESC")
	case sortby.Favorites:
		if f.Type == entity.AlbumFolder {
			s = s.Order("albums.album_favorite DESC, albums.album_path ASC, albums.album_uid DESC")
		} else if f.Type == entity.AlbumMonth {
			s = s.Order("albums.album_favorite DESC, albums.album_year DESC, albums.album_month DESC, albums.album_day DESC, albums.album_title, albums.album_uid DESC")
		} else {
			s = s.Order("albums.album_favorite DESC, albums.album_title ASC, albums.album_uid DESC")
		}
	case sortby.Name:
		if f.Type == entity.AlbumFolder {
			s = s.Order("albums.album_path ASC, albums.album_uid DESC")
		} else {
			s = s.Order("albums.album_title ASC, albums.album_uid DESC")
		}
	case sortby.NameReverse:
		if f.Type == entity.AlbumFolder {
			s = s.Order("albums.album_path DESC, albums.album_uid DESC")
		} else {
			s = s.Order("albums.album_title DESC, albums.album_uid DESC")
		}
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
			searchQuery := strings.Trim(strings.ReplaceAll(f.Query, "\\", "/"), "/")
			for _, where := range LikeAllNames(Cols{"albums.album_title", "albums.album_location", "albums.album_path"}, searchQuery) {
				s = s.Where(where)
			}
		}
	}

	// Albums with public pictures only?
	if f.Public {
		s = s.Where("albums.album_private = 0 AND (albums.album_type <> 'folder' OR albums.album_path IN (SELECT photo_path FROM photos WHERE photo_private = 0 AND photo_quality > -1 AND deleted_at IS NULL))")
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

	// Favorites only?
	if f.Favorite {
		s = s.Where("albums.album_favorite = 1")
	}

	// Filter by year?
	if txt.NotEmpty(f.Year) {
		// Filter by the pictures included if it is a manually managed album, as these do not have an explicit
		// year assigned to them, unlike calendar albums and moments for example.
		if f.Type == entity.AlbumManual {
			s = s.Where("? OR albums.album_uid IN (SELECT DISTINCT pay.album_uid FROM photos_albums pay "+
				"JOIN photos py ON pay.photo_uid = py.photo_uid WHERE py.photo_year IN (?) AND pay.hidden = 0 AND pay.missing = 0)",
				gorm.Expr(AnyInt("albums.album_year", f.Year, txt.Or, entity.UnknownYear, txt.YearMax)), strings.Split(f.Year, txt.Or))
		} else {
			s = s.Where(AnyInt("albums.album_year", f.Year, txt.Or, entity.UnknownYear, txt.YearMax))
		}
	}

	// Filter by month?
	if txt.NotEmpty(f.Month) {
		s = s.Where(AnyInt("albums.album_month", f.Month, txt.Or, entity.UnknownMonth, txt.MonthMax))
	}

	// Filter by day?
	if txt.NotEmpty(f.Day) {
		s = s.Where(AnyInt("albums.album_day", f.Day, txt.Or, entity.UnknownDay, txt.DayMax))
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
