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
	ID               uint      `json:"-" yaml:"-"`
	AlbumUID         string    `json:"UID" yaml:"UID"`
	AlbumSlug        string    `json:"Slug" yaml:"Slug"`
	AlbumType        string    `json:"Type" yaml:"Type,omitempty"`
	AlbumName        string    `json:"Name" yaml:"Name,omitempty"`
	AlbumDescription string    `json:"Description" yaml:"Description,omitempty"`
	AlbumNotes       string    `json:"Notes" yaml:"Notes,omitempty"`
	AlbumOrder       string    `json:"Order" yaml:"Order,omitempty"`
	AlbumFilter      string    `json:"Filter" yaml:"Filter,omitempty"`
	AlbumTemplate    string    `json:"Template" yaml:"Template,omitempty"`
	AlbumFavorite    bool      `json:"Favorite" yaml:"Favorite,omitempty"`
	PhotoCount       int       `json:"PhotoCount" yaml:"-"`
	LinkCount        int       `json:"LinkCount" yaml:"-"`
	CreatedAt        time.Time `json:"CreatedAt" yaml:"-"`
	UpdatedAt        time.Time `json:"UpdatedAt" yaml:"-"`
	DeletedAt        time.Time `json:"DeletedAt,omitempty" yaml:"-"`
}

// AlbumByUID returns a Album based on the UID.
func AlbumByUID(albumUID string) (album entity.Album, err error) {
	if err := Db().Where("album_uid = ?", albumUID).Preload("Links").First(&album).Error; err != nil {
		return album, err
	}

	return album, nil
}

// AlbumThumbByUID returns a album preview file based on the uid.
func AlbumThumbByUID(albumUID string) (file entity.File, err error) {
	if err := Db().Where("files.file_primary = 1 AND files.deleted_at IS NULL").
		Joins("JOIN albums ON albums.album_uid = ?", albumUID).
		Joins("JOIN photos_albums pa ON pa.album_uid = albums.album_uid AND pa.photo_uid = files.photo_uid").
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

	defer log.Debug(capture.Time(time.Now(), fmt.Sprintf("albums: search %s", form.Serialize(f, true))))

	s := Db().NewScope(nil).DB()

	s = s.Table("albums").
		Select(`albums.*, 
			COUNT(photos_albums.album_uid) AS photo_count,
			COUNT(links.link_token) AS link_count`).
		Joins("LEFT JOIN photos_albums ON photos_albums.album_uid = albums.album_uid").
		Joins("LEFT JOIN links ON links.share_uid = albums.album_uid").
		Where("albums.deleted_at IS NULL").
		Group("albums.id")

	if f.ID != "" {
		s = s.Where("albums.album_uid = ?", f.ID)

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
		s = s.Order("albums.album_favorite DESC, photo_count DESC, albums.created_at DESC")
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
