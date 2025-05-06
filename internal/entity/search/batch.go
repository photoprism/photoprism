package search

import (
	"strings"
	"time"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/txt"
)

// BatchResult represents a photo geo search result.
type BatchResult struct {
	ID               uint          `json:"-" select:"photos.id"`
	CompositeID      string        `json:"ID,omitempty" select:"files.photo_id AS composite_id"`
	UUID             string        `json:"DocumentID,omitempty" select:"photos.uuid"`
	PhotoUID         string        `json:"UID" select:"photos.photo_uid"`
	PhotoType        string        `json:"Type" select:"photos.photo_type"`
	TypeSrc          string        `json:"TypeSrc" select:"photos.taken_src"`
	PhotoTitle       string        `json:"Title" select:"photos.photo_title"`
	PhotoCaption     string        `json:"Caption,omitempty" select:"photos.photo_caption"`
	TakenAt          time.Time     `json:"TakenAt" select:"photos.taken_at"`
	TakenAtLocal     time.Time     `json:"TakenAtLocal" select:"photos.taken_at_local"`
	TimeZone         string        `json:"TimeZone" select:"photos.time_zone"`
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

	FileID          uint          `json:"-" select:"files.id AS file_id"` // File
	FileUID         string        `json:"FileUID" select:"files.file_uid"`
	FileRoot        string        `json:"FileRoot" select:"files.file_root"`
	FileName        string        `json:"FileName" select:"files.file_name"`
	OriginalName    string        `json:"OriginalName" select:"files.original_name"`
	FileHash        string        `json:"Hash" select:"files.file_hash"`
	FileWidth       int           `json:"Width" select:"files.file_width"`
	FileHeight      int           `json:"Height" select:"files.file_height"`
	FilePortrait    bool          `json:"Portrait" select:"files.file_portrait"`
	FilePrimary     bool          `json:"-" select:"files.file_primary"`
	FileSidecar     bool          `json:"-" select:"files.file_sidecar"`
	FileMissing     bool          `json:"-" select:"files.file_missing"`
	FileVideo       bool          `json:"-" select:"files.file_video"`
	FileDuration    time.Duration `json:"-" select:"files.file_duration"`
	FileFPS         float64       `json:"-" select:"files.file_fps"`
	FileFrames      int           `json:"-" select:"files.file_frames"`
	FilePages       int           `json:"-" select:"files.file_pages"`
	FileCodec       string        `json:"-" select:"files.file_codec"`
	FileType        string        `json:"-" select:"files.file_type"`
	MediaType       string        `json:"-" select:"files.media_type"`
	FileMime        string        `json:"-" select:"files.file_mime"`
	FileSize        int64         `json:"-" select:"files.file_size"`
	FileOrientation int           `json:"-" select:"files.file_orientation"`
	FileProjection  string        `json:"-" select:"files.file_projection"`
	FileAspectRatio float32       `json:"-" select:"files.file_aspect_ratio"`

	DetailsKeywords  string `json:"DetailsKeywords" select:"details.keywords AS details_keywords"`
	DetailsSubject   string `json:"DetailsSubject" select:"details.subject AS details_subject"`
	DetailsArtist    string `json:"DetailsArtist" select:"details.artist AS details_artist"`
	DetailsCopyright string `json:"DetailsCopyright" select:"details.copyright AS details_copyright"`
	DetailsLicense   string `json:"DetailsLicense" select:"details.license AS details_license"`
}

// BatchCols contains the result column names necessary for the photo viewer.
var BatchCols = SelectString(BatchResult{}, SelectCols(BatchResult{}, []string{"*"}))

// BatchPhotos finds PhotoResults based on the search form without checking rights or permissions.
func BatchPhotos(uids []string, sess *entity.Session) (results PhotoResults, count int, err error) {
	frm := form.SearchPhotos{
		UID:     strings.Join(uids, txt.Or),
		Count:   MaxResults,
		Offset:  0,
		Face:    "",
		Merged:  true,
		Details: true,
	}

	return searchPhotos(frm, sess, BatchCols)
}
