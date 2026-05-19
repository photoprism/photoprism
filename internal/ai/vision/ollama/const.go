package ollama

const (
	// EngineName is the canonical identifier for Ollama-based vision services.
	EngineName = "ollama"
	// ApiFormat identifies Ollama-compatible request and response payloads.
	ApiFormat = "ollama"
	// APIKeyEnv defines the environment variable used for Ollama API tokens.
	APIKeyEnv = "OLLAMA_API_KEY" //nolint:gosec // environment variable name, not a secret
	// APIKeyFileEnv defines the file-based fallback environment variable for Ollama API tokens.
	APIKeyFileEnv = "OLLAMA_API_KEY_FILE" //nolint:gosec // environment variable name, not a secret
	// APIKeyPlaceholder is the `${VAR}` form injected when no explicit key is provided.
	APIKeyPlaceholder = "${" + APIKeyEnv + "}"
	// BaseUrlEnv defines the environment variable used for the Ollama base URL e.g. "https://ollama.com" or "http://ollama:11434".
	BaseUrlEnv = "OLLAMA_BASE_URL" //nolint:gosec // environment variable name, not a secret
	// BaseUrlPlaceholder is the `${VAR}` form injected when no explicit URL is provided.
	BaseUrlPlaceholder = "${" + BaseUrlEnv + "}"
	// DefaultBaseUrl is the local Ollama endpoint used when the environment variable is unset.
	DefaultBaseUrl = "http://ollama:11434"
	// CloudBaseUrl is the base URL for the Ollama Cloud service.
	CloudBaseUrl = "https://ollama.com"
	// DefaultUri is the default service URI for self-hosted Ollama instances.
	DefaultUri = BaseUrlPlaceholder + "/api/generate"
	// DefaultModel names the default caption model bundled with our adapter defaults.
	DefaultModel = "gemma3:latest"
	// CloudModel names the default caption for the Ollama cloud service, see https://ollama.com/cloud.
	CloudModel = "qwen3-vl:235b-instruct-cloud"
	// CaptionPrompt instructs Ollama caption models to emit a single, active-voice sentence.
	CaptionPrompt = "Create a caption with exactly one sentence in the active voice that describes the main visual content. Begin with the main subject and clear action. Avoid text formatting, meta-language, and filler words."
	// LabelConfidenceDefault is used when the model omits the confidence field.
	LabelConfidenceDefault = 0.5
	// LabelSystem defines the system prompt shared by Ollama label models. It aims to ensure that single-word nouns are returned.
	LabelSystem = "You are a PhotoPrism vision model. Output concise JSON that matches the schema. Each label name MUST be a single-word noun in its canonical singular form. Avoid spaces, punctuation, emoji, or descriptive phrases."
	// LabelSystemSimple defines a simple system prompt for Ollama label models that does not strictly require names to be single-word nouns.
	LabelSystemSimple = "You are a PhotoPrism vision model. Output concise JSON that matches the schema."
	// LabelPromptDefault defines a simple user prompt for Ollama label models.
	LabelPromptDefault = "Analyze the image and return label objects with name, confidence (0-1), and topicality (0-1)."
	// LabelPromptStrict asks the model to return scored labels for the provided image. It aims to ensure that single-word nouns are returned.
	LabelPromptStrict = "Analyze the image and return label objects with name (single-word noun), confidence (0-1), and topicality (0-1). Respond with JSON exactly like {\"labels\":[{\"name\":\"sunset\",\"confidence\":0.72,\"topicality\":0.64}]} and adjust the values for this image."
	// LabelPromptNSFW asks the model to return scored labels for the provided image that includes a NSFW flag and score. It aims to ensure that single-word nouns are returned.
	LabelPromptNSFW = "Analyze the image and return label objects with name (single-word noun), confidence (0-1), topicality (0-1), nsfw (true when the label describes sensitive or adult content), and nsfw_confidence (0-1). Respond with JSON exactly like {\"labels\":[{\"name\":\"sunset\",\"confidence\":0.72,\"topicality\":0.64,\"nsfw\":false,\"nsfw_confidence\":0.02}]} and adjust the values for this image."
	// DefaultResolution is the default thumbnail size submitted to Ollama models.
	DefaultResolution = 720
)
