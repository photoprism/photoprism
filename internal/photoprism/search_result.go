package photoprism

import (
	"fmt"
	"strings"
	"time"

	"github.com/gosimple/slug"
)

// PhotoSearchResult contains found photos and their main file plus other meta data.
type PhotoSearchResult struct {
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
	PhotoArtist      string
	PhotoKeywords    string
	PhotoColors      string
	PhotoColor       string
	PhotoFavorite    bool
	PhotoPrivate     bool
	PhotoSensitive   bool
	PhotoStory       bool
	PhotoLat         float64
	PhotoLong        float64
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

	// Country
	CountryID   string
	CountryName string

	// Location
	LocationID        uint
	LocDisplayName    string
	LocName           string
	LocCity           string
	LocPostcode       string
	LocCounty         string
	LocState          string
	LocCountry        string
	LocCountryCode    string
	LocCategory       string
	LocType           string
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

	// List of matching labels and keywords
	Labels string
	Keywords string
}

func (m *PhotoSearchResult) DownloadFileName() string {
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

// LabelSearchResult contains found labels
type LabelSearchResult struct {
	// Label
	ID               uint
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        time.Time
	LabelSlug        string
	LabelName        string
	LabelPriority    int
	LabelCount       int
	LabelFavorite    bool
	LabelDescription string
	LabelNotes       string
}

// AlbumSearchResult contains found albums
type AlbumSearchResult struct {
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
