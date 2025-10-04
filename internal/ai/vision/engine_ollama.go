package vision

import (
	"context"
	"fmt"
	"strings"

	"github.com/photoprism/photoprism/internal/ai/vision/ollama"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/service/http/scheme"
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

	// Register the human-friendly engine name so configuration can simply use
	// `Engine: "ollama"` and inherit adapter defaults.
	RegisterEngineAlias(ollama.EngineName, EngineInfo{
		RequestFormat:     ApiFormatOllama,
		ResponseFormat:    ApiFormatOllama,
		FileScheme:        string(scheme.Base64),
		DefaultResolution: ollama.DefaultResolution,
	})

	CaptionModel.Engine = ollama.EngineName
	CaptionModel.ApplyEngineDefaults()
}

func (ollamaDefaults) SystemPrompt(model *Model) string {
	if model == nil || model.Type != ModelTypeLabels {
		return ""
	}
	return ollama.LabelSystem
}

func (ollamaDefaults) UserPrompt(model *Model) string {
	if model == nil {
		return ""
	}

	switch model.Type {
	case ModelTypeCaption:
		return ollama.CaptionPrompt
	case ModelTypeLabels:
		return ollama.LabelPrompt
	default:
		return ""
	}
}

func (ollamaDefaults) SchemaTemplate(model *Model) string {
	if model == nil {
		return ""
	}

	switch model.Type {
	case ModelTypeLabels:
		return ollama.LabelsSchema()
	}

	return ""
}

func (ollamaDefaults) Options(model *Model) *ApiRequestOptions {
	if model == nil {
		return nil
	}

	switch model.Type {
	case ModelTypeLabels:
		return &ApiRequestOptions{
			Temperature: DefaultTemperature,
			TopP:        0.9,
			Stop:        []string{"\n\n"},
		}
	case ModelTypeCaption:
		return &ApiRequestOptions{
			Temperature: DefaultTemperature,
		}
	default:
		return nil
	}
}

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
		req.Model, _, _ = model.Model()
	} else {
		_, req.Model, req.Version = model.Model()
	}

	return req, nil
}

func (ollamaParser) Parse(ctx context.Context, req *ApiRequest, raw []byte, status int) (*ApiResponse, error) {
	ollamaResp, err := decodeOllamaResponse(raw)
	if err != nil {
		return nil, err
	}

	result := &ApiResponse{
		Id:    req.GetId(),
		Code:  status,
		Model: &Model{Name: ollamaResp.Model},
		Result: ApiResult{
			Labels: append([]LabelResult{}, ollamaResp.Result.Labels...),
			Caption: func() *CaptionResult {
				if ollamaResp.Result.Caption != nil {
					copyCaption := *ollamaResp.Result.Caption
					return &copyCaption
				}
				return nil
			}(),
		},
	}

	parsedLabels := len(result.Result.Labels) > 0

	if !parsedLabels && strings.TrimSpace(ollamaResp.Response) != "" && req.Format == FormatJSON {
		if labels, parseErr := parseOllamaLabels(ollamaResp.Response); parseErr != nil {
			log.Debugf("vision: %s (parse ollama labels)", clean.Error(parseErr))
		} else if len(labels) > 0 {
			result.Result.Labels = append(result.Result.Labels, labels...)
			parsedLabels = true
		}
	}

	if parsedLabels {
		filtered := result.Result.Labels[:0]
		for i := range result.Result.Labels {
			if result.Result.Labels[i].Confidence <= 0 {
				result.Result.Labels[i].Confidence = ollama.DefaultLabelConfidence
			}
			if result.Result.Labels[i].Topicality <= 0 {
				result.Result.Labels[i].Topicality = result.Result.Labels[i].Confidence
			}
			normalizeLabelResult(&result.Result.Labels[i])
			if result.Result.Labels[i].Name == "" {
				continue
			}
			if result.Result.Labels[i].Source == "" {
				result.Result.Labels[i].Source = entity.SrcOllama
			}
			filtered = append(filtered, result.Result.Labels[i])
		}
		result.Result.Labels = filtered
	} else {
		if caption := strings.TrimSpace(ollamaResp.Response); caption != "" {
			result.Result.Caption = &CaptionResult{
				Text:   caption,
				Source: entity.SrcOllama,
			}
		}
	}

	return result, nil
}

var ErrInvalidModel = fmt.Errorf("vision: invalid model")
