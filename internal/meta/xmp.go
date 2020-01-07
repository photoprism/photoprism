package meta

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/ugjka/go-tz.v2/tz"

	_ "trimmer.io/go-xmp/models"
	model "trimmer.io/go-xmp/models/exif"
	"trimmer.io/go-xmp/xmp"
)

// Exif returns parses an XMP and returns its data.
func XMP(filename string) (data Data, err error) {
	defer func() {
		if e := recover(); e != nil {
			data = Data{}
			err = fmt.Errorf("meta: %s", e)
		}
	}()

	data.Xmp = xmp.NewDocument()
	data.Exif = &model.ExifInfo{}
	data.ExifEX = &model.ExifEXInfo{}
	data.ExifAux = &model.ExifAuxInfo{}

	f, err := os.Open(filename)

	if err != nil {
		return data, err
	}

	defer f.Close()

	dec := xmp.NewDecoder(f)

	if err := dec.Decode(data.Xmp); err != nil {
		return data, err
	}

	if err := data.Exif.SyncFromXMP(data.Xmp); err != nil {
		return data, err
	}

	if err := data.ExifEX.SyncFromXMP(data.Xmp); err != nil {
		return data, err
	}

	if err := data.ExifAux.SyncFromXMP(data.Xmp); err != nil {
		return data, err
	}

	data.Artist = data.Exif.Artist
	data.Copyright = data.Exif.Copyright
	data.TakenAtLocal = data.Exif.DateTime.Value()
	data.Description = data.Exif.ImageDescription
	data.Height = data.Exif.ImageLength
	data.Width = data.Exif.ImageWidth
	data.CameraMake = data.Exif.Make
	data.CameraModel = data.Exif.Model
	data.LensModel = data.ExifAux.Lens
	data.Orientation = int(data.Exif.Orientation)
	data.Lat = 0 //data.Exif.GPSLatitudeCoord.Value()
	data.Lng = 0 //data.Exif.GPSLongitudeCoord.Value()
	data.Altitude = int(data.Exif.GPSAltitude.Value())
	data.TimeZone = "UTC"

	if data.Lat != 0 && data.Lng != 0 {
		if zones, err := tz.GetZone(tz.Point{
			Lat: data.Lat,
			Lon: data.Lng,
		}); err == nil {
			data.TimeZone = zones[0]
		}
	}

	loc, err := time.LoadLocation(data.TimeZone)

	if err != nil {
		data.TakenAt = data.TakenAtLocal
		log.Warnf("meta: no location for timezone (%s)", err.Error())
	} else {
		data.TakenAtLocal.In(loc)
		data.TakenAt = data.TakenAtLocal.UTC()
	}

	return data, nil
}

// TODO: Needs to be implemented
func (d *Data) SaveAsXMP(filename string) error {
	return nil
}
