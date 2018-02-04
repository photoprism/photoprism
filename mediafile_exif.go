package photoprism

import (
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
	"time"
	"errors"
	"strings"
	"log"
)

type ExifData struct {
	DateTime    time.Time
	CameraModel string
	Lat         float64
	Long        float64
}

func (mediaFile *MediaFile) GetExifData() (*ExifData, error) {
	if mediaFile.exifData != nil {
		log.Printf("GetExifData() Cache Hit %s", mediaFile.filename)
		return mediaFile.exifData, nil
	}

	log.Printf("GetExifData() Cache Miss %s", mediaFile.filename)

	if !mediaFile.IsJpeg() {
		// EXIF only works for JPEG
		return nil, errors.New("MediaFile is not a JPEG")
	}

	mediaFile.exifData = &ExifData{}

	log.Printf("GetExifData() Open File %s", mediaFile.filename)
	file, err := mediaFile.openFile()

	if err != nil {
		return mediaFile.exifData, err
	}

	defer file.Close()

	exif.RegisterParsers(mknote.All...)

	x, err := exif.Decode(file)

	if err != nil {
		return mediaFile.exifData, err
	}

	camModel, _ := x.Get(exif.Model)
	mediaFile.exifData.CameraModel = strings.Replace(camModel.String(), "\"", "", -1)

	tm, _ := x.DateTime()
	mediaFile.exifData.DateTime = tm

	lat, long, _ := x.LatLong()
	mediaFile.exifData.Lat = lat
	mediaFile.exifData.Long = long

	return mediaFile.exifData, nil
}