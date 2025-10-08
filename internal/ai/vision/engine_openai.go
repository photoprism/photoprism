package vision

import (
	"github.com/photoprism/photoprism/internal/ai/vision/openai"
	"github.com/photoprism/photoprism/pkg/service/http/scheme"
)

// init registers the OpenAI engine alias so models can set Engine: "openai"
// and inherit sensible defaults (request/response formats, file scheme, and
// preferred thumbnail resolution).
func init() {
	RegisterEngineAlias(openai.EngineName, EngineInfo{
		RequestFormat:     ApiFormatOpenAI,
		ResponseFormat:    ApiFormatOpenAI,
		FileScheme:        string(scheme.Base64),
		DefaultResolution: openai.DefaultResolution,
	})
}
