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
	AlbumCount       int
	AlbumFavorite    bool
	AlbumDescription string
	AlbumNotes       string
}

// FindAlbumByUUID returns a Album based on the UUID.
func (s *Repo) FindAlbumByUUID(albumUUID string) (album entity.Album, err error) {
	if err := s.db.Where("album_uuid = ?", albumUUID).First(&album).Error; err != nil {
		return album, err
	}

	return album, nil
}

// FindAlbumThumbByUUID returns a album preview file based on the uuid.
func (s *Repo) FindAlbumThumbByUUID(albumUUID string) (file entity.File, err error) {
	// s.db.LogMode(true)

	if err := s.db.Where("files.file_primary AND files.deleted_at IS NULL").
		Joins("JOIN albums ON albums.album_uuid = ?", albumUUID).
		Joins("JOIN photos_albums pa ON pa.album_uuid = albums.album_uuid AND pa.photo_uuid = files.photo_uuid").
		First(&file).Error; err != nil {
		return file, err
	}

	return file, nil
}

// Albums searches albums based on their name.
func (s *Repo) Albums(f form.AlbumSearch) (results []AlbumResult, err error) {
	if err := f.ParseQueryString(); err != nil {
		return results, err
	}

	defer log.Debug(capture.Time(time.Now(), fmt.Sprintf("albums: %+v", f)))

	q := s.db.NewScope(nil).DB()

	q = q.Table("albums").
		Select(`albums.*, COUNT(photos_albums.album_uuid) AS album_count`).
		Joins("LEFT JOIN photos_albums ON photos_albums.album_uuid = albums.album_uuid").
		Where("albums.deleted_at IS NULL").
		Group("albums.id")

	if f.ID != "" {
		q = q.Where("albums.album_uuid = ?", f.ID)

		if result := q.Scan(&results); result.Error != nil {
			return results, result.Error
		}

		return results, nil
	}

	if f.Query != "" {
		likeString := "%" + strings.ToLower(f.Query) + "%"
		q = q.Where("LOWER(albums.album_name) LIKE ?", likeString)
	}

	if f.Favorites {
		q = q.Where("albums.album_favorite = 1")
	}

	switch f.Order {
	case "slug":
		q = q.Order("albums.album_favorite DESC, album_slug ASC")
	default:
		q = q.Order("albums.album_favorite DESC, album_count DESC, albums.created_at DESC")
	}

	if f.Count > 0 && f.Count <= 1000 {
		q = q.Limit(f.Count).Offset(f.Offset)
	} else {
		q = q.Limit(100).Offset(0)
	}

	if result := q.Scan(&results); result.Error != nil {
		return results, result.Error
	}

	return results, nil
}
