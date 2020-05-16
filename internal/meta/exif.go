package meta

import (
	"fmt"
	"math"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dsoprea/go-exif/v2"
	"github.com/dsoprea/go-exif/v2/common"
	"github.com/dsoprea/go-jpeg-image-structure"
	"github.com/dsoprea/go-png-image-structure"
	"github.com/photoprism/photoprism/pkg/txt"
	"gopkg.in/ugjka/go-tz.v2/tz"
)

const DateTimeZero = "0000:00:00 00:00:00"

// ValidDateTime returns true if a date string looks valid and is not zero.
func ValidDateTime(s string) bool {
	return len(s) == len(DateTimeZero) && s != DateTimeZero
}

// SanitizeString removes unwanted character from an exif value string.
func SanitizeString(value string) string {
	value = strings.TrimSpace(value)
	return strings.Replace(value, "\"", "", -1)
}

// Exif parses an image file for Exif meta data and returns as Data struct.
func Exif(fileName string) (data Data, err error) {
	err = data.Exif(fileName)

	return data, err
}

// Exif parses an image file for Exif meta data and returns as Data struct.
func (data *Data) Exif(fileName string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%s (exif metadata)", e)
		}
	}()

	// Extract raw EXIF block.
	var rawExif []byte
	var parsed bool

	fileExtension := strings.ToLower(path.Ext(fileName))

	if fileExtension == ".jpg" || fileExtension == ".jpeg" {
		jmp := jpegstructure.NewJpegMediaParser()

		sl, err := jmp.ParseFile(fileName)

		if err != nil {
			return err
		}

		_, rawExif, err = sl.Exif()

		if err != nil {
			if err.Error() == "no exif data" {
				return fmt.Errorf("no exif data in %s", txt.Quote(filepath.Base(fileName)))
			} else {
				log.Warnf("exif: %s (parse jpeg)", err)
			}
		} else {
			parsed = true
		}
	} else if fileExtension == ".png" {
		pmp := pngstructure.NewPngMediaParser()

		cs, err := pmp.ParseFile(fileName)

		if err != nil {
			return err
		}

		_, rawExif, err = cs.Exif()

		if err != nil {
			if err.Error() == "file does not have EXIF" {
				return fmt.Errorf("no exif data in %s", txt.Quote(filepath.Base(fileName)))
			} else {
				log.Warnf("exif: %s (parse png)", err)
			}
		} else {
			parsed = true
		}
	}

	if !parsed {
		// Fallback to an optimistic, brute-force search.
		var err error

		rawExif, err = exif.SearchFileAndExtractExif(fileName)

		if err != nil {
			return err
		}
	}

	// Enumerate tags in EXIF block.
	ti := exif.NewTagIndex()

	if err := exif.LoadStandardTags(ti); err != nil {
		return err
	}

	if data.All == nil {
		data.All = make(map[string]string)
	}

	visitor := func(fqIfdPath string, ifdIndex int, ite *exif.IfdTagEntry) (err error) {
		tagId := ite.TagId()
		tagType := ite.TagType()

		ifdPath, err := im.StripPathPhraseIndices(fqIfdPath)

		if err != nil {
			return nil
		}

		it, err := ti.Get(ifdPath, tagId)

		if err != nil {
			return nil
		}

		valueString := ""

		if tagType != exifcommon.TypeUndefined {
			valueString, err = ite.FormatFirst()

			if err != nil {
				log.Errorf("exif: %s", err.Error())

				return nil
			}

			if it.Name != "" && valueString != "" {
				data.All[it.Name] = strings.Split(valueString, "\x00")[0]
			}
		}

		return nil
	}

	_, _, err = exif.Visit(exifcommon.IfdStandard, im, ti, rawExif, visitor)

	if err != nil {
		return err
	}

	tags := data.All

	// Cherry-pick the values that we care about.

	if value, ok := tags["Artist"]; ok {
		data.Artist = SanitizeString(value)
	}

	if value, ok := tags["Copyright"]; ok {
		data.Copyright = SanitizeString(value)
	}

	if value, ok := tags["Model"]; ok {
		data.CameraModel = SanitizeString(value)
	} else if value, ok := tags["CameraModel"]; ok {
		data.CameraModel = SanitizeString(value)
	}

	if value, ok := tags["Make"]; ok {
		data.CameraMake = SanitizeString(value)
	} else if value, ok := tags["CameraMake"]; ok {
		data.CameraMake = SanitizeString(value)
	}

	if value, ok := tags["CameraOwnerName"]; ok {
		data.CameraOwner = SanitizeString(value)
	}

	if value, ok := tags["BodySerialNumber"]; ok {
		data.CameraSerial = SanitizeString(value)
	}

	if value, ok := tags["LensMake"]; ok {
		data.LensMake = SanitizeString(value)
	}

	if value, ok := tags["LensModel"]; ok {
		data.LensModel = SanitizeString(value)
	}

	if value, ok := tags["ExposureTime"]; ok {
		if n := strings.Split(value, "/"); len(n) == 2 {
			if n[0] != "1" && len(n[0]) < len(n[1]) {
				n0, _ := strconv.ParseUint(n[0], 10, 64)
				if n1, err := strconv.ParseUint(n[1], 10, 64); err == nil && n0 > 0 && n1 > 0 {
					value = fmt.Sprintf("1/%d", n1/n0)
				}
			}
		}

		data.Exposure = value
	}

	if value, ok := tags["FNumber"]; ok {
		values := strings.Split(value, "/")

		if len(values) == 2 && values[1] != "0" && values[1] != "" {
			number, _ := strconv.ParseFloat(values[0], 64)
			denom, _ := strconv.ParseFloat(values[1], 64)

			data.FNumber = float32(math.Round((number/denom)*1000) / 1000)
		}
	}

	if value, ok := tags["ApertureValue"]; ok {
		values := strings.Split(value, "/")

		if len(values) == 2 && values[1] != "0" && values[1] != "" {
			number, _ := strconv.ParseFloat(values[0], 64)
			denom, _ := strconv.ParseFloat(values[1], 64)

			data.Aperture = float32(math.Round((number/denom)*1000) / 1000)
		}
	}

	if value, ok := tags["FocalLengthIn35mmFilm"]; ok {
		if i, err := strconv.Atoi(value); err == nil {
			data.FocalLength = i
		}
	} else if value, ok := tags["FocalLength"]; ok {
		values := strings.Split(value, "/")

		if len(values) == 2 && values[1] != "0" && values[1] != "" {
			number, _ := strconv.ParseFloat(values[0], 64)
			denom, _ := strconv.ParseFloat(values[1], 64)

			data.FocalLength = int(math.Round((number/denom)*1000) / 1000)
		}
	}

	if value, ok := tags["ISOSpeedRatings"]; ok {
		if i, err := strconv.Atoi(value); err == nil {
			data.Iso = i
		}
	}

	if value, ok := tags["ImageUniqueID"]; ok {
		data.UniqueID = value
	}

	if value, ok := tags["ImageWidth"]; ok {
		if i, err := strconv.Atoi(value); err == nil {
			data.Width = i
		}
	} else if value, ok := tags["PixelXDimension"]; ok {
		if i, err := strconv.Atoi(value); err == nil {
			data.Width = i
		}
	}

	if value, ok := tags["ImageLength"]; ok {
		if i, err := strconv.Atoi(value); err == nil {
			data.Height = i
		}
	} else if value, ok := tags["PixelYDimension"]; ok {
		if i, err := strconv.Atoi(value); err == nil {
			data.Height = i
		}
	}

	if value, ok := tags["Orientation"]; ok {
		if i, err := strconv.Atoi(value); err == nil {
			data.Orientation = i
		}
	} else {
		data.Orientation = 1
	}

	_, index, err := exif.Collect(im, ti, rawExif)

	if err != nil {
		return err
	}

	if ifd, err := index.RootIfd.ChildWithIfdPath(exifcommon.IfdPathStandardGps); err == nil {
		if gi, err := ifd.GpsInfo(); err == nil {
			data.Lat = float32(gi.Latitude.Decimal())
			data.Lng = float32(gi.Longitude.Decimal())
			data.Altitude = gi.Altitude
		} else {
			log.Warnf("exif: %s (gps info)", err)
		}
	}

	if data.Lat != 0 && data.Lng != 0 {
		zones, err := tz.GetZone(tz.Point{
			Lat: float64(data.Lat),
			Lon: float64(data.Lng),
		})

		if err == nil && len(zones) > 0 {
			data.TimeZone = zones[0]
		}
	}

	var takenAt string

	if value, ok := tags["DateTimeOriginal"]; ok && ValidDateTime(value) {
		takenAt = value
	} else if value, ok := tags["DateTimeDigitized"]; ok && ValidDateTime(value) {
		takenAt = value
	} else if value, ok := tags["CreateDate"]; ok && ValidDateTime(value) {
		takenAt = value
	} else if value, ok := tags["DateTime"]; ok && ValidDateTime(value) {
		takenAt = value
	}

	if ValidDateTime(takenAt) {
		if taken, err := time.Parse("2006:01:02 15:04:05", takenAt); err == nil {
			data.TakenAtLocal = taken.Round(time.Second)
			data.TakenAt = data.TakenAtLocal

			if loc, err := time.LoadLocation(data.TimeZone); err != nil {
				log.Warnf("exif: unknown time zone %s", data.TimeZone)
			} else if tl, err := time.ParseInLocation("2006:01:02 15:04:05", takenAt, loc); err == nil {
				data.TakenAt = tl.Round(time.Second).UTC()
			} else {
				log.Errorf("exif: %s", err.Error()) // this should never happen
			}
		} else {
			log.Warnf("exif: invalid time %s", takenAt)
		}
	}

	if value, ok := tags["Flash"]; ok {
		if i, err := strconv.Atoi(value); err == nil && i&1 == 1 {
			data.Flash = true
		}
	}

	if value, ok := tags["ImageDescription"]; ok {
		data.Description = SanitizeString(value)
	}

	data.All = tags

	return nil
}
