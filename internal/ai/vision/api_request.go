package vision

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"slices"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/ai/vision/openai"
	"github.com/photoprism/photoprism/internal/ai/vision/schema"
	"github.com/photoprism/photoprism/internal/api/download"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/scheme"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Files holds a list of input file paths or URLs for vision requests.
type Files = []string

const (
	// FormatJSON indicates JSON payloads.
	FormatJSON = "json"

	logDataPreviewLength   = 16
	logDataTruncatedSuffix = "... (truncated)"
)

// ApiRequestContext represents a context parameter returned from a previous request.
type ApiRequestContext = []int

// ApiRequest represents a Vision API service request.
type ApiRequest struct {
	Id             string             `form:"id" yaml:"Id,omitempty" json:"id,omitempty"`
	Model          string             `form:"model" yaml:"Model,omitempty" json:"model,omitempty"`
	Version        string             `form:"version" yaml:"Version,omitempty" json:"version,omitempty"`
	System         string             `form:"system" yaml:"System,omitempty" json:"system,omitempty"`
	Prompt         string             `form:"prompt" yaml:"Prompt,omitempty" json:"prompt,omitempty"`
	Suffix         string             `form:"suffix" yaml:"Suffix,omitempty" json:"suffix"`
	Format         string             `form:"format" yaml:"Format,omitempty" json:"format,omitempty"`
	Url            string             `form:"url" yaml:"Url,omitempty" json:"url,omitempty"`
	Org            string             `form:"org" yaml:"Org,omitempty" json:"org,omitempty"`
	Project        string             `form:"project" yaml:"Project,omitempty" json:"project,omitempty"`
	Think          string             `form:"think" yaml:"Think,omitempty" json:"think,omitempty"`
	Options        *ModelOptions      `form:"options" yaml:"Options,omitempty" json:"options,omitempty"`
	Context        *ApiRequestContext `form:"context" yaml:"Context,omitempty" json:"context,omitempty"`
	Stream         bool               `form:"stream" yaml:"Stream,omitempty" json:"stream"`
	Images         Files              `form:"images" yaml:"Images,omitempty" json:"images,omitempty"`
	Schema         json.RawMessage    `form:"schema" yaml:"Schema,omitempty" json:"schema,omitempty"`
	ResponseFormat ApiFormat          `form:"-" yaml:"-" json:"-"`
}

// NewApiRequest returns a new service API request with the specified format and payload.
func NewApiRequest(requestFormat ApiFormat, files Files, fileScheme scheme.Type) (result *ApiRequest, err error) {
	if len(files) == 0 {
		return result, errors.New("missing files")
	}

	switch requestFormat {
	case ApiFormatUrl:
		return NewApiRequestUrl(files[0], fileScheme)
	case ApiFormatImages, ApiFormatVision:
		return NewApiRequestImages(files, fileScheme)
	case ApiFormatOllama:
		return NewApiRequestOllama(files, fileScheme)
	default:
		return result, errors.New("invalid request format")
	}
}

// NewApiRequestUrl returns a new Vision API request with the specified image Url as payload.
func NewApiRequestUrl(fileName string, fileScheme scheme.Type) (result *ApiRequest, err error) {
	var imgUrl string

	switch fileScheme {
	case scheme.Https:
		// Return if no thumbnail filenames were given.
		if !fs.FileExistsNotEmpty(fileName) {
			return result, errors.New("invalid image file name")
		}

		// Generate a random token for the remote service to download the file.
		fileUuid := rnd.UUID()

		if err = download.Register(fileUuid, fileName); err != nil {
			return result, fmt.Errorf("%s (create download url)", err)
		}

		imgUrl = fmt.Sprintf("%s/%s", DownloadUrl, fileUuid)
	case scheme.Data:
		var u *url.URL
		if u, err = url.Parse(fileName); err != nil {
			return result, fmt.Errorf("%s (invalid image url)", err)
		} else if !slices.Contains(scheme.HttpsHttp, u.Scheme) {
			return nil, fmt.Errorf("unsupported image url scheme %s", clean.Log(u.Scheme))
		} else {
			imgUrl = u.String()
		}
	default:
		return nil, fmt.Errorf("unsupported file scheme %s", clean.Log(fileScheme))
	}

	return &ApiRequest{
		Id:             rnd.UUID(),
		Model:          "",
		Url:            imgUrl,
		ResponseFormat: ApiFormatVision,
	}, nil
}

// NewApiRequestImages returns a new Vision API request with the specified images as payload.
func NewApiRequestImages(images Files, fileScheme scheme.Type) (*ApiRequest, error) {
	imageUrls := make(Files, len(images))

	if fileScheme == scheme.Https && !strings.HasPrefix(DownloadUrl, "https://") {
		log.Tracef("vision: file request scheme changed from https to data because https is not configured")
		fileScheme = scheme.Data
	}

	for i := range images {
		switch fileScheme {
		case scheme.Https:
			fileUuid := rnd.UUID()
			if err := download.Register(fileUuid, images[i]); err != nil {
				return nil, fmt.Errorf("%s (create download url)", err)
			} else {
				imageUrls[i] = fmt.Sprintf("%s/%s", DownloadUrl, fileUuid)
			}
		case scheme.Data:
			file, err := os.Open(images[i])
			if err != nil {
				return nil, fmt.Errorf("%s (create data url)", err)
			}
			imageUrls[i] = media.DataUrl(file)
			if err := file.Close(); err != nil {
				return nil, fmt.Errorf("%s (close data url)", err)
			}
		default:
			return nil, fmt.Errorf("unsupported file scheme %s", clean.Log(fileScheme))
		}
	}

	return &ApiRequest{
		Id:             rnd.UUID(),
		Model:          "",
		Images:         imageUrls,
		ResponseFormat: ApiFormatVision,
	}, nil
}

// GetId returns the request ID string and generates a random ID if none was set.
func (r *ApiRequest) GetId() string {
	if r.Id == "" {
		r.Id = rnd.UUID()
	}

	return r.Id
}

// GetResponseFormat returns the expected response format type.
func (r *ApiRequest) GetResponseFormat() ApiFormat {
	if r.ResponseFormat == "" {
		return ApiFormatVision
	}

	return r.ResponseFormat
}

// JSON returns the request data as JSON-encoded bytes.
func (r *ApiRequest) JSON() ([]byte, error) {
	if r == nil {
		return nil, errors.New("api request is nil")
	}

	if r.ResponseFormat == ApiFormatOpenAI {
		return r.openAIJSON()
	}

	data, err := json.Marshal(*r)

	if err != nil {
		return nil, err
	}

	// Normalize "true"/"false" to JSON booleans so Ollama accepts think values
	// configured as strings while still supporting string levels like "low".
	normalizedThink, hasThink := normalizeThinkValue(r.Think)
	if !hasThink {
		return data, nil
	}

	var payload map[string]any

	if err = json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}

	payload["think"] = normalizedThink

	return json.Marshal(payload)
}

// WriteLog logs the request data when trace log mode is enabled.
func (r *ApiRequest) WriteLog() {
	if !log.IsLevelEnabled(logrus.TraceLevel) {
		return
	}

	sanitized := r.sanitizedForLog()

	if data, _ := json.Marshal(sanitized); len(data) > 0 {
		log.Tracef("vision: %s", data)
	}
}

// sanitizedForLog returns a shallow copy of the request with large base64 payloads shortened.
func (r *ApiRequest) sanitizedForLog() ApiRequest {
	if r == nil {
		return ApiRequest{}
	}

	sanitized := *r

	if len(r.Images) > 0 {
		sanitized.Images = make(Files, len(r.Images))

		for i := range r.Images {
			sanitized.Images[i] = sanitizeLogPayload(r.Images[i])
		}
	}

	sanitized.Url = sanitizeLogPayload(r.Url)

	sanitized.Schema = r.Schema

	return sanitized
}

// sanitizeLogPayload shortens base64-encoded data so trace logs remain readable.
func sanitizeLogPayload(value string) string {
	if value == "" {
		return value
	}

	if strings.HasPrefix(value, "data:") {
		if prefix, encoded, found := strings.Cut(value, ","); found {
			sanitized := truncateBase64ForLog(encoded)

			if sanitized != encoded {
				return prefix + "," + sanitized
			}
		}

		return value
	}

	if isLikelyBase64(value) {
		return truncateBase64ForLog(value)
	}

	return value
}

func truncateBase64ForLog(value string) string {
	if len(value) <= logDataPreviewLength {
		return value
	}

	return value[:logDataPreviewLength] + logDataTruncatedSuffix
}

func isLikelyBase64(value string) bool {
	if len(value) < logDataPreviewLength {
		return false
	}

	for i := 0; i < len(value); i++ {
		c := value[i]

		switch {
		case c >= 'A' && c <= 'Z':
		case c >= 'a' && c <= 'z':
		case c >= '0' && c <= '9':
		case c == '+', c == '/', c == '=', c == '-', c == '_':
		case c == '\n' || c == '\r':
			continue
		default:
			return false
		}
	}

	return true
}

// normalizeThinkValue returns the serialized think value and whether it should
// be included in the outgoing payload.
func normalizeThinkValue(value string) (any, bool) {
	value = strings.TrimSpace(value)

	if value == "" {
		return nil, false
	}

	switch strings.ToLower(value) {
	case "true":
		return true, true
	case "false":
		return false, true
	default:
		return value, true
	}
}

// openAIJSON converts the request data into an OpenAI Responses API payload.
func (r *ApiRequest) openAIJSON() ([]byte, error) {
	detail := openai.DefaultDetail

	if opts := r.Options; opts != nil && strings.TrimSpace(opts.Detail) != "" {
		detail = strings.TrimSpace(opts.Detail)
	}

	messages := make([]openai.InputMessage, 0, 2)

	if system := strings.TrimSpace(r.System); system != "" {
		messages = append(messages, openai.InputMessage{
			Role: "system",
			Type: "message",
			Content: []openai.ContentItem{
				{
					Type: openai.ContentTypeText,
					Text: system,
				},
			},
		})
	}

	userContent := make([]openai.ContentItem, 0, len(r.Images)+1)

	if prompt := strings.TrimSpace(r.Prompt); prompt != "" {
		userContent = append(userContent, openai.ContentItem{
			Type: openai.ContentTypeText,
			Text: prompt,
		})
	}

	for _, img := range r.Images {
		if img == "" {
			continue
		}

		userContent = append(userContent, openai.ContentItem{
			Type:     openai.ContentTypeImage,
			ImageURL: img,
			Detail:   detail,
		})
	}

	if len(userContent) > 0 {
		messages = append(messages, openai.InputMessage{
			Role:    "user",
			Type:    "message",
			Content: userContent,
		})
	}

	if len(messages) == 0 {
		return nil, errors.New("openai request requires at least one message")
	}

	payload := openai.HTTPRequest{
		Model: strings.TrimSpace(r.Model),
		Input: messages,
	}

	if payload.Model == "" {
		payload.Model = openai.DefaultModel
	}

	if strings.HasPrefix(strings.ToLower(payload.Model), "gpt-5") {
		payload.Reasoning = &openai.Reasoning{Effort: "low"}
	}

	if opts := r.Options; opts != nil {
		if opts.MaxOutputTokens > 0 {
			payload.MaxOutputTokens = opts.MaxOutputTokens
		}

		if opts.Temperature > 0 {
			payload.Temperature = opts.Temperature
		}

		if opts.TopP > 0 {
			payload.TopP = opts.TopP
		}

		if opts.PresencePenalty != 0 {
			payload.PresencePenalty = opts.PresencePenalty
		}

		if opts.FrequencyPenalty != 0 {
			payload.FrequencyPenalty = opts.FrequencyPenalty
		}
	}

	if format := buildOpenAIResponseFormat(r); format != nil {
		payload.Text = &openai.TextOptions{
			Format: format,
		}
	}

	return json.Marshal(payload)
}

// buildOpenAIResponseFormat determines which response_format to send to OpenAI.
func buildOpenAIResponseFormat(r *ApiRequest) *openai.ResponseFormat {
	if r == nil {
		return nil
	}

	opts := r.Options
	hasSchema := len(r.Schema) > 0

	if !hasSchema && (opts == nil || !opts.ForceJson) {
		return nil
	}

	result := &openai.ResponseFormat{}

	if hasSchema {
		result.Type = openai.ResponseFormatJSONSchema
		result.Schema = r.Schema

		if opts != nil && strings.TrimSpace(opts.SchemaVersion) != "" {
			result.Name = strings.TrimSpace(opts.SchemaVersion)
		} else {
			result.Name = schema.JsonSchemaName(r.Schema, openai.DefaultSchemaVersion)
		}
	} else {
		result.Type = openai.ResponseFormatJSONObject
	}

	return result
}
