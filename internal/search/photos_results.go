package search

import (
	"fmt"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/ulule/deepcopier"
)

// Photo represents a photo search result.
type Photo struct {
	ID               uint          `json:"-"`
	CompositeID      string        `json:"ID"`
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
	PhotoStack       int8          `json:"Stack"`
	PhotoFavorite    bool          `json:"Favorite"`
	PhotoPrivate     bool          `json:"Private"`
	PhotoIso         int           `json:"Iso"`
	PhotoFocalLength int           `json:"FocalLength"`
	PhotoFNumber     float32       `json:"FNumber"`
	PhotoExposure    string        `json:"Exposure"`
	PhotoFaces       int           `json:"Faces,omitempty"`
	PhotoQuality     int           `json:"Quality"`
	PhotoResolution  int           `json:"Resolution"`
	PhotoColor       uint8         `json:"Color"`
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
	InstanceID       string        `json:"InstanceID"`
	FileID           uint          `json:"-"` // File
	FileUID          string        `json:"FileUID"`
	FileRoot         string        `json:"FileRoot"`
	FileName         string        `json:"FileName"`
	FileHash         string        `json:"Hash"`
	FileWidth        int           `json:"Width"`
	FileHeight       int           `json:"Height"`
	FilePortrait     bool          `json:"Portrait"`
	FilePrimary      bool          `json:"-"`
	FileSidecar      bool          `json:"-"`
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

type PhotoResults []Photo

// UIDs returns a slice of photo UIDs.
func (m PhotoResults) UIDs() []string {
	result := make([]string, len(m))

	for i, el := range m {
		result[i] = el.PhotoUID
	}

	return result
}

// Merge consecutive file results that belong to the same photo.
func (m PhotoResults) Merge() (photos PhotoResults, count int, err error) {
	count = len(m)
	photos = make(PhotoResults, 0, count)

	var i int
	var photoId uint

	for _, photo := range m {
		file := entity.File{}

		if err = deepcopier.Copy(&file).From(photo); err != nil {
			return photos, count, err
		}

		file.ID = photo.FileID

		if photoId == photo.ID && i > 0 {
			photos[i-1].Files = append(photos[i-1].Files, file)
			photos[i-1].Merged = true
			continue
		}

		i++
		photoId = photo.ID
		photo.CompositeID = fmt.Sprintf("%d-%d", photoId, file.ID)
		photo.Files = append(photo.Files, file)

		photos = append(photos, photo)
	}

	return photos, count, nil
}

// ShareBase returns a meaningful file name for sharing.
func (m *Photo) ShareBase(seq int) string {
	var name string

	if m.PhotoTitle != "" {
		name = strings.Title(slug.MakeLang(m.PhotoTitle, "en"))
	} else {
		name = m.PhotoUID
	}

	taken := m.TakenAtLocal.Format("20060102-150405")

	if seq > 0 {
		return fmt.Sprintf("%s-%s (%d).%s", taken, name, seq, m.FileType)
	}

	return fmt.Sprintf("%s-%s.%s", taken, name, m.FileType)
}
