package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/pkg/list"
)

// Report returns global config values as a table for reporting.
func (o Options) Report() (rows [][]string, cols []string) {
	v := reflect.ValueOf(o)
	t := v.Type()
	activeTags := []string{Features}

	if o.NodeRole == cluster.RolePortal {
		activeTags = append(activeTags, Portal)
	}

	cols = []string{"Name", "Type", "CLI Flag"}
	rows = make([][]string, 0, v.NumField())

	// Iterate through all config fields.
	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		field := t.Field(i)

		yamlName := strings.SplitN(field.Tag.Get("yaml"), ",", 2)[0]
		flagName := strings.SplitN(field.Tag.Get("flag"), ",", 2)[0]

		if yamlName == "" || yamlName == "-" || flagName == "" || flagName == "-" {
			continue
		}

		// Skip options by feature set if tags are set.
		if tags := field.Tag.Get("tags"); tags == "" {
			// Report.
		} else if !list.ContainsAny(strings.Split(tags, ","), activeTags) {
			// Skip.
			continue
		}

		fieldType := fmt.Sprintf("%T", fieldValue.Interface())

		rows = append(rows, []string{yamlName, fieldType, "--" + flagName})
	}

	return rows, cols
}
