package vision

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/vision/openai"
	"github.com/photoprism/photoprism/internal/ai/vision/schema"
	"github.com/photoprism/photoprism/internal/entity"
)

func TestOpenAIBuilderBuild(t *testing.T) {
	model := &Model{
		Type:   ModelTypeLabels,
		Name:   openai.DefaultModel,
		Engine: openai.EngineName,
	}
	model.ApplyEngineDefaults()

	request, err := openaiBuilder{}.Build(context.Background(), model, Files{examplesPath + "/chameleon_lime.jpg"})
	require.NoError(t, err)
	require.NotNil(t, request)

	assert.Equal(t, ApiFormatOpenAI, request.ResponseFormat)
	assert.NotEmpty(t, request.Images)
	assert.NotNil(t, request.Options)
	assert.Equal(t, openai.DefaultDetail, request.Options.Detail)
	assert.True(t, request.Options.ForceJson)
	assert.GreaterOrEqual(t, request.Options.MaxOutputTokens, openai.LabelsMaxTokens)
}

func TestOpenAIBuilderBuildCaptionDisablesForceJSON(t *testing.T) {
	model := &Model{
		Type:    ModelTypeCaption,
		Name:    openai.DefaultModel,
		Engine:  openai.EngineName,
		Options: &ApiRequestOptions{ForceJson: true},
	}
	model.ApplyEngineDefaults()

	request, err := openaiBuilder{}.Build(context.Background(), model, Files{examplesPath + "/chameleon_lime.jpg"})
	require.NoError(t, err)
	require.NotNil(t, request)
	require.NotNil(t, request.Options)
	assert.False(t, request.Options.ForceJson)
	assert.GreaterOrEqual(t, request.Options.MaxOutputTokens, openai.CaptionMaxTokens)
}

func TestApiRequestJSONForOpenAI(t *testing.T) {
	req := &ApiRequest{
		Model:          "gpt-5-mini",
		System:         "system",
		Prompt:         "describe the scene",
		Images:         []string{"data:image/jpeg;base64,AA=="},
		ResponseFormat: ApiFormatOpenAI,
		Options: &ApiRequestOptions{
			Detail:          openai.DefaultDetail,
			MaxOutputTokens: 128,
			Temperature:     0.2,
			TopP:            0.8,
			ForceJson:       true,
		},
		Schema: json.RawMessage(`{"type":"object","properties":{"caption":{"type":"object"}}}`),
	}

	payload, err := req.JSON()
	require.NoError(t, err)

	var decoded struct {
		Model string `json:"model"`
		Input []struct {
			Role    string `json:"role"`
			Content []struct {
				Type string `json:"type"`
			} `json:"content"`
		} `json:"input"`
		Text struct {
			Format struct {
				Type   string          `json:"type"`
				Name   string          `json:"name"`
				Schema json.RawMessage `json:"schema"`
				Strict bool            `json:"strict"`
			} `json:"format"`
		} `json:"text"`
		Reasoning struct {
			Effort string `json:"effort"`
		} `json:"reasoning"`
		MaxOutputTokens int `json:"max_output_tokens"`
	}

	require.NoError(t, json.Unmarshal(payload, &decoded))
	assert.Equal(t, "gpt-5-mini", decoded.Model)
	require.Len(t, decoded.Input, 2)
	assert.Equal(t, "system", decoded.Input[0].Role)
	assert.Equal(t, openai.ResponseFormatJSONSchema, decoded.Text.Format.Type)
	assert.Equal(t, schema.JsonSchemaName(decoded.Text.Format.Schema, openai.DefaultSchemaVersion), decoded.Text.Format.Name)
	assert.False(t, decoded.Text.Format.Strict)
	assert.NotNil(t, decoded.Text.Format.Schema)
	assert.Equal(t, "low", decoded.Reasoning.Effort)
	assert.Equal(t, 128, decoded.MaxOutputTokens)
}

func TestApiRequestJSONForOpenAIDefaultSchemaName(t *testing.T) {
	req := &ApiRequest{
		Model:          "gpt-5-mini",
		Images:         []string{"data:image/jpeg;base64,AA=="},
		ResponseFormat: ApiFormatOpenAI,
		Options: &ApiRequestOptions{
			Detail:          openai.DefaultDetail,
			MaxOutputTokens: 64,
			ForceJson:       true,
		},
		Schema: json.RawMessage(`{"type":"object"}`),
	}

	payload, err := req.JSON()
	require.NoError(t, err)

	var decoded struct {
		Text struct {
			Format struct {
				Name string `json:"name"`
			} `json:"format"`
		} `json:"text"`
	}

	require.NoError(t, json.Unmarshal(payload, &decoded))
	assert.Equal(t, schema.JsonSchemaName(req.Schema, openai.DefaultSchemaVersion), decoded.Text.Format.Name)
}

func TestOpenAIParserParsesJSONFromTextPayload(t *testing.T) {
	respPayload := `{
		"id": "resp_123",
		"model": "gpt-5-mini",
		"output": [{
			"role": "assistant",
			"content": [{
				"type": "output_text",
				"text": "{\"labels\":[{\"name\":\"deer\",\"confidence\":0.98,\"topicality\":0.99}]}"
			}]
		}]
	}`

	req := &ApiRequest{
		Id:             "test",
		Model:          "gpt-5-mini",
		ResponseFormat: ApiFormatOpenAI,
	}

	resp, err := openaiParser{}.Parse(context.Background(), req, []byte(respPayload), http.StatusOK)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Result.Labels, 1)
	assert.Equal(t, "Deer", resp.Result.Labels[0].Name)
	assert.Nil(t, resp.Result.Caption)
}

func TestParseOpenAISchemaLegacyUpgrade(t *testing.T) {
	legacy := `{
		"labels": [{
			"name": "",
			"confidence": 0,
			"topicality": 0
		}]
	}`

	raw, err := parseOpenAISchema(legacy)
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(raw, &decoded))

	assert.Equal(t, "object", decoded["type"])

	props, ok := decoded["properties"].(map[string]any)
	require.True(t, ok)
	labels, ok := props["labels"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "array", labels["type"])
}

func TestParseOpenAISchemaLegacyUpgradeNSFW(t *testing.T) {
	legacy := `{
		"labels": [{
			"name": "",
			"confidence": 0,
			"topicality": 0,
			"nsfw": false,
			"nsfw_confidence": 0
		}]
	}`

	raw, err := parseOpenAISchema(legacy)
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(raw, &decoded))

	props := decoded["properties"].(map[string]any)
	labels := props["labels"].(map[string]any)
	items := labels["items"].(map[string]any)
	_, hasNSFW := items["properties"].(map[string]any)["nsfw"]
	assert.True(t, hasNSFW)
}

func TestPerformApiRequestOpenAISuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqPayload struct {
			Model string `json:"model"`
		}
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&reqPayload))
		assert.Equal(t, "gpt-5-mini", reqPayload.Model)

		response := map[string]any{
			"id":    "resp_123",
			"model": "gpt-5-mini",
			"output": []any{
				map[string]any{
					"role": "assistant",
					"content": []any{
						map[string]any{
							"type": "output_json",
							"json": map[string]any{
								"caption": map[string]any{
									"text":       "A cat rests on a windowsill.",
									"confidence": 0.91,
								},
								"labels": []map[string]any{
									{
										"name":       "cat",
										"confidence": 0.92,
										"topicality": 0.88,
									},
								},
							},
						},
					},
				},
			},
		}

		assert.NoError(t, json.NewEncoder(w).Encode(response))
	}))
	defer server.Close()

	req := &ApiRequest{
		Id:             "test",
		Model:          "gpt-5-mini",
		Images:         []string{"data:image/jpeg;base64,AA=="},
		ResponseFormat: ApiFormatOpenAI,
		Options: &ApiRequestOptions{
			Detail: openai.DefaultDetail,
		},
		Schema: json.RawMessage(`{"type":"object"}`),
	}

	resp, err := PerformApiRequest(req, server.URL, http.MethodPost, "secret")
	require.NoError(t, err)
	require.NotNil(t, resp)

	require.NotNil(t, resp.Result.Caption)
	assert.Equal(t, entity.SrcOpenAI, resp.Result.Caption.Source)
	assert.Equal(t, "A cat rests on a windowsill.", resp.Result.Caption.Text)

	require.Len(t, resp.Result.Labels, 1)
	assert.Equal(t, entity.SrcOpenAI, resp.Result.Labels[0].Source)
	assert.Equal(t, "Cat", resp.Result.Labels[0].Name)
}

func TestPerformApiRequestOpenAITextFallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]any{
			"id":    "resp_456",
			"model": "gpt-5-mini",
			"output": []any{
				map[string]any{
					"role": "assistant",
					"content": []any{
						map[string]any{
							"type": "output_text",
							"text": "Two hikers reach the summit at sunset.",
						},
					},
				},
			},
		}
		assert.NoError(t, json.NewEncoder(w).Encode(response))
	}))
	defer server.Close()

	req := &ApiRequest{
		Id:             "fallback",
		Model:          "gpt-5-mini",
		Images:         []string{"data:image/jpeg;base64,AA=="},
		ResponseFormat: ApiFormatOpenAI,
		Options: &ApiRequestOptions{
			Detail: openai.DefaultDetail,
		},
		Schema: nil,
	}

	resp, err := PerformApiRequest(req, server.URL, http.MethodPost, "")
	require.NoError(t, err)
	require.NotNil(t, resp.Result.Caption)
	assert.Equal(t, "Two hikers reach the summit at sunset.", resp.Result.Caption.Text)
	assert.Equal(t, entity.SrcOpenAI, resp.Result.Caption.Source)
}

func TestPerformApiRequestOpenAIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"message": "Invalid image payload",
			},
		})
	}))
	defer server.Close()

	req := &ApiRequest{
		Id:             "error",
		Model:          "gpt-5-mini",
		ResponseFormat: ApiFormatOpenAI,
		Schema:         nil,
		Images:         []string{"data:image/jpeg;base64,AA=="},
	}

	_, err := PerformApiRequest(req, server.URL, http.MethodPost, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid image payload")
}
