package photoprism

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
)

// Exif returns information about a single image.
type Exif struct {
	UUID        string
	DateTime    time.Time
	TimeZone    string
	Artist      string
	CameraMake  string
	CameraModel string
	LensMake    string
	LensModel   string
	Aperture    float64
	FocalLength float64
	Lat         float64
	Long        float64
	Thumbnail   []byte
	Width       int
	Height      int
	Orientation int
}

func init() {
	exif.RegisterParsers(mknote.All...)
}

// Exif returns exif meta data of a media file.
func (m *MediaFile) Exif() (result *Exif, err error) {
	defer func() {
		if e := recover(); e != nil {
			result = m.exifData
			err = fmt.Errorf("error while parsing exif data: %s", e)
		}
	}()

	if m == nil {
		return nil, errors.New("can't parse exif data: file instance is null")
	}

	if m.exifData != nil {
		return m.exifData, nil
	}

	if !m.IsJpeg() && !m.IsRaw() {
		return nil, errors.New(fmt.Sprintf("media file not compatible with exif: \"%s\"", m.Filename()))
	}

	m.exifData = &Exif{}

	file, err := m.openFile()

	if err != nil {
		return nil, err
	}

	defer file.Close()

	x, err := exif.Decode(file)

	if err != nil {
		return nil, err
	}

	if artist, err := x.Get(exif.Artist); err == nil {
		m.exifData.Artist = strings.Replace(artist.String(), "\"", "", -1)
	}

	if camModel, err := x.Get(exif.Model); err == nil {
		m.exifData.CameraModel = strings.Replace(camModel.String(), "\"", "", -1)
	}

	if camMake, err := x.Get(exif.Make); err == nil {
		m.exifData.CameraMake = strings.Replace(camMake.String(), "\"", "", -1)
	}

	if lensMake, err := x.Get(exif.LensMake); err == nil {
		m.exifData.LensMake = strings.Replace(lensMake.String(), "\"", "", -1)
	}

	if lensModel, err := x.Get(exif.LensModel); err == nil {
		m.exifData.LensModel = strings.Replace(lensModel.String(), "\"", "", -1)
	}

	if aperture, err := x.Get(exif.ApertureValue); err == nil {
		number, denom, _ := aperture.Rat2(0)

		if denom == 0 {
			denom = 1
		}

		value := float64(number) / float64(denom)

		m.exifData.Aperture = math.Round(value*1000) / 1000
	}

	if focal, err := x.Get(exif.FocalLength); err == nil {
		number, denom, _ := focal.Rat2(0)

		if denom == 0 {
			denom = 1
		}

		value := float64(number) / float64(denom)

		m.exifData.FocalLength = math.Round(value*1000) / 1000
	}

	if tm, err := x.DateTime(); err == nil {
		m.exifData.DateTime = tm
	}

	if tz, err := x.TimeZone(); err == nil {
		m.exifData.TimeZone = tz.String()
	}

	if lat, long, err := x.LatLong(); err == nil {
		m.exifData.Lat = lat
		m.exifData.Long = long
	}

	if thumbnail, err := x.JpegThumbnail(); err == nil {
		m.exifData.Thumbnail = thumbnail
	}

	if uniqueID, err := x.Get(exif.ImageUniqueID); err == nil {
		m.exifData.UUID = uniqueID.String()
	}

	if width, err := x.Get(exif.ImageWidth); err == nil {
		m.exifData.Width, _ = width.Int(0)
	}

	if height, err := x.Get(exif.ImageLength); err == nil {
		m.exifData.Height, _ = height.Int(0)
	}

	if orientation, err := x.Get(exif.Orientation); err == nil {
		m.exifData.Orientation, _ = orientation.Int(0)
	} else {
		m.exifData.Orientation = 1
	}

	return m.exifData, err
}
