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

// XMP parses an XMP file and returns a Data struct.
func XMP(filename string) (data Data, err error) {
	defer func() {
		if e := recover(); e != nil {
			data = Data{}
			err = fmt.Errorf("meta: %s", e)
		}
	}()

	data.xmpDoc = xmp.NewDocument()
	data.exifInfo = &model.ExifInfo{}
	data.exifEx = &model.ExifEXInfo{}
	data.exifAux = &model.ExifAuxInfo{}

	f, err := os.Open(filename)

	if err != nil {
		return data, err
	}

	defer f.Close()

	dec := xmp.NewDecoder(f)

	if err := dec.Decode(data.xmpDoc); err != nil {
		return data, err
	}

	if err := data.exifInfo.SyncFromXMP(data.xmpDoc); err != nil {
		return data, err
	}

	if err := data.exifEx.SyncFromXMP(data.xmpDoc); err != nil {
		return data, err
	}

	if err := data.exifAux.SyncFromXMP(data.xmpDoc); err != nil {
		return data, err
	}

	data.Artist = data.exifInfo.Artist
	data.Copyright = data.exifInfo.Copyright
	data.TakenAtLocal = data.exifInfo.DateTime.Value()
	data.Description = data.exifInfo.ImageDescription
	data.Height = data.exifInfo.ImageLength
	data.Width = data.exifInfo.ImageWidth
	data.CameraMake = data.exifInfo.Make
	data.CameraModel = data.exifInfo.Model
	data.LensModel = data.exifAux.Lens
	data.Orientation = int(data.exifInfo.Orientation)
	data.Lat = 0 //data.exifInfo.GPSLatitudeCoord.Value()
	data.Lng = 0 //data.exifInfo.GPSLongitudeCoord.Value()
	data.Altitude = int(data.exifInfo.GPSAltitude.Value())
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
