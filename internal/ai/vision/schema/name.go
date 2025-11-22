package schema

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/photoprism/photoprism/pkg/clean"
)

const (
	NamePrefix = "photoprism_vision"
)

// JsonSchemaName returns the schema version string to be used for API requests.
func JsonSchemaName(schema json.RawMessage, version string) string {
	var schemaName string

	switch {
	case bytes.Contains(schema, []byte("labels")):
		schemaName = "labels"
	case bytes.Contains(schema, []byte("labels")):
		schemaName = "caption"
	default:
		schemaName = "schema"
	}

	version = clean.TypeLowerUnderscore(version)

	if version == "" {
		version = "v1"
	}

	return fmt.Sprintf("%s_%s_%s", NamePrefix, schemaName, version)

}
