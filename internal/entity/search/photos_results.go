package search

import (
	"fmt"
	"time"

	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/ulule/deepcopier"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/media/video"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Photo represents a photo search result row joined with its primary file and
// related metadata that we surface in the UI and API responses.
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
	PhotoTitle       string        `json:"Title" select:"photos.photo_title"`
	PhotoCaption     string        `json:"Caption" select:"photos.photo_caption"`
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
	PhotoDuration    time.Duration `json:"Duration,omitempty" select:"photos.photo_duration"`
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
	OriginalName     string        `json:"OriginalName" select:"files.original_name"`
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
	FilePages        int           `json:"-" select:"files.file_pages"`
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

	// Additional information from the details table.
	DetailsKeywords  string `json:"DetailsKeywords,omitempty" select:"-"`
	DetailsSubject   string `json:"DetailsSubject,omitempty" select:"-"`
	DetailsArtist    string `json:"DetailsArtist,omitempty" select:"-"`
	DetailsCopyright string `json:"DetailsCopyright,omitempty" select:"-"`
	DetailsLicense   string `json:"DetailsLicense,omitempty" select:"-"`

	// List of files if search results are merged.
	Files []entity.File `json:"Files" select:"-"`
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

// String returns the id or name as string for logging purposes.
func (m *Photo) String() string {
	if m == nil {
		return "Photo<nil>"
	}

	return entity.PhotoLogString(m.PhotoPath, m.PhotoName, m.OriginalName, m.PhotoUID, m.ID)
}

// Approve promotes the photo to quality level 3 and clears review flags if it
// currently sits in review state.
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
		UpdateColumns(entity.Values{
			"deleted_at":    gorm.Expr("NULL"),
			"edited_at":     &edited,
			"photo_quality": 3}).Error; err != nil {
		return err
	}

	m.EditedAt = edited
	m.PhotoQuality = 3
	m.DeletedAt = nil

	// Update precalculated photo and file counts.
	entity.UpdateCountsAsync()

	event.Publish("count.review", event.Data{
		"count": -1,
	})

	return nil
}

// Restore removes the photo from the archive by clearing the soft-delete flag.
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
	case entity.MediaLive, entity.MediaVideo, entity.MediaAudio, entity.MediaAnimated:
		return true
	default:
		return false
	}
}

// MediaInfo returns the best available media hash, codec, mime type, and
// dimensions for the photo based on its media type and merged files.
func (m *Photo) MediaInfo() (mediaHash, mediaCodec, mediaMime string, width, height int) {
	switch m.PhotoType {
	case entity.MediaVideo, entity.MediaLive:
		for _, f := range m.Files {
			if f.FileVideo && f.FileHash != "" {
				return f.FileHash, f.FileCodec, video.ContentType(f.FileMime, f.FileType, f.FileCodec, f.IsHDR()), f.FileWidth, f.FileHeight
			}
		}
	case entity.MediaImage:
		for _, f := range m.Files {
			if f.MediaType == entity.MediaImage && f.FileCodec != "jpeg" && f.FileHash != "" {
				return f.FileHash, f.FileCodec, clean.ContentType(f.FileMime), f.FileWidth, f.FileHeight
			}
		}
		return m.FileHash, m.FileCodec, m.FileMime, m.FileWidth, m.FileHeight
	case entity.MediaAnimated:
		for _, f := range m.Files {
			if f.MediaType == entity.MediaImage && (f.FileType == "gif" || f.FileType == "png" || f.FileDuration > 0 || f.FileFrames > 0) {
				return f.FileHash, f.FileCodec, clean.ContentType(f.FileMime), f.FileWidth, f.FileHeight
			}
		}
	case entity.MediaRaw:
		for _, f := range m.Files {
			if f.MediaType == entity.MediaRaw && f.FileWidth > 0 && f.FileHeight > 0 {
				return m.FileHash, f.FileCodec, clean.ContentType(f.FileMime), f.FileWidth, f.FileHeight
			}
		}
	case entity.MediaVector:
		for _, f := range m.Files {
			if f.MediaType == entity.MediaVector && f.FileHash != "" {
				return f.FileHash, f.FileCodec, clean.ContentType(f.FileMime), f.FileWidth, f.FileHeight
			}
		}
	case entity.MediaDocument:
		for _, f := range m.Files {
			if f.MediaType == entity.MediaDocument && f.FileHash != "" {
				return f.FileHash, f.FileCodec, clean.ContentType(f.FileMime), f.FileWidth, f.FileHeight
			}
		}
	}

	return m.FileHash, "", m.FileMime, m.FileWidth, m.FileHeight
}

// ShareBase returns a deterministic, human friendly file name stem for sharing
// downloads generated from the photo's timestamp and title.
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

// PhotoResults represents a list of photo search results that can be post
// processed (for example merged by file).
type PhotoResults []Photo

// Photos returns the results as a slice of the generic PhotoInterface type so
// callers can interact with shared entity helpers.
func (m PhotoResults) Photos() []entity.PhotoInterface {
	result := make([]entity.PhotoInterface, len(m))

	for i := range m {
		result[i] = &m[i]
	}

	return result
}

// UIDs returns the photo UIDs for all results in order.
func (m PhotoResults) UIDs() []string {
	result := make([]string, len(m))

	for i, el := range m {
		result[i] = el.PhotoUID
	}

	return result
}

// Merge collapses consecutive rows that reference the same photo into a single
// item with an aggregated Files slice.
func (m PhotoResults) Merge() (merged PhotoResults, count int, err error) {
	count = len(m)
	merged = make(PhotoResults, 0, count)

	var i int
	var photoId uint

	for _, photo := range m {
		file := entity.File{OmitMarkers: true}

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
