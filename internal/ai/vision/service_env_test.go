package vision

import (
	"net/url"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/vision/ollama"
	"github.com/photoprism/photoprism/internal/ai/vision/openai"
	"github.com/photoprism/photoprism/pkg/http/safe"
)

// resetRefusedEnvWarned clears the logged refused variables before and after a test.
func resetRefusedEnvWarned(t *testing.T) {
	t.Helper()
	refusedEnvWarned.Clear()
	t.Cleanup(refusedEnvWarned.Clear)
}

// TestExpandEnvSuffix checks that only variables with an allowed name suffix are expanded.
func TestExpandEnvSuffix(t *testing.T) {
	t.Setenv("VISION_TEST_MODEL", "qwen3-vl:8b")
	t.Setenv("VISION_TEST_SECRET", "s3cr3t-value")

	t.Run("Allowed", func(t *testing.T) {
		expanded, refused := expandEnvSuffix("${VISION_TEST_MODEL}", modelEnvSuffixes)
		assert.Equal(t, "qwen3-vl:8b", expanded)
		assert.Empty(t, refused)
	})
	t.Run("AllowedWithoutBraces", func(t *testing.T) {
		expanded, refused := expandEnvSuffix("$VISION_TEST_MODEL", modelEnvSuffixes)
		assert.Equal(t, "qwen3-vl:8b", expanded)
		assert.Empty(t, refused)
	})
	t.Run("Refused", func(t *testing.T) {
		expanded, refused := expandEnvSuffix("${VISION_TEST_SECRET}", modelEnvSuffixes)
		assert.Equal(t, "", expanded)
		assert.Equal(t, []string{"VISION_TEST_SECRET"}, refused)
	})
	t.Run("Mixed", func(t *testing.T) {
		expanded, refused := expandEnvSuffix("${VISION_TEST_MODEL}-${HOME}", modelEnvSuffixes)
		assert.Equal(t, "qwen3-vl:8b-", expanded)
		assert.Equal(t, []string{"HOME"}, refused)
	})
	t.Run("UnsetAllowed", func(t *testing.T) {
		expanded, refused := expandEnvSuffix("${VISION_TEST_UNSET_MODEL}", modelEnvSuffixes)
		assert.Equal(t, "", expanded)
		assert.Empty(t, refused)
	})
	t.Run("DefaultValueSyntax", func(t *testing.T) {
		expanded, refused := expandEnvSuffix("${VISION_TEST_MODEL:-gemma3}", modelEnvSuffixes)
		assert.Equal(t, "", expanded)
		assert.Equal(t, []string{"VISION_TEST_MODEL:-gemma3"}, refused)
	})
	t.Run("LowerCase", func(t *testing.T) {
		expanded, refused := expandEnvSuffix("${vision_test_model}", modelEnvSuffixes)
		assert.Equal(t, "", expanded)
		assert.Equal(t, []string{"vision_test_model"}, refused)
	})
	t.Run("ShellSpecial", func(t *testing.T) {
		expanded, refused := expandEnvSuffix("a$1b$*c", modelEnvSuffixes)
		assert.Equal(t, "abc", expanded)
		assert.Equal(t, []string{"1", "*"}, refused)
	})
	t.Run("NoVariables", func(t *testing.T) {
		expanded, refused := expandEnvSuffix("gemma3:latest", modelEnvSuffixes)
		assert.Equal(t, "gemma3:latest", expanded)
		assert.Empty(t, refused)
	})
}

// TestEnvNameAllowed checks the variable names that the logged service fields expand.
func TestEnvNameAllowed(t *testing.T) {
	t.Run("Model", func(t *testing.T) {
		assert.True(t, envNameAllowed("VISION_MODEL", modelEnvSuffixes))
		assert.True(t, envNameAllowed("OLLAMA_MODEL", modelEnvSuffixes))
		assert.True(t, envNameAllowed("PHOTOPRISM_VISION_CAPTION_MODEL", modelEnvSuffixes))
	})
	t.Run("Uri", func(t *testing.T) {
		assert.True(t, envNameAllowed(ollama.BaseUrlEnv, uriEnvSuffixes))
		assert.True(t, envNameAllowed("PHOTOPRISM_VISION_URI", uriEnvSuffixes))
		assert.True(t, envNameAllowed("OLLAMA_HOST", uriEnvSuffixes))
	})
	t.Run("Refused", func(t *testing.T) {
		assert.False(t, envNameAllowed("OLLAMA_MODELS", modelEnvSuffixes))
		assert.False(t, envNameAllowed(ollama.APIKeyEnv, modelEnvSuffixes))
		assert.False(t, envNameAllowed(openai.APIKeyEnv, uriEnvSuffixes))
		assert.False(t, envNameAllowed("HOME", uriEnvSuffixes))
		assert.False(t, envNameAllowed("_MODEL", modelEnvSuffixes))
		assert.False(t, envNameAllowed("", modelEnvSuffixes))
		assert.False(t, envNameAllowed("vision_model", modelEnvSuffixes))
		assert.False(t, envNameAllowed("OLLAMA_BASE_URL:-http://ollama:11434", uriEnvSuffixes))
		assert.False(t, envNameAllowed("FOO:-x_URL", uriEnvSuffixes))
		assert.False(t, envNameAllowed("FOO:-x_MODEL", modelEnvSuffixes))
	})
}

// TestEnvNameValid checks which references are valid variable names.
func TestEnvNameValid(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		assert.True(t, envNameValid("OLLAMA_BASE_URL"))
		assert.True(t, envNameValid("_x1"))
		assert.True(t, envNameValid("lower_case"))
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.False(t, envNameValid(""))
		assert.False(t, envNameValid("1"))
		assert.False(t, envNameValid("*"))
		assert.False(t, envNameValid("OPENAI_API_KEY:-sk-default"))
		assert.False(t, envNameValid("A B"))
		assert.False(t, envNameValid("ÄB"))
	})
}

// TestWarnRefusedEnv checks the warning for a refused variable and that it is logged once.
func TestWarnRefusedEnv(t *testing.T) {
	t.Run("OncePerFieldValueAndName", func(t *testing.T) {
		resetRefusedEnvWarned(t)
		t.Setenv(ollama.APIKeyEnv, "test-key")
		_, systemHook := captureLogs(t)

		warnRefusedEnv("Service.Model", "${OLLAMA_API_KEY}", []string{"OLLAMA_API_KEY"}, modelEnvSuffixes)
		warnRefusedEnv("Service.Model", "${OLLAMA_API_KEY}", []string{"OLLAMA_API_KEY"}, modelEnvSuffixes)
		warnRefusedEnv("Service.Uri", "${OLLAMA_API_KEY}", []string{"OLLAMA_API_KEY"}, uriEnvSuffixes)

		require.Len(t, systemHook.AllEntries(), 2)
		assert.Equal(t, logrus.WarnLevel, systemHook.AllEntries()[0].Level)
		assert.Equal(t, "vision: Service.Model does not expand OLLAMA_API_KEY, as only variables ending in _MODEL are expanded there", systemHook.AllEntries()[0].Message)
		assert.Equal(t, "vision: Service.Uri does not expand OLLAMA_API_KEY, as only variables ending in _URL, _URI, or _HOST are expanded there", systemHook.AllEntries()[1].Message)
	})
	t.Run("InvalidReference", func(t *testing.T) {
		resetRefusedEnvWarned(t)
		_, systemHook := captureLogs(t)

		value := "${OLLAMA_BASE_URL:-https://user:pa55@ollama.example.com}/api/generate"
		warnRefusedEnv("Service.Uri", value, []string{"OLLAMA_BASE_URL:-https://user:pa55@ollama.example.com"}, uriEnvSuffixes)

		require.Len(t, systemHook.AllEntries(), 1)
		assert.Equal(t, "vision: Service.Uri contains an unset or invalid variable reference, which is not expanded", systemHook.LastEntry().Message)
	})
	t.Run("LiteralDollar", func(t *testing.T) {
		resetRefusedEnvWarned(t)
		_, systemHook := captureLogs(t)

		// A "$" inside an inline password is read as a variable reference by os.Expand.
		//nolint:gosec // G101: test fixture, not a credential.
		s := &Service{Uri: "https://user:Pa$Xyz123word@vision.example.com/api"}
		uri, _ := s.Endpoint()

		assert.Equal(t, "", uri)
		require.Len(t, systemHook.AllEntries(), 1)
		assert.Equal(t, "vision: Service.Uri contains an unset or invalid variable reference, which is not expanded", systemHook.LastEntry().Message)
		assert.NotContains(t, systemHook.LastEntry().Message, "Xyz123word")
	})
	t.Run("None", func(t *testing.T) {
		resetRefusedEnvWarned(t)
		logHook, _ := captureLogs(t)

		warnRefusedEnv("Service.Model", "gemma3", nil, modelEnvSuffixes)

		assert.Empty(t, logHook.AllEntries())
	})
}

// TestService_RefusedEnv checks that refused variables leave the service fields unset and are never logged.
func TestService_RefusedEnv(t *testing.T) {
	const secret = "sk-test-0123456789abcdef"

	t.Setenv(ollama.APIKeyEnv, secret)
	t.Setenv(openai.APIKeyEnv, secret)
	t.Setenv("PHOTOPRISM_VISION_KEY", secret)
	t.Setenv("VISION_TEST_MODEL", "qwen3-vl:8b")

	// noSecret checks that no log entry contains the value of the secret variables.
	noSecret := func(t *testing.T, logHook, systemHook interface{ AllEntries() []*logrus.Entry }) {
		t.Helper()

		for _, entry := range append(logHook.AllEntries(), systemHook.AllEntries()...) {
			assert.NotContains(t, entry.Message, secret)
		}
	}

	t.Run("ModelAllowed", func(t *testing.T) {
		resetRefusedEnvWarned(t)
		logHook, _ := captureLogs(t)

		assert.Equal(t, "qwen3-vl:8b", (&Service{Model: "${VISION_TEST_MODEL}"}).GetModel())
		assert.Empty(t, logHook.AllEntries())
	})
	t.Run("ModelRefused", func(t *testing.T) {
		resetRefusedEnvWarned(t)
		logHook, systemHook := captureLogs(t)

		assert.Equal(t, "", (&Service{Model: "${OLLAMA_API_KEY}"}).GetModel())
		assert.Equal(t, "", (&Service{Model: "${HOME}"}).GetModel())

		require.Len(t, systemHook.AllEntries(), 2)
		assert.Contains(t, systemHook.AllEntries()[0].Message, "does not expand OLLAMA_API_KEY")
		assert.Contains(t, systemHook.AllEntries()[1].Message, "does not expand HOME")
		noSecret(t, systemHook, logHook)
	})
	t.Run("ModelFallback", func(t *testing.T) {
		resetRefusedEnvWarned(t)
		logHook, systemHook := captureLogs(t)

		m := &Model{Type: ModelTypeCaption, Engine: ollama.EngineName, Model: "gemma3:latest", Service: Service{Model: "${OLLAMA_API_KEY}"}}
		model, name, version := m.GetModel()

		assert.Equal(t, "gemma3:latest", model)
		assert.Equal(t, "gemma3", name)
		assert.Equal(t, "latest", version)
		noSecret(t, systemHook, logHook)
	})
	t.Run("UriDocumented", func(t *testing.T) {
		resetRefusedEnvWarned(t)
		t.Setenv(ollama.BaseUrlEnv, "http://ollama:11434")
		logHook, _ := captureLogs(t)

		s := &Service{Uri: ollama.DefaultUri}
		uri, method := s.Endpoint()

		assert.Equal(t, "http://ollama:11434/api/generate", uri)
		assert.Equal(t, ServiceMethod, method)
		assert.False(t, s.UriUnresolved())
		assert.Empty(t, logHook.AllEntries())
	})
	t.Run("UriRefused", func(t *testing.T) {
		resetRefusedEnvWarned(t)
		logHook, systemHook := captureLogs(t)

		s := &Service{Uri: "${OPENAI_API_KEY}/v1/responses"}
		uri, method := s.Endpoint()

		assert.Equal(t, "", uri)
		assert.Equal(t, "", method)
		assert.True(t, s.UriUnresolved())
		require.NotEmpty(t, systemHook.AllEntries())
		assert.Contains(t, systemHook.AllEntries()[0].Message, "Service.Uri does not expand OPENAI_API_KEY")
		noSecret(t, systemHook, logHook)
	})
	t.Run("UriDefaultValueSyntax", func(t *testing.T) {
		resetRefusedEnvWarned(t)
		logHook, systemHook := captureLogs(t)

		s := &Service{Uri: "${OLLAMA_BASE_URL:-https://user:pa55@ollama.example.com}/api/generate"}
		uri, _ := s.Endpoint()

		assert.Equal(t, "", uri)
		require.NotEmpty(t, systemHook.AllEntries())

		for _, entry := range append(systemHook.AllEntries(), logHook.AllEntries()...) {
			assert.NotContains(t, entry.Message, "pa55")
		}
	})
	t.Run("SecretFields", func(t *testing.T) {
		resetRefusedEnvWarned(t)
		logHook, _ := captureLogs(t)

		//nolint:gosec // G101: variable references, not credentials.
		s := &Service{Key: "${PHOTOPRISM_VISION_KEY}", Username: "${VISION_TEST_MODEL}", Password: "${OLLAMA_API_KEY}"}
		username, password := s.BasicAuth()

		assert.Equal(t, secret, s.EndpointKey())
		assert.Equal(t, "qwen3-vl:8b", username)
		assert.Equal(t, secret, password)
		assert.Empty(t, logHook.AllEntries())
	})
	t.Run("UnresolvedModelNamed", func(t *testing.T) {
		resetRefusedEnvWarned(t)
		resetUnresolvedUriWarnings(t)
		logHook, systemHook := captureLogs(t)

		m := &Model{Type: ModelTypeLabels, Model: "${VISION_TEST_MODEL}", Service: Service{Uri: "${OPENAI_API_KEY}/v1/responses", Model: "${VISION_TEST_MODEL}"}}

		err := m.unresolvedUriErr()

		assert.EqualError(t, err, "service uri of labels model qwen3-vl does not resolve")
		assert.False(t, strings.Contains(err.Error(), secret))
		require.Len(t, systemHook.AllEntries(), 2)
		assert.Equal(t, "vision: Service.Uri does not expand OPENAI_API_KEY, as only variables ending in _URL, _URI, or _HOST are expanded there", systemHook.AllEntries()[0].Message)
		assert.Equal(t, "vision: service uri of labels model qwen3-vl does not resolve, so no service is used", systemHook.AllEntries()[1].Message)
		assert.Empty(t, logHook.AllEntries())
	})
}

// TestService_ExpandedUriNotLogged checks that a failed request names neither the credentials nor the path
// of an expanded service URI in the main log.
func TestService_ExpandedUriNotLogged(t *testing.T) {
	t.Setenv("VISION_TEST_SERVICE_URL", "http://user:FAKE-PASS@127.0.0.1:1/secret-path-token")
	resetRefusedEnvWarned(t)
	logHook, _ := captureLogs(t)

	s := &Service{Uri: "${VISION_TEST_SERVICE_URL}/api/generate"}
	uri, method := s.Endpoint()
	require.Equal(t, "http://user:FAKE-PASS@127.0.0.1:1/secret-path-token/api/generate", uri)

	_, err := PerformApiRequest(&ApiRequest{Model: "gemma3:latest", ResponseFormat: ApiFormatOllama}, uri, method, "")
	require.EqualError(t, err, "ollama service request failed (connection error)")

	assert.NotContains(t, err.Error(), "FAKE-PASS")
	assert.NotContains(t, err.Error(), "secret-path-token")

	for _, entry := range logHook.AllEntries() {
		if entry.Level <= logrus.InfoLevel {
			assert.NotContains(t, entry.Message, "FAKE-PASS")
			assert.NotContains(t, entry.Message, "secret-path-token")
		}
	}
}

// TestForgetRefusedEnv checks that only the refusals of the given field and value are removed.
func TestForgetRefusedEnv(t *testing.T) {
	resetRefusedEnvWarned(t)
	refusedEnvWarned.Store("Service.Uri\x00${A_TOKEN}/x\x00A_TOKEN", struct{}{})
	refusedEnvWarned.Store("Service.Uri\x00${A_TOKEN}/xy\x00A_TOKEN", struct{}{})
	refusedEnvWarned.Store("Service.Model\x00${A_TOKEN}/x\x00A_TOKEN", struct{}{})

	forgetRefusedEnv("Service.Uri", "${A_TOKEN}/x")

	_, removed := refusedEnvWarned.Load("Service.Uri\x00${A_TOKEN}/x\x00A_TOKEN")
	assert.False(t, removed)
	_, kept := refusedEnvWarned.Load("Service.Uri\x00${A_TOKEN}/xy\x00A_TOKEN")
	assert.True(t, kept)
	_, kept = refusedEnvWarned.Load("Service.Model\x00${A_TOKEN}/x\x00A_TOKEN")
	assert.True(t, kept)
}

// TestExpandUriEnv checks that the query of an expanded variable is moved to the end of the URI.
func TestExpandUriEnv(t *testing.T) {
	t.Setenv("VISION_TEST_BASE_URL", "https://gw.example.com/v1?token=abc")
	t.Setenv("VISION_TEST_PLAIN_URL", "https://plain.example.com/v1/")
	t.Setenv("VISION_TEST_FRAGMENT_URL", "https://gw.example.com/v1#top")
	t.Setenv("VISION_TEST_OTHER_URL", "https://other.example.com?b=2")
	t.Setenv("VISION_TEST_MULTI_URL", "https://x.example.com/p?a=1&b=2")
	t.Setenv("VISION_TEST_EMPTY_QUERY_URL", "https://gw.example.com/v1?")
	t.Setenv("VISION_TEST_SECRET", "s3cr3t-value")

	cases := []struct {
		name string
		in   string
		want string
	}{
		{"NoVariable", "https://api.example.com/v1/responses?x=1", "https://api.example.com/v1/responses?x=1"},
		{"NoQuery", "${VISION_TEST_PLAIN_URL}/responses", "https://plain.example.com/v1//responses"},
		{"Query", "${VISION_TEST_BASE_URL}/responses", "https://gw.example.com/v1/responses?token=abc"},
		{"TemplateQuery", "${VISION_TEST_BASE_URL}/responses?stream=false", "https://gw.example.com/v1/responses?stream=false&token=abc"},
		{"TemplateEmptyQuery", "${VISION_TEST_BASE_URL}/responses?", "https://gw.example.com/v1/responses?token=abc"},
		{"TemplateFragment", "${VISION_TEST_BASE_URL}/responses#part", "https://gw.example.com/v1/responses?token=abc#part"},
		{"ValueFragment", "${VISION_TEST_FRAGMENT_URL}/responses", "https://gw.example.com/v1/responses"},
		{"TwoQueries", "${VISION_TEST_BASE_URL}/x/${VISION_TEST_OTHER_URL}", "https://gw.example.com/v1/x/https://other.example.com?token=abc&b=2"},
		{"TemplateAmpersand", "${VISION_TEST_BASE_URL}/responses?stream=false&", "https://gw.example.com/v1/responses?stream=false&token=abc"},
		{"ContinuedQuery", "${VISION_TEST_BASE_URL}&stream=false", "https://gw.example.com/v1?token=abc&stream=false"},
		{"ValueInTemplateQuery", "https://h.example.com/api?cb=${VISION_TEST_MULTI_URL}", "https://h.example.com/api?cb=https://x.example.com/p?a=1&b=2"},
		{"MovedAndInTemplateQuery", "${VISION_TEST_BASE_URL}/api?cb=${VISION_TEST_MULTI_URL}", "https://gw.example.com/v1/api?cb=https://x.example.com/p?a=1&b=2&token=abc"},
		{"EmptyQuery", "${VISION_TEST_EMPTY_QUERY_URL}/responses", "https://gw.example.com/v1/responses"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expanded, refused := expandUriEnv(tc.in)
			assert.Equal(t, tc.want, expanded)
			assert.Empty(t, refused)
		})
	}

	t.Run("NullByte", func(t *testing.T) {
		expanded, refused := expandUriEnv("${VISION_TEST_BASE_URL}/x\x00")
		assert.Equal(t, "https://gw.example.com/v1?token=abc/x\x00", expanded)
		assert.Empty(t, refused)
	})
	t.Run("Refused", func(t *testing.T) {
		expanded, refused := expandUriEnv("${VISION_TEST_BASE_URL}/${VISION_TEST_SECRET}")
		assert.Equal(t, "https://gw.example.com/v1/?token=abc", expanded)
		assert.Equal(t, []string{"VISION_TEST_SECRET"}, refused)
	})
}

// TestExpandEnvValues checks that the expanded values pass through the function.
func TestExpandEnvValues(t *testing.T) {
	t.Setenv("VISION_TEST_MODEL", "qwen3-vl:8b")
	t.Setenv("VISION_TEST_SECRET", "s3cr3t-value")

	expanded, refused := expandEnvValues("${VISION_TEST_MODEL}-${VISION_TEST_SECRET}", modelEnvSuffixes, strings.ToUpper)
	assert.Equal(t, "QWEN3-VL:8B-", expanded)
	assert.Equal(t, []string{"VISION_TEST_SECRET"}, refused)

	expanded, refused = expandEnvValues("plain", modelEnvSuffixes, strings.ToUpper)
	assert.Equal(t, "plain", expanded)
	assert.Empty(t, refused)
}

// TestUriPathEnd checks which positions in a URI end its path part.
func TestUriPathEnd(t *testing.T) {
	assert.False(t, uriPathEnd("", ""))
	assert.False(t, uriPathEnd("https:", "//evil.example.com/api"))
	assert.True(t, uriPathEnd("https://h", "/api"))
	assert.True(t, uriPathEnd("https://h/v1", "?x=1"))
	assert.True(t, uriPathEnd("https://h/v1", "#top"))
	assert.False(t, uriPathEnd("https://h/api?cb=", ""))
	assert.False(t, uriPathEnd("https://h/api#", ""))
	assert.False(t, uriPathEnd("https://", ":8443/api"))
	assert.False(t, uriPathEnd("https://", ".example.com/api"))
	assert.False(t, uriPathEnd("https://h/v1", "&stream=false"))
}

// TestExpandUriEnv_Authority checks that moving a query never changes the scheme, user, or host of the URI
// compared with expanding the values as written.
func TestExpandUriEnv_Authority(t *testing.T) {
	values := []string{
		"good.example.com?token=X",
		"https://good.example.com/v1?token=X",
		"https://good.example.com/v1/?token=X#frag",
		"?token=X",
		"https://u:p@good.example.com?token=X",
		"https://good.example.com?a=1@evil.example.com",
		"https://[fe80::1%25eth0]:8080?token=X",
		"https://good.example.com/v1",
		"https://good.example.com#x",
		"https://good.example.com#x?token=X",
		"good.example.com#x",
		"https://good.example.com#%zz",
		"https:?token=X",
	}
	templates := []string{
		"${VISION_TEST_V_URL}/responses",
		"${VISION_TEST_V_URL}",
		"https://${VISION_TEST_V_URL}:8443/api",
		"https://${VISION_TEST_V_URL}.example.com/api",
		"https://${VISION_TEST_V_URL}@good.example.com/api",
		"https://${VISION_TEST_V_URL}${VISION_TEST_W_URL}/api",
		"${VISION_TEST_V_URL}${VISION_TEST_W_URL}",
		"https://h.example.com/api#${VISION_TEST_V_URL}",
		"https://h.example.com/api?cb=${VISION_TEST_V_URL}",
		"${VISION_TEST_V_URL}&stream=false",
		"${VISION_TEST_V_URL}/x?y=1#z",
		"${VISION_TEST_V_URL}@evil.example.com/api",
		"${VISION_TEST_V_URL}.evil.example.com/api",
		"${VISION_TEST_V_URL}//evil.example.com/api",
		"https:${VISION_TEST_V_URL}//evil.example.com/api",
	}

	t.Setenv("VISION_TEST_W_URL", "https://other.example.com/w?w=1")

	for _, value := range values {
		for _, tmpl := range templates {
			t.Setenv("VISION_TEST_V_URL", value)

			inline, _ := expandEnvSuffix(tmpl, uriEnvSuffixes)
			moved, _ := expandUriEnv(tmpl)
			inlineUrl, inlineErr := url.Parse(inline)
			movedUrl, movedErr := url.Parse(moved)

			// A URI that cannot be sent as written may only be sent to the host of the value itself, e.g.
			// once an invalid fragment was dropped.
			if inlineErr != nil || inlineUrl.Host == "" {
				_, safeErr := safe.URL(moved)
				assert.True(t, movedErr != nil || safeErr != nil || movedUrl.Host == "" ||
					strings.HasPrefix(value, movedUrl.Scheme+"://"+movedUrl.Host), "%s with %s", tmpl, value)
				continue
			}

			require.NoError(t, movedErr, "%s with %s", tmpl, value)
			assert.Equal(t, inlineUrl.Scheme, movedUrl.Scheme, "%s with %s", tmpl, value)
			assert.Equal(t, inlineUrl.User.String(), movedUrl.User.String(), "%s with %s", tmpl, value)
			assert.Equal(t, inlineUrl.Host, movedUrl.Host, "%s with %s", tmpl, value)
		}
	}
}
