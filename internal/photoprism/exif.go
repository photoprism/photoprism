package photoprism

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/dsoprea/go-exif"
	"gopkg.in/ugjka/go-tz.v2/tz"
)

// Exif returns information about a single image.
type Exif struct {
	UUID         string
	TakenAt      time.Time
	TakenAtLocal time.Time
	TimeZone     string
	Artist       string
	CameraMake   string
	CameraModel  string
	Description  string
	LensMake     string
	LensModel    string
	Flash        bool
	FocalLength  int
	Exposure     string
	Aperture     float64
	FNumber      float64
	Iso          int
	Lat          float64
	Long         float64
	Altitude     int
	Width        int
	Height       int
	Orientation  int
	All          map[string]string
}

var im *exif.IfdMapping

func IfdMapping() *exif.IfdMapping {
	if im != nil {
		return im
	}

	im = exif.NewIfdMapping()

	if err := exif.LoadStandardIfds(im); err != nil {
		log.Errorf("could not parse exif config: %s", err.Error())
	}

	return im
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
		return nil, errors.New("can't parse exif data: file instance is nil")
	}

	if m.exifData != nil {
		return m.exifData, nil
	}

	if !m.IsJpeg() && !m.IsRaw() && !m.IsHEIF() {
		return nil, errors.New(fmt.Sprintf("media file not compatible with exif: \"%s\"", m.Filename()))
	}

	m.exifData = &Exif{}

	rawExif, err := exif.SearchFileAndExtractExif(m.Filename())

	if err != nil {
		return m.exifData, err
	}

	ti := exif.NewTagIndex()

	tags := make(map[string]string)
	im := IfdMapping()

	visitor := func(fqIfdPath string, ifdIndex int, tagId uint16, tagType exif.TagType, valueContext exif.ValueContext) (err error) {
		ifdPath, err := im.StripPathPhraseIndices(fqIfdPath)

		if err != nil {
			return nil
		}

		it, err := ti.Get(ifdPath, tagId)

		if err != nil {
			return nil
		}

		valueString := ""

		if tagType.Type() != exif.TypeUndefined {
			valueString, err = tagType.ResolveAsString(valueContext, true)

			if err != nil {
				log.Error(err)

				return nil
			}

			if it.Name != "" && valueString != "" {
				tags[it.Name] = valueString
			}
		}

		return nil
	}

	_, err = exif.Visit(exif.IfdStandard, im, ti, rawExif, visitor)

	if err != nil {
		return m.exifData, err
	}

	if value, ok := tags["Artist"]; ok {
		m.exifData.Artist = strings.Replace(value, "\"", "", -1)
	}

	if value, ok := tags["Model"]; ok {
		m.exifData.CameraModel = strings.Replace(value, "\"", "", -1)
	}

	if value, ok := tags["Make"]; ok {
		m.exifData.CameraMake = strings.Replace(value, "\"", "", -1)
	}

	if value, ok := tags["LensMake"]; ok {
		m.exifData.LensMake = strings.Replace(value, "\"", "", -1)
	}

	if value, ok := tags["LensModel"]; ok {
		m.exifData.LensModel = strings.Replace(value, "\"", "", -1)
	}

	if value, ok := tags["ExposureTime"]; ok {
		m.exifData.Exposure = value
	}

	if value, ok := tags["FNumber"]; ok {
		values := strings.Split(value, "/")

		if len(values) == 2 && values[1] != "0" && values[1] != "" {
			number, _ := strconv.ParseFloat(values[0], 64)
			denom, _ := strconv.ParseFloat(values[1], 64)

			m.exifData.FNumber = math.Round((number/denom)*1000) / 1000
		}
	}

	if value, ok := tags["ApertureValue"]; ok {
		values := strings.Split(value, "/")

		if len(values) == 2 && values[1] != "0" && values[1] != "" {
			number, _ := strconv.ParseFloat(values[0], 64)
			denom, _ := strconv.ParseFloat(values[1], 64)

			m.exifData.Aperture = math.Round((number/denom)*1000) / 1000
		}
	}

	if value, ok := tags["FocalLengthIn35mmFilm"]; ok {
		if i, err := strconv.Atoi(value); err == nil {
			m.exifData.FocalLength = i
		}
	} else if value, ok := tags["FocalLength"]; ok {
		values := strings.Split(value, "/")

		if len(values) == 2 && values[1] != "0" && values[1] != "" {
			number, _ := strconv.ParseFloat(values[0], 64)
			denom, _ := strconv.ParseFloat(values[1], 64)

			m.exifData.FocalLength = int(math.Round((number/denom)*1000) / 1000)
		}
	}

	if value, ok := tags["ISOSpeedRatings"]; ok {
		if i, err := strconv.Atoi(value); err == nil {
			m.exifData.Iso = i
		}
	}

	if value, ok := tags["ImageUniqueID"]; ok {
		m.exifData.UUID = value
	}

	if value, ok := tags["ImageWidth"]; ok {
		if i, err := strconv.Atoi(value); err == nil {
			m.exifData.Width = i
		}
	}

	if value, ok := tags["ImageLength"]; ok {
		if i, err := strconv.Atoi(value); err == nil {
			m.exifData.Width = i
		}
	}

	if value, ok := tags["Orientation"]; ok {
		if i, err := strconv.Atoi(value); err == nil {
			m.exifData.Orientation = i
		}
	} else {
		m.exifData.Orientation = 1
	}

	_, index, err := exif.Collect(im, ti, rawExif)

	if err != nil {
		return m.exifData, err
	}

	if ifd, err := index.RootIfd.ChildWithIfdPath(exif.IfdPathStandardGps); err == nil {
		if gi, err := ifd.GpsInfo(); err == nil {
			m.exifData.Lat = gi.Latitude.Decimal()
			m.exifData.Long = gi.Longitude.Decimal()
			m.exifData.Altitude = gi.Altitude
		}
	}

	if m.exifData.Lat != 0 && m.exifData.Long != 0 {
		zones, err := tz.GetZone(tz.Point{
			Lat: m.exifData.Lat,
			Lon: m.exifData.Long,
		})

		if err != nil {
			m.exifData.TimeZone = "UTC"
		}

		m.exifData.TimeZone = zones[0]
	}

	if value, ok := tags["DateTimeOriginal"]; ok {
		m.exifData.TakenAtLocal, _ = time.Parse("2006:01:02 15:04:05", value)

		loc, err := time.LoadLocation(m.exifData.TimeZone)

		if err != nil {
			m.exifData.TakenAt = m.exifData.TakenAtLocal
			log.Warnf("no location for timezone: %s", err.Error())
		} else if tl, err := time.ParseInLocation("2006:01:02 15:04:05", value, loc); err == nil {
			m.exifData.TakenAt = tl.UTC()
		} else {
			log.Warnf("could parse time: %s", err.Error())
		}
	}

	if value, ok := tags["Flash"]; ok {
		if i, err := strconv.Atoi(value); err == nil && i&1 == 1 {
			m.exifData.Flash = true
		}
	}

	if value, ok := tags["ImageDescription"]; ok {
		m.exifData.Description = strings.Replace(value, "\"", "", -1)
	}

	m.exifData.All = tags

	return m.exifData, nil
}
