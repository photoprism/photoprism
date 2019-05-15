package photoprism

import (
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/forms"
	"github.com/photoprism/photoprism/internal/models"

	log "github.com/sirupsen/logrus"
)

// Search searches given an originals path and a db instance.
type Search struct {
	originalsPath string
	db            *gorm.DB
}

// SearchCount is the total number of search hits.
type SearchCount struct {
	Total int
}

// PhotoSearchResult is a found mediafile.
type PhotoSearchResult struct {
	// Photo
	ID                 uint
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          time.Time
	TakenAt            time.Time
	PhotoTitle         string
	PhotoDescription   string
	PhotoArtist        string
	PhotoKeywords      string
	PhotoColors        string
	PhotoColor         string
	PhotoCanonicalName string
	PhotoLat           float64
	PhotoLong          float64
	PhotoFavorite      bool

	// Camera
	CameraID    uint
	CameraModel string
	CameraMake  string

	// Lens
	LensID    uint
	LensModel string
	LensMake  string

	// Country
	CountryID   string
	CountryName string

	// Location
	LocationID     uint
	LocDisplayName string
	LocName        string
	LocCity        string
	LocPostcode    string
	LocCounty      string
	LocState       string
	LocCountry     string
	LocCountryCode string
	LocCategory    string
	LocType        string

	// File
	FileID             uint
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

	// Tags
	Tags string
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
func (s *Search) Photos(form forms.PhotoSearchForm) (results []PhotoSearchResult, err error) {
	if err := form.ParseQueryString(); err != nil {
		return results, err
	}

	log.Infof("%+v\n", form)

	q := s.db.NewScope(nil).DB()
	q = q.Table("photos").
		Select(`SQL_CALC_FOUND_ROWS photos.*,
		files.id AS file_id, files.file_primary, files.file_missing, files.file_name, files.file_hash, 
		files.file_type, files.file_mime, files.file_width, files.file_height, files.file_aspect_ratio, 
		files.file_orientation, files.file_main_color, files.file_colors, files.file_luminance, files.file_saturation,
		cameras.camera_make, cameras.camera_model,
		lenses.lens_make, lenses.lens_model,
		countries.country_name,
		locations.loc_display_name, locations.loc_name, locations.loc_city, locations.loc_postcode, locations.loc_county, 
		locations.loc_state, locations.loc_country, locations.loc_country_code, locations.loc_category, locations.loc_type,
		GROUP_CONCAT(tags.tag_label) AS tags`).
		Joins("JOIN files ON files.photo_id = photos.id AND files.file_primary AND files.deleted_at IS NULL").
		Joins("JOIN cameras ON cameras.id = photos.camera_id").
		Joins("JOIN lenses ON lenses.id = photos.lens_id").
		Joins("LEFT JOIN countries ON countries.id = photos.country_id").
		Joins("LEFT JOIN locations ON locations.id = photos.location_id").
		Joins("LEFT JOIN photo_tags ON photo_tags.photo_id = photos.id").
		Joins("LEFT JOIN tags ON photo_tags.tag_id = tags.id").
		Where("photos.deleted_at IS NULL AND files.file_missing = 0").
		Group("photos.id, files.id")

	if form.Location == true {
		q = q.Where("location_id > 0")

		if form.Query != "" {
			likeString := "%" + strings.ToLower(form.Query) + "%"
			q = q.Where("LOWER(locations.loc_display_name) LIKE ?", likeString)
		}
	} else if form.Query != "" {
		likeString := "%" + strings.ToLower(form.Query) + "%"
		q = q.Where("tags.tag_label LIKE ? OR LOWER(photo_title) LIKE ? OR LOWER(files.file_main_color) LIKE ?", likeString, likeString, likeString)
	}

	if form.Camera > 0 {
		q = q.Where("photos.camera_id = ?", form.Camera)
	}

	if form.Color != "" {
		q = q.Where("files.file_main_color = ?", form.Color)
	}

	if form.Favorites {
		q = q.Where("photos.photo_favorite = 1")
	}

	if form.Country != "" {
		q = q.Where("locations.loc_country_code = ?", form.Country)
	}

	switch form.Cat {
	case "amenity":
		q = q.Where("locations.loc_category = 'amenity'")
	case "bank":
		q = q.Where("locations.loc_type = 'bank'")
	case "building":
		q = q.Where("locations.loc_category = 'building'")
	case "school":
		q = q.Where("locations.loc_type = 'school'")
	case "supermarket":
		q = q.Where("locations.loc_type = 'supermarket'")
	case "shop":
		q = q.Where("locations.loc_category = 'shop'")
	case "hotel":
		q = q.Where("locations.loc_type = 'hotel'")
	case "bar":
		q = q.Where("locations.loc_type = 'bar'")
	case "parking":
		q = q.Where("locations.loc_type = 'parking'")
	case "airport":
		q = q.Where("locations.loc_category = 'aeroway'")
	case "historic":
		q = q.Where("locations.loc_category = 'historic'")
	case "tourism":
		q = q.Where("locations.loc_category = 'tourism'")
	default:
	}

	switch form.Order {
	case "newest":
		q = q.Order("taken_at DESC")
	case "oldest":
		q = q.Order("taken_at")
	case "imported":
		q = q.Order("created_at DESC")
	default:
		q = q.Order("taken_at DESC")
	}

	if form.Count > 0 && form.Count <= 1000 {
		q = q.Limit(form.Count).Offset(form.Offset)
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

// FindFileByID returns a mediafile given a certain ID.
func (s *Search) FindFileByID(id string) (file models.File, err error) {
	if err := s.db.Where("id = ?", id).First(&file).Error; err != nil {
		return file, err
	}

	return file, nil
}

// FindFileByHash finds a file with a given hash string.
func (s *Search) FindFileByHash(fileHash string) (file models.File, err error) {
	if err := s.db.Where("file_hash = ?", fileHash).First(&file).Error; err != nil {
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
