package photoprism

import (
	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/forms"
	"strings"
	"time"
)

type Search struct {
	originalsPath string
	db            *gorm.DB
}

type SearchCount struct {
	Total int
}

type PhotoSearchResult struct {
	// Photo
	ID                  uint
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           time.Time
	TakenAt             time.Time
	PhotoTitle          string
	PhotoDescription    string
	PhotoArtist         string
	PhotoKeywords       string
	PhotoColors         string
	PhotoVibrantColor   string
	PhotoMutedColor     string
	PhotoCanonicalName  string
	PhotoPerceptualHash string
	PhotoLat            float64
	PhotoLong           float64
	PhotoFavorite       bool

	// Camera
	CameraID    uint
	CameraModel string

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
	FileID          uint
	FileName        string
	FileType        string
	FileMime        string
	FileWidth       int
	FileHeight      int
	FileOrientation int
	FileAspectRatio float64

	// Tags
	Tags string
}

func NewSearch(originalsPath string, db *gorm.DB) *Search {
	instance := &Search{
		originalsPath: originalsPath,
		db:            db,
	}

	return instance
}

func (s *Search) Photos(form forms.PhotoSearchForm) ([]PhotoSearchResult, int, error) {
	q := s.db.NewScope(nil).DB()
	q = q.Table("photos").
		Select(`SQL_CALC_FOUND_ROWS photos.*,
		files.id AS file_id, files.file_name, files.file_type, files.file_mime, files.file_width, files.file_height, files.file_aspect_ratio, files.file_orientation,
		cameras.camera_model,
		locations.loc_display_name, locations.loc_name, locations.loc_city, locations.loc_postcode, locations.loc_country, locations.loc_country_code, locations.loc_category, locations.loc_type,
		GROUP_CONCAT(tags.tag_label) AS tags`).
		Joins("JOIN files ON files.photo_id = photos.id AND files.file_primary AND files.deleted_at IS NULL").
		Joins("JOIN cameras ON cameras.id = photos.camera_id").
		Joins("LEFT JOIN locations ON locations.id = photos.location_id").
		Joins("LEFT JOIN photo_tags ON photo_tags.photo_id = photos.id").
		Joins("LEFT JOIN tags ON photo_tags.tag_id = tags.id").
		Where("photos.deleted_at IS NULL").
		Group("photos.id, files.id")

	if form.Query != "" {
		q = q.Where("tags.tag_label LIKE ? OR MATCH (photo_title, photo_description, photo_artist, photo_colors) AGAINST (?)", "%"+strings.ToLower(form.Query)+"%", form.Query)
	}

	if form.CameraID > 0 {
		q = q.Where("camera_id = ?", form.CameraID)
	}

	q = q.Order(form.Order).Limit(form.Count).Offset(form.Offset)

	results := make([]PhotoSearchResult, 0, form.Count)

	rows, err := q.Rows()

	if err != nil {
		return results, 0, err
	}

	defer rows.Close()

	for rows.Next() {
		var result PhotoSearchResult
		s.db.ScanRows(rows, &result)
		results = append(results, result)
	}

	// TODO: Check if this works properly with concurrent requests and caching
	count := &SearchCount{}
	s.db.Raw("SELECT FOUND_ROWS() AS total").Scan(&count)
	total := count.Total

	return results, total, nil
}

func (s *Search) FindFiles(count int, offset int) (files []File) {
	s.db.Where(&File{}).Limit(count).Offset(offset).Find(&files)

	return files
}

func (s *Search) FindFile(id string) (file File) {
	s.db.Where("id = ?", id).First(&file)

	return file
}
