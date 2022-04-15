package meta

import (
	"reflect"
	"strings"
	"time"

	"github.com/photoprism/photoprism/pkg/projection"

	"github.com/photoprism/photoprism/pkg/media"

	"github.com/photoprism/photoprism/pkg/clean"
)

// Report returns form fields as table rows for reports.
func Report(f interface{}) (rows [][]string, cols []string) {
	cols = []string{"Tag", "Type", "Exiftool", "Adobe XMP", "Dublin Core"}

	v := reflect.ValueOf(f)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return rows, cols
	}

	rows = make([][]string, 0, v.NumField())

	// Iterate through all form fields.
	for i := 0; i < v.NumField(); i++ {
		if !v.Type().Field(i).IsExported() {
			continue
		}

		fieldValue := v.Field(i)

		fieldName := v.Type().Field(i).Name
		metaTags := v.Type().Field(i).Tag.Get("meta")
		xmpTags := v.Type().Field(i).Tag.Get("xmp")
		dcTags := v.Type().Field(i).Tag.Get("dc")

		// Serialize field values as string.
		if metaTags != "" && metaTags != "-" {
			typeName := "any"

			switch t := fieldValue.Interface().(type) {
			case Keywords:
				typeName = "list"
			case projection.Type, media.Type:
				typeName = "type"
			case time.Duration:
				typeName = "duration"
			case time.Time:
				typeName = "timestamp"
			case int, int8, int16, int32, int64:
				typeName = "number"
			case uint, uint8, uint16, uint32, uint64:
				typeName = "number"
			case float32, float64:
				typeName = "decimal"
			case string:
				typeName = "text"
			case bool:
				typeName = "flag"
			default:
				log.Warnf("failed reporting on %T %s", t, clean.Token(fieldName))
				continue
			}

			metaTags = strings.ReplaceAll(metaTags, ",", ", ")
			xmpTags = strings.ReplaceAll(xmpTags, ",", ", ")
			dcTags = strings.ReplaceAll(dcTags, ",", ", ")

			rows = append(rows, []string{fieldName, typeName, metaTags, xmpTags, dcTags})
		}
	}

	return rows, cols
}
