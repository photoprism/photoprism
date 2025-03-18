package search

import (
	"strings"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/sortby"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Albums finds AlbumResults based on the search form without checking rights or permissions.
func Albums(frm form.SearchAlbums) (results AlbumResults, err error) {
	return UserAlbums(frm, nil)
}

// UserAlbums finds AlbumResults based on the search form and user session.
func UserAlbums(frm form.SearchAlbums, sess *entity.Session) (results AlbumResults, err error) {
	start := time.Now()

	if err = frm.ParseQueryString(); err != nil {
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
		switch frm.Type {
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
		if acl.Rules.DenyAll(aclResource, aclRole, acl.Permissions{acl.AccessAll, acl.AccessLibrary, acl.AccessShared, acl.AccessOwn}) {
			return AlbumResults{}, ErrForbidden
		}

		// Limit results by UID, owner and path.
		if sess.IsVisitor() || sess.NotRegistered() {
			s = s.Where("albums.album_uid IN (?) OR albums.published_at > ?", sess.SharedUIDs(), entity.Now())
		} else if acl.Rules.DenyAll(aclResource, aclRole, acl.Permissions{acl.AccessAll, acl.AccessLibrary}) {
			s = s.Where("albums.album_uid IN (?) OR albums.created_by = ? OR albums.published_at > ?", sess.SharedUIDs(), user.UserUID, entity.Now())
		}

		// Exclude private content?
		if acl.Rules.Deny(acl.ResourcePhotos, aclRole, acl.AccessPrivate) || acl.Rules.Deny(aclResource, aclRole, acl.AccessPrivate) {
			frm.Public = true
			frm.Private = false
		}
	}

	// Set sort order.
	switch frm.Order {
	case sortby.Count:
		s = s.Order("photo_count DESC, albums.album_title, albums.album_uid DESC")
	case sortby.Moment, sortby.Newest:
		if frm.Type == entity.AlbumManual || frm.Type == entity.AlbumState {
			s = s.Order("albums.album_uid DESC")
		} else if frm.Type == entity.AlbumMoment {
			s = s.Order("has_year, albums.album_year DESC, albums.album_month DESC, albums.album_day DESC, albums.album_title, albums.album_uid DESC")
		} else {
			s = s.Order("albums.album_year DESC, albums.album_month DESC, albums.album_day DESC, albums.album_title, albums.album_uid DESC")
		}
	case sortby.Oldest:
		if frm.Type == entity.AlbumManual || frm.Type == entity.AlbumState {
			s = s.Order("albums.album_uid ASC")
		} else if frm.Type == entity.AlbumMoment {
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
		if frm.Type == entity.AlbumFolder {
			s = s.Order("albums.album_favorite DESC, albums.album_path ASC, albums.album_uid DESC")
		} else if frm.Type == entity.AlbumMonth {
			s = s.Order("albums.album_favorite DESC, albums.album_year DESC, albums.album_month DESC, albums.album_day DESC, albums.album_title, albums.album_uid DESC")
		} else {
			s = s.Order("albums.album_favorite DESC, albums.album_title ASC, albums.album_uid DESC")
		}
	case sortby.Name:
		if frm.Type == entity.AlbumFolder {
			s = s.Order("albums.album_path ASC, albums.album_uid DESC")
		} else {
			s = s.Order("albums.album_title ASC, albums.album_uid DESC")
		}
	case sortby.NameReverse:
		if frm.Type == entity.AlbumFolder {
			s = s.Order("albums.album_path DESC, albums.album_uid DESC")
		} else {
			s = s.Order("albums.album_title DESC, albums.album_uid DESC")
		}
	default:
		s = s.Order("albums.album_favorite DESC, albums.album_title ASC, albums.album_uid DESC")
	}

	// Find specific UIDs only?
	if txt.NotEmpty(frm.UID) {
		ids := SplitOr(strings.ToLower(frm.UID))

		if rnd.ContainsUID(ids, entity.AlbumUID) {
			s = s.Where("albums.album_uid IN (?)", ids)
		}
	}

	// Filter by title or path?
	if txt.NotEmpty(frm.Query) {
		q := "%" + strings.Trim(frm.Query, " *%") + "%"

		if frm.Type == entity.AlbumFolder {
			s = s.Where("albums.album_title LIKE ? OR albums.album_location LIKE ? OR albums.album_path LIKE ?", q, q, q)
		} else {
			s = s.Where("albums.album_title LIKE ? OR albums.album_location LIKE ?", q, q)
		}
	}

	// Albums with public pictures only?
	if frm.Public {
		s = s.Where("albums.album_private = 0 AND (albums.album_type <> 'folder' OR albums.album_path IN (SELECT photo_path FROM photos WHERE photo_private = 0 AND photo_quality > -1 AND deleted_at IS NULL))")
	} else {
		s = s.Where("albums.album_type <> 'folder' OR albums.album_path IN (SELECT photo_path FROM photos WHERE photo_quality > -1 AND deleted_at IS NULL)")
	}

	if txt.NotEmpty(frm.Type) {
		s = s.Where("albums.album_type IN (?)", strings.Split(frm.Type, txt.Or))
	}

	if txt.NotEmpty(frm.Category) {
		s = s.Where("albums.album_category IN (?)", strings.Split(frm.Category, txt.Or))
	}

	if txt.NotEmpty(frm.Location) {
		s = s.Where("albums.album_location IN (?)", strings.Split(frm.Location, txt.Or))
	}

	if txt.NotEmpty(frm.Country) {
		s = s.Where("albums.album_country IN (?)", strings.Split(frm.Country, txt.Or))
	}

	// Favorites only?
	if frm.Favorite {
		s = s.Where("albums.album_favorite = 1")
	}

	// Filter by year?
	if txt.NotEmpty(frm.Year) {
		// Filter by the pictures included if it is a manually managed album, as these do not have an explicit
		// year assigned to them, unlike calendar albums and moments for example.
		if frm.Type == entity.AlbumManual {
			s = s.Where("? OR albums.album_uid IN (SELECT DISTINCT pay.album_uid FROM photos_albums pay "+
				"JOIN photos py ON pay.photo_uid = py.photo_uid WHERE py.photo_year IN (?) AND pay.hidden = 0 AND pay.missing = 0)",
				gorm.Expr(AnyInt("albums.album_year", frm.Year, txt.Or, entity.UnknownYear, txt.YearMax)), strings.Split(frm.Year, txt.Or))
		} else {
			s = s.Where(AnyInt("albums.album_year", frm.Year, txt.Or, entity.UnknownYear, txt.YearMax))
		}
	}

	// Filter by month?
	if txt.NotEmpty(frm.Month) {
		s = s.Where(AnyInt("albums.album_month", frm.Month, txt.Or, entity.UnknownMonth, txt.MonthMax))
	}

	// Filter by day?
	if txt.NotEmpty(frm.Day) {
		s = s.Where(AnyInt("albums.album_day", frm.Day, txt.Or, entity.UnknownDay, txt.DayMax))
	}

	// Limit result count.
	if frm.Count > 0 && frm.Count <= MaxResults {
		s = s.Limit(frm.Count).Offset(frm.Offset)
	} else {
		s = s.Limit(MaxResults).Offset(frm.Offset)
	}

	// Query database.
	if result := s.Scan(&results); result.Error != nil {
		return results, result.Error
	}

	// Log number of results.
	log.Debugf("albums: found %s [%s]", english.Plural(len(results), "result", "results"), time.Since(start))

	return results, nil
}
