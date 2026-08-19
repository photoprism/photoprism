package query

import (
	"fmt"
	"time"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/mutex"

	"github.com/photoprism/photoprism/internal/entity/sortby"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Albums returns a slice of albums.
func Albums(offset, limit int) (results entity.Albums, err error) {
	err = UnscopedDb().Table("albums").Select("*").Order("album_type, album_uid").Offset(offset).Limit(limit).Find(&results).Error
	return results, err
}

// AlbumsByUID returns albums by UID.
func AlbumsByUID(albumUIDs []string, includeDeleted bool) (results entity.Albums, err error) {
	if includeDeleted {
		err = UnscopedDb().Where("album_uid IN (?)", albumUIDs).Find(&results).Error
	} else {
		err = UnscopedDb().Where("album_uid IN (?) AND deleted_at IS NULL", albumUIDs).Find(&results).Error
	}
	return results, err
}

// AlbumsByType returns albums by AlbumType.
func AlbumsByType(albumType string, includeDeleted bool) (results entity.Albums, err error) {
	if includeDeleted {
		err = UnscopedDb().Where("album_type = ?", albumType).Find(&results).Error
	} else {
		err = Db().Where("album_type = ?", albumType).Find(&results).Error
	}
	return results, err
}

// AlbumByUID returns a Album based on the UID.
func AlbumByUID(albumUID string) (album entity.Album, err error) {
	if rnd.InvalidUID(albumUID, entity.AlbumUID) {
		return album, fmt.Errorf("invalid album uid")
	}

	return entity.CachedAlbumByUID(albumUID)
}

// AlbumHasThumb tests if a usable cover file has been assigned to the album with the specified UID.
// Requires the file to still resolve, so a stale hash falls back to a cover query.
func AlbumHasThumb(albumUID string) bool {
	if rnd.InvalidUID(albumUID, entity.AlbumUID) {
		return false
	}

	var result []string

	if err := Db().Model(entity.Album{}).
		Joins("JOIN files ON files.file_hash = albums.thumb AND files.file_missing = 0 AND files.file_error = '' AND files.deleted_at IS NULL").
		Where("albums.album_uid = ?", albumUID).
		Limit(1).
		Pluck("albums.thumb", &result).Error; err != nil {
		return false
	} else if len(result) == 0 {
		return false
	}

	return result[0] != ""
}

// AlbumCoverByUID returns an album cover file based on the uid.
func AlbumCoverByUID(uid string, public bool) (file entity.File, err error) {
	if rnd.InvalidUID(uid, entity.AlbumUID) {
		return file, fmt.Errorf("invalid album uid")
	}

	a := entity.Album{}

	// Find album.
	if a, err = AlbumByUID(uid); err != nil {
		return file, err
	} else if !a.HasID() {
		return file, fmt.Errorf("album uid %s is invalid", clean.Log(uid))
	} else if a.AlbumType != entity.AlbumManual { // TODO: Optimize
		if a.AlbumFilter == "" {
			return file, fmt.Errorf("smart album %s has no filter specified", a.AlbumUID)
		}

		f := form.SearchPhotos{Album: a.AlbumUID, Filter: a.AlbumFilter, Order: sortby.Relevance, Count: 1, Offset: 0, Merged: false}

		if err = f.ParseQueryString(); err != nil {
			return file, err
		}

		// Public private only?
		if !public {
			f.Public = false
		}

		if photos, _, err := search.Photos(f); err != nil {
			return file, err
		} else if len(photos) > 0 {
			for _, photo := range photos {
				if err := Db().Where("photo_uid = ? AND file_primary = 1", photo.PhotoUID).First(&file).Error; err != nil {
					return file, err
				} else {
					return file, nil
				}
			}
		}

		// Automatically hide empty months.
		switch a.AlbumType {
		case entity.AlbumMonth, entity.AlbumState:
			if err = a.Delete(); err != nil {
				log.Errorf("%s: %s (hide)", a.AlbumType, err)
			} else {
				log.Infof("%s: %s hidden", a.AlbumType, clean.Log(a.AlbumTitle))
			}
		}

		// Return without album cover.
		return file, fmt.Errorf("no cover found")
	}

	// Build query.
	stmt := Db().Where("files.file_primary = 1 AND files.file_missing = 0 AND files.file_type IN (?) AND files.deleted_at IS NULL", media.PreviewExpr).
		Joins("JOIN albums a ON a.album_uid = ?", uid).
		Joins("JOIN photos_albums pa ON pa.album_uid = a.album_uid AND pa.photo_uid = files.photo_uid AND pa.hidden = 0 AND pa.missing = 0").
		Joins("JOIN photos ON photos.id = files.photo_id AND photos.deleted_at IS NULL")

	// Public pictures only?
	if public {
		stmt = stmt.Where("photos.photo_private = 0")
	}

	// Find first picture.
	if err = stmt.Order("photos.photo_quality DESC, photos.taken_at DESC").
		First(&file).Error; err != nil {
		return file, err
	}

	return file, nil
}

// UpdateAlbumDates updates the year, month and day of the album based on the indexed photo metadata.
func UpdateAlbumDates() (updated int, err error) {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	switch DbDialect() {
	case dsn.DriverMySQL:
		result := UnscopedDb().Exec(`UPDATE albums INNER JOIN (
             SELECT photo_path, MAX(taken_at_local) AS taken_max
			 FROM photos WHERE taken_src = 'meta' AND photos.photo_quality >= 3 AND photos.deleted_at IS NULL
			 GROUP BY photo_path
	    ) AS p ON albums.album_path = p.photo_path
		SET albums.album_year = YEAR(taken_max), albums.album_month = MONTH(taken_max), albums.album_day = DAY(taken_max)
		WHERE albums.album_type = 'folder' AND albums.album_path IS NOT NULL AND p.taken_max IS NOT NULL
		AND (album_year = 0 OR album_month = 0 OR album_day = 0
			OR DATE(p.taken_max) <> COALESCE(STR_TO_DATE(CONCAT(album_year, '-', album_month, '-', album_day), '%Y-%c-%e'), DATE('1000-01-01'))
		)`)
		return int(result.RowsAffected), result.Error
	case dsn.DriverSQLite3:
		// SQLite has potential locking issues if the update is done on all albums at once.
		var albums entity.Albums
		if albums, err = AlbumsByType(entity.AlbumFolder, true); err != nil {
			log.Errorf("album: get folders (%v)", err)
			return updated, err
		}
		for _, album := range albums {
			albumDate := time.Date(1000, 1, 1, 0, 0, 0, 0, time.UTC)
			// Clip the allowable range to what MariaDB supports (most restrictive of supported DBMS') so a database migration in the future won't break
			if album.AlbumYear > 999 && album.AlbumYear < 10000 && album.AlbumMonth > 0 && album.AlbumMonth < 13 && album.AlbumDay > 0 && album.AlbumDay < 32 {
				albumDate = time.Date(album.AlbumYear, time.Month(album.AlbumMonth), album.AlbumDay, 0, 0, 0, 0, time.UTC)
			}
			result := UnscopedDb().Exec(`UPDATE albums
			SET album_year = strftime('%Y', taken_max), album_month = strftime('%m', taken_max), album_day = strftime('%d', taken_max)
			FROM (SELECT photo_path, MAX(DATE(taken_at_local)) AS taken_max
	 			FROM photos WHERE taken_src = 'meta' AND photos.photo_quality >= 3 AND photos.deleted_at IS NULL
				AND photo_path = ?
	 			GROUP BY photo_path
			) AS p
			WHERE albums.album_path = p.photo_path AND albums.album_type = 'folder' AND albums.album_path IS NOT NULL AND date(p.taken_max) IS NOT NULL AND DATE(p.taken_max) <> DATE(?)`, album.AlbumPath, albumDate)
			if err = result.Error; err != nil {
				log.Errorf("album: set album dates on %s (%v)", album.AlbumTitle, err)
				return updated, err
			} else {
				updated += int(result.RowsAffected)
			}
		}
		return updated, err
	default:
		return updated, err
	}
}

// UpdateMissingAlbumEntries sets a flag for missing photo album entries.
func UpdateMissingAlbumEntries() error {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	switch DbDialect() {
	default:
		return UnscopedDb().Exec(`UPDATE photos_albums SET missing = 1
            WHERE photo_uid IN (SELECT photo_uid FROM photos WHERE deleted_at IS NOT NULL OR photo_quality < 0)
            OR photo_uid IN (SELECT pa.photo_uid FROM photos_albums pa LEFT JOIN photos p ON pa.photo_uid = p.photo_uid WHERE p.photo_uid IS NULL)`).Error
	}
}

// AlbumEntryFound removes the missing flag from album entries.
func AlbumEntryFound(uid string) error {
	if rnd.InvalidUID(uid, entity.PhotoUID) {
		return fmt.Errorf("invalid photo uid")
	}

	switch DbDialect() {
	default:
		return UnscopedDb().Exec(`UPDATE photos_albums SET missing = 0 WHERE photo_uid = ?`, uid).Error
	}
}

// AlbumsPhotoUIDs returns up to 100000 photo UIDs that belong to the specified albums.
func AlbumsPhotoUIDs(albums []string, includeDefault, includePrivate bool) (photos []string, err error) {
	for _, albumUid := range albums {
		if rnd.InvalidUID(albumUid, entity.AlbumUID) {
			// Should never happen.
			log.Debugf("query: album uid %s is invalid", clean.Log(albumUid))
			continue
		}

		a, err := AlbumByUID(albumUid)

		if err != nil {
			log.Warnf("query: album %s not found (%s)", albumUid, err.Error())
			continue
		}

		if a.IsDefault() && !includeDefault || !a.HasID() {
			continue
		} else if !a.IsDefault() && a.AlbumFilter == "" {
			// Should never happen.
			log.Debugf("query: smart album %s has empty filter", clean.Log(a.AlbumUID))
			continue
		}

		frm := form.SearchPhotos{
			Album:    a.AlbumUID,
			Filter:   a.AlbumFilter,
			Count:    100000,
			Offset:   0,
			Public:   !includePrivate,
			Hidden:   false,
			Archived: false,
			Quality:  1,
		}

		if err := frm.ParseQueryString(); err != nil {
			return photos, err
		}

		res, count, err := search.PhotoIds(frm)

		if err != nil {
			return photos, err
		} else if count == 0 {
			continue
		}

		ids := make([]string, 0, count)

		for _, r := range res {
			ids = append(ids, r.PhotoUID)
		}

		if len(ids) > 0 {
			photos = append(photos, ids...)
		}
	}

	return photos, nil
}
