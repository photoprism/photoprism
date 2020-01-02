package photoprism

import (
	"reflect"
)

type IndexOptions struct {
	UpdateDate     bool
	UpdateColors   bool
	UpdateSize     bool
	UpdateTitle    bool
	UpdateLocation bool
	UpdateCamera   bool
	UpdateLabels   bool
	UpdateKeywords bool
	UpdateXMP      bool
	UpdateExif     bool
}

func (o *IndexOptions) UpdateAny() bool {
	v := reflect.ValueOf(o).Elem()

	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).Bool() {
			return true
		}
	}

	return false
}

func (o *IndexOptions) SkipUnchanged() bool {
	return !o.UpdateAny()
}

// IndexOptionsAll returns new index options with all options set to true.
func IndexOptionsAll() IndexOptions {
	result := IndexOptions{
		UpdateDate:     true,
		UpdateColors:   true,
		UpdateSize:     true,
		UpdateTitle:    true,
		UpdateLocation: true,
		UpdateCamera:   true,
		UpdateLabels:   true,
		UpdateKeywords: true,
		UpdateXMP:      true,
		UpdateExif:     true,
	}

	return result
}

// IndexOptionsNone returns new index options with all options set to false.
func IndexOptionsNone() IndexOptions {
	result := IndexOptions{}

	return result
}
