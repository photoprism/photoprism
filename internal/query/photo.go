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
	PhotoYear        int
	PhotoMonth       int
	PhotoCountry     string
	PhotoFavorite    bool
	PhotoPrivate     bool
	PhotoSensitive   bool
	PhotoStory       bool
	PhotoLat         float64
	PhotoLng         float64
	PhotoAltitude    int
	PhotoIso         int
	PhotoFocalLength int
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
	LocationID string
	PlaceID    string
	LocLabel   string
	LocCity    string
	LocState   string
	LocCountry string

	// File
	FileID          uint
	FileUUID        string
	FilePrimary     bool
	FileMissing     bool
	FileName        string
	FileHash        string
	FileType        string
	FileMime        string
	FileWidth       int
	FileHeight      int
	FileOrientation int
	FileAspectRatio float64
	FileColors      string // todo: remove from result?
	FileChroma      uint8  // todo: remove from result?
	FileLuminance   string // todo: remove from result?
	FileDiff        uint32 // todo: remove from result?
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
func (q *Query) Photos(f form.PhotoSearch) (results []PhotoResult, err error) {
	if err := f.ParseQueryString(); err != nil {
		return results, err
	}

	defer log.Debug(capture.Time(time.Now(), fmt.Sprintf("photos: %+v", f)))

	s := q.db.NewScope(nil).DB()

	// s.LogMode(true)

	s = s.Table("photos").
		Select(`photos.*,
		files.id AS file_id, files.file_uuid, files.file_primary, files.file_missing, files.file_name, files.file_hash, 
		files.file_type, files.file_mime, files.file_width, files.file_height, files.file_aspect_ratio, 
		files.file_orientation, files.file_main_color, files.file_colors, files.file_luminance, files.file_chroma,
		files.file_diff,
		cameras.camera_make, cameras.camera_model,
		lenses.lens_make, lenses.lens_model,
		places.loc_label, places.loc_city, places.loc_state, places.loc_country
		`).
		Joins("JOIN files ON files.photo_id = photos.id AND files.file_primary AND files.deleted_at IS NULL").
		Joins("JOIN cameras ON cameras.id = photos.camera_id").
		Joins("JOIN lenses ON lenses.id = photos.lens_id").
		Joins("JOIN places ON photos.place_id = places.id").
		Joins("LEFT JOIN photos_labels ON photos_labels.photo_id = photos.id AND photos_labels.uncertainty < 100").
		Where("files.file_missing = 0").
		Group("photos.id, files.id")

	if f.ID != "" {
		s = s.Where("photos.photo_uuid = ?", f.ID)

		if result := s.Scan(&results); result.Error != nil {
			return results, result.Error
		}

		return results, nil
	}

	var categories []entity.Category
	var label entity.Label
	var labelIds []uint

	if f.Label != "" {
		slugString := strings.ToLower(f.Label)
		if result := q.db.First(&label, "label_slug =? OR custom_slug = ?", slugString, slugString); result.Error != nil {
			log.Errorf("search: label \"%s\" not found", f.Label)
			return results, fmt.Errorf("label \"%s\" not found", f.Label)
		} else {
			labelIds = append(labelIds, label.ID)

			q.db.Where("category_id = ?", label.ID).Find(&categories)

			for _, category := range categories {
				labelIds = append(labelIds, category.LabelID)
			}

			s = s.Where("photos_labels.label_id IN (?)", labelIds)
		}
	}

	if f.Location == true {
		s = s.Where("location_id > 0")

		if f.Query != "" {
			s = s.Joins("LEFT JOIN photos_keywords ON photos_keywords.photo_id = photos.id").
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

		s = s.Joins("LEFT JOIN photos_keywords ON photos_keywords.photo_id = photos.id").
			Joins("LEFT JOIN keywords ON photos_keywords.keyword_id = keywords.id")

		if result := q.db.First(&label, "label_slug = ? OR custom_slug = ?", slugString, slugString); result.Error != nil {
			log.Infof("search: label \"%s\" not found, using fuzzy search", f.Query)

			s = s.Where("keywords.keyword LIKE ?", likeString)
		} else {
			labelIds = append(labelIds, label.ID)

			q.db.Where("category_id = ?", label.ID).Find(&categories)

			for _, category := range categories {
				labelIds = append(labelIds, category.LabelID)
			}

			log.Infof("search: label \"%s\" includes %d categories", label.LabelName, len(labelIds))

			s = s.Where("photos_labels.label_id IN (?) OR keywords.keyword LIKE ?", labelIds, likeString)
		}
	}

	if f.Archived {
		s = s.Where("photos.deleted_at IS NOT NULL")
	} else {
		s = s.Where("photos.deleted_at IS NULL")
	}

	if f.Error {
		s = s.Where("files.file_error <> ''")
	}

	if f.Album != "" {
		s = s.Joins("JOIN photos_albums ON photos_albums.photo_uuid = photos.photo_uuid").Where("photos_albums.album_uuid = ?", f.Album)
	}

	if f.Camera > 0 {
		s = s.Where("photos.camera_id = ?", f.Camera)
	}

	if f.Lens > 0 {
		s = s.Where("photos.lens_id = ?", f.Lens)
	}

	if f.Year > 0 {
		s = s.Where("photos.photo_year = ?", f.Year)
	}

	if f.Month > 0 {
		s = s.Where("photos.photo_month = ?", f.Month)
	}

	if f.Color != "" {
		s = s.Where("files.file_main_color = ?", strings.ToLower(f.Color))
	}

	if f.Favorites {
		s = s.Where("photos.photo_favorite = 1")
	}

	if f.Public {
		s = s.Where("photos.photo_private = 0")
	}

	if f.Safe {
		s = s.Where("photos.photo_nsfw = 0")
	}

	if f.Nsfw {
		s = s.Where("photos.photo_nsfw = 1")
	}

	if f.Story {
		s = s.Where("photos.photo_story = 1")
	}

	if f.Country != "" {
		s = s.Where("photos.photo_country = ?", f.Country)
	}

	if f.Title != "" {
		s = s.Where("LOWER(photos.photo_title) LIKE ?", fmt.Sprintf("%%%s%%", strings.ToLower(f.Title)))
	}

	if f.Hash != "" {
		s = s.Where("files.file_hash = ?", f.Hash)
	}

	if f.Duplicate {
		s = s.Where("files.file_duplicate = 1")
	}

	if f.Portrait {
		s = s.Where("files.file_portrait = 1")
	}

	if f.Mono {
		s = s.Where("files.file_chroma = 0")
	} else if f.Chroma > 9 {
		s = s.Where("files.file_chroma > ?", f.Chroma)
	} else if f.Chroma > 0 {
		s = s.Where("files.file_chroma > 0 AND files.file_chroma <= ?", f.Chroma)
	}

	if f.Diff != 0 {
		s = s.Where("files.file_diff = ?", f.Diff)
	}

	if f.Fmin > 0 {
		s = s.Where("photos.photo_f_number >= ?", f.Fmin)
	}

	if f.Fmax > 0 {
		s = s.Where("photos.photo_f_number <= ?", f.Fmax)
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
		s = s.Where("photos.photo_lat BETWEEN ? AND ?", latMin, latMax)
	}

	if f.Lng > 0 {
		lngMin := f.Lng - SearchRadius*float64(f.Dist)
		lngMax := f.Lng + SearchRadius*float64(f.Dist)
		s = s.Where("photos.photo_lng BETWEEN ? AND ?", lngMin, lngMax)
	}

	if !f.Before.IsZero() {
		s = s.Where("photos.taken_at <= ?", f.Before.Format("2006-01-02"))
	}

	if !f.After.IsZero() {
		s = s.Where("photos.taken_at >= ?", f.After.Format("2006-01-02"))
	}

	switch f.Order {
	case "relevance":
		s = s.Order("photo_story DESC, photo_favorite DESC, taken_at DESC")
	case "newest":
		s = s.Order("taken_at DESC, photos.photo_uuid")
	case "oldest":
		s = s.Order("taken_at, photos.photo_uuid")
	case "imported":
		s = s.Order("photos.id DESC")
	case "similar":
		s = s.Order("files.file_main_color, photos.location_id, files.file_diff, taken_at DESC")
	default:
		s = s.Order("taken_at DESC, photos.photo_uuid")
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

// PhotoByID returns a Photo based on the ID.
func (q *Query) PhotoByID(photoID uint64) (photo entity.Photo, err error) {
	if err := q.db.Unscoped().Where("id = ?", photoID).
		Preload("Links").
		Preload("Description").
		Preload("Location").
		Preload("Location.Place").
		Preload("Labels", func(db *gorm.DB) *gorm.DB {
			return db.Order("photos_labels.uncertainty ASC, photos_labels.label_id DESC")
		}).
		Preload("Labels.Label").
		First(&photo).Error; err != nil {
		return photo, err
	}

	return photo, nil
}

// PhotoByUUID returns a Photo based on the UUID.
func (q *Query) PhotoByUUID(photoUUID string) (photo entity.Photo, err error) {
	if err := q.db.Unscoped().Where("photo_uuid = ?", photoUUID).
		Preload("Links").
		Preload("Description").
		Preload("Location").
		Preload("Location.Place").
		Preload("Labels", func(db *gorm.DB) *gorm.DB {
			return db.Order("photos_labels.uncertainty ASC, photos_labels.label_id DESC")
		}).
		Preload("Labels.Label").
		First(&photo).Error; err != nil {
		return photo, err
	}

	return photo, nil
}

// PreloadPhotoByUUID returns a Photo based on the UUID with all dependencies preloaded.
func (q *Query) PreloadPhotoByUUID(photoUUID string) (photo entity.Photo, err error) {
	if err := q.db.Unscoped().Where("photo_uuid = ?", photoUUID).
		Preload("Labels", func(db *gorm.DB) *gorm.DB {
			return db.Order("photos_labels.uncertainty ASC, photos_labels.label_id DESC")
		}).
		Preload("Labels.Label").
		Preload("Camera").
		Preload("Lens").
		Preload("Links").
		Preload("Location").
		Preload("Location.Place").
		Preload("Description").
		First(&photo).Error; err != nil {
		return photo, err
	}

	photo.PreloadMany(q.db)

	return photo, nil
}
