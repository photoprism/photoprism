package search

import (
	"fmt"
	"time"

	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/ulule/deepcopier"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Photo represents a photo search result.
type Photo struct {
	ID               uint          `json:"-" select:"photos.id"`
	CompositeID      string        `json:"ID" select:"files.photo_id AS composite_id"`
	UUID             string        `json:"DocumentID,omitempty" select:"photos.uuid"`
	PhotoUID         string        `json:"UID" select:"photos.photo_uid"`
	PhotoType        string        `json:"Type" select:"photos.photo_type"`
	TypeSrc          string        `json:"TypeSrc" select:"photos.taken_src"`
	TakenAt          time.Time     `json:"TakenAt" select:"photos.taken_at"`
	TakenAtLocal     time.Time     `json:"TakenAtLocal" select:"photos.taken_at_local"`
	TakenSrc         string        `json:"TakenSrc" select:"photos.taken_src"`
	TimeZone         string        `json:"TimeZone" select:"photos.time_zone"`
	PhotoPath        string        `json:"Path" select:"photos.photo_path"`
	PhotoName        string        `json:"Name" select:"photos.photo_name"`
	OriginalName     string        `json:"OriginalName" select:"photos.original_name"`
	PhotoTitle       string        `json:"Title" select:"photos.photo_title"`
	PhotoDescription string        `json:"Description" select:"photos.photo_description"`
	PhotoYear        int           `json:"Year" select:"photos.photo_year"`
	PhotoMonth       int           `json:"Month" select:"photos.photo_month"`
	PhotoDay         int           `json:"Day" select:"photos.photo_day"`
	PhotoCountry     string        `json:"Country" select:"photos.photo_country"`
	PhotoStack       int8          `json:"Stack" select:"photos.photo_stack"`
	PhotoFavorite    bool          `json:"Favorite" select:"photos.photo_favorite"`
	PhotoPrivate     bool          `json:"Private" select:"photos.photo_private"`
	PhotoIso         int           `json:"Iso" select:"photos.photo_iso"`
	PhotoFocalLength int           `json:"FocalLength" select:"photos.photo_focal_length"`
	PhotoFNumber     float32       `json:"FNumber" select:"photos.photo_f_number"`
	PhotoExposure    string        `json:"Exposure" select:"photos.photo_exposure"`
	PhotoFaces       int           `json:"Faces,omitempty" select:"photos.photo_faces"`
	PhotoQuality     int           `json:"Quality" select:"photos.photo_quality"`
	PhotoResolution  int           `json:"Resolution" select:"photos.photo_resolution"`
	PhotoDuration    time.Duration `json:"Duration,omitempty" yaml:"photos.photo_duration"`
	PhotoColor       int16         `json:"Color" select:"photos.photo_color"`
	PhotoScan        bool          `json:"Scan" select:"photos.photo_scan"`
	PhotoPanorama    bool          `json:"Panorama" select:"photos.photo_panorama"`
	CameraID         uint          `json:"CameraID" select:"photos.camera_id"` // Camera
	CameraSrc        string        `json:"CameraSrc,omitempty" select:"photos.camera_src"`
	CameraSerial     string        `json:"CameraSerial,omitempty" select:"photos.camera_serial"`
	CameraMake       string        `json:"CameraMake,omitempty" select:"cameras.camera_make"`
	CameraModel      string        `json:"CameraModel,omitempty" select:"cameras.camera_model"`
	CameraType       string        `json:"CameraType,omitempty" select:"cameras.camera_type"`
	LensID           uint          `json:"LensID" select:"photos.lens_id"` // Lens
	LensMake         string        `json:"LensMake,omitempty" select:"lenses.lens_model"`
	LensModel        string        `json:"LensModel,omitempty" select:"lenses.lens_make"`
	PhotoAltitude    int           `json:"Altitude,omitempty" select:"photos.photo_altitude"`
	PhotoLat         float64       `json:"Lat" select:"photos.photo_lat"`
	PhotoLng         float64       `json:"Lng" select:"photos.photo_lng"`
	CellID           string        `json:"CellID" select:"photos.cell_id"` // Cell
	CellAccuracy     int           `json:"CellAccuracy,omitempty" select:"photos.cell_accuracy"`
	PlaceID          string        `json:"PlaceID" select:"photos.place_id"`
	PlaceSrc         string        `json:"PlaceSrc" select:"photos.place_src"`
	PlaceLabel       string        `json:"PlaceLabel" select:"places.place_label"`
	PlaceCity        string        `json:"PlaceCity" select:"places.place_city"`
	PlaceState       string        `json:"PlaceState" select:"places.place_state"`
	PlaceCountry     string        `json:"PlaceCountry" select:"places.place_country"`
	InstanceID       string        `json:"InstanceID" select:"files.instance_id"`
	FileID           uint          `json:"-" select:"files.id AS file_id"` // File
	FileUID          string        `json:"FileUID" select:"files.file_uid"`
	FileRoot         string        `json:"FileRoot" select:"files.file_root"`
	FileName         string        `json:"FileName" select:"files.file_name"`
	FileHash         string        `json:"Hash" select:"files.file_hash"`
	FileWidth        int           `json:"Width" select:"files.file_width"`
	FileHeight       int           `json:"Height" select:"files.file_height"`
	FilePortrait     bool          `json:"Portrait" select:"files.file_portrait"`
	FilePrimary      bool          `json:"-" select:"files.file_primary"`
	FileSidecar      bool          `json:"-" select:"files.file_sidecar"`
	FileMissing      bool          `json:"-" select:"files.file_missing"`
	FileVideo        bool          `json:"-" select:"files.file_video"`
	FileDuration     time.Duration `json:"-" select:"files.file_duration"`
	FileFPS          float64       `json:"-" select:"files.file_fps"`
	FileFrames       int           `json:"-" select:"files.file_frames"`
	FileCodec        string        `json:"-" select:"files.file_codec"`
	FileType         string        `json:"-" select:"files.file_type"`
	MediaType        string        `json:"-" select:"files.media_type"`
	FileMime         string        `json:"-" select:"files.file_mime"`
	FileSize         int64         `json:"-" select:"files.file_size"`
	FileOrientation  int           `json:"-" select:"files.file_orientation"`
	FileProjection   string        `json:"-" select:"files.file_projection"`
	FileAspectRatio  float32       `json:"-" select:"files.file_aspect_ratio"`
	FileColors       string        `json:"-" select:"files.file_colors"`
	FileDiff         int           `json:"-" select:"files.file_diff"`
	FileChroma       int16         `json:"-" select:"files.file_chroma"`
	FileLuminance    string        `json:"-" select:"files.file_luminance"`
	Merged           bool          `json:"Merged" select:"-"`
	CreatedAt        time.Time     `json:"CreatedAt" select:"photos.created_at"`
	UpdatedAt        time.Time     `json:"UpdatedAt" select:"photos.updated_at"`
	EditedAt         time.Time     `json:"EditedAt,omitempty" select:"photos.edited_at"`
	CheckedAt        time.Time     `json:"CheckedAt,omitempty" select:"photos.checked_at"`
	DeletedAt        *time.Time    `json:"DeletedAt,omitempty" select:"photos.deleted_at"`

	Files []entity.File `json:"Files"`
}

// GetID returns the numeric entity ID.
func (m *Photo) GetID() uint {
	return m.ID
}

// HasID checks if the photo has an id and uid assigned to it.
func (m *Photo) HasID() bool {
	return m.ID > 0 && m.PhotoUID != ""
}

// GetUID returns the unique entity id.
func (m *Photo) GetUID() string {
	return m.PhotoUID
}

// Approve approves the photo if it is in review.
func (m *Photo) Approve() error {
	if !m.HasID() {
		return fmt.Errorf("photo has no id")
	} else if m.PhotoQuality >= 3 {
		// Nothing to do.
		return nil
	}

	// Restore photo if archived.
	if err := m.Restore(); err != nil {
		return err
	}

	edited := entity.Now()

	if err := UnscopedDb().
		Table(entity.Photo{}.TableName()).
		Where("photo_uid = ?", m.GetUID()).
		UpdateColumns(entity.Map{
			"deleted_at":    gorm.Expr("NULL"),
			"edited_at":     &edited,
			"photo_quality": 3}).Error; err != nil {
		return err
	}

	m.EditedAt = edited
	m.PhotoQuality = 3
	m.DeletedAt = nil

	// Update precalculated photo and file counts.
	if err := entity.UpdateCounts(); err != nil {
		log.Warnf("index: %s (update counts)", err)
	}

	event.Publish("count.review", event.Data{
		"count": -1,
	})

	return nil
}

// Restore removes the photo from the archive (reverses soft delete).
func (m *Photo) Restore() error {
	if !m.HasID() {
		return fmt.Errorf("photo has no id")
	} else if m.DeletedAt == nil {
		return nil
	}

	if err := UnscopedDb().
		Table(entity.Photo{}.TableName()).
		Where("photo_uid = ?", m.GetUID()).
		UpdateColumn("deleted_at", gorm.Expr("NULL")).Error; err != nil {
		return err
	}

	m.DeletedAt = nil

	return nil
}

// IsPlayable returns true if the photo has a related video/animation that is playable.
func (m *Photo) IsPlayable() bool {
	switch m.PhotoType {
	case entity.MediaVideo, entity.MediaLive, entity.MediaAnimated:
		return true
	default:
		return false
	}
}

// ShareBase returns a meaningful file name for sharing.
func (m *Photo) ShareBase(seq int) string {
	var name string

	if m.PhotoTitle != "" {
		name = txt.Title(slug.MakeLang(m.PhotoTitle, "en"))
	} else {
		name = m.PhotoUID
	}

	taken := m.TakenAtLocal.Format("20060102-150405")

	if seq > 0 {
		return fmt.Sprintf("%s-%s (%d).%s", taken, name, seq, m.FileType)
	}

	return fmt.Sprintf("%s-%s.%s", taken, name, m.FileType)
}

type PhotoResults []Photo

// Photos returns the result as a slice of Photo.
func (m PhotoResults) Photos() []entity.PhotoInterface {
	result := make([]entity.PhotoInterface, len(m))

	for i := range m {
		result[i] = &m[i]
	}

	return result
}

// UIDs returns a slice of photo UIDs.
func (m PhotoResults) UIDs() []string {
	result := make([]string, len(m))

	for i, el := range m {
		result[i] = el.PhotoUID
	}

	return result
}

// Merge consecutive file results that belong to the same photo.
func (m PhotoResults) Merge() (merged PhotoResults, count int, err error) {
	count = len(m)
	merged = make(PhotoResults, 0, count)

	var i int
	var photoId uint

	for _, photo := range m {
		file := entity.File{}

		if err = deepcopier.Copy(&file).From(photo); err != nil {
			return merged, count, err
		}

		file.ID = photo.FileID

		if photoId == photo.ID && i > 0 {
			merged[i-1].Files = append(merged[i-1].Files, file)
			merged[i-1].Merged = true
			continue
		}

		i++
		photoId = photo.ID
		photo.CompositeID = fmt.Sprintf("%d-%d", photoId, file.ID)
		photo.Files = append(photo.Files, file)

		merged = append(merged, photo)
	}

	return merged, count, nil
}
