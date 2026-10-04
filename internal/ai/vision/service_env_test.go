package vision

import (
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/vision/ollama"
	"github.com/photoprism/photoprism/internal/ai/vision/openai"
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
		logHook, _ := captureLogs(t)

		warnRefusedEnv("Service.Model", "${OLLAMA_API_KEY}", []string{"OLLAMA_API_KEY"}, modelEnvSuffixes)
		warnRefusedEnv("Service.Model", "${OLLAMA_API_KEY}", []string{"OLLAMA_API_KEY"}, modelEnvSuffixes)
		warnRefusedEnv("Service.Uri", "${OLLAMA_API_KEY}", []string{"OLLAMA_API_KEY"}, uriEnvSuffixes)

		require.Len(t, logHook.AllEntries(), 2)
		assert.Equal(t, logrus.WarnLevel, logHook.AllEntries()[0].Level)
		assert.Equal(t, "vision: Service.Model does not expand OLLAMA_API_KEY, as only variables ending in _MODEL are expanded there", logHook.AllEntries()[0].Message)
		assert.Equal(t, "vision: Service.Uri does not expand OLLAMA_API_KEY, as only variables ending in _URL, _URI, or _HOST are expanded there", logHook.AllEntries()[1].Message)
	})
	t.Run("InvalidReference", func(t *testing.T) {
		resetRefusedEnvWarned(t)
		logHook, _ := captureLogs(t)

		value := "${OLLAMA_BASE_URL:-https://user:pa55@ollama.example.com}/api/generate"
		warnRefusedEnv("Service.Uri", value, []string{"OLLAMA_BASE_URL:-https://user:pa55@ollama.example.com"}, uriEnvSuffixes)

		require.Len(t, logHook.AllEntries(), 1)
		assert.Equal(t, "vision: Service.Uri contains an unset or invalid variable reference, which is not expanded", logHook.LastEntry().Message)
	})
	t.Run("LiteralDollar", func(t *testing.T) {
		resetRefusedEnvWarned(t)
		logHook, _ := captureLogs(t)

		// A "$" inside an inline password is read as a variable reference by os.Expand.
		//nolint:gosec // G101: test fixture, not a credential.
		s := &Service{Uri: "https://user:Pa$Xyz123word@vision.example.com/api"}
		uri, _ := s.Endpoint()

		assert.Equal(t, "", uri)
		require.Len(t, logHook.AllEntries(), 1)
		assert.Equal(t, "vision: Service.Uri contains an unset or invalid variable reference, which is not expanded", logHook.LastEntry().Message)
		assert.NotContains(t, logHook.LastEntry().Message, "Xyz123word")
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

		require.Len(t, logHook.AllEntries(), 2)
		assert.Contains(t, logHook.AllEntries()[0].Message, "does not expand OLLAMA_API_KEY")
		assert.Contains(t, logHook.AllEntries()[1].Message, "does not expand HOME")
		noSecret(t, logHook, systemHook)
	})
	t.Run("ModelFallback", func(t *testing.T) {
		resetRefusedEnvWarned(t)
		logHook, systemHook := captureLogs(t)

		m := &Model{Type: ModelTypeCaption, Engine: ollama.EngineName, Model: "gemma3:latest", Service: Service{Model: "${OLLAMA_API_KEY}"}}
		model, name, version := m.GetModel()

		assert.Equal(t, "gemma3:latest", model)
		assert.Equal(t, "gemma3", name)
		assert.Equal(t, "latest", version)
		noSecret(t, logHook, systemHook)
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
		require.NotEmpty(t, logHook.AllEntries())
		assert.Contains(t, logHook.AllEntries()[0].Message, "Service.Uri does not expand OPENAI_API_KEY")
		noSecret(t, logHook, systemHook)
	})
	t.Run("UriDefaultValueSyntax", func(t *testing.T) {
		resetRefusedEnvWarned(t)
		logHook, systemHook := captureLogs(t)

		s := &Service{Uri: "${OLLAMA_BASE_URL:-https://user:pa55@ollama.example.com}/api/generate"}
		uri, _ := s.Endpoint()

		assert.Equal(t, "", uri)
		require.NotEmpty(t, logHook.AllEntries())

		for _, entry := range append(logHook.AllEntries(), systemHook.AllEntries()...) {
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
		require.Len(t, logHook.AllEntries(), 2)
		assert.Equal(t, "vision: Service.Uri does not expand OPENAI_API_KEY, as only variables ending in _URL, _URI, or _HOST are expanded there", logHook.AllEntries()[0].Message)
		assert.Equal(t, "vision: service uri of labels model qwen3-vl does not resolve, so no service is used", logHook.AllEntries()[1].Message)
		assert.Empty(t, systemHook.AllEntries())
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
