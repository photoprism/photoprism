package search

import (
	"time"

	"github.com/gin-gonic/gin"
	geojson "github.com/paulmach/go.geojson"

	"github.com/photoprism/photoprism/internal/entity"
)

// GeoResult represents a photo geo search result.
type GeoResult struct {
	ID               string    `json:"-"`
	PhotoUID         string    `json:"UID"`
	PhotoType        string    `json:"Type,omitempty"`
	PhotoLat         float32   `json:"Lat"`
	PhotoLng         float32   `json:"Lng"`
	PhotoTitle       string    `json:"Title"`
	PhotoDescription string    `json:"Description,omitempty"`
	PhotoFavorite    bool      `json:"Favorite,omitempty"`
	FileHash         string    `json:"Hash"`
	FileWidth        int       `json:"Width"`
	FileHeight       int       `json:"Height"`
	TakenAtLocal     time.Time `json:"TakenAt"`
}

// Lat returns the position latitude.
func (photo GeoResult) Lat() float64 {
	return float64(photo.PhotoLat)
}

// Lng returns the position longitude.
func (photo GeoResult) Lng() float64 {
	return float64(photo.PhotoLng)
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
			"Width":   p.FileWidth,
			"Height":  p.FileHeight,
			"TakenAt": p.TakenAtLocal,
			"Title":   p.PhotoTitle,
		}

		if p.PhotoDescription != "" {
			props["Description"] = p.PhotoDescription
		}

		if p.PhotoType != entity.TypeImage && p.PhotoType != entity.TypeDefault {
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
