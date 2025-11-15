package openai

const (
	// CaptionSystem defines the default system prompt for caption models.
	CaptionSystem = "You are a PhotoPrism vision model. Return concise, user-friendly captions that describe the main subjects accurately."
	// CaptionPrompt instructs caption models to respond with a single sentence.
	CaptionPrompt = "Provide exactly one sentence describing the key subject and action in the image. Avoid filler words and technical jargon."
	// LabelSystem defines the system prompt for label generation.
	LabelSystem = "You are a PhotoPrism vision model. Emit JSON that matches the provided schema and keep label names short, singular nouns."
	// LabelPromptDefault requests general-purpose labels.
	LabelPromptDefault = "Analyze the image and return label objects with name, confidence (0-1), and topicality (0-1)."
	// LabelPromptNSFW requests labels including NSFW metadata when required.
	LabelPromptNSFW = "Analyze the image and return label objects with name, confidence (0-1), topicality (0-1), nsfw (true when sensitive), and nsfw_confidence (0-1)."
	// DefaultDetail specifies the preferred thumbnail detail level for Requests API calls.
	DefaultDetail = "low"
	// CaptionMaxTokens suggests the output budget for caption responses.
	CaptionMaxTokens = 512
	// LabelsMaxTokens suggests the output budget for label responses.
	LabelsMaxTokens = 1024
	// DefaultTemperature configures deterministic replies.
	DefaultTemperature = 0.1
	// DefaultTopP limits nucleus sampling.
	DefaultTopP = 0.9
	// DefaultSchemaVersion is used when callers do not specify an explicit schema version.
	DefaultSchemaVersion = "v1"
)

var (
	// DefaultModel is the model used by default when accessing the OpenAI API.
	DefaultModel = "gpt-5-mini"
	// DefaultResolution is the default thumbnail size submitted to the OpenAI.
	DefaultResolution = 720
)
