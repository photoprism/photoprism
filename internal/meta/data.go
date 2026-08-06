package meta

import (
	"math"
	"time"

	"github.com/photoprism/photoprism/pkg/geo/s2"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Extended image type constants extracted from vendor-specific metadata.
const (
	ImageTypeHDR = 3 // see https://exiftool.org/TagNames/Apple.html
)

// Data represents image metadata.
//
// Note: the meta:"…", xmp:"…", and dc:"…" struct tags below are read by
// internal/meta/report.go via reflection to render the metadata-source
// columns of `photoprism show metadata`. Do not delete them as
// "vestigial" — they are documentation that ships in the CLI report.
//
// The xmp:"…" tags list the namespaced properties the XMP reader consults,
// in the priority order its chains apply (internal/meta/xmp_document.go);
// the dc:"…" tags name the Dublin Core properties among them.
type Data struct {
	FileName         string        `meta:"FileName"`
	MimeType         string        `meta:"MIMEType" report:"-"`
	DocumentID       string        `meta:"ContentIdentifier,MediaGroupUUID,BurstUUID,OriginalDocumentID,DocumentID,ImageUniqueID,DigitalImageGUID" xmp:"xmpMM:OriginalDocumentID,xmpMM:DocumentID,dc:identifier" dc:"identifier"` // see https://exiftool.org/forum/index.php?topic=14874.0
	InstanceID       string        `meta:"InstanceID,DocumentID" xmp:"xmpMM:InstanceID"`
	CreatedAt        time.Time     `meta:"SubSecCreateDate,CreationTime,CreationDate,CreateDate,MediaCreateDate,ContentCreateDate,TrackCreateDate" xmp:"xmp:CreateDate,xmpDM:CreationDate"`
	TakenAt          time.Time     `meta:"SubSecDateTimeOriginal,SubSecDateTimeCreated,DateTimeOriginal,CreationTime,CreationDate,DateTimeCreated,DateTime,DateTimeDigitized" xmp:"photoshop:DateCreated,exif:DateTimeOriginal,xmp:CreateDate"`
	TakenAtLocal     time.Time     `meta:"SubSecDateTimeOriginal,SubSecDateTimeCreated,DateTimeOriginal,CreationDate,DateTimeCreated,DateTime,DateTimeDigitized"`
	TakenGps         time.Time     `meta:"GPSDateTime,GPSDateStamp" xmp:"exif:GPSTimeStamp,exif:GPSDateStamp"`
	TakenNs          int           `meta:"-"`
	TimeZone         string        `meta:"-"`
	TimeOffset       string        `meta:"OffsetTime,OffsetTimeOriginal,OffsetTimeDigitized" xmp:"exif:OffsetTimeOriginal,exif:OffsetTime,exif:OffsetTimeDigitized"`
	MediaType        media.Type    `meta:"-"`
	HasThumbEmbedded bool          `meta:"ThumbnailImage,PhotoshopThumbnail" report:"-"`
	HasVideoEmbedded bool          `meta:"EmbeddedVideoFile,MotionPhoto,MotionPhotoVideo,MicroVideo" report:"-"`
	Duration         time.Duration `meta:"Duration,MediaDuration,TrackDuration,PreviewDuration"`
	FPS              float64       `meta:"VideoFrameRate,VideoAvgFrameRate"`
	Frames           int           `meta:"FrameCount,AnimationFrames"`
	Pages            int           `meta:"PageCount,NPages,Pages"`
	Codec            string        `meta:"CompressorID,VideoCodecID,CodecID,OtherFormat,VideoCodec,FileType"`
	Title            string        `meta:"Title,Headline" xmp:"dc:title,photoshop:Headline" dc:"title,title.Alt"`
	Caption          string        `meta:"Description,ImageDescription,Caption,Caption-Abstract" xmp:"dc:description,tiff:ImageDescription" dc:"description,description.Alt"`
	Subject          string        `meta:"Subject,PersonInImage,ObjectName,HierarchicalSubject,CatalogSets" xmp:"dc:subject,Iptc4xmpExt:PersonInImage,lr:hierarchicalSubject" dc:"subject"`
	Keywords         Keywords      `meta:"Keywords"`
	Faces            []Face        `meta:"-"`
	FacesDeclared    bool          `meta:"-"`
	FacesPartial     bool          `meta:"-"`
	Favorite         bool          `meta:"Favorite" xmp:"fstop:favorite"`
	Notes            string        `meta:"Comment,UserComment" xmp:"exif:UserComment"`
	Artist           string        `meta:"Artist,Creator,By-line,OwnerName,Owner" xmp:"dc:creator" dc:"creator"`
	Copyright        string        `meta:"Rights,Copyright,CopyrightNotice,WebStatement" xmp:"dc:rights,tiff:Copyright,xmpRights:WebStatement" dc:"rights,rights.Alt"`
	License          string        `meta:"UsageTerms,License" xmp:"xmpRights:UsageTerms"`
	Projection       string        `meta:"ProjectionType" xmp:"GPano:ProjectionType"`
	ColorProfile     string        `meta:"ICCProfileName,ProfileDescription" xmp:"photoshop:ICCProfile"`
	CameraMake       string        `meta:"CameraMake,Make" xmp:"tiff:Make"`
	CameraModel      string        `meta:"CameraModel,Model,CameraID,UniqueCameraModel" xmp:"tiff:Model"`
	CameraOwner      string        `meta:"OwnerName" xmp:"aux:OwnerName"`
	CameraSerial     string        `meta:"SerialNumber" xmp:"exifEX:BodySerialNumber,aux:SerialNumber"`
	LensMake         string        `meta:"LensMake" xmp:"exifEX:LensMake"`
	LensModel        string        `meta:"LensModel,Lens,LensID" xmp:"exifEX:LensModel,aux:Lens,aux:LensID"`
	Software         string        `meta:"Software,Producer,CreatorTool,CreatorSubTool,HistorySoftwareAgent,ProcessingSoftware" xmp:"xmp:CreatorTool,tiff:Software"`
	Flash            bool          `meta:"FlashFired" xmp:"exif:Flash/exif:Fired"`
	FocalLength      int           `meta:"FocalLength,FocalLengthIn35mmFormat" xmp:"exif:FocalLength,exif:FocalLengthIn35mmFilm"`
	FocalDistance    float64       `meta:"HyperfocalDistance"`
	Exposure         string        `meta:"ExposureTime,ShutterSpeedValue,ShutterSpeed,TargetExposureTime" xmp:"exif:ExposureTime,exif:ShutterSpeedValue"`
	Aperture         float32       `meta:"ApertureValue,Aperture" xmp:"exif:ApertureValue"`
	FNumber          float32       `meta:"FNumber" xmp:"exif:FNumber"`
	Iso              int           `meta:"ISO" xmp:"exifEX:PhotographicSensitivity,exif:ISOSpeedRatings"`
	ImageType        int           `meta:"HDRImageType"`
	GPSPosition      string        `meta:"GPSPosition"`
	GPSLatitude      string        `meta:"GPSLatitude" xmp:"exif:GPSLatitude"`
	GPSLongitude     string        `meta:"GPSLongitude" xmp:"exif:GPSLongitude"`
	Lat              float64       `meta:"-"`
	Lng              float64       `meta:"-"`
	Altitude         float64       `meta:"GlobalAltitude,GPSAltitude" xmp:"exif:GPSAltitude"`
	Width            int           `meta:"ImageWidth,PixelXDimension,ExifImageWidth,SourceImageWidth"`
	Height           int           `meta:"ImageHeight,ImageLength,PixelYDimension,ExifImageHeight,SourceImageHeight"`
	Orientation      int           `meta:"-"`
	Rotation         int           `meta:"Rotation"`
	Views            int           `meta:"-"`
	Albums           []string      `meta:"-"`
	Warning          string        `meta:"Warning" report:"-"`
	Error            error         `meta:"-"`
	json             map[string]string
	exif             map[string]string
}

// NewData returns a new Data struct with default values.
func NewData() Data {
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
