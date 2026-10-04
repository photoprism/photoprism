package vision

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/photoprism/photoprism/internal/ai/vision/openai"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/http/scheme"
	"github.com/photoprism/photoprism/pkg/media"
)

// openaiDefaults provides canned prompts, schema templates, and options for OpenAI engines.
type openaiDefaults struct{}

// openaiBuilder prepares ApiRequest objects for OpenAI's Responses API.
type openaiBuilder struct{}

// openaiParser converts Responses API payloads into ApiResponse instances.
type openaiParser struct{}

func init() {
	RegisterEngine(ApiFormatOpenAI, Engine{
		Builder:  openaiBuilder{},
		Parser:   openaiParser{},
		Defaults: openaiDefaults{},
	})

	registerOpenAIEngineDefaults()
}

// registerOpenAIEngineDefaults registers the OpenAI engine alias with the base URL and default model
// from OPENAI_BASE_URL and OPENAI_MODEL, if set.
func registerOpenAIEngineDefaults() {
	ensureEnv()

	defaultModel := openai.DefaultModel

	if envDefault := envModel(openai.ModelEnv); envDefault != "" {
		defaultModel = envDefault
	}

	RegisterEngineAlias(openai.EngineName, EngineInfo{
		Uri:               openai.DefaultUri,
		RequestFormat:     ApiFormatOpenAI,
		ResponseFormat:    ApiFormatOpenAI,
		FileScheme:        scheme.Data,
		DefaultModel:      defaultModel,
		DefaultResolution: openai.DefaultResolution,
		DefaultKey:        openai.APIKeyPlaceholder,
	})
}

// logOpenAIBaseUrl writes the OpenAI base URL to the system log if it is not the default, as it receives the
// requests and keys of models without their own service URI.
func logOpenAIBaseUrl() {
	baseUrl := os.Getenv(openai.BaseUrlEnv)

	if baseUrl == "" || baseUrl == openai.DefaultBaseUrl {
		return
	}

	if redacted := clean.UriRedacted(baseUrl); redacted != "" {
		event.SystemInfo([]string{"vision", "openai engine uses base url %s"}, redacted)
	} else {
		event.SystemWarn([]string{"vision", "openai engine uses an invalid base url"})
	}
}

// usesOpenAIDefaultUri reports whether the model sends its requests to the base URL of the OpenAI engine.
func (m *Model) usesOpenAIDefaultUri() bool {
	return m != nil && !m.Disabled && !m.Service.Disabled && m.Engine == openai.EngineName &&
		strings.TrimSpace(m.Service.Uri) == openai.DefaultUri
}

// openaiDefaultModel returns the registered default model of the OpenAI engine.
func openaiDefaultModel() string {
	if info, ok := EngineInfoFor(openai.EngineName); ok && info.DefaultModel != "" {
		return info.DefaultModel
	}

	return openai.DefaultModel
}

// SystemPrompt returns the default OpenAI system prompt for the specified model type.
func (openaiDefaults) SystemPrompt(model *Model) string {
	if model == nil {
		return ""
	}

	switch model.Type {
	case ModelTypeCaption:
		return openai.CaptionSystem
	case ModelTypeLabels:
		return openai.LabelSystem
	default:
		return ""
	}
}

// UserPrompt returns the default OpenAI user prompt for the specified model type.
func (openaiDefaults) UserPrompt(model *Model) string {
	if model == nil {
		return ""
	}

	switch model.Type {
	case ModelTypeCaption:
		return openai.CaptionPrompt
	case ModelTypeLabels:
		if DetectNSFWLabels {
			return openai.LabelPromptNSFW
		}
		return openai.LabelPromptDefault
	default:
		return ""
	}
}

// SchemaTemplate returns the JSON schema template for the model, if applicable.
func (openaiDefaults) SchemaTemplate(model *Model) string {
	if model == nil {
		return ""
	}

	switch model.Type {
	case ModelTypeLabels:
		return string(openai.SchemaLabels(model.PromptContains("nsfw")))
	default:
		return ""
	}
}

// Options returns default OpenAI request options for the model.
func (openaiDefaults) Options(model *Model) *ModelOptions {
	if model == nil {
		return nil
	}

	switch model.Type {
	case ModelTypeCaption:
		return &ModelOptions{
			Detail:          openai.DefaultDetail,
			MaxOutputTokens: openai.CaptionMaxTokens,
		}
	case ModelTypeLabels:
		return &ModelOptions{
			Detail:          openai.DefaultDetail,
			MaxOutputTokens: openai.LabelsMaxTokens,
			ForceJson:       true,
		}
	default:
		return nil
	}
}

// Build constructs an OpenAI request payload using base64-encoded thumbnails.
func (openaiBuilder) Build(ctx context.Context, model *Model, files Files, mediaSrc media.Src) (*ApiRequest, error) {
	if model == nil {
		return nil, ErrInvalidModel
	}

	dataReq, err := NewApiRequestImages(files, scheme.Data, mediaSrc)
	if err != nil {
		return nil, err
	}

	req := &ApiRequest{
		Id:             dataReq.Id,
		Images:         append(Files(nil), dataReq.Images...),
		ResponseFormat: ApiFormatOpenAI,
	}

	if opts := model.GetOptions(); opts != nil {
		req.Options = cloneOptions(opts)

		switch model.Type {
		case ModelTypeCaption:
			// Captions default to plain text responses; structured JSON is optional.
			req.Options.ForceJson = false
			if req.Options.MaxOutputTokens < openai.CaptionMaxTokens {
				req.Options.MaxOutputTokens = openai.CaptionMaxTokens
			}
		case ModelTypeLabels:
			if req.Options.MaxOutputTokens < openai.LabelsMaxTokens {
				req.Options.MaxOutputTokens = openai.LabelsMaxTokens
			}
		}

		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(model.Name)), "gpt-5") {
			req.Options.Temperature = 0
			req.Options.TopP = 0
		}
	}

	if schema := strings.TrimSpace(model.SchemaTemplate()); schema != "" {
		if raw, parseErr := parseOpenAISchema(schema); parseErr != nil {
			log.Warnf("vision: failed to parse OpenAI schema template (%s)", clean.Error(parseErr))
		} else {
			req.Schema = raw
		}
	}

	return req, nil
}

// Parse converts an OpenAI Responses API payload into the internal ApiResponse representation.
func (openaiParser) Parse(ctx context.Context, req *ApiRequest, raw []byte, status int) (*ApiResponse, error) {
	// Return an error for a failed request, as its response text is not a result.
	if err := serviceStatusError(openai.EngineName, ApiFormatOpenAI, openaiRequestModel(req), status); err != nil {
		return nil, err
	}

	var resp openai.Response
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}

	if resp.Error != nil && resp.Error.Message != "" {
		return nil, errors.New(resp.Error.Message)
	}

	normalize := req.GetNormalize()

	result := ApiResult{}
	if jsonPayload := resp.FirstJSON(); len(jsonPayload) > 0 {
		if err := populateOpenAIJSONResult(&result, jsonPayload, normalize); err != nil {
			log.Debugf("vision: %s (parse openai json payload)", clean.Error(err))
		}
	}

	if result.Caption == nil {
		if text := resp.FirstText(); text != "" {
			trimmed := strings.TrimSpace(text)
			var parsedJSON bool

			if len(trimmed) > 0 && (trimmed[0] == '{' || trimmed[0] == '[') {
				if err := populateOpenAIJSONResult(&result, json.RawMessage(trimmed), normalize); err != nil {
					log.Debugf("vision: %s (parse openai json text payload)", clean.Error(err))
				} else {
					parsedJSON = true
				}
			}

			if !parsedJSON && trimmed != "" {
				result.Caption = &CaptionResult{
					Text:   trimmed,
					Source: entity.SrcOpenAI,
				}
			}
		}
	}

	var responseID string
	if req != nil {
		responseID = req.GetId()
	}

	modelName := strings.TrimSpace(resp.Model)
	if modelName == "" && req != nil {
		modelName = strings.TrimSpace(req.Model)
	}

	return &ApiResponse{
		Id:     responseID,
		Code:   status,
		Model:  &Model{Name: modelName},
		Result: result,
	}, nil
}

// parseOpenAISchema validates the provided JSON schema and returns it as a raw message.
func parseOpenAISchema(schema string) (json.RawMessage, error) {
	var raw json.RawMessage
	if err := json.Unmarshal([]byte(schema), &raw); err != nil {
		return nil, err
	}
	return normalizeOpenAISchema(raw)
}

// normalizeOpenAISchema upgrades legacy label schema definitions so they comply with
// OpenAI's json_schema format requirements.
func normalizeOpenAISchema(raw json.RawMessage) (json.RawMessage, error) {
	if len(raw) == 0 {
		return raw, nil
	}

	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		// Fallback to the original payload if it isn't a JSON object.
		return raw, nil
	}

	if t, ok := doc["type"]; ok {
		if typeStr, ok := t.(string); ok && strings.TrimSpace(typeStr) != "" {
			return raw, nil
		}
	}

	if _, ok := doc["properties"]; ok {
		return raw, nil
	}

	labels, ok := doc["labels"]
	if !ok {
		return raw, nil
	}

	nsfw := false

	if items, ok := labels.([]any); ok && len(items) > 0 {
		if first, ok := items[0].(map[string]any); ok {
			if _, hasNSFW := first["nsfw"]; hasNSFW {
				nsfw = true
			}
			if _, hasNSFWConfidence := first["nsfw_confidence"]; hasNSFWConfidence {
				nsfw = true
			}
		}
	}

	return openai.SchemaLabels(nsfw), nil
}

// populateOpenAIJSONResult unmarshals a structured OpenAI response into ApiResult fields.
func populateOpenAIJSONResult(result *ApiResult, payload json.RawMessage, normalize NormalizeType) error {
	if result == nil || len(payload) == 0 {
		return nil
	}

	var envelope struct {
		Caption *struct {
			Text       string  `json:"text"`
			Confidence float32 `json:"confidence"`
		} `json:"caption"`
		Labels []LabelResult `json:"labels"`
	}

	if err := json.Unmarshal(payload, &envelope); err != nil {
		return err
	}

	if envelope.Caption != nil {
		text := strings.TrimSpace(envelope.Caption.Text)
		if text != "" {
			result.Caption = &CaptionResult{
				Text:       text,
				Confidence: envelope.Caption.Confidence,
				Source:     entity.SrcOpenAI,
			}
		}
	}

	if len(envelope.Labels) > 0 {
		filtered := envelope.Labels[:0]

		for i := range envelope.Labels {
			if envelope.Labels[i].Source == "" {
				envelope.Labels[i].Source = entity.SrcOpenAI
			}

			normalizeLabelResult(&envelope.Labels[i], normalize)

			if envelope.Labels[i].Name == "" {
				continue
			}

			filtered = append(filtered, envelope.Labels[i])
		}

		result.Labels = append(result.Labels, filtered...)
	}

	return nil
}

// openaiRequestModel returns the model a request was sent for, which is the default model if none is set.
func openaiRequestModel(req *ApiRequest) string {
	if req != nil {
		if model := strings.TrimSpace(req.Model); model != "" {
			return model
		}
	}

	return openaiDefaultModel()
}
