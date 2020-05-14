package query

import (
	"fmt"
	"strings"
	"time"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/capture"
)

// AlbumResult contains found albums
type AlbumResult struct {
	ID               uint
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        time.Time
	AlbumUUID        string
	AlbumSlug        string
	AlbumName        string
	AlbumDescription string
	AlbumNotes       string
	AlbumOrder       string
	AlbumTemplate    string
	AlbumCount       int
	AlbumFavorite    bool
	LinkCount        int
}

// AlbumByUUID returns a Album based on the UUID.
func AlbumByUUID(albumUUID string) (album entity.Album, err error) {
	if err := Db().Where("album_uuid = ?", albumUUID).Preload("Links").First(&album).Error; err != nil {
		return album, err
	}

	return album, nil
}

// AlbumThumbByUUID returns a album preview file based on the uuid.
func AlbumThumbByUUID(albumUUID string) (file entity.File, err error) {
	if err := Db().Where("files.file_primary = 1 AND files.deleted_at IS NULL").
		Joins("JOIN albums ON albums.album_uuid = ?", albumUUID).
		Joins("JOIN photos_albums pa ON pa.album_uuid = albums.album_uuid AND pa.photo_uuid = files.photo_uuid").
		Joins("JOIN photos ON photos.id = files.photo_id AND photos.photo_private = 0 AND photos.deleted_at IS NULL").
		Order("photos.photo_quality DESC, photos.taken_at DESC").
		First(&file).Error; err != nil {
		return file, err
	}

	return file, nil
}

// Albums searches albums based on their name.
func Albums(f form.AlbumSearch) (results []AlbumResult, err error) {
	if err := f.ParseQueryString(); err != nil {
		return results, err
	}

	defer log.Debug(capture.Time(time.Now(), fmt.Sprintf("albums: %+v", f)))

	s := Db().NewScope(nil).DB()

	s = s.Table("albums").
		Select(`albums.*, 
			COUNT(photos_albums.album_uuid) AS album_count,
			COUNT(links.link_token) AS link_count`).
		Joins("LEFT JOIN photos_albums ON photos_albums.album_uuid = albums.album_uuid").
		Joins("LEFT JOIN links ON links.share_uuid = albums.album_uuid").
		Where("albums.deleted_at IS NULL").
		Group("albums.id")

	if f.ID != "" {
		s = s.Where("albums.album_uuid = ?", f.ID)

		if result := s.Scan(&results); result.Error != nil {
			return results, result.Error
		}

		return results, nil
	}

	if f.Query != "" {
		likeString := "%" + strings.ToLower(f.Query) + "%"
		s = s.Where("LOWER(albums.album_name) LIKE ?", likeString)
	}

	if f.Favorite {
		s = s.Where("albums.album_favorite = 1")
	}

	switch f.Order {
	case "slug":
		s = s.Order("albums.album_favorite DESC, album_slug ASC")
	default:
		s = s.Order("albums.album_favorite DESC, album_count DESC, albums.created_at DESC")
	}

	if f.Count > 0 && f.Count <= 1000 {
		s = s.Limit(f.Count).Offset(f.Offset)
	} else {
		s = s.Limit(100).Offset(0)
	}

	if result := s.Scan(&results); result.Error != nil {
		return results, result.Error
	}

	return results, nil
}
