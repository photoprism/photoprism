package openai

import "github.com/photoprism/photoprism/internal/ai/vision/schema"

var (
	// DefaultModel is the model used by default when accessing the OpenAI API.
	DefaultModel = "gpt-5-mini"
	// DefaultResolution is the default thumbnail size submitted to the OpenAI.
	DefaultResolution = 720
)

// LabelsSchema returns the canonical label schema string consumed by OpenAI models.
func LabelsSchema() string {
	return schema.LabelsDefault
}
