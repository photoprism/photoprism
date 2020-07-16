package meta

import (
	"math"
	"time"
)

// Data represents image meta data.
type Data struct {
	DocumentID   string        `meta:"ImageUniqueID,OriginalDocumentID,DocumentID"`
	InstanceID   string        `meta:"InstanceID,DocumentID"`
	TakenAt      time.Time     `meta:"DateTimeOriginal,CreateDate,MediaCreateDate,DateTimeDigitized,DateTime"`
	TakenAtLocal time.Time     `meta:"DateTimeOriginal,CreateDate,MediaCreateDate,DateTimeDigitized,DateTime"`
	TimeZone     string        `meta:"-"`
	Duration     time.Duration `meta:"Duration,MediaDuration,TrackDuration"`
	Codec        string        `meta:"CompressorID,Compression,FileType"`
	Title        string        `meta:"Title"`
	Subject      string        `meta:"Subject,PersonInImage,ObjectName"`
	Keywords     string        `meta:"Keywords"`
	Comment      string        `meta:"-"`
	Artist       string        `meta:"Artist,Creator"`
	Description  string        `meta:"Description"`
	Copyright    string        `meta:"Rights,Copyright"`
	Projection   string        `meta:"ProjectionType"`
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
	Width        int           `meta:"PixelXDimension,ImageWidth,ExifImageWidth,SourceImageWidth"`
	Height       int           `meta:"PixelYDimension,ImageHeight,ImageLength,ExifImageHeight,SourceImageHeight"`
	Orientation  int           `meta:"-"`
	Rotation     int           `meta:"Rotation"`
	Views        int           `meta:"-"`
	Albums       []string      `meta:"-"`
	Error        error         `meta:"-"`
	All          map[string]string
}

// NewData creates a new metadata struct.
func NewData() Data {
	return Data{
		All: make(map[string]string),
	}
}

// AspectRatio returns the aspect ratio based on width and height.
func (data Data) AspectRatio() float32 {
	width := float64(data.ActualWidth())
	height := float64(data.ActualHeight())

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

// HasDocumentID returns true if a DocumentID exists.
func (data Data) HasDocumentID() bool {
	return len(data.DocumentID) >= 15
}

// HasInstanceID returns true if an InstanceID exists.
func (data Data) HasInstanceID() bool {
	return len(data.InstanceID) >= 15
}

// HasTimeAndPlace if data contains a time and gps position.
func (data Data) HasTimeAndPlace() bool {
	return !data.TakenAt.IsZero() && data.Lat != 0 && data.Lng != 0
}

// ActualWidth is the width after rotating the media file if needed.
func (data Data) ActualWidth() int {
	if data.Orientation > 4 {
		return data.Height
	}

	return data.Width
}

// ActualHeight is the height after rotating the media file if needed.
func (data Data) ActualHeight() int {
	if data.Orientation > 4 {
		return data.Width
	}

	return data.Height
}
