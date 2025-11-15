package openai

import (
	"encoding/json"

	"github.com/photoprism/photoprism/internal/ai/vision/schema"
)

// SchemaLabels returns the canonical labels JSON Schema string consumed by Ollama models.
//
// Related documentation and references:
// - https://platform.openai.com/docs/guides/structured-outputs
// - https://json-schema.org/learn/miscellaneous-examples
func SchemaLabels(nsfw bool) json.RawMessage {
	return schema.LabelsJsonSchema(nsfw)
}
