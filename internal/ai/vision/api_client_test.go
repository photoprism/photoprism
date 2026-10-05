package vision

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/vision/ollama"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/http/scheme"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/txt"
)

func TestNewApiRequest(t *testing.T) {
	t.Run("Data", func(t *testing.T) {
		thumbnails := Files{samplesPath + "/chameleon_lime.jpg"}
		result, err := NewApiRequestImages(thumbnails, scheme.Data, media.SrcLocal)

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
		result, err := NewApiRequestImages(thumbnails, scheme.Https, media.SrcLocal)

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

func TestPerformApiRequestVisionStatus(t *testing.T) {
	newServer := func(code int, body string) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(header.ContentType, header.ContentTypeJson)
			w.WriteHeader(code)
			_, _ = w.Write([]byte(body))
		}))
	}

	request := func() *ApiRequest {
		return &ApiRequest{Id: "3487da77-246e-4b4d-b1b2-2b5d5ee7b5a8", Images: []string{"data:image/jpeg;base64,AA=="}, ResponseFormat: ApiFormatVision}
	}

	t.Run("Forbidden", func(t *testing.T) {
		// A service that refuses the request, e.g. because its Vision API is disabled.
		server := newServer(http.StatusForbidden, `{"id":"3487da77-246e-4b4d-b1b2-2b5d5ee7b5a8","code":403,"error":"Forbidden","result":{}}`)
		defer server.Close()

		resp, err := PerformApiRequest(request(), server.URL, http.MethodPost, "")
		assert.EqualError(t, err, "vision service request failed (status 403)")
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusForbidden, resp.Code)
	})
	t.Run("NoErrorText", func(t *testing.T) {
		server := newServer(http.StatusUnauthorized, `{"code":401}`)
		defer server.Close()

		_, err := PerformApiRequest(request(), server.URL, http.MethodPost, "")
		assert.EqualError(t, err, "vision service request failed (status 401)")
	})
	t.Run("ErrorTextOmitted", func(t *testing.T) {
		server := newServer(http.StatusInternalServerError, `{"code":500,"error":"a\nb\u001b[31m"}`)
		defer server.Close()

		_, err := PerformApiRequest(request(), server.URL, http.MethodPost, "")
		assert.EqualError(t, err, "vision service request failed (status 500)")
	})
	t.Run("MultipleChoices", func(t *testing.T) {
		server := newServer(http.StatusMultipleChoices, `{}`)
		defer server.Close()

		_, err := PerformApiRequest(request(), server.URL, http.MethodPost, "")
		assert.EqualError(t, err, "vision service request failed (status 300)")
	})
	t.Run("Success", func(t *testing.T) {
		server := newServer(http.StatusOK, `{"id":"3487da77-246e-4b4d-b1b2-2b5d5ee7b5a8","code":200,"result":{"labels":[{"name":"cat","confidence":0.9}]}}`)
		defer server.Close()

		resp, err := PerformApiRequest(request(), server.URL, http.MethodPost, "")
		assert.NoError(t, err)
		assert.Len(t, resp.Result.Labels, 1)
	})
}

// captureLogs replaces the package logger and the system log with test loggers, disables the audit logger,
// and returns their hooks.
func captureLogs(t *testing.T) (logHook, systemHook *logtest.Hook) {
	t.Helper()

	prevLog, prevSystem, prevAudit := log, event.SystemLog, event.AuditLog
	t.Cleanup(func() { log, event.SystemLog, event.AuditLog = prevLog, prevSystem, prevAudit })
	event.AuditLog = nil

	logger, logHook := logtest.NewNullLogger()
	logger.SetLevel(logrus.TraceLevel)
	log = logger

	systemLogger, systemHook := logtest.NewNullLogger()
	systemLogger.SetLevel(logrus.TraceLevel)
	event.SystemLog = systemLogger

	return logHook, systemHook
}

// TestPerformApiRequestErrorLog checks that the text of a failed response is only written to the system log.
func TestPerformApiRequestErrorLog(t *testing.T) {
	const marker = "remote-body-marker"

	resetServiceFailures(t)

	for _, tc := range []struct {
		name   string
		format ApiFormat
		code   int
		body   string
		err    string
	}{
		{"VisionJson", ApiFormatVision, http.StatusInternalServerError, `{"code":500,"error":"` + marker + `"}`, "vision service request failed (status 500)"},
		{"VisionText", ApiFormatVision, http.StatusInternalServerError, `<html>` + marker + `</html>`, "vision service request failed (status 500)"},
		{"VisionRedirect", ApiFormatVision, http.StatusMultipleChoices, `{"error":"` + marker + `"}`, "vision service request failed (status 300)"},
		{"Ollama", ApiFormatOllama, http.StatusInternalServerError, `{"code":500,"error":"` + marker + `"}`, "ollama service request failed (status 500)"},
		{"OllamaCaption", ApiFormatOllama, http.StatusInternalServerError, `{"model":"qwen2.5vl:latest","response":"` + marker + `"}`, "ollama service request failed (status 500)"},
		{"OllamaText", ApiFormatOllama, http.StatusInternalServerError, `<html>` + marker + `</html>`, "ollama service request failed (status 500)"},
		{"OpenAI", ApiFormatOpenAI, http.StatusBadRequest, `{"error":{"message":"` + marker + `"}}`, "openai service request failed (status 400)"},
		{"OpenAIText", ApiFormatOpenAI, http.StatusUnauthorized, marker, "openai service request failed (status 401)"},
		{"OpenAISuccessError", ApiFormatOpenAI, http.StatusOK, `{"error":{"message":"` + marker + `"}}`, "openai service returned an invalid response (status 200)"},
		{"VisionDecodeError", ApiFormatVision, http.StatusOK, `{"code":200,"model":{"tensorflow":{"input":{"resizeOperation":"` + marker + `"}}}}`, "vision service returned an invalid response (status 200)"},
		{"VisionNumberLiteral", ApiFormatVision, http.StatusOK, `{"code":1` + strings.Repeat("0", 200) + `,"error":"` + marker + `"}`, "vision service returned an invalid response (status 200)"},
		{"OllamaDecodeError", ApiFormatOllama, http.StatusOK, `{"model":"m","created_at":"` + marker + `"}`, "ollama service returned an invalid response (status 200)"},
		{"OpenAIDecodeError", ApiFormatOpenAI, http.StatusOK, `{"output":"` + marker + `"}`, "openai service returned an invalid response (status 200)"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			logHook, systemHook := captureLogs(t)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set(header.ContentType, header.ContentTypeJson)
				w.WriteHeader(tc.code)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()

			request := &ApiRequest{Id: "3487da77-246e-4b4d-b1b2-2b5d5ee7b5a8", Model: "qwen2.5vl:latest", Images: []string{"data:image/jpeg;base64,AA=="}, ResponseFormat: tc.format}
			_, err := PerformApiRequest(request, server.URL, http.MethodPost, "")

			if tc.err == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.err)
			}

			for _, entry := range logHook.AllEntries() {
				assert.NotContains(t, entry.Message, marker, entry.Level.String())
			}

			logged := tc.err
			if logged == "" {
				logged = serviceError(tc.format, tc.code).Error()
			}

			require.Len(t, systemHook.AllEntries(), 1)
			entry := systemHook.LastEntry()
			assert.Equal(t, logrus.ErrorLevel, entry.Level)
			assert.True(t, strings.HasPrefix(entry.Message, "vision: "+logged+" › "), entry.Message)
			assert.NotContains(t, entry.Message, "\x1b[")

			if tc.name != "VisionNumberLiteral" && tc.name != "OpenAIDecodeError" {
				assert.Contains(t, entry.Message, marker)
			}
			assert.NotContains(t, entry.Message, "truncated")
		})
	}
	t.Run("Success", func(t *testing.T) {
		logHook, systemHook := captureLogs(t)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(header.ContentType, header.ContentTypeJson)
			_, _ = w.Write([]byte(`{"model":"qwen2.5vl:latest","response":"{\"labels\":[{\"name\":\"` + marker + `\",\"confidence\":0.9}]}"}`))
		}))
		defer server.Close()

		request := &ApiRequest{Id: "3487da77-246e-4b4d-b1b2-2b5d5ee7b5a8", Model: "qwen2.5vl:latest", Images: []string{"data:image/jpeg;base64,AA=="}, ResponseFormat: ApiFormatOllama}
		_, err := PerformApiRequest(request, server.URL, http.MethodPost, "")
		require.NoError(t, err)

		assert.Empty(t, systemHook.AllEntries())

		for _, entry := range logHook.AllEntries() {
			if entry.Level != logrus.TraceLevel {
				assert.NotContains(t, entry.Message, marker, entry.Level.String())
			}
		}
	})
	t.Run("LargeBody", func(t *testing.T) {
		logHook, systemHook := captureLogs(t)

		body := strings.Repeat("\u00e9", 5*1024*1024/2)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(body))
		}))
		defer server.Close()

		request := &ApiRequest{Id: "3487da77-246e-4b4d-b1b2-2b5d5ee7b5a8", Images: []string{"data:image/jpeg;base64,AA=="}, ResponseFormat: ApiFormatVision}
		_, err := PerformApiRequest(request, server.URL, http.MethodPost, "")
		assert.EqualError(t, err, "vision service request failed (status 502)")

		for _, entry := range logHook.AllEntries() {
			assert.NotContains(t, entry.Message, "\u00e9", entry.Level.String())
		}

		require.Len(t, systemHook.AllEntries(), 1)
		msg := systemHook.LastEntry().Message
		assert.Less(t, len(msg), txt.ClipLongText+200)
		assert.True(t, strings.HasSuffix(msg, " (truncated from 5.2 MB)"), msg[len(msg)-60:])
	})
}

// TestLogServiceResponse checks the system log entry written for the text of a failed response.
func TestLogServiceResponse(t *testing.T) {
	err := serviceError(ApiFormatOpenAI, http.StatusBadGateway)

	t.Run("Success", func(t *testing.T) {
		logHook, systemHook := captureLogs(t)
		logServiceResponse(err, []byte("<html>bad gateway</html>"))
		assert.Empty(t, logHook.AllEntries())
		require.Len(t, systemHook.AllEntries(), 1)
		assert.Equal(t, logrus.ErrorLevel, systemHook.LastEntry().Level)
		assert.Equal(t, `vision: openai service request failed (status 502) › "<html>bad gateway</html>"`, systemHook.LastEntry().Message)
	})
	t.Run("Escaped", func(t *testing.T) {
		_, systemHook := captureLogs(t)
		logServiceResponse(err, []byte("a\nb\x1b[31m\u202e"))
		require.Len(t, systemHook.AllEntries(), 1)
		assert.Equal(t, `vision: openai service request failed (status 502) › "a\nb\x1b[31m\u202e"`, systemHook.LastEntry().Message)
	})
	t.Run("Truncated", func(t *testing.T) {
		_, systemHook := captureLogs(t)
		text := strings.Repeat("a", txt.ClipLongText-1) + "\u00e9" + strings.Repeat("b", 1000)
		logServiceResponse(err, []byte(text))
		require.Len(t, systemHook.AllEntries(), 1)
		assert.Equal(t, `vision: openai service request failed (status 502) › "`+strings.Repeat("a", txt.ClipLongText-1)+`" (truncated from 5.1 kB)`, systemHook.LastEntry().Message)
	})
	t.Run("Limit", func(t *testing.T) {
		_, systemHook := captureLogs(t)
		logServiceResponse(err, []byte(strings.Repeat("a", txt.ClipLongText)+"\n"))
		require.Len(t, systemHook.AllEntries(), 1)
		assert.Equal(t, `vision: openai service request failed (status 502) › "`+strings.Repeat("a", txt.ClipLongText)+`"`, systemHook.LastEntry().Message)
	})
	t.Run("OverLimit", func(t *testing.T) {
		_, systemHook := captureLogs(t)
		logServiceResponse(err, []byte(strings.Repeat("a", txt.ClipLongText+1)))
		require.Len(t, systemHook.AllEntries(), 1)
		assert.Equal(t, `vision: openai service request failed (status 502) › "`+strings.Repeat("a", txt.ClipLongText)+`" (truncated from 4.1 kB)`, systemHook.LastEntry().Message)
	})
	t.Run("Empty", func(t *testing.T) {
		_, systemHook := captureLogs(t)
		logServiceResponse(err, nil)
		require.Len(t, systemHook.AllEntries(), 1)
		assert.Equal(t, `vision: openai service request failed (status 502) › ""`, systemHook.LastEntry().Message)
	})
}

// TestServiceName checks that service names are derived from sanitized response formats.
func TestServiceName(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		assert.Equal(t, "openai", serviceName(ApiFormatOpenAI))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "remote", serviceName(""))
	})
	t.Run("Sanitized", func(t *testing.T) {
		assert.Equal(t, "a_b", serviceName("a\nb"))
	})
}

// TestServiceError checks that the error for a failed request names the service and status only.
func TestServiceError(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		assert.EqualError(t, serviceError(ApiFormatOpenAI, http.StatusBadRequest), "openai service request failed (status 400)")
	})
	t.Run("UnknownFormat", func(t *testing.T) {
		assert.EqualError(t, serviceError("", http.StatusBadGateway), "remote service request failed (status 502)")
	})
}

// TestTransportError checks the error returned for a request without a response, and what it logs.
func TestTransportError(t *testing.T) {
	t.Run("ConnectionError", func(t *testing.T) {
		logHook, systemHook := captureLogs(t)
		cause := &url.Error{Op: "Get", URL: "https://vision.example.com/api", Err: errors.New("remote-marker")}
		assert.EqualError(t, transportError(ApiFormatOpenAI, cause), "openai service request failed (connection error)")
		assert.Empty(t, logHook.AllEntries())
		require.Len(t, systemHook.AllEntries(), 1)
		assert.Contains(t, systemHook.LastEntry().Message, "remote-marker")
	})
	t.Run("RedactsUri", func(t *testing.T) {
		_, systemHook := captureLogs(t)
		cause := &url.Error{Op: "Post", URL: "https://scoped@vision.example.com/api?key=test-secret", Err: errors.New("remote-marker")}
		assert.EqualError(t, transportError(ApiFormatOpenAI, cause), "openai service request failed (connection error)")
		require.Len(t, systemHook.AllEntries(), 1)
		assert.Contains(t, systemHook.LastEntry().Message, "vision.example.com/api")
		assert.Contains(t, systemHook.LastEntry().Message, "remote-marker")
		assert.NotContains(t, systemHook.LastEntry().Message, "test-secret")
		assert.NotContains(t, systemHook.LastEntry().Message, "scoped")
	})
	t.Run("Timeout", func(t *testing.T) {
		_, systemHook := captureLogs(t)
		cause := &url.Error{Op: "Post", URL: "https://vision.example.com/api", Err: context.DeadlineExceeded}
		assert.EqualError(t, transportError(ApiFormatVision, cause), "vision service request failed (timeout)")
		require.Len(t, systemHook.AllEntries(), 1)
	})
}

// TestPerformApiRequestTransportError checks that a redirect the client cannot follow returns no response text.
func TestPerformApiRequestTransportError(t *testing.T) {
	const marker = "remote-location-marker"

	logHook, systemHook := captureLogs(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "%zz "+marker)
		w.WriteHeader(http.StatusFound)
	}))
	defer server.Close()

	request := &ApiRequest{Id: "3487da77-246e-4b4d-b1b2-2b5d5ee7b5a8", Images: []string{"data:image/jpeg;base64,AA=="}, ResponseFormat: ApiFormatVision}
	_, err := PerformApiRequest(request, server.URL, http.MethodPost, "")
	assert.EqualError(t, err, "vision service request failed (connection error)")

	for _, entry := range logHook.AllEntries() {
		assert.NotContains(t, entry.Message, marker, entry.Level.String())
	}

	require.Len(t, systemHook.AllEntries(), 1)
	assert.Contains(t, systemHook.LastEntry().Message, marker)
}

// TestPerformApiRequestBodyReadError checks that a response body that cannot be read returns no response text.
func TestPerformApiRequestBodyReadError(t *testing.T) {
	logHook, systemHook := captureLogs(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, buf, err := http.NewResponseController(w).Hijack()
		require.NoError(t, err)
		_, _ = buf.WriteString("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: 100\r\n\r\n{\"code\":")
		_ = buf.Flush()
		_ = conn.Close()
	}))
	defer server.Close()

	request := &ApiRequest{Id: "3487da77-246e-4b4d-b1b2-2b5d5ee7b5a8", Images: []string{"data:image/jpeg;base64,AA=="}, ResponseFormat: ApiFormatVision}
	_, err := PerformApiRequest(request, server.URL, http.MethodPost, "")
	assert.EqualError(t, err, "vision service request failed (connection error)")
	assert.Empty(t, logHook.AllEntries())
	require.Len(t, systemHook.AllEntries(), 1)
	assert.Contains(t, systemHook.LastEntry().Message, "unexpected EOF")
}

// TestResponseError checks the error returned for a failed or unparsable response, and what it logs.
func TestResponseError(t *testing.T) {
	t.Run("FailedStatus", func(t *testing.T) {
		_, systemHook := captureLogs(t)
		assert.EqualError(t, responseError(ApiFormatVision, http.StatusForbidden, errors.New("remote-marker")), "vision service request failed (status 403)")
		assert.Empty(t, systemHook.AllEntries())
	})
	t.Run("InvalidResponse", func(t *testing.T) {
		logHook, systemHook := captureLogs(t)
		assert.EqualError(t, responseError(ApiFormatOllama, http.StatusOK, errors.New("remote-marker\n")), "ollama service returned an invalid response (status 200)")
		assert.Empty(t, logHook.AllEntries())
		require.Len(t, systemHook.AllEntries(), 1)
		assert.Equal(t, `vision: ollama service returned an invalid response (status 200) › "remote-marker"`, systemHook.LastEntry().Message)
	})
	t.Run("NoParseError", func(t *testing.T) {
		_, systemHook := captureLogs(t)
		assert.EqualError(t, responseError(ApiFormatOllama, http.StatusOK, nil), "ollama service returned an invalid response (status 200)")
		assert.Empty(t, systemHook.AllEntries())
	})
}

// TestPerformApiRequestRedirect checks that service redirects are reported rather than followed.
func TestPerformApiRequestRedirect(t *testing.T) {
	request := func() *ApiRequest {
		return &ApiRequest{Id: "3487da77-246e-4b4d-b1b2-2b5d5ee7b5a8", Images: []string{"data:image/jpeg;base64,AA=="}, ResponseFormat: ApiFormatVision}
	}

	for _, code := range []int{http.StatusMovedPermanently, http.StatusFound, http.StatusTemporaryRedirect, http.StatusPermanentRedirect} {
		t.Run(http.StatusText(code), func(t *testing.T) {
			logHook, systemHook := captureLogs(t)

			var targetHits atomic.Int32

			target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				targetHits.Add(1)
				w.WriteHeader(http.StatusOK)
			}))
			defer target.Close()

			location := strings.Replace(target.URL, "http://", "http://user:pass@", 1) + "/next?sig=abc123"

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set(header.Location, location)
				w.WriteHeader(code)
			}))
			defer server.Close()

			_, err := PerformApiRequest(request(), server.URL, http.MethodPost, "service-key")

			assert.EqualError(t, err, fmt.Sprintf("vision service request failed (status %d, redirect not followed)", code))
			assert.Zero(t, targetHits.Load())
			assert.Empty(t, logHook.AllEntries())
			require.Len(t, systemHook.AllEntries(), 1)

			assert.Contains(t, systemHook.LastEntry().Message, "location "+target.URL+"/next?***")
		})
	}
}

// TestRedirectError checks the error and the system log entry for a service redirect.
func TestRedirectError(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		logHook, systemHook := captureLogs(t)

		err := redirectError(ApiFormatOpenAI, http.StatusFound, "https://user:pass@example.com/v1?sig=abc123#part")

		assert.EqualError(t, err, "openai service request failed (status 302, redirect not followed)")
		assert.Empty(t, logHook.AllEntries())
		require.Len(t, systemHook.AllEntries(), 1)
		assert.Contains(t, systemHook.LastEntry().Message, `location https://example.com/v1?***`)
		assert.NotContains(t, systemHook.LastEntry().Message, "part")
	})
	t.Run("RelativeLocation", func(t *testing.T) {
		_, systemHook := captureLogs(t)
		assert.EqualError(t, redirectError("", http.StatusMovedPermanently, "/next?sig=abc123"), "remote service request failed (status 301, redirect not followed)")
		require.Len(t, systemHook.AllEntries(), 1)
		assert.Contains(t, systemHook.LastEntry().Message, "location /next?***")
	})
}

// TestNoRedirect checks that the redirect policy returns the redirect response instead of following it.
func TestNoRedirect(t *testing.T) {
	assert.ErrorIs(t, noRedirect(nil, nil), http.ErrUseLastResponse)
}

// TestRedirectTarget checks that a redirect target is logged without its userinfo, query, and fragment.
func TestRedirectTarget(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		assert.Equal(t, "https://example.com/v1/next", redirectTarget("https://user:pass@example.com/v1/next"))
		assert.Equal(t, "https://example.com/v1?***", redirectTarget("https://example.com/v1?sig=abc123#part"))
		assert.Equal(t, "//example.com/next?***", redirectTarget("//user:pass@example.com/next?a='x'&sig=abc123"))
		assert.Equal(t, "/next?***", redirectTarget("/next?"))
	})
	t.Run("Unparseable", func(t *testing.T) {
		assert.Equal(t, "https://example.com/%zz?***", redirectTarget("https://example.com/%zz?sig=abc123"))
	})
}

// TestInvalidUriError checks the error and the system log entry for a service URI that is not a valid request URL.
func TestInvalidUriError(t *testing.T) {
	logHook, systemHook := captureLogs(t)

	err := invalidUriError(ApiFormatOllama, errors.New(`parse "://models.example.com/api?key=abc123": missing protocol scheme`))

	assert.EqualError(t, err, "ollama service request failed (invalid service uri)")
	assert.Empty(t, logHook.AllEntries())
	require.Len(t, systemHook.AllEntries(), 1)
	assert.Contains(t, systemHook.LastEntry().Message, "models.example.com/api?***")
	assert.NotContains(t, systemHook.LastEntry().Message, "abc123")
}

// TestPerformApiRequestInvalidUri checks that an invalid service URI returns a fixed error.
func TestPerformApiRequestInvalidUri(t *testing.T) {
	logHook, systemHook := captureLogs(t)

	request := &ApiRequest{Id: "3487da77-246e-4b4d-b1b2-2b5d5ee7b5a8", Images: []string{"data:image/jpeg;base64,AA=="}, ResponseFormat: ApiFormatOpenAI}
	_, err := PerformApiRequest(request, "://models.example.com/api?key=abc123", http.MethodPost, "")

	assert.EqualError(t, err, "openai service request failed (invalid service uri)")
	assert.Empty(t, logHook.AllEntries())
	require.Len(t, systemHook.AllEntries(), 1)
	assert.NotContains(t, systemHook.LastEntry().Message, "abc123")
}

// TestTransportErrorQuery checks that the query of a URI in a transport error is not written to the system log.
func TestTransportErrorQuery(t *testing.T) {
	logHook, systemHook := captureLogs(t)

	err := transportError(ApiFormatOpenAI, errors.New(`Post "https://fn.example.net/api?code=abc123": dial tcp: connection refused`))

	assert.EqualError(t, err, "openai service request failed (connection error)")
	assert.Empty(t, logHook.AllEntries())
	require.Len(t, systemHook.AllEntries(), 1)
	assert.Contains(t, systemHook.LastEntry().Message, "fn.example.net/api?***")
	assert.NotContains(t, systemHook.LastEntry().Message, "abc123")
}
