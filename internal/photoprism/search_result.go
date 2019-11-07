package photoprism

import "time"

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

	// List of matching labels (tags)
	Labels string
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
