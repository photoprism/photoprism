package photoprism

import (
	"fmt"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/models"
	"github.com/photoprism/photoprism/internal/util"
)

// About 1km ('good enough' for now)
const SearchRadius = 0.009

// Search searches given an originals path and a db instance.
type Search struct {
	originalsPath string
	db            *gorm.DB
}

// SearchCount is the total number of search hits.
type SearchCount struct {
	Total int
}

// NewSearch returns a new Search type with a given path and db instance.
func NewSearch(originalsPath string, db *gorm.DB) *Search {
	instance := &Search{
		originalsPath: originalsPath,
		db:            db,
	}

	return instance
}

// Photos searches for photos based on a Form and returns a PhotoSearchResult slice.
func (s *Search) Photos(f form.PhotoSearch) (results []PhotoSearchResult, err error) {
	if err := f.ParseQueryString(); err != nil {
		return results, err
	}

	defer util.ProfileTime(time.Now(), fmt.Sprintf("search: %+v", f))

	q := s.db.NewScope(nil).DB()

	// q.LogMode(true)

	q = q.Table("photos").
		Select(`photos.*,
		files.id AS file_id, files.file_uuid, files.file_primary, files.file_missing, files.file_name, files.file_hash, 
		files.file_type, files.file_mime, files.file_width, files.file_height, files.file_aspect_ratio, 
		files.file_orientation, files.file_main_color, files.file_colors, files.file_luminance, files.file_chroma,
		cameras.camera_make, cameras.camera_model,
		lenses.lens_make, lenses.lens_model,
		countries.country_name,
		locations.loc_display_name, locations.loc_name, locations.loc_city, locations.loc_postcode, locations.loc_county, 
		locations.loc_state, locations.loc_country, locations.loc_country_code, locations.loc_category, locations.loc_type,
		GROUP_CONCAT(labels.label_name) AS labels`).
		Joins("JOIN files ON files.photo_id = photos.id AND files.file_primary AND files.deleted_at IS NULL").
		Joins("JOIN cameras ON cameras.id = photos.camera_id").
		Joins("JOIN lenses ON lenses.id = photos.lens_id").
		Joins("LEFT JOIN countries ON countries.id = photos.country_id").
		Joins("LEFT JOIN locations ON locations.id = photos.location_id").
		Joins("LEFT JOIN photos_labels ON photos_labels.photo_id = photos.id").
		Joins("LEFT JOIN labels ON photos_labels.label_id = labels.id").
		Where("photos.deleted_at IS NULL AND files.file_missing = 0").
		Group("photos.id, files.id")
	var categories []models.Category
	var label models.Label
	var labelIds []uint

	if f.Label != "" {
		if result := s.db.First(&label, "label_slug = ?", strings.ToLower(f.Label)); result.Error != nil {
			log.Errorf("search: label \"%s\" not found", f.Label)
			return results, fmt.Errorf("label \"%s\" not found", f.Label)
		} else {
			labelIds = append(labelIds, label.ID)

			s.db.Where("category_id = ?", label.ID).Find(&categories)

			for _, category := range categories {
				labelIds = append(labelIds, category.LabelID)
			}

			q = q.Where("labels.id IN (?)", labelIds)
		}
	}

	if f.Location == true {
		q = q.Where("location_id > 0")

		if f.Query != "" {
			likeString := "%" + strings.ToLower(f.Query) + "%"
			q = q.Where("LOWER(locations.loc_display_name) LIKE ?", likeString)
		}
	} else if f.Query != "" {
		slugString := slug.Make(f.Query)
		lowerString := strings.ToLower(f.Query)
		likeString := "%" + lowerString + "%"

		if result := s.db.First(&label, "label_slug = ?", slugString); result.Error != nil {
			log.Infof("search: label \"%s\" not found", f.Query)

			q = q.Where("labels.label_slug = ? OR LOWER(photo_title) LIKE ? OR files.file_main_color = ?", slugString, likeString, lowerString)
		} else {
			labelIds = append(labelIds, label.ID)

			s.db.Where("category_id = ?", label.ID).Find(&categories)

			for _, category := range categories {
				labelIds = append(labelIds, category.LabelID)
			}

			log.Infof("search: label \"%s\" includes %d categories", label.LabelName, len(labelIds))

			q = q.Where("labels.id IN (?) OR LOWER(photo_title) LIKE ? OR files.file_main_color = ?", labelIds, likeString, lowerString)
		}

	}

	if f.Album != "" {
		q = q.Joins("JOIN photos_albums ON photos_albums.photo_uuid = photos.photo_uuid").Where("photos_albums.album_uuid = ?", f.Album)
	}

	if f.Camera > 0 {
		q = q.Where("photos.camera_id = ?", f.Camera)
	}

	if f.Color != "" {
		q = q.Where("files.file_main_color = ?", strings.ToLower(f.Color))
	}

	if f.Favorites {
		q = q.Where("photos.photo_favorite = 1")
	}

	if f.Country != "" {
		q = q.Where("locations.loc_country_code = ?", f.Country)
	}

	if f.Title != "" {
		q = q.Where("LOWER(photos.photo_title) LIKE ?", fmt.Sprintf("%%%s%%", strings.ToLower(f.Title)))
	}

	if f.Description != "" {
		q = q.Where("LOWER(photos.photo_description) LIKE ?", fmt.Sprintf("%%%s%%", strings.ToLower(f.Description)))
	}

	if f.Notes != "" {
		q = q.Where("LOWER(photos.photo_notes) LIKE ?", fmt.Sprintf("%%%s%%", strings.ToLower(f.Notes)))
	}

	if f.Hash != "" {
		q = q.Where("files.file_hash = ?", f.Hash)
	}

	if f.Duplicate {
		q = q.Where("files.file_duplicate = 1")
	}

	if f.Portrait {
		q = q.Where("files.file_portrait = 1")
	}

	if f.Mono {
		q = q.Where("files.file_chroma = 0")
	} else if f.Chroma > 9 {
		q = q.Where("files.file_chroma > ?", f.Chroma)
	} else if f.Chroma > 0 {
		q = q.Where("files.file_chroma > 0 AND files.file_chroma <= ?", f.Chroma)
	}

	if f.Fmin > 0 {
		q = q.Where("photos.photo_f_number >= ?", f.Fmin)
	}

	if f.Fmax > 0 {
		q = q.Where("photos.photo_f_number <= ?", f.Fmax)
	}

	if f.Dist == 0 {
		f.Dist = 20
	} else if f.Dist > 1000 {
		f.Dist = 1000
	}

	// Inaccurate distance search, but probably 'good enough' for now
	if f.Lat > 0 {
		latMin := f.Lat - SearchRadius*float64(f.Dist)
		latMax := f.Lat + SearchRadius*float64(f.Dist)
		q = q.Where("photos.photo_lat BETWEEN ? AND ?", latMin, latMax)
	}

	if f.Long > 0 {
		longMin := f.Long - SearchRadius*float64(f.Dist)
		longMax := f.Long + SearchRadius*float64(f.Dist)
		q = q.Where("photos.photo_long BETWEEN ? AND ?", longMin, longMax)
	}

	if !f.Before.IsZero() {
		q = q.Where("photos.taken_at <= ?", f.Before.Format("2006-01-02"))
	}

	if !f.After.IsZero() {
		q = q.Where("photos.taken_at >= ?", f.After.Format("2006-01-02"))
	}

	switch f.Order {
	case "newest":
		q = q.Order("taken_at DESC")
	case "oldest":
		q = q.Order("taken_at")
	case "imported":
		q = q.Order("created_at DESC")
	default:
		q = q.Order("taken_at DESC")
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

// FindFiles finds files returning maximum results defined by limit
// and finding them from an offest defined by offset.
func (s *Search) FindFiles(limit int, offset int) (files []models.File, err error) {
	if err := s.db.Where(&models.File{}).Limit(limit).Offset(offset).Find(&files).Error; err != nil {
		return files, err
	}

	return files, nil
}

// FindFilesByUUID
func (s *Search) FindFilesByUUID(u []string, limit int, offset int) (files []models.File, err error) {
	if err := s.db.Where("(photo_uuid IN (?) AND file_primary = 1) OR file_uuid IN (?)", u, u).Preload("Photo").Limit(limit).Offset(offset).Find(&files).Error; err != nil {
		return files, err
	}

	return files, nil
}

// FindFileByPhotoUUID
func (s *Search) FindFileByPhotoUUID(u string) (file models.File, err error) {
	if err := s.db.Where("photo_uuid = ? AND file_primary = 1", u).Preload("Photo").First(&file).Error; err != nil {
		return file, err
	}

	return file, nil
}

// FindFileByID returns a mediafile given a certain ID.
func (s *Search) FindFileByID(id string) (file models.File, err error) {
	if err := s.db.Where("id = ?", id).Preload("Photo").First(&file).Error; err != nil {
		return file, err
	}

	return file, nil
}

// FindFileByHash finds a file with a given hash string.
func (s *Search) FindFileByHash(fileHash string) (file models.File, err error) {
	if err := s.db.Where("file_hash = ?", fileHash).Preload("Photo").First(&file).Error; err != nil {
		return file, err
	}

	return file, nil
}

// FindPhotoByID returns a Photo based on the ID.
func (s *Search) FindPhotoByID(photoID uint64) (photo models.Photo, err error) {
	if err := s.db.Where("id = ?", photoID).First(&photo).Error; err != nil {
		return photo, err
	}

	return photo, nil
}

// FindPhotoByUUID returns a Photo based on the UUID.
func (s *Search) FindPhotoByUUID(photoUUID string) (photo models.Photo, err error) {
	if err := s.db.Where("photo_uuid = ?", photoUUID).First(&photo).Error; err != nil {
		return photo, err
	}

	return photo, nil
}

// FindLabelBySlug returns a Label based on the slug name.
func (s *Search) FindLabelBySlug(labelSlug string) (label models.Label, err error) {
	if err := s.db.Where("label_slug = ?", labelSlug).First(&label).Error; err != nil {
		return label, err
	}

	return label, nil
}

// FindLabelThumbBySlug returns a label preview file based on the slug name.
func (s *Search) FindLabelThumbBySlug(labelSlug string) (file models.File, err error) {
	// s.db.LogMode(true)

	if err := s.db.Where("files.file_primary AND files.deleted_at IS NULL").
		Joins("JOIN labels ON labels.label_slug = ?", labelSlug).
		Joins("JOIN photos_labels ON photos_labels.label_id = labels.id AND photos_labels.photo_id = files.photo_id").
		Order("photos_labels.label_uncertainty ASC").
		First(&file).Error; err != nil {
		return file, err
	}

	return file, nil
}

// Labels searches labels based on their name.
func (s *Search) Labels(f form.LabelSearch) (results []LabelSearchResult, err error) {
	if err := f.ParseQueryString(); err != nil {
		return results, err
	}

	defer util.ProfileTime(time.Now(), fmt.Sprintf("search: %+v", f))

	q := s.db.NewScope(nil).DB()

	// q.LogMode(true)

	q = q.Table("labels").
		Select(`labels.*, COUNT(photos_labels.label_id) AS label_count`).
		Joins("JOIN photos_labels ON photos_labels.label_id = labels.id").
		Where("labels.deleted_at IS NULL").
		Group("labels.id")

	if f.Query != "" {
		var labelIds []uint
		var categories []models.Category
		var label models.Label

		likeString := "%" + strings.ToLower(f.Query) + "%"

		if result := s.db.First(&label, "LOWER(label_name) LIKE LOWER(?)", f.Query); result.Error != nil {
			log.Infof("search: label \"%s\" not found", f.Query)

			q = q.Where("LOWER(labels.label_name) LIKE ?", likeString)
		} else {
			labelIds = append(labelIds, label.ID)

			s.db.Where("category_id = ?", label.ID).Find(&categories)

			for _, category := range categories {
				labelIds = append(labelIds, category.LabelID)
			}

			log.Infof("search: label \"%s\" includes %d categories", label.LabelName, len(labelIds))

			q = q.Where("labels.id IN (?) OR LOWER(labels.label_name) LIKE ?", labelIds, likeString)
		}
	}

	if f.Favorites {
		q = q.Where("labels.label_favorite = 1")
	}

	if f.Priority != 0 {
		q = q.Where("labels.label_priority > ?", f.Priority)
	} else {
		q = q.Where("labels.label_priority >= -1")
	}

	switch f.Order {
	case "slug":
		q = q.Order("labels.label_favorite DESC, label_slug ASC")
	default:
		q = q.Order("labels.label_favorite DESC, labels.label_priority DESC, label_count DESC, labels.created_at DESC")
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

/***************** Albums  *****************/

// FindAlbumByUUID returns a Album based on the UUID.
func (s *Search) FindAlbumByUUID(albumUUID string) (album models.Album, err error) {
	if err := s.db.Where("album_uuid = ?", albumUUID).First(&album).Error; err != nil {
		return album, err
	}

	return album, nil
}

// FindAlbumThumbByUUID returns a album preview file based on the uuid.
func (s *Search) FindAlbumThumbByUUID(albumUUID string) (file models.File, err error) {
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
func (s *Search) Albums(f form.AlbumSearch) (results []AlbumSearchResult, err error) {
	if err := f.ParseQueryString(); err != nil {
		return results, err
	}

	defer util.ProfileTime(time.Now(), fmt.Sprintf("search: %+v", f))

	q := s.db.NewScope(nil).DB()

	// q.LogMode(true)

	q = q.Table("albums").
		Select(`albums.*, COUNT(photos_albums.album_uuid) AS album_count`).
		Joins("LEFT JOIN photos_albums ON photos_albums.album_uuid = albums.album_uuid").
		Where("albums.deleted_at IS NULL").
		Group("albums.id")

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
