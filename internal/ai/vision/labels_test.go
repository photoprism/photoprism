package vision

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/ai/vision/ollama"
	"github.com/photoprism/photoprism/internal/ai/vision/openai"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/http/scheme"
	"github.com/photoprism/photoprism/pkg/media"
)

func TestGenerateLabels(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		result, err := GenerateLabels(Files{samplesPath + "/chameleon_lime.jpg"}, media.SrcLocal, entity.SrcAuto)

		assert.NoError(t, err)
		assert.IsType(t, classify.Labels{}, result)
		assert.Equal(t, 1, len(result))

		t.Log(result)

		assert.Equal(t, "chameleon", result[0].Name)
		assert.InDelta(t, 7, result[0].Uncertainty, 3)
	})
	t.Run("Cat224", func(t *testing.T) {
		result, err := GenerateLabels(Files{samplesPath + "/cat_224.jpeg"}, media.SrcLocal, entity.SrcAuto)

		assert.NoError(t, err)
		assert.IsType(t, classify.Labels{}, result)
		assert.Equal(t, 1, len(result))

		t.Log(result)

		assert.Equal(t, "cat", result[0].Name)
		assert.InDelta(t, 59, result[0].Uncertainty, 10)
		assert.InDelta(t, float32(0.41), result[0].Confidence(), 0.1)
	})
	t.Run("Cat720", func(t *testing.T) {
		result, err := GenerateLabels(Files{samplesPath + "/cat_720.jpeg"}, media.SrcLocal, entity.SrcAuto)

		assert.NoError(t, err)
		assert.IsType(t, classify.Labels{}, result)
		assert.Equal(t, 1, len(result))

		t.Log(result)

		assert.Equal(t, "cat", result[0].Name)
		assert.InDelta(t, 60, result[0].Uncertainty, 10)
		assert.InDelta(t, float32(0.4), result[0].Confidence(), 0.1)
	})
	t.Run("CustomSourceLocal", func(t *testing.T) {
		labels, err := GenerateLabels(Files{samplesPath + "/cat_224.jpeg"}, media.SrcLocal, entity.SrcManual)
		if err != nil {
			t.Fatalf("GenerateLabels error: %v", err)
		}
		for _, label := range labels {
			if label.Source != entity.SrcManual {
				t.Fatalf("expected custom source %q, got %q", entity.SrcManual, label.Source)
			}
		}
	})
	t.Run("InvalidFile", func(t *testing.T) {
		_, err := GenerateLabels(Files{samplesPath + "/notexisting.jpg"}, media.SrcLocal, entity.SrcAuto)
		assert.Error(t, err)
	})
}

func TestGenerateLabelsRequestShapingForStructuredOutputIdea(t *testing.T) {
	prevConfig := Config
	t.Cleanup(func() {
		Config = prevConfig
	})
	t.Run("OllamaUsesJsonFormatWithSchemaPromptInstructions", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req ApiRequest
			require.NoError(t, json.NewDecoder(r.Body).Decode(&req))

			assert.Equal(t, FormatJSON, req.Format)
			assert.Contains(t, req.Prompt, "Return JSON that matches this schema:")
			assert.Empty(t, req.Schema, "Ollama structured schema payload is not sent yet")

			require.NoError(t, json.NewEncoder(w).Encode(ollama.Response{
				Model:    "gemma3:4b",
				Response: `{"labels":[{"name":"cat","confidence":0.92,"topicality":0.88}]}`,
			}))
		}))
		defer server.Close()

		model := &Model{
			Type:   ModelTypeLabels,
			Name:   "gemma3:4b",
			Engine: ollama.EngineName,
			Service: Service{
				Uri:            server.URL,
				Method:         http.MethodPost,
				RequestFormat:  ApiFormatOllama,
				ResponseFormat: ApiFormatOllama,
				FileScheme:     scheme.Base64,
			},
		}
		model.ApplyEngineDefaults()

		Config = &ConfigValues{
			Models:     Models{model},
			Thresholds: DefaultThresholds,
		}

		labels, err := GenerateLabels(Files{samplesPath + "/cat_224.jpeg"}, media.SrcLocal, entity.SrcAuto)
		require.NoError(t, err)
		require.Len(t, labels, 1)
		assert.Equal(t, "Cat", labels[0].Name)
		assert.Equal(t, entity.SrcOllama, labels[0].Source)
	})
	t.Run("OpenAIUsesStructuredOutputAndStillAppendsSchemaPrompt", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var reqPayload openai.HTTPRequest
			require.NoError(t, json.NewDecoder(r.Body).Decode(&reqPayload))

			require.NotNil(t, reqPayload.Text)
			require.NotNil(t, reqPayload.Text.Format)
			assert.Equal(t, openai.ResponseFormatJSONSchema, reqPayload.Text.Format.Type)
			assert.NotEmpty(t, reqPayload.Text.Format.Schema)

			var promptText string
			for i := range reqPayload.Input {
				if reqPayload.Input[i].Role != "user" {
					continue
				}

				for j := range reqPayload.Input[i].Content {
					if reqPayload.Input[i].Content[j].Type == openai.ContentTypeText {
						promptText = reqPayload.Input[i].Content[j].Text
						break
					}
				}
			}

			if strings.TrimSpace(promptText) == "" {
				t.Fatal("expected user text prompt in OpenAI request")
			}

			assert.Contains(t, promptText, "Return JSON that matches this schema:")

			response := map[string]any{
				"id":    "resp_5450",
				"model": "gpt-5-mini",
				"output": []any{
					map[string]any{
						"role": "assistant",
						"content": []any{
							map[string]any{
								"type": "output_json",
								"json": map[string]any{
									"labels": []map[string]any{
										{
											"name":       "cat",
											"confidence": 0.94,
											"topicality": 0.89,
										},
									},
								},
							},
						},
					},
				},
			}

			require.NoError(t, json.NewEncoder(w).Encode(response))
		}))
		defer server.Close()

		model := &Model{
			Type:   ModelTypeLabels,
			Name:   "gpt-5-mini",
			Engine: openai.EngineName,
			Service: Service{
				Uri:            server.URL,
				Method:         http.MethodPost,
				RequestFormat:  ApiFormatOpenAI,
				ResponseFormat: ApiFormatOpenAI,
				FileScheme:     scheme.Data,
			},
		}
		model.ApplyEngineDefaults()

		Config = &ConfigValues{
			Models:     Models{model},
			Thresholds: DefaultThresholds,
		}

		labels, err := GenerateLabels(Files{samplesPath + "/cat_224.jpeg"}, media.SrcLocal, entity.SrcAuto)
		require.NoError(t, err)
		require.Len(t, labels, 1)
		assert.Equal(t, "Cat", labels[0].Name)
		assert.Equal(t, entity.SrcOpenAI, labels[0].Source)
	})
}
