package openai

import (
	"encoding/json"
	"strings"
)

const (
	// ContentTypeText identifies text input segments for the Responses API.
	ContentTypeText = "input_text"
	// ContentTypeImage identifies image input segments for the Responses API.
	ContentTypeImage = "input_image"

	// ResponseFormatJSONSchema requests JSON constrained by a schema.
	ResponseFormatJSONSchema = "json_schema"
	// ResponseFormatJSONObject requests a free-form JSON object.
	ResponseFormatJSONObject = "json_object"
)

// HTTPRequest represents the payload expected by OpenAI's Responses API.
type HTTPRequest struct {
	Model            string         `json:"model"`
	Input            []InputMessage `json:"input"`
	Text             *TextOptions   `json:"text,omitempty"`
	Reasoning        *Reasoning     `json:"reasoning,omitempty"`
	MaxOutputTokens  int            `json:"max_output_tokens,omitempty"`
	Temperature      float64        `json:"temperature,omitempty"`
	TopP             float64        `json:"top_p,omitempty"`
	PresencePenalty  float64        `json:"presence_penalty,omitempty"`
	FrequencyPenalty float64        `json:"frequency_penalty,omitempty"`
}

// TextOptions carries formatting preferences for textual responses.
type TextOptions struct {
	Format *ResponseFormat `json:"format,omitempty"`
}

// Reasoning configures the effort level for reasoning models.
type Reasoning struct {
	Effort string `json:"effort,omitempty"`
}

// InputMessage captures a single system or user message in the request.
type InputMessage struct {
	Role    string        `json:"role"`
	Type    string        `json:"type,omitempty"`
	Content []ContentItem `json:"content"`
}

// ContentItem represents a text or image entry within a message.
type ContentItem struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
	Detail   string `json:"detail,omitempty"`
}

// ResponseFormat describes how OpenAI should format its response.
type ResponseFormat struct {
	Type        string          `json:"type"`
	Name        string          `json:"name,omitempty"`
	Schema      json.RawMessage `json:"schema,omitempty"`
	Description string          `json:"description,omitempty"`
	Strict      bool            `json:"strict,omitempty"`
}

// Response mirrors the subset of the Responses API response we need.
type Response struct {
	ID     string           `json:"id"`
	Model  string           `json:"model"`
	Output []ResponseOutput `json:"output"`
	Error  *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

// ResponseOutput captures assistant messages within the response.
type ResponseOutput struct {
	Role    string            `json:"role"`
	Content []ResponseContent `json:"content"`
}

// ResponseContent contains individual message parts (JSON or text).
type ResponseContent struct {
	Type string          `json:"type"`
	Text string          `json:"text,omitempty"`
	JSON json.RawMessage `json:"json,omitempty"`
}

// FirstJSON returns the first JSON payload contained in the response.
func (r *Response) FirstJSON() json.RawMessage {
	if r == nil {
		return nil
	}

	for i := range r.Output {
		for j := range r.Output[i].Content {
			if len(r.Output[i].Content[j].JSON) > 0 {
				return r.Output[i].Content[j].JSON
			}
		}
	}

	return nil
}

// FirstText returns the first textual payload contained in the response.
func (r *Response) FirstText() string {
	if r == nil {
		return ""
	}

	for i := range r.Output {
		for j := range r.Output[i].Content {
			if text := strings.TrimSpace(r.Output[i].Content[j].Text); text != "" {
				return text
			}
		}
	}

	return ""
}

// ParseErrorMessage extracts a human readable error message from a Responses API payload.
func ParseErrorMessage(raw []byte) string {
	var errResp struct {
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(raw, &errResp); err != nil {
		return ""
	}

	if errResp.Error != nil {
		return strings.TrimSpace(errResp.Error.Message)
	}

	return ""
}
