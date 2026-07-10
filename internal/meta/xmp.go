package meta

import (
	"fmt"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media/projection"
)

// XMP parses an XMP file and returns a Data struct.
func XMP(fileName string) (data Data, err error) {
	err = data.XMP(fileName)

	return data, err
}

// applyTimeOffset reinterprets t's wall-clock components as local time in the
// fixed zone described by an EXIF OffsetTime string ("+02:00", "-05:30").
// Returns t unchanged if the offset cannot be parsed.
func applyTimeOffset(t time.Time, offset string) time.Time {
	z, err := time.Parse("-07:00", offset)
	if err != nil {
		return t
	}

	_, secs := z.Zone()
	loc := time.FixedZone(offset, secs)

	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc)
}

// XMP parses an XMP file and returns a Data struct.
func (data *Data) XMP(fileName string) (err error) {
	logName := clean.Log(filepath.Base(fileName))

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("metadata: %s in %s (xmp panic)\nstack: %s", e, logName, debug.Stack())
		}
	}()

	// Resolve file name e.g. in case it's a symlink.
	if fileName, err = fs.Resolve(fileName); err != nil {
		return fmt.Errorf("metadata: %s %s (xmp)", err, logName)
	}

	doc := XmpDocument{}

	if err = doc.Load(fileName); err != nil {
		return fmt.Errorf("metadata: cannot read %s (xmp)", logName)
	}

	if v := doc.Title(); v != "" {
		data.Title = v
	}
	if v := doc.Artist(); v != "" {
		data.Artist = v
	}
	if v := doc.Description(); v != "" {
		data.Caption = v
	}
	if v := doc.Copyright(); v != "" {
		data.Copyright = v
	}
	if v := doc.License(); v != "" {
		data.License = v
	}
	if v := doc.Software(); v != "" {
		data.Software = v
	}
	if v := doc.DocumentID(); v != "" {
		data.DocumentID = v
	}
	if v := doc.InstanceID(); v != "" {
		data.InstanceID = v
	}

	// GPS Lat/Lng pass through NormalizeGPS for parity with the embedded
	// path's clamp/normalize behavior.
	if lat, lng := doc.Lat(), doc.Lng(); lat != 0 || lng != 0 {
		data.Lat, data.Lng = NormalizeGPS(lat, lng)
	}
	if v := doc.Altitude(); v != 0 {
		data.Altitude = v
	}
	if v := doc.TakenGps(); !v.IsZero() {
		data.TakenGps = v.UTC()
	}

	if v := doc.CameraMake(); v != "" {
		data.CameraMake = v
	}
	if v := doc.CameraModel(); v != "" {
		data.CameraModel = v
	}
	if v := doc.LensMake(); v != "" {
		data.LensMake = v
	}
	if v := doc.LensModel(); v != "" {
		data.LensModel = v
	}
	if v := doc.CameraSerial(); v != "" {
		data.CameraSerial = v
	}
	if v := doc.CameraOwner(); v != "" {
		data.CameraOwner = v
	}
	if v := doc.Projection(); v != "" {
		data.Projection = v
	}
	if v := doc.ColorProfile(); v != "" {
		data.ColorProfile = v
	}

	if v := doc.Aperture(); v != 0 {
		data.Aperture = v
	}
	if v := doc.FNumber(); v != 0 {
		data.FNumber = v
	}
	if v := doc.FocalLength(); v != 0 {
		data.FocalLength = v
	}
	if v := doc.Iso(); v != 0 {
		data.Iso = v
	}
	if v := doc.Exposure(); v != "" {
		data.Exposure = v
	}
	if doc.Flash() {
		// Mirror the embedded-EXIF flow: set the bool *and* add the
		// "flash" keyword, so XMP-sidecar-only photos surface in the
		// same searches as photos indexed via the EXIF path.
		data.AddKeywords(KeywordFlash)
		data.Flash = true
	}
	if v := doc.Notes(); v != "" {
		data.Notes = v
	}

	if v := doc.TakenAt(data.TimeZone); !v.IsZero() {
		// Keep the wall-clock value (carrying any FixedZone parsed from the
		// XMP timestamp) on both fields so the shared resolver can normalize
		// UTC and local time consistently — mirrors how the EXIF reflection
		// loop fills TakenAt and TakenAtLocal from the same source tag.
		data.TakenAt = v
		data.TakenAtLocal = v
	}
	if v := doc.TakenNs(); v > 0 {
		data.TakenNs = v
	}
	// CreatedAt only serves as a capture-time fallback for video/audio
	// sidecars (xmpDM:CreationDate), where TakenAt has no source tag. When
	// TakenAt is already set from photoshop:DateCreated, leaving CreatedAt
	// empty stops the shared resolver from overwriting the higher-priority
	// capture instant with the less-specific xmp:CreateDate.
	if data.TakenAt.IsZero() {
		if v := doc.CreatedAt(data.TimeZone); !v.IsZero() {
			data.CreatedAt = v.UTC()
		}
	}
	if v := doc.TimeOffset(); v != "" {
		data.TimeOffset = v
	}

	// XMP capture-time tags carry no inline UTC offset, whereas the EXIF 2.31
	// OffsetTime* value lives in a separate tag. Attach it to TakenAt and
	// TakenAtLocal so the shared resolver — which assumes TakenAtLocal knows
	// its own zone — can derive the correct UTC instant. Skip when the parsed
	// timestamp already carries an offset (non-UTC zone).
	if data.TimeOffset != "" && !data.TakenAt.IsZero() {
		if _, off := data.TakenAt.Zone(); off == 0 {
			data.TakenAt = applyTimeOffset(data.TakenAt, data.TimeOffset)
			data.TakenAtLocal = data.TakenAt
		}
	}

	// dc:subject populates the descriptive Subject field only, never the
	// keyword list: dc:subject is Adobe's "Keywords" panel, but PhotoPrism's
	// Keywords, Labels, and Subject fields serve distinct purposes. This
	// matches the embedded/ExifTool path, where data.Subject comes from the
	// dc:subject-backed Subject tag and data.Keywords from IPTC Keywords.
	if v := doc.Subject(); v != "" {
		data.Subject = SanitizeMeta(v)
	}

	data.Favorite = doc.Favorite()

	// Auto-derive keywords from projection and caption text so XMP-only
	// photos surface in the same searches as EXIF-indexed photos. Mirrors
	// the EXIF and ExifTool JSON flows (exif.go's ProjectionType/ImageDescription
	// keyword block, json_exiftool.go's panorama/caption keyword block).
	// AutoAddKeywords also sets data.ImageType to ImageTypeHDR when the
	// caption mentions "hdr".
	if projection.Equirectangular.Equal(data.Projection) {
		data.AddKeywords(KeywordPanorama)
	}
	if data.Caption != "" {
		data.AutoAddKeywords(data.Caption)
	}

	// Normalize capture time, local time, and time zone using the shared
	// resolver so the XMP sidecar path produces the same entity state the
	// ExifTool JSON path would for identical metadata.
	data.ResolveTimeZone(logName)

	return nil
}
