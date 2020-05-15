package meta

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/photoprism/photoprism/pkg/txt"
	"github.com/tidwall/gjson"
	"gopkg.in/ugjka/go-tz.v2/tz"
)

// JSON parses a json sidecar file (as used by Exiftool) and returns a Data struct.
func JSON(fileName string) (data Data, err error) {
	err = data.JSON(fileName)

	return data, err
}

// JSON parses a json sidecar file (as used by Exiftool) and returns a Data struct.
func (data *Data) JSON(fileName string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%s (json metadata)", e)
		}
	}()

	if data.All == nil {
		data.All = make(map[string]string)
	}

	jsonString, err := ioutil.ReadFile(fileName)

	if err != nil {
		log.Warnf("meta: %s", err.Error())
		return fmt.Errorf("can't read %s (json)", txt.Quote(filepath.Base(fileName)))
	}

	j := gjson.GetBytes(jsonString, "@flatten|@join")

	if !j.IsObject() {
		return fmt.Errorf("data is not an object in %s (json)", txt.Quote(filepath.Base(fileName)))
	}

	jsonValues := j.Map()

	for key, val := range jsonValues {
		data.All[key] = val.String()
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
				if tv, err := time.Parse("2006:01:02 15:04:05", strings.TrimSpace(jsonValue.String())); err == nil {
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
				log.Warnf("meta: can't assign value of type %s to %s", t, tagValue)
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
				log.Warnf("meta: unknown time zone %s", data.TimeZone)
			} else if tl, err := time.ParseInLocation("2006:01:02 15:04:05", data.TakenAtLocal.Format("2006:01:02 15:04:05"), loc); err == nil {
				data.TakenAt = tl.Round(time.Second).UTC()
			} else {
				log.Errorf("meta: %s", err.Error()) // this should never happen
			}
		}
	}

	// Fix rotation.
	if data.Rotation == 90 || data.Rotation == 270 || data.Rotation == -90 {
		data.Width, data.Height = data.Height, data.Width
		data.Rotation = 0
	}

	// Normalize compression information.
	data.Codec = strings.ToLower(data.Codec)
	if strings.Contains(data.Codec, CodecJpeg) {
		data.Codec = CodecJpeg
	}

	return nil
}
