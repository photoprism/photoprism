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

// Data represents image metadata.
type Data struct {
	FileName      string        `meta:"FileName"`
	DocumentID    string        `meta:"BurstUUID,MediaGroupUUID,ImageUniqueID,OriginalDocumentID,DocumentID,DigitalImageGUID"`
	InstanceID    string        `meta:"InstanceID,DocumentID"`
	CreatedAt     time.Time     `meta:"SubSecCreateDate,CreationDate,CreateDate,MediaCreateDate,ContentCreateDate,TrackCreateDate"`
	TakenAt       time.Time     `meta:"SubSecDateTimeOriginal,SubSecDateTimeCreated,DateTimeOriginal,CreationDate,DateTimeCreated,DateTime,DateTimeDigitized" xmp:"DateCreated"`
	TakenAtLocal  time.Time     `meta:"SubSecDateTimeOriginal,SubSecDateTimeCreated,DateTimeOriginal,CreationDate,DateTimeCreated,DateTime,DateTimeDigitized"`
	TakenGps      time.Time     `meta:"GPSDateTime,GPSDateStamp"`
	TakenNs       int           `meta:"-"`
	TimeZone      string        `meta:"-"`
	Duration      time.Duration `meta:"Duration,MediaDuration,TrackDuration,PreviewDuration"`
	FPS           float64       `meta:"VideoFrameRate,VideoAvgFrameRate"`
	Frames        int           `meta:"FrameCount,AnimationFrames"`
	Codec         string        `meta:"CompressorID,VideoCodecID,CodecID,OtherFormat,FileType"`
	Title         string        `meta:"Title,Headline" xmp:"dc:title" dc:"title,title.Alt"`
	Description   string        `meta:"Description,ImageDescription,Caption,Caption-Abstract" xmp:"Description,Description.Alt"`
	Subject       string        `meta:"Subject,PersonInImage,ObjectName,HierarchicalSubject,CatalogSets" xmp:"Subject"`
	Keywords      Keywords      `meta:"Keywords"`
	Favorite      bool          `meta:"Favorite"`
	Notes         string        `meta:"Comment,UserComment"`
	Artist        string        `meta:"Artist,Creator,By-line,OwnerName,Owner" xmp:"Creator"`
	Copyright     string        `meta:"Rights,Copyright,CopyrightNotice,WebStatement" xmp:"Rights,Rights.Alt"`
	License       string        `meta:"UsageTerms,License"`
	Projection    string        `meta:"ProjectionType"`
	ColorProfile  string        `meta:"ICCProfileName,ProfileDescription"`
	CameraMake    string        `meta:"CameraMake,Make" xmp:"Make"`
	CameraModel   string        `meta:"CameraModel,Model" xmp:"Model"`
	CameraOwner   string        `meta:"OwnerName"`
	CameraSerial  string        `meta:"SerialNumber"`
	LensMake      string        `meta:"LensMake"`
	LensModel     string        `meta:"Lens,LensModel" xmp:"LensModel"`
	Software      string        `meta:"Software,CreatorTool,HistorySoftwareAgent,ProcessingSoftware"`
	Flash         bool          `meta:"FlashFired"`
	FocalLength   int           `meta:"FocalLength,FocalLengthIn35mmFormat"`
	FocalDistance float64       `meta:"HyperfocalDistance"`
	Exposure      string        `meta:"ExposureTime,ShutterSpeedValue,ShutterSpeed,TargetExposureTime"`
	Aperture      float32       `meta:"ApertureValue,Aperture"`
	FNumber       float32       `meta:"FNumber"`
	Iso           int           `meta:"ISO"`
	ImageType     int           `meta:"HDRImageType"`
	GPSPosition   string        `meta:"GPSPosition"`
	GPSLatitude   string        `meta:"GPSLatitude"`
	GPSLongitude  string        `meta:"GPSLongitude"`
	Lat           float32       `meta:"-"`
	Lng           float32       `meta:"-"`
	Altitude      float64       `meta:"GlobalAltitude,GPSAltitude"`
	Width         int           `meta:"ImageWidth,PixelXDimension,ExifImageWidth,SourceImageWidth"`
	Height        int           `meta:"ImageHeight,ImageLength,PixelYDimension,ExifImageHeight,SourceImageHeight"`
	Orientation   int           `meta:"-"`
	Rotation      int           `meta:"Rotation"`
	Views         int           `meta:"-"`
	Albums        []string      `meta:"-"`
	EmbeddedVideo string        `meta:"EmbeddedVideo"`
	Error         error         `meta:"-"`
	json          map[string]string
	exif          map[string]string
}

// New returns a new metadata struct.
func New() Data {
	return Data{}
}

// AspectRatio returns the aspect ratio based on width and height.
func (data Data) AspectRatio() float32 {
	w := float64(data.ActualWidth())
	h := float64(data.ActualHeight())

	if w <= 0 || h <= 0 {
		return 0
	}

	return float32(math.Round((w/h)*100) / 100)
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
