package meta

import (
	"fmt"
	"math"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/sanitize"
	"gopkg.in/photoprism/go-tz.v2/tz"
)

var exifIfdMapping *exifcommon.IfdMapping
var exifTagIndex = exif.NewTagIndex()
var exifMutex = sync.Mutex{}

const DateTimeZero = "0000:00:00 00:00:00"

func init() {
	exifIfdMapping = exifcommon.NewIfdMapping()

	if err := exifcommon.LoadStandardIfds(exifIfdMapping); err != nil {
		log.Errorf("metadata: %s", err.Error())
	}
}

// ValidDateTime returns true if a date string looks valid and is not zero.
func ValidDateTime(s string) bool {
	return len(s) == len(DateTimeZero) && s != DateTimeZero
}

// Exif parses an image file for Exif meta data and returns as Data struct.
func Exif(fileName string, fileType fs.FileFormat) (data Data, err error) {
	err = data.Exif(fileName, fileType)

	return data, err
}

// Exif parses an image file for Exif meta data and returns as Data struct.
func (data *Data) Exif(fileName string, fileType fs.FileFormat) (err error) {
	exifMutex.Lock()
	defer exifMutex.Unlock()

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("metadata: %s in %s (exif panic)\nstack: %s", e, sanitize.Log(filepath.Base(fileName)), debug.Stack())
		}
	}()

	// Extract raw Exif block.
	rawExif, err := RawExif(fileName, fileType)

	if err != nil {
		return err
	}

	logName := sanitize.Log(filepath.Base(fileName))

	if data.All == nil {
		data.All = make(map[string]string)
	}

	// Enumerate tags in Exif block.
	opt := exif.ScanOptions{}
	entries, _, err := exif.GetFlatExifData(rawExif, &opt)

	for _, entry := range entries {
		if entry.TagName != "" && entry.Formatted != "" {
			data.All[entry.TagName] = strings.Split(entry.FormattedFirst, "\x00")[0]
		}
	}

	tags := data.All

	_, index, err := exif.Collect(exifIfdMapping, exifTagIndex, rawExif)

	if err != nil {
		log.Debugf("exif: %s in %s (collect)", err.Error(), logName)
	} else {
		if ifd, err := index.RootIfd.ChildWithIfdPath(exifcommon.IfdGpsInfoStandardIfdIdentity); err == nil {
			if gi, err := ifd.GpsInfo(); err != nil {
				log.Debugf("exif: %s in %s (gps info)", err, logName)
				log.Infof("metadata: failed parsing GPS coordinates in %s (exif)", logName)
			} else if math.IsNaN(gi.Latitude.Decimal()) || math.IsNaN(gi.Longitude.Decimal()) {
				log.Warnf("metadata: invalid GPS coordinates in %s (exif)", logName)
			} else {
				data.Lat = float32(gi.Latitude.Decimal())
				data.Lng = float32(gi.Longitude.Decimal())
				data.Altitude = gi.Altitude
			}
		}
	}

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
		if id := rnd.SanitizeUUID(value); id != "" {
			data.DocumentID = id
		}
	}

	if value, ok := tags["PixelXDimension"]; ok {
		if i, err := strconv.Atoi(value); err == nil {
			data.Width = i
		}
	} else if value, ok := tags["ImageWidth"]; ok {
		if i, err := strconv.Atoi(value); err == nil {
			data.Width = i
		}
	}

	if value, ok := tags["PixelYDimension"]; ok {
		if i, err := strconv.Atoi(value); err == nil {
			data.Height = i
		}
	} else if value, ok := tags["ImageLength"]; ok {
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
		takenAt = strings.ReplaceAll(takenAt, "/", ":")
		takenAt = strings.ReplaceAll(takenAt, "-", ":")

		if taken, err := time.Parse("2006:01:02 15:04:05", takenAt); err == nil {
			data.TakenAtLocal = taken.Round(time.Second)
			data.TakenAt = data.TakenAtLocal

			if loc, err := time.LoadLocation(data.TimeZone); err != nil {
				log.Warnf("metadata: unknown time zone %s in %s (exif)", data.TimeZone, logName)
			} else if tl, err := time.ParseInLocation("2006:01:02 15:04:05", takenAt, loc); err == nil {
				data.TakenAt = tl.Round(time.Second).UTC()
			} else {
				log.Errorf("metadata: %s in %s (exif time)", err.Error(), logName) // this should never happen
			}
		} else {
			log.Warnf("metadata: invalid time %s in %s (exif)", takenAt, logName)
		}
	}

	if value, ok := tags["Flash"]; ok {
		if i, err := strconv.Atoi(value); err == nil && i&1 == 1 {
			data.AddKeywords(KeywordFlash)
			data.Flash = true
		}
	}

	if value, ok := tags["ImageDescription"]; ok {
		data.AutoAddKeywords(value)
		data.Description = SanitizeDescription(value)
	}

	if value, ok := tags["ProjectionType"]; ok {
		data.AddKeywords(KeywordPanorama)
		data.Projection = SanitizeString(value)
	}

	data.Subject = SanitizeMeta(data.Subject)
	data.Artist = SanitizeMeta(data.Artist)

	data.All = tags

	return nil
}
