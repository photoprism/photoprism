package config

import (
	"reflect"
	"time"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/pkg/txt"
)

// ApplyCliContext applies the values of the cli context based on the "flag" annotations.
func ApplyCliContext(c interface{}, ctx *cli.Context) error {
	v := reflect.ValueOf(c).Elem()

	// Iterate through all config fields.
	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)

		tagValue := v.Type().Field(i).Tag.Get("flag")

		// Assign value to field with "flag" tag.
		if tagValue != "" && tagValue != "-" {
			switch t := fieldValue.Interface().(type) {
			case time.Duration:
				var s string

				// Get duration string.
				if ctx.IsSet(tagValue) {
					s = ctx.String(tagValue)
				} else if ctx.GlobalIsSet(tagValue) || fieldValue.Interface().(time.Duration) == 0 {
					s = ctx.GlobalString(tagValue)
				}

				// Parse duration string.
				if s == "" {
					// Omit.
				} else if sec := txt.UInt(s); sec > 0 {
					fieldValue.Set(reflect.ValueOf(time.Duration(sec) * time.Second))
				} else if d, err := time.ParseDuration(s); err == nil {
					fieldValue.Set(reflect.ValueOf(d))
				}
			case float64:
				// Only if explicitly set or current value is empty (use default).
				if ctx.IsSet(tagValue) {
					f := ctx.Float64(tagValue)
					fieldValue.SetFloat(f)
				} else if ctx.GlobalIsSet(tagValue) || fieldValue.Float() == 0 {
					f := ctx.GlobalFloat64(tagValue)
					fieldValue.SetFloat(f)
				}
			case int, int64:
				// Only if explicitly set or current value is empty (use default).
				if ctx.IsSet(tagValue) {
					f := ctx.Int64(tagValue)
					fieldValue.SetInt(f)
				} else if ctx.GlobalIsSet(tagValue) || fieldValue.Int() == 0 {
					f := ctx.GlobalInt64(tagValue)
					fieldValue.SetInt(f)
				}
			case uint, uint64:
				// Only if explicitly set or current value is empty (use default).
				if ctx.IsSet(tagValue) {
					f := ctx.Uint64(tagValue)
					fieldValue.SetUint(f)
				} else if ctx.GlobalIsSet(tagValue) || fieldValue.Uint() == 0 {
					f := ctx.GlobalUint64(tagValue)
					fieldValue.SetUint(f)
				}
			case string:
				// Only if explicitly set or current value is empty (use default)
				if ctx.IsSet(tagValue) {
					f := ctx.String(tagValue)
					fieldValue.SetString(f)
				} else if ctx.GlobalIsSet(tagValue) || fieldValue.String() == "" {
					f := ctx.GlobalString(tagValue)
					fieldValue.SetString(f)
				}
			case []string:
				if ctx.IsSet(tagValue) {
					f := reflect.ValueOf(ctx.StringSlice(tagValue))
					fieldValue.Set(f)
				} else if ctx.GlobalIsSet(tagValue) || fieldValue.Len() == 0 {
					f := reflect.ValueOf(ctx.GlobalStringSlice(tagValue))
					fieldValue.Set(f)
				}
			case bool:
				if ctx.IsSet(tagValue) {
					f := ctx.Bool(tagValue)
					fieldValue.SetBool(f)
				} else if ctx.GlobalIsSet(tagValue) {
					f := ctx.GlobalBool(tagValue)
					fieldValue.SetBool(f)
				}
			default:
				log.Warnf("cannot assign value of type %s from cli flag %s", t, tagValue)
			}
		}
	}

	return nil
}
