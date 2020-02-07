package meta

import (
	"time"
)

// Data represents image meta data.
type Data struct {
	UniqueID     string
	TakenAt      time.Time
	TakenAtLocal time.Time
	TimeZone     string
	Title        string
	Subject      string
	Keywords     string
	Comment      string
	Artist       string
	Description  string
	Copyright    string
	CameraMake   string
	CameraModel  string
	CameraOwner  string
	CameraSerial string
	LensMake     string
	LensModel    string
	Flash        bool
	FocalLength  int
	Exposure     string
	Aperture     float64
	FNumber      float64
	Iso          int
	Lat          float64
	Lng          float64
	Altitude     int
	Width        int
	Height       int
	Orientation  int
	All          map[string]string
}
