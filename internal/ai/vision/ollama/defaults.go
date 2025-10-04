package ollama

import "github.com/photoprism/photoprism/internal/ai/vision/schema"

const (
	// CaptionPrompt instructs Ollama caption models to emit a single, active-voice sentence.
	CaptionPrompt = "Create a caption with exactly one sentence in the active voice that describes the main visual content. Begin with the main subject and clear action. Avoid text formatting, meta-language, and filler words."
	// CaptionModel names the default caption model bundled with our adapter defaults.
	CaptionModel = "gemma3"
	// DefaultLabelConfidence is used when the model omits the confidence field.
	DefaultLabelConfidence = 0.5
	// LabelSystemSimple defines a simple system prompt for Ollama label models.
	LabelSystemSimple = "You are a PhotoPrism vision model. Output concise JSON that matches the schema."
	// LabelPromptSimple defines a simple user prompt for Ollama label models.
	LabelPromptSimple = "Analyze the image and return label objects with name, confidence (0-1), and topicality (0-1)."
	// LabelSystem defines the system prompt shared by Ollama label models. It aims to ensure that single-word nouns are returned.
	LabelSystem = "You are a PhotoPrism vision model. Output concise JSON that matches the schema. Each label name MUST be a single-word noun in its canonical singular form. Avoid spaces, punctuation, emoji, or descriptive phrases."
	// LabelPrompt asks the model to return scored labels for the provided image. It aims to ensure that single-word nouns are returned.
	LabelPrompt = "Analyze the image and return label objects with name (single-word noun), confidence (0-1), and topicality (0-1). Respond with JSON exactly like {\"labels\":[{\"name\":\"sunset\",\"confidence\":0.72,\"topicality\":0.64}]} and adjust the values for this photo."
	// DefaultResolution is the default thumbnail size submitted to Ollama models.
	DefaultResolution = 720
)

// LabelsSchema returns the canonical label schema string consumed by Ollama models.
func LabelsSchema() string {
	return schema.LabelsDefaultV1
}
