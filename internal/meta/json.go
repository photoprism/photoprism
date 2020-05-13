package meta

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/photoprism/photoprism/pkg/txt"
	"github.com/tidwall/gjson"
)

// JSON parses a json sidecar file (as used by Exiftool) and returns a Data struct.
func JSON(filename string) (data Data, err error) {
	err = data.JSON(filename)

	return data, err
}

// JSON parses a json sidecar file (as used by Exiftool) and returns a Data struct.
func (data *Data) JSON(filename string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("meta: %s", e)
		}
	}()

	if data.All == nil {
		data.All = make(map[string]string)
	}

	jsonString, err := ioutil.ReadFile(filename)

	if err != nil {
		return err
	}

	j := gjson.GetBytes(jsonString, "@flatten|@join")

	if !j.IsObject() {
		return fmt.Errorf("meta: json is not an object (%s)", txt.Quote(filepath.Base(filename)))
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
					fieldValue.Set(reflect.ValueOf(tv))
				}
			case time.Duration:
				if n := strings.Split(strings.TrimSpace(jsonValue.String()), ":"); len(n) == 3 {
					h, _ := strconv.Atoi(n[0])
					m, _ := strconv.Atoi(n[1])
					s, _ := strconv.Atoi(n[2])

					dv := time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(s)*time.Second

					fieldValue.Set(reflect.ValueOf(dv))
				}
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

	return nil
}
