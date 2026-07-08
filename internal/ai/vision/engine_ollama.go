package vision

import (
	"context"
	"os"
	"strings"

	"github.com/photoprism/photoprism/internal/ai/vision/ollama"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/http/scheme"
)

type ollamaDefaults struct{}

type ollamaBuilder struct{}

type ollamaParser struct{}

func init() {
	RegisterEngine(ApiFormatOllama, Engine{
		Builder:  ollamaBuilder{},
		Parser:   ollamaParser{},
		Defaults: ollamaDefaults{},
	})

	registerOllamaEngineDefaults()
}

// registerOllamaEngineDefaults selects the default Ollama endpoint based on the
// configured base URL and registers the engine alias accordingly. When
// OLLAMA_BASE_URL points at the cloud host we only switch the default model to
// the cloud preset; the actual base URL continues to come from
// OLLAMA_BASE_URL (or falls back to the local compose default) so we don't
// accidentally talk to the hosted service without an explicit endpoint.
func registerOllamaEngineDefaults() {
	ensureEnv()

	defaultModel := ollama.DefaultModel

	// Use different default model for the Ollama cloud service.
	if baseUrl := os.Getenv(ollama.BaseUrlEnv); baseUrl == ollama.CloudBaseUrl {
		defaultModel = ollama.CloudModel
	}

	// Register the human-friendly engine name so configuration can simply use
	// `Engine: "ollama"` and inherit adapter defaults.
	RegisterEngineAlias(ollama.EngineName, EngineInfo{
		Uri:               ollama.DefaultUri,
		RequestFormat:     ApiFormatOllama,
		ResponseFormat:    ApiFormatOllama,
		FileScheme:        scheme.Base64,
		DefaultModel:      defaultModel,
		DefaultResolution: ollama.DefaultResolution,
		DefaultKey:        ollama.APIKeyPlaceholder,
	})

	// Keep the default caption model config aligned with the defaults.
	CaptionModel.ApplyEngineDefaults()
}

// SystemPrompt returns the Ollama system prompt for the specified model type.
func (ollamaDefaults) SystemPrompt(model *Model) string {
	if model == nil || model.Type != ModelTypeLabels {
		return ""
	}
	return ollama.LabelSystem
}

// UserPrompt returns the Ollama user prompt for the specified model type.
func (ollamaDefaults) UserPrompt(model *Model) string {
	if model == nil {
		return ""
	}

	switch model.Type {
	case ModelTypeCaption:
		return ollama.CaptionPrompt
	case ModelTypeLabels:
		if DetectNSFWLabels {
			return ollama.LabelPromptNSFW
		} else {
			return ollama.LabelPromptDefault
		}
	default:
		return ""
	}
}

// SchemaTemplate returns the Ollama JSON schema template.
func (ollamaDefaults) SchemaTemplate(model *Model) string {
	if model == nil {
		return ""
	}

	if model.Type == ModelTypeLabels {
		return ollama.SchemaLabels(model.PromptContains("nsfw"))
	}

	return ""
}

// Options returns the Ollama service request options.
func (ollamaDefaults) Options(model *Model) *ModelOptions {
	if model == nil {
		return nil
	}

	switch model.Type {
	case ModelTypeLabels:
		return &ModelOptions{
			Temperature: DefaultTemperature,
			TopP:        0.9,
			Stop:        []string{"\n\n"},
		}
	case ModelTypeCaption:
		return &ModelOptions{
			Temperature: DefaultTemperature,
		}
	default:
		return nil
	}
}

// Build builds the Ollama service request.
func (ollamaBuilder) Build(ctx context.Context, model *Model, files Files) (*ApiRequest, error) {
	if model == nil {
		return nil, ErrInvalidModel
	}

	req, err := NewApiRequest(model.EndpointRequestFormat(), files, model.EndpointFileScheme())
	if err != nil {
		return nil, err
	}

	req.ResponseFormat = model.EndpointResponseFormat()

	if format := model.GetFormat(); format != "" {
		req.Format = format
	}

	if model.Service.RequestFormat == ApiFormatOllama {
		req.Model, _, _ = model.GetModel()
	} else {
		_, req.Model, req.Version = model.GetModel()
	}

	return req, nil
}

// Parse processes the Ollama service response.
func (ollamaParser) Parse(ctx context.Context, req *ApiRequest, raw []byte, status int) (*ApiResponse, error) {
	ollamaResp, err := decodeOllamaResponse(raw)

	if err != nil {
		return nil, err
	}

	response := &ApiResponse{
		Id:    req.GetId(),
		Code:  status,
		Model: &Model{Name: ollamaResp.Model},
		Result: ApiResult{
			Labels:  convertOllamaLabels(ollamaResp.Result.Labels),
			Caption: convertOllamaCaption(ollamaResp.Result.Caption),
		},
	}

	parsedLabels := len(response.Result.Labels) > 0

	// Qwen3-VL models stream their JSON payload in the "Thinking" field.
	fallbackResponse := strings.TrimSpace(ollamaResp.Response)
	fallbackThinking := strings.TrimSpace(ollamaResp.Thinking)

	fallbackJSON := fallbackResponse
	if fallbackJSON == "" {
		fallbackJSON = fallbackThinking
	}

	fallbackCaption := fallbackResponse
	if fallbackCaption == "" {
		fallbackCaption = fallbackThinking
	}

	if !parsedLabels && fallbackJSON != "" && (req.Format == FormatJSON || strings.HasPrefix(fallbackJSON, "{")) {
		if labels, parseErr := parseOllamaLabels(fallbackJSON); parseErr != nil {
			log.Warnf("vision: %s (parse ollama labels)", clean.Error(parseErr))
		} else if len(labels) > 0 {
			response.Result.Labels = append(response.Result.Labels, labels...)
			parsedLabels = true
		}
	}

	if parsedLabels {
		filtered := response.Result.Labels[:0]
		for i := range response.Result.Labels {
			if response.Result.Labels[i].Confidence <= 0 {
				response.Result.Labels[i].Confidence = ollama.LabelConfidenceDefault
			}

			if response.Result.Labels[i].Topicality <= 0 {
				response.Result.Labels[i].Topicality = response.Result.Labels[i].Confidence
			}

			// Apply thresholds and canonicalize the name.
			normalizeLabelResult(&response.Result.Labels[i])

			if response.Result.Labels[i].Name == "" {
				continue
			}

			if response.Result.Labels[i].Source == "" {
				response.Result.Labels[i].Source = entity.SrcOllama
			}

			filtered = append(filtered, response.Result.Labels[i])
		}
		response.Result.Labels = filtered
	} else if fallbackCaption != "" {
		response.Result.Caption = &CaptionResult{
			Text:   fallbackCaption,
			Source: entity.SrcOllama,
		}
	}

	return response, nil
}

func convertOllamaLabels(payload []ollama.LabelPayload) []LabelResult {
	if len(payload) == 0 {
		return nil
	}

	labels := make([]LabelResult, len(payload))

	for i := range payload {
		labels[i] = LabelResult{
			Name:           payload[i].Name,
			Source:         payload[i].Source,
			Priority:       payload[i].Priority,
			Confidence:     payload[i].Confidence,
			Topicality:     payload[i].Topicality,
			Categories:     payload[i].Categories,
			NSFW:           payload[i].NSFW,
			NSFWConfidence: payload[i].NSFWConfidence,
		}
	}

	return labels
}

func convertOllamaCaption(payload *ollama.CaptionPayload) *CaptionResult {
	if payload == nil {
		return nil
	}

	return &CaptionResult{
		Text:       payload.Text,
		Source:     payload.Source,
		Confidence: payload.Confidence,
	}
}
