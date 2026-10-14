package config

import (
	"reflect"
	"strconv"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/pkg/txt"
)

// LocalFlag marks a command flag that has the name of a config flag but a different meaning, so that
// its value never applies to the config.
type LocalFlag interface {
	cli.Flag
	LocalFlag()
}

// flagContext returns the context that holds the value of a config flag: the nearest one that sets it,
// or else the outermost, so that a command flag of the same name that is not set does not hide the
// global value or its environment variable. A LocalFlag is skipped.
func flagContext(ctx *cli.Context, name string) *cli.Context {
	outer := ctx

	for _, c := range ctx.Lineage() {
		if c.App == nil {
			continue
		}

		outer = c

		if isLocalFlag(c.Command, name) {
			continue
		} else if c.IsSet(name) {
			return c
		}
	}

	return outer
}

// isLocalFlag reports whether a command defines the named flag as a LocalFlag.
func isLocalFlag(cmd *cli.Command, name string) bool {
	if cmd == nil {
		return false
	}

	for _, f := range cmd.Flags {
		if _, ok := f.(LocalFlag); !ok {
			continue
		}

		for _, n := range f.Names() {
			if n == name {
				return true
			}
		}
	}

	return false
}

// ApplyCliContext applies the values of the cli context based on the "flag" annotations.
func ApplyCliContext(c any, ctx *cli.Context) error {
	v := reflect.ValueOf(c).Elem()

	// Iterate through all config fields.
	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)

		tagValue := v.Type().Field(i).Tag.Get("flag")

		// Assign value to field with "flag" tag.
		if tagValue != "" && tagValue != "-" {
			fc := flagContext(ctx, tagValue)

			switch t := fieldValue.Interface().(type) {
			case time.Duration:
				var s string

				// Get duration string.
				if fc.IsSet(tagValue) || fieldValue.Interface().(time.Duration) == 0 {
					s = fc.String(tagValue)
				}

				// Parse duration string.
				if s == "" {
					// Omit.
				} else if sec := txt.UInt(s); sec > 0 {
					fieldValue.Set(reflect.ValueOf(time.Duration(sec) * time.Second)) //nolint:gosec // txt.UInt is bounded; duration uses int64 on supported platforms
				} else if d, err := time.ParseDuration(s); err == nil {
					fieldValue.Set(reflect.ValueOf(d))
				}
			case float64:
				// Only if explicitly set or current value is empty (use default).
				if fc.IsSet(tagValue) || fieldValue.Float() == 0 {
					f := fc.Float64(tagValue)
					fieldValue.SetFloat(f)
				}
			case int, int64:
				// Only if explicitly set or current value is empty (use default).
				if fc.IsSet(tagValue) || fieldValue.Int() == 0 {
					f := fc.Int64(tagValue)
					fieldValue.SetInt(f)
				}
			case uint, uint64:
				// Only if explicitly set or current value is empty (use default).
				if fc.IsSet(tagValue) || fieldValue.Uint() == 0 {
					f := fc.Uint64(tagValue)
					fieldValue.SetUint(f)
				}
			case string:
				// Only if explicitly set or current value is empty (use default)
				if fc.IsSet(tagValue) || fieldValue.String() == "" {
					f := fc.String(tagValue)
					fieldValue.SetString(f)
				}
			case []string:
				if fc.IsSet(tagValue) || fieldValue.Len() == 0 {
					f := reflect.ValueOf(fc.StringSlice(tagValue))
					fieldValue.Set(f)
				}
			case []float64:
				if fc.IsSet(tagValue) || fieldValue.Len() == 0 {
					var floats []float64

					for _, s := range fc.StringSlice(tagValue) {
						if s == "" {
							continue
						}

						if f, err := strconv.ParseFloat(s, 64); err != nil {
							log.Warnf("cannot parse value %q for cli flag %s (%s)", s, tagValue, err)
							continue
						} else {
							floats = append(floats, f)
						}
					}

					fieldValue.Set(reflect.ValueOf(floats))
				}
			case bool:
				if fc.IsSet(tagValue) {
					f := fc.Bool(tagValue)
					fieldValue.SetBool(f)
				}
			default:
				log.Warnf("cannot assign value of type %s from cli flag %s", t, tagValue)
			}
		}
	}

	return nil
}
