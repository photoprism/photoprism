package vision

import (
	"github.com/photoprism/photoprism/internal/ai/vision/ollama"
	"github.com/photoprism/photoprism/internal/ai/vision/openai"
)

type ApiFormat = string

const (
	ApiFormatUrl    ApiFormat = "url"
	ApiFormatImages ApiFormat = "images"
	ApiFormatVision ApiFormat = "vision"
	ApiFormatOllama ApiFormat = ollama.ApiFormat
	ApiFormatOpenAI ApiFormat = openai.ApiFormat
)
