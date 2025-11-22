package schema

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJsonSchemaName(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		assert.Equal(t, "photoprism_vision_schema_v1", JsonSchemaName(nil, ""))
	})
	t.Run("Labels", func(t *testing.T) {
		assert.Equal(t, "photoprism_vision_labels_v1", JsonSchemaName(json.RawMessage(LabelsJsonSchemaDefault), ""))
	})
	t.Run("LabelsV1", func(t *testing.T) {
		assert.Equal(t, "photoprism_vision_labels_v2", JsonSchemaName([]byte("labels"), "v2"))
	})
	t.Run("LabelsJsonSchema", func(t *testing.T) {
		assert.Equal(t, "photoprism_vision_labels_v1", JsonSchemaName(LabelsJsonSchema(false), "v1"))
	})
}
