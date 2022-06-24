package meta

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/photoprism/photoprism/pkg/video"

	"github.com/photoprism/photoprism/pkg/projection"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
	"github.com/tidwall/gjson"
	"gopkg.in/photoprism/go-tz.v2/tz"
)

const MimeVideoMP4 = "video/mp4"
const MimeQuicktime = "video/quicktime"

// Exiftool parses JSON sidecar data as created by Exiftool.
func (data *Data) Exiftool(jsonData []byte, originalName string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("metadata: %s (exiftool panic)\nstack: %s", e, debug.Stack())
		}
	}()

	j := gjson.GetBytes(jsonData, "@flatten|@join")

	if !j.IsObject() {
		return fmt.Errorf("metadata: data is not an object in %s (exiftool)", clean.Log(filepath.Base(originalName)))
	}

	jsonStrings := make(map[string]string)
	jsonValues := j.Map()

	for key, val := range jsonValues {
		jsonStrings[key] = val.String()
	}

	if fileName, ok := jsonStrings["FileName"]; ok && fileName != "" && originalName != "" && fileName != originalName {
		return fmt.Errorf("metadata: original name %s does not match %s (exiftool)", clean.Log(originalName), clean.Log(fileName))
	}

	v := reflect.ValueOf(data).Elem()

	// Iterate through all config fields
	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)

		tagData := v.Type().Field(i).Tag.Get("meta")

		// Automatically assign values to fields with "flag" tag
		if tagData != "" {
			tagValues := strings.Split(tagData, ",")

			var jsonValue gjson.Result
			var tagValue string

			for _, tagValue = range tagValues {
				if r, ok := jsonValues[tagValue]; !ok {
					continue
				} else {
					jsonValue = r
					break
				}
			}

			// Skip empty values.
			if !jsonValue.Exists() {
				continue
			}

			switch t := fieldValue.Interface().(type) {
			case time.Time:
				if !fieldValue.IsZero() {
					continue
				}

				if dateTime := txt.DateTime(jsonValue.String(), ""); !dateTime.IsZero() {
					fieldValue.Set(reflect.ValueOf(dateTime))
				}
			case time.Duration:
				if !fieldValue.IsZero() {
					continue
				}

				fieldValue.Set(reflect.ValueOf(StringToDuration(jsonValue.String())))
			case int, int64:
				if !fieldValue.IsZero() {
					continue
				}

				fieldValue.SetInt(jsonValue.Int())
			case float32, float64:
				if !fieldValue.IsZero() {
					continue
				}

				fieldValue.SetFloat(jsonValue.Float())
			case uint, uint64:
				if !fieldValue.IsZero() {
					continue
				}

				fieldValue.SetUint(jsonValue.Uint())
			case []string:
				existing := fieldValue.Interface().([]string)
				fieldValue.Set(reflect.ValueOf(txt.AddToWords(existing, strings.TrimSpace(jsonValue.String()))))
			case Keywords:
				existing := fieldValue.Interface().(Keywords)
				fieldValue.Set(reflect.ValueOf(txt.AddToWords(existing, strings.TrimSpace(jsonValue.String()))))
			case projection.Type:
				if !fieldValue.IsZero() {
					continue
				}

				fieldValue.Set(reflect.ValueOf(projection.Type(strings.TrimSpace(jsonValue.String()))))
			case string:
				if !fieldValue.IsZero() {
					continue
				}

				fieldValue.SetString(strings.TrimSpace(jsonValue.String()))
			case bool:
				if !fieldValue.IsZero() {
					continue
				}

				fieldValue.SetBool(jsonValue.Bool())
			default:
				log.Warnf("metadata: cannot assign value of type %s to %s (exiftool)", t, tagValue)
			}
		}
	}

	// Nanoseconds.
	if data.TakenNs <= 0 {
		for _, name := range exifSubSecTags {
			if s := jsonStrings[name]; txt.IsPosInt(s) {
				data.TakenNs = txt.Int(s + strings.Repeat("0", 9-len(s)))
				break
			}
		}
	}

	// Set latitude and longitude if known and not already set.
	if data.Lat == 0 && data.Lng == 0 {
		if data.GPSPosition != "" {
			data.Lat, data.Lng = GpsToLatLng(data.GPSPosition)
		} else if data.GPSLatitude != "" && data.GPSLongitude != "" {
			data.Lat = GpsToDecimal(data.GPSLatitude)
			data.Lng = GpsToDecimal(data.GPSLongitude)
		}
	}

	if data.Altitude == 0 {
		// Parseable floating point number?
		if fl := GpsFloatRegexp.FindAllString(jsonStrings["GPSAltitude"], -1); len(fl) != 1 {
			// Ignore.
		} else if alt, err := strconv.ParseFloat(fl[0], 64); err == nil && alt != 0 {
			data.Altitude = int(alt)
		}
	}

	hasTimeOffset := false

	// Fallback to GPS timestamp.
	if data.TakenAt.IsZero() && data.TakenAtLocal.IsZero() && !data.TakenGps.IsZero() {
		data.TimeZone = time.UTC.String()
		data.TakenAt = data.TakenGps.UTC()
		data.TakenAtLocal = time.Time{}
	}

	if _, offset := data.TakenAtLocal.Zone(); offset != 0 && !data.TakenAtLocal.IsZero() {
		hasTimeOffset = true
	} else if mt, ok := jsonStrings["MIMEType"]; ok && (mt == MimeVideoMP4 || mt == MimeQuicktime) {
		// Assume default time zone for MP4 & Quicktime videos is UTC.
		// see https://exiftool.org/TagNames/QuickTime.html
		data.TimeZone = time.UTC.String()
		data.TakenAt = data.TakenAt.UTC()
		data.TakenAtLocal = time.Time{}
	}

	// Set time zone and calculate UTC time.
	if data.Lat != 0 && data.Lng != 0 {
		zones, err := tz.GetZone(tz.Point{
			Lat: float64(data.Lat),
			Lon: float64(data.Lng),
		})

		if err == nil && len(zones) > 0 {
			data.TimeZone = zones[0]
		}

		if loc, err := time.LoadLocation(data.TimeZone); err != nil {
			log.Warnf("metadata: unknown time zone %s (exiftool)", data.TimeZone)
		} else if !data.TakenAtLocal.IsZero() {
			if tl, err := time.ParseInLocation("2006:01:02 15:04:05", data.TakenAtLocal.Format("2006:01:02 15:04:05"), loc); err == nil {
				if localUtc, err := time.ParseInLocation("2006:01:02 15:04:05", data.TakenAtLocal.Format("2006:01:02 15:04:05"), time.UTC); err == nil {
					data.TakenAtLocal = localUtc
				}

				data.TakenAt = tl.Truncate(time.Second).UTC()
			} else {
				log.Errorf("metadata: %s (exiftool)", err.Error()) // this should never happen
			}
		} else if !data.TakenAt.IsZero() {
			if localUtc, err := time.ParseInLocation("2006:01:02 15:04:05", data.TakenAt.In(loc).Format("2006:01:02 15:04:05"), time.UTC); err == nil {
				data.TakenAtLocal = localUtc
				data.TakenAt = data.TakenAt.UTC()
			} else {
				log.Errorf("metadata: %s (exiftool)", err.Error()) // this should never happen
			}
		}
	} else if hasTimeOffset {
		if localUtc, err := time.ParseInLocation("2006:01:02 15:04:05", data.TakenAtLocal.Format("2006:01:02 15:04:05"), time.UTC); err == nil {
			data.TakenAtLocal = localUtc
		}

		data.TakenAt = data.TakenAt.Truncate(time.Second).UTC()
	}

	// Set local time if still empty.
	if data.TakenAtLocal.IsZero() && !data.TakenAt.IsZero() {
		if loc, err := time.LoadLocation(data.TimeZone); data.TimeZone == "" || err != nil {
			data.TakenAtLocal = data.TakenAt
		} else if localUtc, err := time.ParseInLocation("2006:01:02 15:04:05", data.TakenAt.In(loc).Format("2006:01:02 15:04:05"), time.UTC); err == nil {
			data.TakenAtLocal = localUtc
			data.TakenAt = data.TakenAt.UTC()
		} else {
			log.Errorf("metadata: %s (exiftool)", err.Error()) // this should never happen
		}
	}

	// Add nanoseconds to the calculated UTC and local time.
	if data.TakenAt.Nanosecond() == 0 {
		if ns := time.Duration(data.TakenNs); ns > 0 && ns <= time.Second {
			data.TakenAt.Truncate(time.Second).UTC().Add(ns)
			data.TakenAtLocal.Truncate(time.Second).Add(ns)
		}
	}

	// Use actual image width and height if available, see issue #2447.
	if jsonValues["ImageWidth"].Exists() && jsonValues["ImageHeight"].Exists() {
		if val := jsonValues["ImageWidth"].Int(); val > 0 {
			data.Width = int(val)
		}

		if val := jsonValues["ImageHeight"].Int(); val > 0 {
			data.Height = int(val)
		}
	}

	// Image orientation, see https://www.daveperrett.com/articles/2012/07/28/exif-orientation-handling-is-a-ghetto/.
	if orientation, ok := jsonStrings["Orientation"]; ok && orientation != "" {
		switch orientation {
		case "1", "Horizontal (normal)":
			data.Orientation = 1
		case "2":
			data.Orientation = 2
		case "3", "Rotate 180 CW":
			data.Orientation = 3
		case "4":
			data.Orientation = 4
		case "5":
			data.Orientation = 5
		case "6", "Rotate 90 CW":
			data.Orientation = 6
		case "7":
			data.Orientation = 7
		case "8", "Rotate 270 CW":
			data.Orientation = 8
		}
	}

	if data.Orientation == 0 {
		// Set orientation based on rotation.
		switch data.Rotation {
		case 0:
			data.Orientation = 1
		case -180, 180:
			data.Orientation = 3
		case 90:
			data.Orientation = 6
		case -90, 270:
			data.Orientation = 8
		}
	}

	// Normalize codec name.
	data.Codec = strings.ToLower(data.Codec)
	if strings.Contains(data.Codec, CodecJpeg) { // JPEG Image?
		data.Codec = CodecJpeg
	} else if c, ok := video.Codecs[data.Codec]; ok { // Video codec?
		data.Codec = string(c)
	} else if strings.HasPrefix(data.Codec, "a_") { // Audio codec?
		data.Codec = ""
	}

	// Validate and normalize optional DocumentID.
	if data.DocumentID != "" {
		data.DocumentID = rnd.SanitizeUUID(data.DocumentID)
	}

	// Validate and normalize optional InstanceID.
	if data.InstanceID != "" {
		data.InstanceID = rnd.SanitizeUUID(data.InstanceID)
	}

	if projection.Equirectangular.Equal(data.Projection) {
		data.AddKeywords(KeywordPanorama)
	}

	if data.Description != "" {
		data.AutoAddKeywords(data.Description)
		data.Description = SanitizeDescription(data.Description)
	}

	data.Title = SanitizeTitle(data.Title)
	data.Subject = SanitizeMeta(data.Subject)
	data.Artist = SanitizeMeta(data.Artist)

	return nil
}
