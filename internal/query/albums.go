package query

import (
	"fmt"
	"strings"
	"time"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/photoprism/photoprism/pkg/txt"
)

// AlbumResult contains found albums
type AlbumResult struct {
	ID               uint      `json:"-"`
	AlbumUID         string    `json:"UID"`
	CoverUID         string    `json:"CoverUID"`
	FolderUID        string    `json:"FolderUID"`
	AlbumSlug        string    `json:"Slug"`
	AlbumType        string    `json:"Type"`
	AlbumTitle       string    `json:"Title"`
	AlbumLocation    string    `json:"Location"`
	AlbumCategory    string    `json:"Category"`
	AlbumCaption     string    `json:"Caption"`
	AlbumDescription string    `json:"Description"`
	AlbumNotes       string    `json:"Notes"`
	AlbumFilter      string    `json:"Filter"`
	AlbumOrder       string    `json:"Order"`
	AlbumTemplate    string    `json:"Template"`
	AlbumPath        string    `json:"Path"`
	AlbumCountry     string    `json:"Country"`
	AlbumYear        int       `json:"Year"`
	AlbumMonth       int       `json:"Month"`
	AlbumDay         int       `json:"Day"`
	AlbumFavorite    bool      `json:"Favorite"`
	AlbumPrivate     bool      `json:"Private"`
	PhotoCount       int       `json:"PhotoCount"`
	LinkCount        int       `json:"LinkCount"`
	CreatedAt        time.Time `json:"CreatedAt"`
	UpdatedAt        time.Time `json:"UpdatedAt"`
	DeletedAt        time.Time `json:"DeletedAt,omitempty"`
}

type AlbumResults []AlbumResult

// AlbumByUID returns a Album based on the UID.
func AlbumByUID(albumUID string) (album entity.Album, err error) {
	if err := Db().Where("album_uid = ?", albumUID).First(&album).Error; err != nil {
		return album, err
	}

	return album, nil
}

// AlbumCoverByUID returns a album preview file based on the uid.
func AlbumCoverByUID(uid string) (file entity.File, err error) {
	a := entity.Album{}

	if err := Db().Where("album_uid = ?", uid).First(&a).Error; err != nil {
		return file, err
	} else if a.AlbumType != entity.AlbumDefault { // TODO: Optimize
		f := form.PhotoSearch{Album: a.AlbumUID, Filter: a.AlbumFilter, Order: entity.SortOrderRelevance, Count: 1, Offset: 0, Merged: false}

		if photos, _, err := PhotoSearch(f); err != nil {
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

		return file, fmt.Errorf("found no cover for moment")
	}

	if err := Db().Where("files.file_primary = 1 AND files.file_missing = 0 AND files.file_type = 'jpg' AND files.deleted_at IS NULL").
		Joins("JOIN albums ON albums.album_uid = ?", uid).
		Joins("JOIN photos_albums pa ON pa.album_uid = albums.album_uid AND pa.photo_uid = files.photo_uid AND pa.hidden = 0").
		Joins("JOIN photos ON photos.id = files.photo_id AND photos.photo_private = 0 AND photos.deleted_at IS NULL").
		Order("photos.photo_quality DESC, photos.taken_at DESC").
		First(&file).Error; err != nil {
		return file, err
	}

	return file, nil
}

// AlbumPhotos returns up to count photos from an album.
func AlbumPhotos(a entity.Album, count int) (results PhotoResults, err error) {
	results, _, err = PhotoSearch(form.PhotoSearch{
		Album:  a.AlbumUID,
		Filter: a.AlbumFilter,
		Count:  count,
		Offset: 0,
	})

	return results, err
}

// AlbumSearch searches albums based on their name.
func AlbumSearch(f form.AlbumSearch) (results AlbumResults, err error) {
	if err := f.ParseQueryString(); err != nil {
		return results, err
	}

	defer log.Debug(capture.Time(time.Now(), fmt.Sprintf("albums: search %s", form.Serialize(f, true))))

	// Base query.
	s := UnscopedDb().Table("albums").
		Select("albums.*, cp.photo_count,	cl.link_count").
		Joins("LEFT JOIN (SELECT album_uid, count(photo_uid) AS photo_count FROM photos_albums WHERE hidden = 0 AND missing = 0 GROUP BY album_uid) AS cp ON cp.album_uid = albums.album_uid").
		Joins("LEFT JOIN (SELECT share_uid, count(share_uid) AS link_count FROM links GROUP BY share_uid) AS cl ON cl.share_uid = albums.album_uid").
		Where("albums.album_type <> 'folder' OR albums.album_path IN (SELECT photo_path FROM photos WHERE photo_private = 0 AND photo_quality > -1 AND deleted_at IS NULL)").
		Where("albums.deleted_at IS NULL")

	// Limit result count.
	if f.Count > 0 && f.Count <= MaxResults {
		s = s.Limit(f.Count).Offset(f.Offset)
	} else {
		s = s.Limit(MaxResults).Offset(f.Offset)
	}

	// Set sort order.
	switch f.Order {
	case "slug":
		s = s.Order("albums.album_favorite DESC, album_slug ASC")
	default:
		s = s.Order("albums.album_favorite DESC, albums.album_year DESC, albums.album_month DESC, albums.album_day DESC, albums.album_title, albums.created_at DESC")
	}

	if f.ID != "" {
		s = s.Where("albums.album_uid IN (?)", strings.Split(f.ID, Or))

		if result := s.Scan(&results); result.Error != nil {
			return results, result.Error
		}

		return results, nil
	}

	if f.Query != "" {
		likeString := "%" + f.Query + "%"
		s = s.Where("albums.album_title LIKE ? OR albums.album_location LIKE ?", likeString, likeString)
	}

	if f.Type != "" {
		s = s.Where("albums.album_type IN (?)", strings.Split(f.Type, Or))
	}

	if f.Category != "" {
		s = s.Where("albums.album_category IN (?)", strings.Split(f.Category, Or))
	}

	if f.Location != "" {
		s = s.Where("albums.album_location IN (?)", strings.Split(f.Location, Or))
	}

	if f.Country != "" {
		s = s.Where("albums.album_country IN (?)", strings.Split(f.Country, Or))
	}

	if f.Favorite {
		s = s.Where("albums.album_favorite = 1")
	}

	if (f.Year > 0 && f.Year <= txt.YearMax) || f.Year == entity.YearUnknown {
		s = s.Where("albums.album_year = ?", f.Year)
	}

	if (f.Month >= txt.MonthMin && f.Month <= txt.MonthMax) || f.Month == entity.MonthUnknown {
		s = s.Where("albums.album_month = ?", f.Month)
	}

	if (f.Day >= txt.DayMin && f.Month <= txt.DayMax) || f.Day == entity.DayUnknown {
		s = s.Where("albums.album_day = ?", f.Day)
	}

	if result := s.Scan(&results); result.Error != nil {
		return results, result.Error
	}

	return results, nil
}

// UpdateAlbumDates updates album year, month and day based on indexed photo metadata.
func UpdateAlbumDates() error {
	switch DbDialect() {
	case MySQL:
		return UnscopedDb().Exec(`UPDATE albums
		INNER JOIN
			(SELECT photo_path, MAX(taken_at_local) AS taken_max
			 FROM photos WHERE taken_src = 'meta' AND photos.photo_quality >= 3 AND photos.deleted_at IS NULL
			 GROUP BY photo_path) AS p ON albums.album_path = p.photo_path
		SET albums.album_year = YEAR(taken_max), albums.album_month = MONTH(taken_max), albums.album_day = DAY(taken_max)
		WHERE albums.album_type = 'folder' AND albums.album_path IS NOT NULL AND p.taken_max IS NOT NULL`).Error
	default:
		return nil
	}
}

// UpdateMissingAlbumEntries sets a flag for missing photo album entries.
func UpdateMissingAlbumEntries() error {
	switch DbDialect() {
	default:
		return UnscopedDb().Exec(`UPDATE photos_albums SET missing = 1 WHERE photo_uid IN
		(SELECT photo_uid FROM photos WHERE deleted_at IS NOT NULL OR photo_quality < 0)`).Error
	}
}

// AlbumEntryFound removes the missing flag from album entries.
func AlbumEntryFound(uid string) error {
	switch DbDialect() {
	default:
		return UnscopedDb().Exec(`UPDATE photos_albums SET missing = 0 WHERE photo_uid = ?`, uid).Error
	}
}

// GetAlbums returns a slice of albums.
func GetAlbums(offset, limit int) (results entity.Albums, err error) {
	err = UnscopedDb().Table("albums").Select("*").Offset(offset).Limit(limit).Find(&results).Error
	return results, err
}
