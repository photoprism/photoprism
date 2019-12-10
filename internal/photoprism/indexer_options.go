package photoprism

import (
	"reflect"
)

type IndexerOptions struct {
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

func (o *IndexerOptions) UpdateAny() bool {
	v := reflect.ValueOf(o).Elem()

	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).Bool() {
			return true
		}
	}

	return false
}

func (o *IndexerOptions) SkipUnchanged() bool {
	return !o.UpdateAny()
}

// IndexerOptionsAll returns new indexer options with all options set to true.
func IndexerOptionsAll() IndexerOptions {
	instance := IndexerOptions{
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

	return instance
}

// IndexerOptionsNone returns new indexer options with all options set to false.
func IndexerOptionsNone() IndexerOptions {
	instance := IndexerOptions{}

	return instance
}
