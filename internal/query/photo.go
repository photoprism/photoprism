package query

import (
	"fmt"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/capture"
)

// PhotoResult contains found photos and their main file plus other meta data.
type PhotoResult struct {
	// Photo
	ID               uint
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        time.Time
	TakenAt          time.Time
	TakenAtLocal     time.Time
	TimeZone         string
	PhotoUUID        string
	PhotoPath        string
	PhotoName        string
	PhotoTitle       string
	PhotoDescription string
	PhotoYear        int
	PhotoMonth       int
	PhotoCountry     string
	PhotoArtist      string
	PhotoKeywords    string
	PhotoColors      string
	PhotoColor       string
	PhotoFavorite    bool
	PhotoPrivate     bool
	PhotoSensitive   bool
	PhotoStory       bool
	PhotoLat         float64
	PhotoLng         float64
	PhotoAltitude    int
	PhotoFocalLength int
	PhotoIso         int
	PhotoFNumber     float64
	PhotoExposure    string

	// Camera
	CameraID    uint
	CameraModel string
	CameraMake  string

	// Lens
	LensID    uint
	LensModel string
	LensMake  string

	// Location
	LocationID        string
	PlaceID           string
	LocLabel          string
	LocCity           string
	LocState          string
	LocCountry        string
	LocationChanged   bool
	LocationEstimated bool

	// File
	FileID             uint
	FileUUID           string
	FilePrimary        bool
	FileMissing        bool
	FileName           string
	FileHash           string
	FilePerceptualHash string
	FileType           string
	FileMime           string
	FileWidth          int
	FileHeight         int
	FileOrientation    int
	FileAspectRatio    float64
}

func (m *PhotoResult) DownloadFileName() string {
	var name string

	if m.PhotoTitle != "" {
		name = strings.Title(slug.MakeLang(m.PhotoTitle, "en"))
	} else {
		name = m.PhotoUUID
	}

	taken := m.TakenAt.Format("20060102-150405")

	result := fmt.Sprintf("%s-%s.%s", taken, name, m.FileType)

	return result
}

// Photos searches for photos based on a Form and returns a PhotoResult slice.
func (s *Repo) Photos(f form.PhotoSearch) (results []PhotoResult, err error) {
	if err := f.ParseQueryString(); err != nil {
		return results, err
	}

	defer log.Debug(capture.Time(time.Now(), fmt.Sprintf("photos: %+v", f)))

	q := s.db.NewScope(nil).DB()

	// q.LogMode(true)

	q = q.Table("photos").
		Select(`photos.*,
		files.id AS file_id, files.file_uuid, files.file_primary, files.file_missing, files.file_name, files.file_hash, 
		files.file_type, files.file_mime, files.file_width, files.file_height, files.file_aspect_ratio, 
		files.file_orientation, files.file_main_color, files.file_colors, files.file_luminance, files.file_chroma,
		cameras.camera_make, cameras.camera_model,
		lenses.lens_make, lenses.lens_model,
		places.loc_label, places.loc_city, places.loc_state, places.loc_country
		`).
		Joins("JOIN files ON files.photo_id = photos.id AND files.file_primary AND files.deleted_at IS NULL").
		Joins("JOIN cameras ON cameras.id = photos.camera_id").
		Joins("JOIN lenses ON lenses.id = photos.lens_id").
		Joins("JOIN places ON photos.place_id = places.id").
		Joins("LEFT JOIN photos_labels ON photos_labels.photo_id = photos.id").
		Where("files.file_missing = 0").
		Group("photos.id, files.id")

	if f.ID != "" {
		q = q.Where("photos.photo_uuid = ?", f.ID)

		if result := q.Scan(&results); result.Error != nil {
			return results, result.Error
		}

		return results, nil
	}

	var categories []entity.Category
	var label entity.Label
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

			q = q.Where("photos_labels.label_id IN (?)", labelIds)
		}
	}

	if f.Location == true {
		q = q.Where("location_id > 0")

		if f.Query != "" {
			q = q.Joins("LEFT JOIN photos_keywords ON photos_keywords.photo_id = photos.id").
				Joins("LEFT JOIN keywords ON photos_keywords.keyword_id = keywords.id").
				Where("keywords.keyword LIKE ?", strings.ToLower(f.Query)+"%")
		}
	} else if f.Query != "" {
		if len(f.Query) < 2 {
			return results, fmt.Errorf("query too short")
		}

		slugString := slug.Make(f.Query)
		lowerString := strings.ToLower(f.Query)
		likeString := lowerString + "%"

		q = q.Joins("LEFT JOIN photos_keywords ON photos_keywords.photo_id = photos.id").
			Joins("LEFT JOIN keywords ON photos_keywords.keyword_id = keywords.id")

		if result := s.db.First(&label, "label_slug = ?", slugString); result.Error != nil {
			log.Infof("search: label \"%s\" not found, using fuzzy search", f.Query)

			q = q.Where("keywords.keyword LIKE ?", likeString)
		} else {
			labelIds = append(labelIds, label.ID)

			s.db.Where("category_id = ?", label.ID).Find(&categories)

			for _, category := range categories {
				labelIds = append(labelIds, category.LabelID)
			}

			log.Infof("search: label \"%s\" includes %d categories", label.LabelName, len(labelIds))

			q = q.Where("photos_labels.label_id IN (?) OR keywords.keyword LIKE ?", labelIds, likeString)
		}
	}

	if f.Archived {
		q = q.Where("photos.deleted_at IS NOT NULL")
	} else {
		q = q.Where("photos.deleted_at IS NULL")
	}

	if f.Error {
		q = q.Where("files.file_error <> ''")
	}

	if f.Album != "" {
		q = q.Joins("JOIN photos_albums ON photos_albums.photo_uuid = photos.photo_uuid").Where("photos_albums.album_uuid = ?", f.Album)
	}

	if f.Camera > 0 {
		q = q.Where("photos.camera_id = ?", f.Camera)
	}

	if f.Lens > 0 {
		q = q.Where("photos.lens_id = ?", f.Lens)
	}

	if f.Year > 0 {
		q = q.Where("photos.photo_year = ?", f.Year)
	}

	if f.Month > 0 {
		q = q.Where("photos.photo_month = ?", f.Month)
	}

	if f.Color != "" {
		q = q.Where("files.file_main_color = ?", strings.ToLower(f.Color))
	}

	if f.Favorites {
		q = q.Where("photos.photo_favorite = 1")
	}

	if f.Public {
		q = q.Where("photos.photo_private = 0")
	}

	if f.Safe {
		q = q.Where("photos.photo_nsfw = 0")
	}

	if f.Nsfw {
		q = q.Where("photos.photo_nsfw = 1")
	}

	if f.Story {
		q = q.Where("photos.photo_story = 1")
	}

	if f.Country != "" {
		q = q.Where("photos.photo_country = ?", f.Country)
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
	} else if f.Dist > 5000 {
		f.Dist = 5000
	}

	// Inaccurate distance search, but probably 'good enough' for now
	if f.Lat > 0 {
		latMin := f.Lat - SearchRadius*float64(f.Dist)
		latMax := f.Lat + SearchRadius*float64(f.Dist)
		q = q.Where("photos.photo_lat BETWEEN ? AND ?", latMin, latMax)
	}

	if f.Lng > 0 {
		lngMin := f.Lng - SearchRadius*float64(f.Dist)
		lngMax := f.Lng + SearchRadius*float64(f.Dist)
		q = q.Where("photos.photo_lng BETWEEN ? AND ?", lngMin, lngMax)
	}

	if !f.Before.IsZero() {
		q = q.Where("photos.taken_at <= ?", f.Before.Format("2006-01-02"))
	}

	if !f.After.IsZero() {
		q = q.Where("photos.taken_at >= ?", f.After.Format("2006-01-02"))
	}

	switch f.Order {
	case "relevance":
		q = q.Order("photo_story DESC, photo_favorite DESC, taken_at DESC")
	case "newest":
		q = q.Order("taken_at DESC, photos.photo_uuid")
	case "oldest":
		q = q.Order("taken_at, photos.photo_uuid")
	case "imported":
		q = q.Order("photos.id DESC")
	default:
		q = q.Order("taken_at DESC, photos.photo_uuid")
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

// FindPhotoByID returns a Photo based on the ID.
func (s *Repo) FindPhotoByID(photoID uint64) (photo entity.Photo, err error) {
	if err := s.db.Where("id = ?", photoID).First(&photo).Error; err != nil {
		return photo, err
	}

	return photo, nil
}

// FindPhotoByUUID returns a Photo based on the UUID.
func (s *Repo) FindPhotoByUUID(photoUUID string) (photo entity.Photo, err error) {
	if err := s.db.Where("photo_uuid = ?", photoUUID).First(&photo).Error; err != nil {
		return photo, err
	}

	return photo, nil
}

// PreloadPhotoByUUID returns a Photo based on the UUID with all dependencies preloaded.
func (s *Repo) PreloadPhotoByUUID(photoUUID string) (photo entity.Photo, err error) {
	if err := s.db.Where("photo_uuid = ?", photoUUID).
		Preload("Labels", func(db *gorm.DB) *gorm.DB {
			return db.Order("photos_labels.label_uncertainty ASC, photos_labels.label_id DESC")
		}).
		Preload("Labels.Label").
		Preload("Camera").
		Preload("Lens").
		First(&photo).Error; err != nil {
		return photo, err
	}

	photo.PreloadMany(s.db)

	return photo, nil
}
