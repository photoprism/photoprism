package search

import (
	"time"

	"github.com/gin-gonic/gin"
	geojson "github.com/paulmach/go.geojson"

	"github.com/photoprism/photoprism/internal/entity"
)

// GeoResult represents a photo geo search result.
type GeoResult struct {
	ID            string        `json:"-" select:"photos.id"`
	PhotoUID      string        `json:"UID" select:"photos.photo_uid"`
	PhotoType     string        `json:"Type,omitempty" select:"photos.photo_type"`
	PhotoTitle    string        `json:"Title" select:"photos.photo_title"`
	PhotoCaption  string        `json:"Caption,omitempty" select:"photos.photo_caption"`
	PhotoLat      float64       `json:"Lat" select:"photos.photo_lat"`
	PhotoLng      float64       `json:"Lng" select:"photos.photo_lng"`
	TakenAt       time.Time     `json:"TakenAt" select:"photos.taken_at"`
	TakenAtLocal  time.Time     `json:"TakenAtLocal" select:"photos.taken_at_local"`
	PhotoFavorite bool          `json:"Favorite,omitempty" select:"photos.photo_favorite"`
	PhotoDuration time.Duration `json:"Duration,omitempty" select:"photos.photo_duration"`
	FileID        uint          `json:"-" select:"files.id AS file_id"` // File
	FileWidth     int           `json:"Width" select:"files.file_width"`
	FileHeight    int           `json:"Height" select:"files.file_height"`
	FileHash      string        `json:"Hash" select:"files.file_hash"`
	FileCodec     string        `json:"-" select:"files.file_codec"`
	FileMime      string        `json:"-" select:"files.file_mime"`
	FileVideo     bool          `json:"-" select:"files.file_video"`
	MediaType     string        `json:"-" select:"files.media_type"`
}

// Lat returns the position latitude.
func (m GeoResult) Lat() float64 {
	return m.PhotoLat
}

// Lng returns the position longitude.
func (m GeoResult) Lng() float64 {
	return m.PhotoLng
}

// IsPlayable returns true if the photo has a related video/animation that is playable.
func (m GeoResult) IsPlayable() bool {
	switch m.PhotoType {
	case entity.MediaVideo, entity.MediaLive, entity.MediaAnimated:
		return true
	default:
		return false
	}
}

// GeoResults represents a list of geo search results.
type GeoResults []GeoResult

// GeoJSON returns results as specified on https://geojson.org/.
func (photos GeoResults) GeoJSON() ([]byte, error) {
	fc := geojson.NewFeatureCollection()

	bbox := make([]float64, 4)

	bboxMin := func(pos int, val float64) {
		if bbox[pos] == 0.0 || bbox[pos] > val {
			bbox[pos] = val
		}
	}

	bboxMax := func(pos int, val float64) {
		if bbox[pos] == 0.0 || bbox[pos] < val {
			bbox[pos] = val
		}
	}

	for _, p := range photos {
		bboxMin(0, p.Lng())
		bboxMin(1, p.Lat())
		bboxMax(2, p.Lng())
		bboxMax(3, p.Lat())

		props := gin.H{
			"UID":     p.PhotoUID,
			"Hash":    p.FileHash,
			"TakenAt": p.TakenAt,
			"Title":   p.PhotoTitle,
		}

		if p.PhotoType != entity.MediaImage && p.PhotoType != entity.MediaUnknown {
			props["Type"] = p.PhotoType
		}

		if p.PhotoFavorite {
			props["Favorite"] = true
		}

		feat := geojson.NewPointFeature([]float64{p.Lng(), p.Lat()})
		feat.ID = p.ID
		feat.Properties = props
		fc.AddFeature(feat)
	}

	fc.BoundingBox = bbox

	return fc.MarshalJSON()
}
