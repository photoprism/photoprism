package feat

// Feature flags for the ai/vision package:
var (
	VisionModelGenerate = false // controls exposure of the generate endpoint and CLI commands
	VisionModelMarkers  = false // gates marker generation/return until downstream UI and reconciliation paths are ready
	VisionServiceOpenAI = true  // controls whether users are able to configure OpenAI as a vision service engine
)
