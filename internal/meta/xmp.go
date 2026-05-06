package meta

import (
	"fmt"
	"path/filepath"
	"runtime/debug"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/time/tz"
)

// XMP parses an XMP file and returns a Data struct.
func XMP(fileName string) (data Data, err error) {
	err = data.XMP(fileName)

	return data, err
}

// XMP parses an XMP file and returns a Data struct.
func (data *Data) XMP(fileName string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("metadata: %s in %s (xmp panic)\nstack: %s", e, clean.Log(filepath.Base(fileName)), debug.Stack())
		}
	}()

	// Resolve file name e.g. in case it's a symlink.
	if fileName, err = fs.Resolve(fileName); err != nil {
		return fmt.Errorf("metadata: %s %s (xmp)", err, clean.Log(filepath.Base(fileName)))
	}

	doc := XmpDocument{}

	if err = doc.Load(fileName); err != nil {
		return fmt.Errorf("metadata: cannot read %s (xmp)", clean.Log(filepath.Base(fileName)))
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
	// path's clamp/normalise behaviour.
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
		data.Flash = true
	}
	if v := doc.Notes(); v != "" {
		data.Notes = v
	}

	if v := doc.TakenAt(data.TimeZone); !v.IsZero() {
		data.TakenAt = v.UTC()
		if data.TimeZone == "" {
			data.TimeZone = tz.UTC
		}
	}
	if v := doc.TakenNs(); v > 0 {
		data.TakenNs = v
	}
	if v := doc.CreatedAt(data.TimeZone); !v.IsZero() {
		data.CreatedAt = v.UTC()
	}
	if v := doc.TimeOffset(); v != "" {
		data.TimeOffset = v
	}

	if v := doc.Keywords(); len(v) != 0 {
		data.AddKeywords(v)
	}

	data.Favorite = doc.Favorite()

	return nil
}
