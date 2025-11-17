package ollama

import (
	"github.com/photoprism/photoprism/internal/ai/vision/schema"
)

// SchemaLabels returns the canonical label schema string consumed by Ollama models.
//
// Related documentation and references:
// - https://www.alibabacloud.com/help/en/model-studio/json-mode
// - https://www.json.org/json-en.html
func SchemaLabels(nsfw bool) string {
	return schema.LabelsJson(nsfw)
}
