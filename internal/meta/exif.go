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
	"gopkg.in/photoprism/go-tz.v2/tz"

	exifcommon "github.com/dsoprea/go-exif/v3/common"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media/projection"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

var exifIfdMapping *exifcommon.IfdMapping
var exifTagIndex = exif.NewTagIndex()
var exifMutex = sync.Mutex{}
var exifDateTimeTags = []string{"DateTimeOriginal", "DateTimeCreated", "CreateDate", "DateTime", "DateTimeDigitized"}
var exifSubSecTags = []string{"SubSecTimeOriginal", "SubSecTime", "SubSecTimeDigitized"}

func init() {
	exifIfdMapping = exifcommon.NewIfdMapping()

	if err := exifcommon.LoadStandardIfds(exifIfdMapping); err != nil {
		log.Errorf("metadata: %s", err.Error())
	}
}

// Exif parses an image file for Exif metadata and returns as Data struct.
func Exif(fileName string, fileType fs.Type, bruteForce bool) (data Data, err error) {
	err = data.Exif(fileName, fileType, bruteForce)

	return data, err
}

// Exif parses an image file for Exif metadata and returns as Data struct.
func (data *Data) Exif(fileName string, fileFormat fs.Type, bruteForce bool) (err error) {
	exifMutex.Lock()
	defer exifMutex.Unlock()

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("metadata: %s in %s (exif panic)\nstack: %s", e, clean.Log(filepath.Base(fileName)), debug.Stack())
		}
	}()

	// Resolve file name e.g. in case it's a symlink.
	if fileName, err = fs.Resolve(fileName); err != nil {
		return err
	}

	// Extract raw Exif block.
	rawExif, err := RawExif(fileName, fileFormat, bruteForce)

	if err != nil {
		return err
	}

	logName := clean.Log(filepath.Base(fileName))

	// Enumerate data.exif in Exif block.
	opt := exif.ScanOptions{}
	entries, _, err := exif.GetFlatExifData(rawExif, &opt)

	// Create large enough map for values.
	if data.exif == nil {
		data.exif = make(map[string]string, len(entries))
	}

	// Ignore IFD1 data.exif with existing IFD0 values.
	// see https://github.com/photoprism/photoprism/issues/2231
	for _, tag := range entries {
		s := strings.Split(tag.FormattedFirst, "\x00")
		if tag.TagName == "" || len(s) == 0 {
			// Do nothing.
		} else if s[0] != "" && (data.exif[tag.TagName] == "" || tag.IfdPath != exif.ThumbnailFqIfdPath) {
			data.exif[tag.TagName] = s[0]
		}
	}

	// Abort if no values were found.
	if len(data.exif) == 0 {
		return fmt.Errorf("metadata: no exif data in %s", logName)
	}

	var ifdIndex exif.IfdIndex
	_, ifdIndex, err = exif.Collect(exifIfdMapping, exifTagIndex, rawExif)

	// Find and parse GPS coordinates.
	if err != nil {
		log.Debugf("metadata: %s in %s (exif collect)", err, logName)
	} else {
		var ifd *exif.Ifd
		if ifd, err = ifdIndex.RootIfd.ChildWithIfdPath(exifcommon.IfdGpsInfoStandardIfdIdentity); err == nil {
			var gi *exif.GpsInfo
			if gi, err = ifd.GpsInfo(); err != nil {
				log.Debugf("metadata: %s in %s (exif gps-info)", err, logName)
			} else {
				if !math.IsNaN(gi.Latitude.Decimal()) && !math.IsNaN(gi.Longitude.Decimal()) {
					data.Lat, data.Lng = NormalizeGPS(gi.Latitude.Decimal(), gi.Longitude.Decimal())
				} else if gi.Altitude != 0 || !gi.Timestamp.IsZero() {
					log.Warnf("metadata: invalid exif gps coordinates in %s (%s)", logName, clean.Log(gi.String()))
				}

				if gi.Altitude != 0 {
					data.Altitude = float64(gi.Altitude)
				}

				if !gi.Timestamp.IsZero() {
					data.TakenGps = gi.Timestamp
				}
			}
		}
	}

	if value, ok := data.exif["Artist"]; ok {
		data.Artist = SanitizeString(value)
	}

	if value, ok := data.exif["Copyright"]; ok {
		data.Copyright = SanitizeString(value)
	}

	// Ignore numeric model names as they are probably invalid.
	if value, ok := data.exif["CameraModel"]; ok && !txt.IsUInt(value) {
		data.CameraModel = SanitizeString(value)
	} else if value, ok = data.exif["Model"]; ok && !txt.IsUInt(value) {
		data.CameraModel = SanitizeString(value)
	} else if value, ok = data.exif["UniqueCameraModel"]; ok && !txt.IsUInt(value) {
		data.CameraModel = SanitizeString(value)
	}

	if value, ok := data.exif["CameraMake"]; ok && !txt.IsUInt(value) {
		data.CameraMake = SanitizeString(value)
	} else if value, ok = data.exif["Make"]; ok && !txt.IsUInt(value) {
		data.CameraMake = SanitizeString(value)
	}

	if value, ok := data.exif["CameraOwnerName"]; ok {
		data.CameraOwner = SanitizeString(value)
	}

	if value, ok := data.exif["BodySerialNumber"]; ok {
		data.CameraSerial = SanitizeString(value)
	}

	if value, ok := data.exif["LensMake"]; ok && !txt.IsUInt(value) {
		data.LensMake = SanitizeString(value)
	}

	// Ignore numeric model names as they are probably invalid.
	if value, ok := data.exif["LensModel"]; ok && !txt.IsUInt(value) {
		data.LensModel = SanitizeString(value)
	} else if value, ok = data.exif["Lens"]; ok && !txt.IsUInt(value) {
		data.LensModel = SanitizeString(value)
	}

	if value, ok := data.exif["Software"]; ok {
		data.Software = SanitizeString(value)
	}

	if value, ok := data.exif["ExposureTime"]; ok {
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

	if value, ok := data.exif["FNumber"]; ok {
		values := strings.Split(value, "/")

		if len(values) == 2 && values[1] != "0" && values[1] != "" {
			number, _ := strconv.ParseFloat(values[0], 64)
			denom, _ := strconv.ParseFloat(values[1], 64)

			data.FNumber = float32(math.Round((number/denom)*1000) / 1000)
		}
	}

	if value, ok := data.exif["ApertureValue"]; ok {
		values := strings.Split(value, "/")

		if len(values) == 2 && values[1] != "0" && values[1] != "" {
			number, _ := strconv.ParseFloat(values[0], 64)
			denom, _ := strconv.ParseFloat(values[1], 64)

			data.Aperture = float32(math.Round((number/denom)*1000) / 1000)
		}
	}

	if value, ok := data.exif["FocalLengthIn35mmFilm"]; ok {
		if i, err := strconv.Atoi(value); err == nil {
			data.FocalLength = i
		}
	} else if v, ok := data.exif["FocalLength"]; ok {
		values := strings.Split(v, "/")

		if len(values) == 2 && values[1] != "0" && values[1] != "" {
			number, _ := strconv.ParseFloat(values[0], 64)
			denom, _ := strconv.ParseFloat(values[1], 64)

			data.FocalLength = int(math.Round((number/denom)*1000) / 1000)
		}
	}

	if value, ok := data.exif["ISOSpeedRatings"]; ok {
		if i, err := strconv.Atoi(value); err == nil {
			data.Iso = i
		}
	}

	if value, ok := data.exif["ImageUniqueID"]; ok {
		if id := rnd.SanitizeUUID(value); id != "" {
			data.DocumentID = id
		}
	}

	if value, ok := data.exif["PixelXDimension"]; ok {
		if i, err := strconv.Atoi(value); err == nil {
			data.Width = i
		}
	} else if value, ok := data.exif["ImageWidth"]; ok {
		if i, err := strconv.Atoi(value); err == nil {
			data.Width = i
		}
	}

	if value, ok := data.exif["PixelYDimension"]; ok {
		if i, err := strconv.Atoi(value); err == nil {
			data.Height = i
		}
	} else if value, ok := data.exif["ImageLength"]; ok {
		if i, err := strconv.Atoi(value); err == nil {
			data.Height = i
		}
	}

	if value, ok := data.exif["Orientation"]; ok {
		if i, err := strconv.Atoi(value); err == nil {
			data.Orientation = i
		}
	} else {
		data.Orientation = 1
	}

	if data.Lat != 0 && data.Lng != 0 {
		zones, err := tz.GetZone(tz.Point{
			Lat: data.Lat,
			Lon: data.Lng,
		})

		if err == nil && len(zones) > 0 {
			data.TimeZone = zones[0]
		}
	}

	takenAt := time.Time{}

	for _, name := range exifDateTimeTags {
		if dateTime := txt.ParseTime(data.exif[name], data.TimeZone); !dateTime.IsZero() {
			takenAt = dateTime
			break
		}
	}

	// Fallback to GPS timestamp.
	if takenAt.IsZero() && !data.TakenGps.IsZero() {
		takenAt = data.TakenGps.UTC()
	}

	// Nanoseconds.
	if data.TakenNs <= 0 {
		for _, name := range exifSubSecTags {
			if s := data.exif[name]; txt.IsPosInt(s) {
				data.TakenNs = txt.Int(s + strings.Repeat("0", 9-len(s)))
				break
			}
		}
	}

	// UniqueID time found in Exif metadata?
	if !takenAt.IsZero() {
		if takenAtLocal, err := time.ParseInLocation("2006-01-02T15:04:05", takenAt.Format("2006-01-02T15:04:05"), time.UTC); err == nil {
			data.TakenAtLocal = takenAtLocal
		} else {
			data.TakenAtLocal = takenAt
		}

		data.TakenAt = takenAt.UTC()
	}

	// Add nanoseconds to the calculated UTC and local time.
	if data.TakenAt.Nanosecond() == 0 {
		if ns := time.Duration(data.TakenNs); ns > 0 && ns <= time.Second {
			data.TakenAt.Truncate(time.Second).UTC().Add(ns)
			data.TakenAtLocal.Truncate(time.Second).Add(ns)
		}
	}

	if value, ok := data.exif["Flash"]; ok {
		if i, err := strconv.Atoi(value); err == nil && i&1 == 1 {
			data.AddKeywords(KeywordFlash)
			data.Flash = true
		}
	}

	if value, ok := data.exif["ImageDescription"]; ok {
		data.AutoAddKeywords(value)
		data.Description = SanitizeDescription(value)
	}

	if value, ok := data.exif["ProjectionType"]; ok {
		data.AddKeywords(KeywordPanorama)
		data.Projection = projection.New(SanitizeString(value)).String()
	}

	data.Subject = SanitizeMeta(data.Subject)
	data.Artist = SanitizeMeta(data.Artist)

	return nil
}
