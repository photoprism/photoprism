package vision

import (
	"github.com/photoprism/photoprism/internal/ai/vision/ollama"
	"github.com/photoprism/photoprism/internal/ai/vision/openai"
)

// ApiFormat defines the payload format accepted by the Vision API.
type ApiFormat = string

const (
	// ApiFormatUrl treats inputs as HTTP(S) URLs.
	ApiFormatUrl ApiFormat = "url"
	// ApiFormatImages sends images in the native Vision format.
	ApiFormatImages ApiFormat = "images"
	// ApiFormatVision represents a Vision-internal payload.
	ApiFormatVision ApiFormat = "vision"
	// ApiFormatOllama proxies requests to Ollama models.
	ApiFormatOllama ApiFormat = ollama.ApiFormat
	// ApiFormatOpenAI proxies requests to OpenAI vision models.
	ApiFormatOpenAI ApiFormat = openai.ApiFormat
)
