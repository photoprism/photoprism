package meta

import (
	"math"
	"time"
)

// Data represents image meta data.
type Data struct {
	UniqueID     string        `meta:"ImageUniqueID"`
	TakenAt      time.Time     `meta:"DateTimeOriginal,CreateDate,MediaCreateDate,DateTimeDigitized,DateTime"`
	TakenAtLocal time.Time     `meta:"DateTimeOriginal,CreateDate,MediaCreateDate,DateTimeDigitized,DateTime"`
	TimeZone     string        `meta:"-"`
	Duration     time.Duration `meta:"Duration,MediaDuration,TrackDuration"`
	Codec        string        `meta:"CompressorID,Compression"`
	Title        string        `meta:"Title"`
	Subject      string        `meta:"Subject,PersonInImage"`
	Keywords     string        `meta:"Keywords"`
	Comment      string        `meta:"-"`
	Artist       string        `meta:"Artist,Creator"`
	Description  string        `meta:"Description"`
	Copyright    string        `meta:"Rights,Copyright"`
	CameraMake   string        `meta:"CameraMake,Make"`
	CameraModel  string        `meta:"CameraModel,Model"`
	CameraOwner  string        `meta:"OwnerName"`
	CameraSerial string        `meta:"SerialNumber"`
	LensMake     string        `meta:"LensMake"`
	LensModel    string        `meta:"Lens,LensModel"`
	Flash        bool          `meta:"-"`
	FocalLength  int           `meta:"-"`
	Exposure     string        `meta:"ExposureTime"`
	Aperture     float32       `meta:"ApertureValue"`
	FNumber      float32       `meta:"FNumber"`
	Iso          int           `meta:"ISO"`
	GPSPosition  string        `meta:"GPSPosition"`
	GPSLatitude  string        `meta:"GPSLatitude"`
	GPSLongitude string        `meta:"GPSLongitude"`
	Lat          float32       `meta:"-"`
	Lng          float32       `meta:"-"`
	Altitude     int           `meta:"GlobalAltitude"`
	Width        int           `meta:"ImageWidth"`
	Height       int           `meta:"ImageHeight"`
	Orientation  int           `meta:"-"`
	Rotation     int           `meta:"Rotation"`
	All          map[string]string
}

// AspectRatio returns the aspect ratio based on width and height.
func (data Data) AspectRatio() float32 {
	width := float64(data.Width)
	height := float64(data.Height)

	if width <= 0 || height <= 0 {
		return 0
	}

	aspectRatio := float32(width / height)

	return aspectRatio
}

// Portrait returns true if it's a portrait picture or video based on width and height.
func (data Data) Portrait() bool {
	return data.Width < data.Height
}

// Megapixels returns the resolution in megapixels.
func (data Data) Megapixels() int {
	return int(math.Round(float64(data.Width*data.Height) / 1000000))
}
