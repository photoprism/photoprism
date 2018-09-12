package photoprism

import (
	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/forms"
	"time"
)

type Search struct {
	originalsPath string
	db            *gorm.DB
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
}

func NewQuery(originalsPath string, db *gorm.DB) *Search {
	instance := &Search{
		originalsPath: originalsPath,
		db:            db,
	}

	return instance
}

func (s *Search) Photos(form forms.PhotoSearchForm) ([]PhotoSearchResult, error) {
	q := s.db.Preload("Tags").Preload("Files").Preload("Location").Preload("Albums")

	q = q.Table("photos").
		Select(`photos.*,
		files.id AS file_id, files.file_name, files.file_type, files.file_mime, files.file_width, files.file_height, files.file_aspect_ratio, files.file_orientation,
		cameras.camera_model,
		locations.loc_display_name, locations.loc_name, locations.loc_city, locations.loc_postcode, locations.loc_country, locations.loc_country_code, locations.loc_category, locations.loc_type`).
		Joins("JOIN files ON files.photo_id = photos.id AND files.file_primary AND files.deleted_at IS NULL").
		Joins("JOIN cameras ON cameras.id = photos.camera_id").
		Joins("LEFT JOIN locations ON locations.id = photos.location_id").
		Where("photos.deleted_at IS NULL")

	if form.Query != "" {
		q = q.Where("MATCH (photo_title, photo_description, photo_artist, photo_keywords, photo_colors) AGAINST (? IN BOOLEAN MODE)", form.Query)
	}

	q = q.Order("taken_at").Limit(form.Count).Offset(form.Offset)

	results := make([]PhotoSearchResult, 0, form.Count)

	rows, err := q.Rows()

	if err != nil {
		return results, err
	}

	defer rows.Close()

	for rows.Next() {
		var result PhotoSearchResult
		s.db.ScanRows(rows, &result)
		results = append(results, result)
	}

	return results, nil
}

func (s *Search) FindFiles(count int, offset int) (files []File) {
	s.db.Where(&File{}).Limit(count).Offset(offset).Find(&files)

	return files
}

func (s *Search) FindFile(id string) (file File) {
	s.db.Where("id = ?", id).First(&file)

	return file
}
