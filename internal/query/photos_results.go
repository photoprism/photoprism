package query

import (
	"fmt"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/ulule/deepcopier"
)

// PhotosResult contains found photos and their main file plus other meta data.
type PhotosResult struct {
	// Photo
	ID               uint
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        time.Time
	TakenAt          time.Time
	TakenAtLocal     time.Time
	TakenSrc         string
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
	PhotoVideo       bool
	PhotoLat         float32
	PhotoLng         float32
	PhotoAltitude    int
	PhotoIso         int
	PhotoFocalLength int
	PhotoFNumber     float32
	PhotoExposure    string
	PhotoQuality     int
	PhotoResolution  int
	Merged           bool

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
	FileVideo       bool
	FileDuration    time.Duration
	FileName        string
	FileHash        string
	FileCodec       string
	FileType        string
	FileMime        string
	FileWidth       int
	FileHeight      int
	FileSize        int64
	FileOrientation int
	FileAspectRatio float32
	FileColors      string // todo: remove from result?
	FileChroma      uint8  // todo: remove from result?
	FileLuminance   string // todo: remove from result?
	FileDiff        uint32 // todo: remove from result?

	Files []entity.File
}

type PhotosResults []PhotosResult

func (m PhotosResults) Merged() (PhotosResults, int, error) {
	count := len(m)
	merged := make([]PhotosResult, 0, count)

	var lastId uint
	var i int

	for _, res := range m {
		file := entity.File{}

		if err := deepcopier.Copy(&file).From(res); err != nil {
			return merged, count, err
		}

		file.ID = res.FileID

		if lastId == res.ID && i > 0 {
			merged[i-1].Files = append(merged[i-1].Files, file)
			merged[i-1].Merged = true
			continue
		}

		lastId = res.ID

		res.Files = append(res.Files, file)
		merged = append(merged, res)

		i++
	}

	return merged, count, nil
}

func (m *PhotosResult) ShareFileName() string {
	var name string

	if m.PhotoTitle != "" {
		name = strings.Title(slug.MakeLang(m.PhotoTitle, "en"))
	} else {
		name = m.PhotoUUID
	}

	taken := m.TakenAtLocal.Format("20060102-150405")
	token := rnd.Token(3)

	result := fmt.Sprintf("%s-%s-%s.%s", taken, name, token, m.FileType)

	return result
}
