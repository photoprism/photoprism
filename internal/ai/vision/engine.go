package vision

import (
	"context"
	"strings"
	"sync"

	"github.com/photoprism/photoprism/internal/ai/vision/openai"
	"github.com/photoprism/photoprism/pkg/service/http/scheme"
)

// ModelEngine represents the canonical identifier for a computer vision service engine.
type ModelEngine = string

const (
	// EngineVision represents the default PhotoPrism vision service endpoints.
	EngineVision ModelEngine = "vision"
	// EngineTensorFlow represents on-device TensorFlow models.
	EngineTensorFlow ModelEngine = "tensorflow"
	// EngineLocal is used when no explicit engine can be determined.
	EngineLocal ModelEngine = "local"
)

// RequestBuilder builds an API request for an engine based on the model configuration and input files.
type RequestBuilder interface {
	Build(ctx context.Context, model *Model, files Files) (*ApiRequest, error)
}

// ResponseParser parses a raw engine response into the generic ApiResponse structure.
type ResponseParser interface {
	Parse(ctx context.Context, req *ApiRequest, raw []byte, status int) (*ApiResponse, error)
}

// EngineDefaults supplies engine-specific prompt and schema defaults when they are not configured explicitly.
type EngineDefaults interface {
	SystemPrompt(model *Model) string
	UserPrompt(model *Model) string
	SchemaTemplate(model *Model) string
	Options(model *Model) *ApiRequestOptions
}

// Engine groups the callbacks required to integrate a third-party vision service.
type Engine struct {
	Builder  RequestBuilder
	Parser   ResponseParser
	Defaults EngineDefaults
}

var (
	engineRegistry   = make(map[ApiFormat]Engine)
	engineAliasIndex = make(map[string]EngineInfo)
	engineMu         sync.RWMutex
)

// init wires up the built-in aliases so configuration files can reference the
// human-friendly engine names without duplicating adapter metadata.
func init() {
	RegisterEngineAlias(EngineVision, EngineInfo{
		RequestFormat:     ApiFormatVision,
		ResponseFormat:    ApiFormatVision,
		FileScheme:        string(scheme.Data),
		DefaultResolution: DefaultResolution,
	})

	RegisterEngineAlias(openai.EngineName, EngineInfo{
		RequestFormat:     ApiFormatOpenAI,
		ResponseFormat:    ApiFormatOpenAI,
		FileScheme:        string(scheme.Data),
		DefaultResolution: openai.DefaultResolution,
	})
}

// RegisterEngine adds/overrides an engine implementation for a specific API format.
func RegisterEngine(format ApiFormat, engine Engine) {
	engineMu.Lock()
	defer engineMu.Unlock()
	engineRegistry[format] = engine
}

// EngineInfo describes metadata that can be associated with an engine alias.
type EngineInfo struct {
	RequestFormat     ApiFormat
	ResponseFormat    ApiFormat
	FileScheme        string
	DefaultResolution int
}

// RegisterEngineAlias maps a logical engine name (e.g., "ollama") to a
// request/response format pair.
func RegisterEngineAlias(name string, info EngineInfo) {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" || info.RequestFormat == "" {
		return
	}

	if info.ResponseFormat == "" {
		info.ResponseFormat = info.RequestFormat
	}

	engineMu.Lock()
	engineAliasIndex[name] = info
	engineMu.Unlock()
}

// EngineInfoFor returns the metadata associated with a logical engine name.
func EngineInfoFor(name string) (EngineInfo, bool) {
	name = strings.TrimSpace(strings.ToLower(name))
	engineMu.RLock()
	info, ok := engineAliasIndex[name]
	engineMu.RUnlock()
	return info, ok
}

// EngineFor returns the registered engine implementation for the given API
// format, if any.
func EngineFor(format ApiFormat) (Engine, bool) {
	engineMu.RLock()
	defer engineMu.RUnlock()
	engine, ok := engineRegistry[format]
	return engine, ok
}
