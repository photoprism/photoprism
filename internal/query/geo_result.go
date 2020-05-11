package query

import (
	"time"
)

// GeoResult represents a photo for displaying it on a map.
type GeoResult struct {
	ID            string    `json:"ID"`
	PhotoLat      float32   `json:"Lat"`
	PhotoLng      float32   `json:"Lng"`
	PhotoUUID     string    `json:"PhotoUUID"`
	PhotoTitle    string    `json:"PhotoTitle"`
	PhotoFavorite bool      `json:"PhotoFavorite"`
	FileHash      string    `json:"FileHash"`
	FileWidth     int       `json:"FileWidth"`
	FileHeight    int       `json:"FileHeight"`
	TakenAt       time.Time `json:"TakenAt"`
}

func (g GeoResult) Lat() float64 {
	return float64(g.PhotoLat)
}

func (g GeoResult) Lng() float64 {
	return float64(g.PhotoLng)
}

type GeoResults []GeoResult
