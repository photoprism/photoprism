package photoprism

import (
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
	"time"
	"errors"
	"strings"
)

type ExifData struct {
	DateTime    time.Time
	CameraModel string
	UniqueID    string
	Lat         float64
	Long        float64
	Thumbnail   []byte
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

	return m.exifData, nil
}