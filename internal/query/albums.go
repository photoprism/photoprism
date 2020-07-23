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
func AlbumCoverByUID(albumUID string) (file entity.File, err error) {
	a := entity.Album{}

	if err := Db().Where("album_uid = ?", albumUID).First(&a).Error; err != nil {
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
		Joins("JOIN albums ON albums.album_uid = ?", albumUID).
		Joins("JOIN photos_albums pa ON pa.album_uid = albums.album_uid AND pa.photo_uid = files.photo_uid").
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

	s := UnscopedDb().Table("albums").
		Select("albums.*, cp.photo_count,	cl.link_count").
		Joins("LEFT JOIN (SELECT album_uid, count(photo_uid) AS photo_count FROM photos_albums WHERE hidden = 0 GROUP BY album_uid) AS cp ON cp.album_uid = albums.album_uid").
		Joins("LEFT JOIN (SELECT share_uid, count(share_uid) AS link_count FROM links GROUP BY share_uid) AS cl ON cl.share_uid = albums.album_uid").
		Where("albums.deleted_at IS NULL")

	if f.ID != "" {
		s = s.Where("albums.album_uid IN (?)", strings.Split(f.ID, ","))

		if result := s.Scan(&results); result.Error != nil {
			return results, result.Error
		}

		return results, nil
	}

	if f.Query != "" {
		likeString := "%" + strings.ToLower(f.Query) + "%"
		s = s.Where("LOWER(albums.album_title) LIKE ? OR LOWER(albums.album_location) LIKE ?", likeString, likeString)
	}

	if f.Type != "" {
		s = s.Where("albums.album_type IN (?)", strings.Split(f.Type, ","))
	}

	if f.Category != "" {
		s = s.Where("albums.album_category IN (?)", strings.Split(f.Category, ","))
	}

	if f.Location != "" {
		s = s.Where("albums.album_location IN (?)", strings.Split(f.Location, ","))
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

	switch f.Order {
	case "slug":
		s = s.Order("albums.album_favorite DESC, album_slug ASC")
	default:
		s = s.Order("albums.album_favorite DESC, albums.album_year DESC, albums.album_month DESC, albums.album_day DESC, albums.album_title, albums.created_at DESC")
	}

	if f.Count > 0 && f.Count <= MaxResults {
		s = s.Limit(f.Count).Offset(f.Offset)
	} else {
		s = s.Limit(MaxResults).Offset(f.Offset)
	}

	if result := s.Scan(&results); result.Error != nil {
		return results, result.Error
	}

	return results, nil
}
