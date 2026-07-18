package vision

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ai/vision/ollama"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/http/scheme"
)

func TestNewApiRequest(t *testing.T) {
	t.Run("Data", func(t *testing.T) {
		thumbnails := Files{samplesPath + "/chameleon_lime.jpg"}
		result, err := NewApiRequestImages(thumbnails, scheme.Data)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		// t.Logf("request: %#v", result)

		if result != nil {
			json, jsonErr := result.JSON()
			assert.NoError(t, jsonErr)
			assert.NotEmpty(t, json)
			// t.Logf("json: %s", json)
		}
	})
	t.Run("Https", func(t *testing.T) {
		thumbnails := Files{samplesPath + "/chameleon_lime.jpg"}
		result, err := NewApiRequestImages(thumbnails, scheme.Https)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		// t.Logf("request: %#v", result)
		if result != nil {
			json, jsonErr := result.JSON()
			assert.NoError(t, jsonErr)
			assert.NotEmpty(t, json)
			t.Logf("json: %s", json)
		}
	})
}

func TestPerformApiRequestOllama(t *testing.T) {
	t.Run("Labels", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req ApiRequest
			assert.NoError(t, json.NewDecoder(r.Body).Decode(&req))
			assert.Equal(t, FormatJSON, req.Format)
			assert.NoError(t, json.NewEncoder(w).Encode(ollama.Response{
				Model:    "qwen2.5vl:latest",
				Response: `{"labels":[{"name":"test","confidence":0.9,"topicality":0.8}]}`,
			}))
		}))
		defer server.Close()

		apiRequest := &ApiRequest{
			Id:             "test",
			Model:          "qwen2.5vl:latest",
			Format:         FormatJSON,
			Images:         []string{"data:image/jpeg;base64,AA=="},
			ResponseFormat: ApiFormatOllama,
		}

		resp, err := PerformApiRequest(apiRequest, server.URL, http.MethodPost, "")
		assert.NoError(t, err)
		assert.Len(t, resp.Result.Labels, 1)
		assert.Equal(t, "Test", resp.Result.Labels[0].Name)
		assert.Nil(t, resp.Result.Caption)
	})
	t.Run("LabelsWithCodeFence", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NoError(t, json.NewEncoder(w).Encode(ollama.Response{
				Model:    "gemma3:latest",
				Response: "```json\n{\"labels\":[{\"name\":\"lingerie\",\"confidence\":0.81,\"topicality\":0.73}]}\n```\nThe model provided additional commentary.",
			}))
		}))
		defer server.Close()

		apiRequest := &ApiRequest{
			Id:             "fenced",
			Model:          "gemma3:latest",
			Format:         FormatJSON,
			Images:         []string{"data:image/jpeg;base64,AA=="},
			ResponseFormat: ApiFormatOllama,
		}

		resp, err := PerformApiRequest(apiRequest, server.URL, http.MethodPost, "")
		assert.NoError(t, err)
		if assert.Len(t, resp.Result.Labels, 1) {
			assert.Equal(t, "Lingerie", resp.Result.Labels[0].Name)
		}
	})
	t.Run("CaptionFallback", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NoError(t, json.NewEncoder(w).Encode(ollama.Response{
				Model:    "qwen2.5vl:latest",
				Response: "plain text",
			}))
		}))
		defer server.Close()

		apiRequest := &ApiRequest{
			Id:             "test2",
			Model:          "qwen2.5vl:latest",
			Format:         FormatJSON,
			Images:         []string{"data:image/jpeg;base64,AA=="},
			ResponseFormat: ApiFormatOllama,
		}

		resp, err := PerformApiRequest(apiRequest, server.URL, http.MethodPost, "")
		assert.NoError(t, err)
		assert.Len(t, resp.Result.Labels, 0)
		if assert.NotNil(t, resp.Result.Caption) {
			assert.Equal(t, "plain text", resp.Result.Caption.Text)
		}
	})
	t.Run("CaptionThinkingFallback", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NoError(t, json.NewEncoder(w).Encode(ollama.Response{
				Model:    "qwen3-vl:4b",
				Response: "",
				Thinking: "A tabby cat with a white chest stares upward.",
			}))
		}))
		defer server.Close()

		apiRequest := &ApiRequest{
			Id:             "test3",
			Model:          "qwen3-vl:4b",
			Format:         FormatJSON,
			Images:         []string{"data:image/jpeg;base64,AA=="},
			ResponseFormat: ApiFormatOllama,
		}

		resp, err := PerformApiRequest(apiRequest, server.URL, http.MethodPost, "")
		assert.NoError(t, err)
		assert.Len(t, resp.Result.Labels, 0)
		if assert.NotNil(t, resp.Result.Caption) {
			assert.Equal(t, "A tabby cat with a white chest stares upward.", resp.Result.Caption.Text)
		}
	})
}

func TestPerformApiRequestOpenAIHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "org-123", r.Header.Get(header.OpenAIOrg))
		assert.Equal(t, "proj-abc", r.Header.Get(header.OpenAIProject))

		response := map[string]any{
			"id":    "resp_123",
			"model": "gpt-5-mini",
			"output": []any{
				map[string]any{
					"role": "assistant",
					"content": []any{
						map[string]any{
							"type": "output_text",
							"text": "A scenic mountain view.",
						},
					},
				},
			},
		}

		assert.NoError(t, json.NewEncoder(w).Encode(response))
	}))
	defer server.Close()

	req := &ApiRequest{
		Id:             "headers",
		Model:          "gpt-5-mini",
		Images:         []string{"data:image/jpeg;base64,AA=="},
		ResponseFormat: ApiFormatOpenAI,
		Org:            "org-123",
		Project:        "proj-abc",
	}

	resp, err := PerformApiRequest(req, server.URL, http.MethodPost, "")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Result.Caption)
	assert.Equal(t, "A scenic mountain view.", resp.Result.Caption.Text)
}

// shrinkRetryDelay speeds up 429 retry tests by using a tiny backoff and
// restores the package defaults afterwards.
func shrinkRetryDelay(t *testing.T) {
	prevDelay, prevMax := ServiceRetryDelay, ServiceRetryMaxDelay
	ServiceRetryDelay = time.Millisecond
	ServiceRetryMaxDelay = 5 * time.Millisecond
	t.Cleanup(func() {
		ServiceRetryDelay = prevDelay
		ServiceRetryMaxDelay = prevMax
	})
}

func TestPerformApiRequestRetry(t *testing.T) {
	t.Run("OllamaRetryThenSuccess", func(t *testing.T) {
		shrinkRetryDelay(t)
		var calls int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if atomic.AddInt32(&calls, 1) == 1 {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			assert.NoError(t, json.NewEncoder(w).Encode(ollama.Response{
				Model:    "qwen2.5vl:latest",
				Response: `{"labels":[{"name":"test","confidence":0.9,"topicality":0.8}]}`,
			}))
		}))
		defer server.Close()

		apiRequest := &ApiRequest{
			Id:             "retry-ollama",
			Model:          "qwen2.5vl:latest",
			Format:         FormatJSON,
			Images:         []string{"data:image/jpeg;base64,AA=="},
			ResponseFormat: ApiFormatOllama,
		}

		resp, err := PerformApiRequest(apiRequest, server.URL, http.MethodPost, "")
		assert.NoError(t, err)
		assert.Len(t, resp.Result.Labels, 1)
		assert.Equal(t, int32(2), atomic.LoadInt32(&calls))
	})
	t.Run("OpenAIRetryThenSuccess", func(t *testing.T) {
		shrinkRetryDelay(t)
		var calls int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if atomic.AddInt32(&calls, 1) == 1 {
				w.Header().Set(header.RetryAfter, "0")
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			response := map[string]any{
				"id":    "resp_123",
				"model": "gpt-5-mini",
				"output": []any{
					map[string]any{
						"role": "assistant",
						"content": []any{
							map[string]any{"type": "output_text", "text": "A scenic mountain view."},
						},
					},
				},
			}
			assert.NoError(t, json.NewEncoder(w).Encode(response))
		}))
		defer server.Close()

		req := &ApiRequest{
			Id:             "retry-openai",
			Model:          "gpt-5-mini",
			Images:         []string{"data:image/jpeg;base64,AA=="},
			ResponseFormat: ApiFormatOpenAI,
		}

		resp, err := PerformApiRequest(req, server.URL, http.MethodPost, "")
		assert.NoError(t, err)
		assert.NotNil(t, resp.Result.Caption)
		assert.Equal(t, "A scenic mountain view.", resp.Result.Caption.Text)
		assert.Equal(t, int32(2), atomic.LoadInt32(&calls))
	})
	t.Run("NonRetryableStatusStaysTerminal", func(t *testing.T) {
		shrinkRetryDelay(t)
		var calls int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&calls, 1)
			w.WriteHeader(http.StatusBadRequest)
			assert.NoError(t, json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]any{"message": "bad request"},
			}))
		}))
		defer server.Close()

		req := &ApiRequest{
			Id:             "terminal",
			Model:          "gpt-5-mini",
			Images:         []string{"data:image/jpeg;base64,AA=="},
			ResponseFormat: ApiFormatOpenAI,
		}

		_, err := PerformApiRequest(req, server.URL, http.MethodPost, "")
		assert.Error(t, err)
		assert.Equal(t, int32(1), atomic.LoadInt32(&calls))
	})
	t.Run("RetriesExhausted", func(t *testing.T) {
		shrinkRetryDelay(t)
		var calls int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&calls, 1)
			w.WriteHeader(http.StatusTooManyRequests)
			assert.NoError(t, json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]any{"message": "rate limited"},
			}))
		}))
		defer server.Close()

		req := &ApiRequest{
			Id:             "exhausted",
			Model:          "gpt-5-mini",
			Images:         []string{"data:image/jpeg;base64,AA=="},
			ResponseFormat: ApiFormatOpenAI,
		}

		_, err := PerformApiRequest(req, server.URL, http.MethodPost, "")
		assert.Error(t, err)
		assert.Equal(t, ServiceMaxRetries+1, int(atomic.LoadInt32(&calls)))
	})
}

func TestValidateApiRequestURL(t *testing.T) {
	t.Run("AcceptHttpAndHttps", func(t *testing.T) {
		assert.NoError(t, validateApiRequestURL("http://localhost:1234/api"))
		assert.NoError(t, validateApiRequestURL("https://api.example.com/v1"))
	})
	t.Run("RejectUnsupportedScheme", func(t *testing.T) {
		assert.Error(t, validateApiRequestURL("file:///tmp/payload.json"))
	})
	t.Run("RejectMissingHost", func(t *testing.T) {
		assert.Error(t, validateApiRequestURL("https:///v1"))
	})
}

func TestPerformApiRequestResponseLimit(t *testing.T) {
	// Shrink the cap so the test does not allocate the 32 MiB default.
	prev := MaxResponseBytes
	MaxResponseBytes = 1024
	t.Cleanup(func() { MaxResponseBytes = prev })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		//nolint:gosec // test fixture writes a locally generated payload only
		_, _ = w.Write(make([]byte, int(MaxResponseBytes)+512))
	}))
	defer server.Close()

	apiRequest := &ApiRequest{
		Id:             "toolarge",
		Model:          "qwen2.5vl:latest",
		Format:         FormatJSON,
		Images:         []string{"data:image/jpeg;base64,AA=="},
		ResponseFormat: ApiFormatOllama,
	}

	resp, err := PerformApiRequest(apiRequest, server.URL, http.MethodPost, "")
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "exceeds the maximum size")
}
