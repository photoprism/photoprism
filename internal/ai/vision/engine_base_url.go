package vision

import (
	"os"
	"strings"

	"github.com/photoprism/photoprism/internal/ai/vision/ollama"
	"github.com/photoprism/photoprism/internal/ai/vision/openai"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
)

// engineBaseUrl describes an engine whose default service URI is built from a base URL variable.
type engineBaseUrl struct {
	Engine     string // Engine name, e.g. "openai".
	Env        string // Variable that holds the base URL.
	DefaultUrl string // Base URL used when the variable is not set.
	DefaultUri string // Service URI of models without their own.
}

// engineBaseUrls lists the engines whose models send their requests to a base URL from the environment
// unless they have a service URI of their own.
var engineBaseUrls = []engineBaseUrl{
	{Engine: openai.EngineName, Env: openai.BaseUrlEnv, DefaultUrl: openai.DefaultBaseUrl, DefaultUri: openai.DefaultUri},
	{Engine: ollama.EngineName, Env: ollama.BaseUrlEnv, DefaultUrl: ollama.DefaultBaseUrl, DefaultUri: ollama.DefaultUri},
}

// Log writes the base URL to the system log if it is not the default, as it receives the requests and
// keys of models without their own service URI.
func (e engineBaseUrl) Log() {
	baseUrl := os.Getenv(e.Env)

	if baseUrl == "" || baseUrl == e.DefaultUrl {
		return
	}

	if redacted := clean.UriRedacted(baseUrl); redacted != "" {
		event.SystemInfo([]string{"vision", "%s engine uses base url %s"}, e.Engine, redacted)
	} else {
		event.SystemWarn([]string{"vision", "%s engine uses an invalid base url"}, e.Engine)
	}
}

// UsedBy reports whether the model sends its requests to the base URL.
func (e engineBaseUrl) UsedBy(m *Model) bool {
	return m != nil && !m.Disabled && !m.Service.Disabled && m.Engine == e.Engine &&
		strings.TrimSpace(m.Service.Uri) == e.DefaultUri
}
