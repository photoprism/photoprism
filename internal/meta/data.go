package meta

import (
	"time"

	model "trimmer.io/go-xmp/models/exif"
	"trimmer.io/go-xmp/xmp"
)

// Data represents image meta data.
type Data struct {
	UUID         string
	TakenAt      time.Time
	TakenAtLocal time.Time
	TimeZone     string
	Artist       string
	Copyright    string
	CameraMake   string
	CameraModel  string
	Description  string
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
	Xmp          *xmp.Document
	Exif         *model.ExifInfo
	ExifEX       *model.ExifEXInfo
	ExifAux      *model.ExifAuxInfo
}
