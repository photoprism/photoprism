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

// ApiRequestOptions represents additional model parameters listed in the documentation.
type ApiRequestOptions struct {
	NumKeep          int      `yaml:"NumKeep,omitempty" json:"num_keep,omitempty"`
	Seed             int      `yaml:"Seed,omitempty" json:"seed,omitempty"`
	NumPredict       int      `yaml:"NumPredict,omitempty" json:"num_predict,omitempty"`
	TopK             int      `yaml:"TopK,omitempty" json:"top_k,omitempty"`
	TopP             float64  `yaml:"TopP,omitempty" json:"top_p,omitempty"`
	MinP             float64  `yaml:"MinP,omitempty" json:"min_p,omitempty"`
	TfsZ             float64  `yaml:"TfsZ,omitempty" json:"tfs_z,omitempty"`
	TypicalP         float64  `yaml:"TypicalP,omitempty" json:"typical_p,omitempty"`
	RepeatLastN      int      `yaml:"RepeatLastN,omitempty" json:"repeat_last_n,omitempty"`
	Temperature      float64  `yaml:"Temperature,omitempty" json:"temperature,omitempty"`
	RepeatPenalty    float64  `yaml:"RepeatPenalty,omitempty" json:"repeat_penalty,omitempty"`
	PresencePenalty  float64  `yaml:"PresencePenalty,omitempty" json:"presence_penalty,omitempty"`
	FrequencyPenalty float64  `yaml:"FrequencyPenalty,omitempty" json:"frequency_penalty,omitempty"`
	Mirostat         int      `yaml:"Mirostat,omitempty" json:"mirostat,omitempty"`
	MirostatTau      float64  `yaml:"MirostatTau,omitempty" json:"mirostat_tau,omitempty"`
	MirostatEta      float64  `yaml:"MirostatEta,omitempty" json:"mirostat_eta,omitempty"`
	PenalizeNewline  bool     `yaml:"PenalizeNewline,omitempty" json:"penalize_newline,omitempty"`
	Stop             []string `yaml:"Stop,omitempty" json:"stop,omitempty"`
	Numa             bool     `yaml:"Numa,omitempty" json:"numa,omitempty"`
	NumCtx           int      `yaml:"NumCtx,omitempty" json:"num_ctx,omitempty"`
	NumBatch         int      `yaml:"NumBatch,omitempty" json:"num_batch,omitempty"`
	NumGpu           int      `yaml:"NumGpu,omitempty" json:"num_gpu,omitempty"`
	MainGpu          int      `yaml:"MainGpu,omitempty" json:"main_gpu,omitempty"`
	LowVram          bool     `yaml:"LowVram,omitempty" json:"low_vram,omitempty"`
	VocabOnly        bool     `yaml:"VocabOnly,omitempty" json:"vocab_only,omitempty"`
	UseMmap          bool     `yaml:"UseMmap,omitempty" json:"use_mmap,omitempty"`
	UseMlock         bool     `yaml:"UseMlock,omitempty" json:"use_mlock,omitempty"`
	NumThread        int      `yaml:"NumThread,omitempty" json:"num_thread,omitempty"`
	MaxOutputTokens  int      `yaml:"MaxOutputTokens,omitempty" json:"max_output_tokens,omitempty"`
	Detail           string   `yaml:"Detail,omitempty" json:"detail,omitempty"`
	ForceJson        bool     `yaml:"ForceJson,omitempty" json:"force_json,omitempty"`
	SchemaVersion    string   `yaml:"SchemaVersion,omitempty" json:"schema_version,omitempty"`
	CombineOutputs   string   `yaml:"CombineOutputs,omitempty" json:"combine_outputs,omitempty"`
}

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
	Options        *ApiRequestOptions `form:"options" yaml:"Options,omitempty" json:"options,omitempty"`
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
			if file, err := os.Open(images[i]); err != nil {
				return nil, fmt.Errorf("%s (create data url)", err)
			} else {
				imageUrls[i] = media.DataUrl(file)
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

	return json.Marshal(*r)
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
