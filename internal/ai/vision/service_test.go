package vision

import (
	"sync"
	"testing"
)

func TestServiceEndpoint(t *testing.T) {
	//nolint:gosec // G101: Credential-style URLs are intentional test fixtures.
	tests := []struct {
		name       string
		svc        Service
		wantURI    string
		wantMethod string
	}{
		{
			name:       "Disabled",
			svc:        Service{Disabled: true, Uri: "https://vision.example.com"},
			wantURI:    "",
			wantMethod: "",
		},
		{
			name:       "WithBasicAuth",
			svc:        Service{Uri: "https://vision.example.com/api", Username: "user", Password: "secret"},
			wantURI:    "https://user:secret@vision.example.com/api",
			wantMethod: ServiceMethod,
		},
		{
			name:       "UsernameOnly",
			svc:        Service{Uri: "https://vision.example.com/", Username: "scoped"},
			wantURI:    "https://scoped@vision.example.com/",
			wantMethod: ServiceMethod,
		},
		{
			name:       "PreserveExistingUser",
			svc:        Service{Uri: "https://keep:me@vision.example.com", Username: "ignored", Password: "ignored"},
			wantURI:    "https://keep:me@vision.example.com",
			wantMethod: ServiceMethod,
		},
		{
			name:       "Unresolved",
			svc:        Service{Uri: "${VISION_TEST_MISSING_URI}"},
			wantURI:    "",
			wantMethod: "",
		},
		{
			name:       "UnresolvedWithBasicAuth",
			svc:        Service{Uri: "${VISION_TEST_MISSING_URI}", Username: "user", Password: "secret"},
			wantURI:    "",
			wantMethod: "",
		},
		{
			name:       "NestedPlaceholder",
			svc:        Service{Uri: "${VISION_TEST_NESTED_URI}"},
			wantURI:    "",
			wantMethod: "",
		},
		{
			name:       "ExpandsBaseUrlEnv",
			svc:        Service{Uri: "${OLLAMA_BASE_URL}/api/generate"},
			wantURI:    "http://custom:11434/api/generate",
			wantMethod: ServiceMethod,
		},
		{
			name:       "FallbacksWhenEnvMissing",
			svc:        Service{Uri: "${OLLAMA_BASE_URL}/api/generate"},
			wantURI:    "http://ollama:11434/api/generate",
			wantMethod: ServiceMethod,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.name {
			case "ExpandsBaseUrlEnv":
				t.Setenv("OLLAMA_BASE_URL", "http://custom:11434")
			case "FallbacksWhenEnvMissing":
				t.Setenv("OLLAMA_BASE_URL", "http://ollama:11434")
			case "NestedPlaceholder":
				t.Setenv("VISION_TEST_NESTED_URI", "${VISION_TEST_NESTED}")
			}

			uri, method := tt.svc.Endpoint()
			if uri != tt.wantURI {
				t.Fatalf("uri: got %q want %q", uri, tt.wantURI)
			}
			if method != tt.wantMethod {
				t.Fatalf("method: got %q want %q", method, tt.wantMethod)
			}
		})
	}
}

func TestServiceCredentialsAndHeaders(t *testing.T) {
	t.Setenv("VISION_USER", "alice")
	t.Setenv("VISION_PASS", "hunter2")
	t.Setenv("VISION_MODEL", "QuantTrio/Qwen3-VL-30B-A3B-Instruct-AWQ")
	t.Setenv("VISION_ORG", "org-123")
	t.Setenv("VISION_PROJECT", "proj-abc")
	t.Setenv("VISION_THINK", "false")
	t.Setenv("VISION_TIER", "flex")

	svc := Service{
		Username: "${VISION_USER}",
		Password: "${VISION_PASS}",
		Model:    "${VISION_MODEL}",
		Org:      "${VISION_ORG}",
		Project:  "${VISION_PROJECT}",
		Think:    "${VISION_THINK}",
		Tier:     "${VISION_TIER}",
	}

	user, pass := svc.BasicAuth()
	if user != "alice" || pass != "hunter2" {
		t.Fatalf("basic auth: got %q/%q", user, pass)
	}

	if got := svc.GetModel(); got != "QuantTrio/Qwen3-VL-30B-A3B-Instruct-AWQ" {
		t.Fatalf("model override: got %q", got)
	}

	if got := svc.EndpointOrg(); got != "org-123" {
		t.Fatalf("org: got %q", got)
	}

	if got := svc.EndpointProject(); got != "proj-abc" {
		t.Fatalf("project: got %q", got)
	}

	if got := svc.EndpointThink(); got != "false" {
		t.Fatalf("think: got %q", got)
	}

	if got := svc.EndpointTier(); got != "flex" {
		t.Fatalf("tier: got %q", got)
	}
}

// TestServiceEndpoint_BaseUrlQuery checks that the query of an expanded base URL is moved to the end of the
// service URI, for engine and custom URIs.
func TestServiceEndpoint_BaseUrlQuery(t *testing.T) {
	cases := []struct {
		name string
		env  string
		val  string
		uri  string
		want string
	}{
		{"OpenAIApiVersion", "OPENAI_BASE_URL", "https://example.openai.azure.com/openai/v1?api-version=preview", "${OPENAI_BASE_URL}/responses", "https://example.openai.azure.com/openai/v1/responses?api-version=preview"},
		{"OpenAIToken", "OPENAI_BASE_URL", "https://gateway.example.com/v1?token=abc", "${OPENAI_BASE_URL}/responses", "https://gateway.example.com/v1/responses?token=abc"},
		{"OpenAINoQuery", "OPENAI_BASE_URL", "https://api.openai.com/v1", "${OPENAI_BASE_URL}/responses", "https://api.openai.com/v1/responses"},
		{"OllamaToken", "OLLAMA_BASE_URL", "http://ollama:11434?key=abc", "${OLLAMA_BASE_URL}/api/generate", "http://ollama:11434/api/generate?key=abc"},
		{"CustomTemplateQuery", "VISION_TEST_GATEWAY_URL", "https://gw.example.com/v1?token=abc#top", "${VISION_TEST_GATEWAY_URL}/chat?stream=false", "https://gw.example.com/v1/chat?stream=false&token=abc"},
		{"CustomHost", "VISION_TEST_GATEWAY_HOST", "gw.example.com", "https://${VISION_TEST_GATEWAY_HOST}/api?x=1", "https://gw.example.com/api?x=1"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(tc.env, tc.val)

			uri, method := (&Service{Uri: tc.uri}).Endpoint()

			if uri != tc.want {
				t.Fatalf("uri: got %q want %q", uri, tc.want)
			} else if method != ServiceMethod {
				t.Fatalf("method: got %q want %q", method, ServiceMethod)
			}
		})
	}

	t.Run("WithBasicAuth", func(t *testing.T) {
		t.Setenv("VISION_TEST_GATEWAY_URL", "https://gw.example.com/v1?token=abc")

		uri, _ := (&Service{Uri: "${VISION_TEST_GATEWAY_URL}/chat", Username: "user", Password: "secret"}).Endpoint()

		//nolint:gosec // G101: Credential-style URLs are intentional test fixtures.
		if want := "https://user:secret@gw.example.com/v1/chat?token=abc"; uri != want {
			t.Fatalf("uri: got %q want %q", uri, want)
		}
	})
}

// TestServiceEndpoint_NormalizedBaseUrl checks the URI of a base URL that is normalized by ensureEnv.
func TestServiceEndpoint_NormalizedBaseUrl(t *testing.T) {
	t.Setenv("OPENAI_BASE_URL", "https://gw.example.com/v1/?token=abc#top")
	ensureEnvOnce = sync.Once{}
	t.Cleanup(func() { ensureEnvOnce = sync.Once{} })

	uri, _ := (&Service{Uri: "${OPENAI_BASE_URL}/responses"}).Endpoint()

	if want := "https://gw.example.com/v1/responses?token=abc"; uri != want {
		t.Fatalf("uri: got %q want %q", uri, want)
	}
}
