package ollama

import "github.com/photoprism/photoprism/internal/ai/vision/schema"

const (
	// CaptionPrompt instructs Ollama caption models to emit a single, active-voice sentence.
	CaptionPrompt = "Create a caption with exactly one sentence in the active voice that describes the main visual content. Begin with the main subject and clear action. Avoid text formatting, meta-language, and filler words."
	// CaptionModel names the default caption model bundled with our adapter defaults.
	CaptionModel = "gemma3"
	// LabelSystem defines the system prompt shared by Ollama label models.
	LabelSystem = "You are a PhotoPrism vision model. Output concise JSON that matches the schema."
	// LabelPrompt asks the model to return scored labels for the provided image.
	LabelPrompt = "Analyze the image and return label objects with name, confidence (0-1), and topicality (0-1)."
	// DefaultResolution is the default thumbnail size submitted to Ollama models.
	DefaultResolution = 720
)

// LabelsSchema returns the canonical label schema string consumed by Ollama models.
func LabelsSchema() string {
	return schema.LabelsDefaultV1
}
