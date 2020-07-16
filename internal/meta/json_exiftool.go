package meta

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime/debug"
	"strings"
	"time"

	"github.com/photoprism/photoprism/pkg/txt"
	"github.com/tidwall/gjson"
	"gopkg.in/ugjka/go-tz.v2/tz"
)

// Parses JSON sidecar data as created by Exiftool.
func (data *Data) Exiftool(jsonData []byte, originalName string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("metadata: %s (exiftool panic)\nstack: %s", e, debug.Stack())
		}
	}()

	j := gjson.GetBytes(jsonData, "@flatten|@join")

	if !j.IsObject() {
		return fmt.Errorf("metadata: data is not an object in %s (exiftool)", txt.Quote(filepath.Base(originalName)))
	}

	jsonStrings := make(map[string]string)
	jsonValues := j.Map()

	for key, val := range jsonValues {
		jsonStrings[key] = val.String()
	}

	if fileName, ok := jsonStrings["FileName"]; ok && fileName != "" && originalName != "" && fileName != originalName {
		return fmt.Errorf("metadata: original name %s does not match %s (exiftool)", txt.Quote(originalName), txt.Quote(fileName))
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

			if !jsonValue.Exists() {
				continue
			}

			switch t := fieldValue.Interface().(type) {
			case time.Time:
				s := strings.TrimSpace(jsonValue.String())
				s = strings.ReplaceAll(s, "/", ":")
				s = strings.ReplaceAll(s, "-", ":")

				if tv, err := time.Parse("2006:01:02 15:04:05", s); err == nil {
					fieldValue.Set(reflect.ValueOf(tv.Round(time.Second).UTC()))
				}
			case time.Duration:
				fieldValue.Set(reflect.ValueOf(StringToDuration(jsonValue.String())))
			case int, int64:
				fieldValue.SetInt(jsonValue.Int())
			case float32, float64:
				fieldValue.SetFloat(jsonValue.Float())
			case uint, uint64:
				fieldValue.SetUint(jsonValue.Uint())
			case string:
				fieldValue.SetString(strings.TrimSpace(jsonValue.String()))
			case bool:
				fieldValue.SetBool(jsonValue.Bool())
			default:
				log.Warnf("metadata: can't assign value of type %s to %s (exiftool)", t, tagValue)
			}
		}
	}

	// Calculate latitude and longitude if exists.
	if data.GPSPosition != "" {
		data.Lat, data.Lng = GpsToLatLng(data.GPSPosition)
	} else if data.GPSLatitude != "" && data.GPSLongitude != "" {
		data.Lat = GpsToDecimal(data.GPSLatitude)
		data.Lng = GpsToDecimal(data.GPSLongitude)
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

		if !data.TakenAtLocal.IsZero() {
			if loc, err := time.LoadLocation(data.TimeZone); err != nil {
				log.Warnf("metadata: unknown time zone %s (exiftool)", data.TimeZone)
			} else if tl, err := time.ParseInLocation("2006:01:02 15:04:05", data.TakenAtLocal.Format("2006:01:02 15:04:05"), loc); err == nil {
				data.TakenAt = tl.Round(time.Second).UTC()
			} else {
				log.Errorf("metadata: %s (exiftool)", err.Error()) // this should never happen
			}
		}
	}

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

	// Normalize compression information.
	data.Codec = strings.ToLower(data.Codec)
	if strings.Contains(data.Codec, CodecJpeg) {
		data.Codec = CodecJpeg
	}

	// Validate and normalize optional DocumentID.
	if len(data.DocumentID) > 0 {
		data.DocumentID = SanitizeUID(data.DocumentID)
	}

	// Validate and normalize optional InstanceID.
	if len(data.InstanceID) > 0 {
		data.InstanceID = SanitizeUID(data.InstanceID)
	}

	if data.Projection == "equirectangular" {
		data.AddKeyword(KeywordPanorama)
	}

	data.Title = SanitizeTitle(data.Title)
	data.Description = SanitizeDescription(data.Description)

	return nil
}
