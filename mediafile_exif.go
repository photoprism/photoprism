package photoprism

import (
	"errors"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
	"strings"
	"time"
)

type ExifData struct {
	DateTime    time.Time
	CameraModel string
	UniqueID    string
	Lat         float64
	Long        float64
	Thumbnail   []byte
	Width       int
	Height      int
	Orientation int
}

func (m *MediaFile) GetExifData() (*ExifData, error) {
	if m == nil {
		return nil, errors.New("media file is null")
	}

	if m.exifData != nil {
		return m.exifData, nil
	}

	if !m.IsPhoto() {
		return nil, errors.New("not a JPEG or Raw file")
	}

	m.exifData = &ExifData{}

	file, err := m.openFile()

	if err != nil {
		return m.exifData, err
	}

	defer file.Close()

	exif.RegisterParsers(mknote.All...)

	x, err := exif.Decode(file)

	if err != nil {
		return m.exifData, err
	}

	if camModel, err := x.Get(exif.Model); err == nil {
		m.exifData.CameraModel = strings.Replace(camModel.String(), "\"", "", -1)
	}

	if tm, err := x.DateTime(); err == nil {
		m.exifData.DateTime = tm
	}

	if lat, long, err := x.LatLong(); err == nil {
		m.exifData.Lat = lat
		m.exifData.Long = long
	}

	if thumbnail, err := x.JpegThumbnail(); err == nil {
		m.exifData.Thumbnail = thumbnail
	}

	if uniqueId, err := x.Get(exif.ImageUniqueID); err == nil {
		m.exifData.UniqueID = uniqueId.String()
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

	return m.exifData, nil
}
