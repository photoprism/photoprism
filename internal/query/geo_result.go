package query

import (
	"time"
)

// GeoResult represents a photo for displaying it on a map.
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
	TakenAt          time.Time `json:"TakenAt"`
}

func (g GeoResult) Lat() float64 {
	return float64(g.PhotoLat)
}

func (g GeoResult) Lng() float64 {
	return float64(g.PhotoLng)
}

type GeoResults []GeoResult
