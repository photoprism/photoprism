package meta

import (
	"math"
	"time"

	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/s2"
)

const (
	ImageTypeHDR = 3 // see https://exiftool.org/TagNames/Apple.html
)

// Data represents image meta data.
type Data struct {
	FileName     string        `meta:"FileName"`
	DocumentID   string        `meta:"BurstUUID,MediaGroupUUID,ImageUniqueID,OriginalDocumentID,DocumentID"`
	InstanceID   string        `meta:"InstanceID,DocumentID"`
	TakenAt      time.Time     `meta:"DateTimeOriginal,CreationDate,CreateDate,MediaCreateDate,ContentCreateDate,DateTimeDigitized,DateTime"`
	TakenAtLocal time.Time     `meta:"DateTimeOriginal,CreationDate,CreateDate,MediaCreateDate,ContentCreateDate,DateTimeDigitized,DateTime"`
	TimeZone     string        `meta:"-"`
	Duration     time.Duration `meta:"Duration,MediaDuration,TrackDuration"`
	Codec        string        `meta:"CompressorID,FileType"`
	Title        string        `meta:"Title"`
	Subject      string        `meta:"Subject,PersonInImage,ObjectName,HierarchicalSubject,CatalogSets"`
	Keywords     Keywords      `meta:"Keywords"`
	Notes        string        `meta:"-"`
	Artist       string        `meta:"Artist,Creator,OwnerName"`
	Description  string        `meta:"Description"`
	Copyright    string        `meta:"Rights,Copyright"`
	Projection   string        `meta:"ProjectionType"`
	ColorProfile string        `meta:"ICCProfileName,ProfileDescription"`
	CameraMake   string        `meta:"CameraMake,Make"`
	CameraModel  string        `meta:"CameraModel,Model"`
	CameraOwner  string        `meta:"OwnerName"`
	CameraSerial string        `meta:"SerialNumber"`
	LensMake     string        `meta:"LensMake"`
	LensModel    string        `meta:"Lens,LensModel"`
	Flash        bool          `meta:"-"`
	FocalLength  int           `meta:"FocalLength"`
	Exposure     string        `meta:"ExposureTime"`
	Aperture     float32       `meta:"ApertureValue"`
	FNumber      float32       `meta:"FNumber"`
	Iso          int           `meta:"ISO"`
	ImageType    int           `meta:"HDRImageType"`
	GPSPosition  string        `meta:"GPSPosition"`
	GPSLatitude  string        `meta:"GPSLatitude"`
	GPSLongitude string        `meta:"GPSLongitude"`
	Lat          float32       `meta:"-"`
	Lng          float32       `meta:"-"`
	Altitude     int           `meta:"GlobalAltitude,GPSAltitude"`
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

	aspectRatio := float32(math.Round((width/height)*100) / 100)

	return aspectRatio
}

// Portrait returns true if it is a portrait picture or video based on width and height.
func (data Data) Portrait() bool {
	return data.ActualWidth() < data.ActualHeight()
}

// IsHDR tests if it is a high dynamic range file.
func (data Data) IsHDR() bool {
	return data.ImageType == ImageTypeHDR
}

// Megapixels returns the resolution in megapixels.
func (data Data) Megapixels() int {
	return int(math.Round(float64(data.Width*data.Height) / 1000000))
}

// HasDocumentID returns true if a DocumentID exists.
func (data Data) HasDocumentID() bool {
	return rnd.IsUUID(data.DocumentID)
}

// HasInstanceID returns true if an InstanceID exists.
func (data Data) HasInstanceID() bool {
	return rnd.IsUUID(data.InstanceID)
}

// HasTimeAndPlace if data contains a time and GPS position.
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

// CellID returns the S2 cell ID.
func (data Data) CellID() string {
	return s2.PrefixedToken(float64(data.Lat), float64(data.Lng))
}
