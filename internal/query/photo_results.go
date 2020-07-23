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

// PhotoResult contains found photos and their main file plus other meta data.
type PhotoResult struct {
	ID               uint          `json:"-"`
	UUID             string        `json:"DocumentID,omitempty"`
	PhotoUID         string        `json:"UID"`
	PhotoType        string        `json:"Type"`
	TypeSrc          string        `json:"TypeSrc"`
	TakenAt          time.Time     `json:"TakenAt"`
	TakenAtLocal     time.Time     `json:"TakenAtLocal"`
	TakenSrc         string        `json:"TakenSrc"`
	TimeZone         string        `json:"TimeZone"`
	PhotoPath        string        `json:"Path"`
	PhotoName        string        `json:"Name"`
	OriginalName     string        `json:"OriginalName"`
	PhotoTitle       string        `json:"Title"`
	PhotoDescription string        `json:"Description"`
	PhotoYear        int           `json:"Year"`
	PhotoMonth       int           `json:"Month"`
	PhotoDay         int           `json:"Day"`
	PhotoCountry     string        `json:"Country"`
	PhotoFavorite    bool          `json:"Favorite"`
	PhotoPrivate     bool          `json:"Private"`
	PhotoIso         int           `json:"Iso"`
	PhotoFocalLength int           `json:"FocalLength"`
	PhotoFNumber     float32       `json:"FNumber"`
	PhotoExposure    string        `json:"Exposure"`
	PhotoQuality     int           `json:"Quality"`
	PhotoResolution  int           `json:"Resolution"`
	PhotoScan        bool          `json:"Scan"`
	PhotoPanorama    bool          `json:"Panorama"`
	CameraID         uint          `json:"CameraID"` // Camera
	CameraSerial     string        `json:"CameraSerial,omitempty"`
	CameraSrc        string        `json:"CameraSrc,omitempty"`
	CameraModel      string        `json:"CameraModel"`
	CameraMake       string        `json:"CameraMake"`
	LensID           uint          `json:"LensID"` // Lens
	LensModel        string        `json:"LensModel"`
	LensMake         string        `json:"LensMake"`
	PhotoAltitude    int           `json:"Altitude,omitempty"`
	PhotoLat         float32       `json:"Lat"`
	PhotoLng         float32       `json:"Lng"`
	CellID           string        `json:"CellID"` // Cell
	CellAccuracy     int           `json:"CellAccuracy,omitempty"`
	PlaceID          string        `json:"PlaceID"`
	PlaceSrc         string        `json:"PlaceSrc"`
	PlaceLabel       string        `json:"PlaceLabel"`
	PlaceCity        string        `json:"PlaceCity"`
	PlaceState       string        `json:"PlaceState"`
	PlaceCountry     string        `json:"PlaceCountry"`
	FileID           uint          `json:"-"` // File
	FileUID          string        `json:"FileUID"`
	FileRoot         string        `json:"FileRoot"`
	FileName         string        `json:"FileName"`
	FileHash         string        `json:"Hash"`
	FileWidth        int           `json:"Width"`
	FileHeight       int           `json:"Height"`
	FilePrimary      bool          `json:"-"`
	FileMissing      bool          `json:"-"`
	FileVideo        bool          `json:"-"`
	FileDuration     time.Duration `json:"-"`
	FileCodec        string        `json:"-"`
	FileType         string        `json:"-"`
	FileMime         string        `json:"-"`
	FileSize         int64         `json:"-"`
	FileOrientation  int           `json:"-"`
	FileProjection   string        `json:"-"`
	FileAspectRatio  float32       `json:"-"`
	FileColors       string        `json:"-"`
	FileChroma       uint8         `json:"-"`
	FileLuminance    string        `json:"-"`
	FileDiff         uint32        `json:"-"`
	Merged           bool          `json:"Merged"`
	CreatedAt        time.Time     `json:"CreatedAt"`
	UpdatedAt        time.Time     `json:"UpdatedAt"`
	EditedAt         time.Time     `json:"EditedAt,omitempty"`
	CheckedAt        time.Time     `json:"CheckedAt,omitempty"`
	DeletedAt        time.Time     `json:"DeletedAt,omitempty"`

	Files []entity.File `json:"Files"`
}

type PhotoResults []PhotoResult

// UIDs returns a slice of photo UIDs.
func (m PhotoResults) UIDs() []string {
	result := make([]string, len(m))

	for i, el := range m {
		result[i] = el.PhotoUID
	}

	return result
}

func (m PhotoResults) Merged() (PhotoResults, int, error) {
	count := len(m)
	merged := make([]PhotoResult, 0, count)

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

func (m *PhotoResult) ShareFileName() string {
	var name string

	if m.PhotoTitle != "" {
		name = strings.Title(slug.MakeLang(m.PhotoTitle, "en"))
	} else {
		name = m.PhotoUID
	}

	taken := m.TakenAtLocal.Format("20060102-150405")
	token := rnd.Token(3)

	result := fmt.Sprintf("%s-%s-%s.%s", taken, name, token, m.FileType)

	return result
}
