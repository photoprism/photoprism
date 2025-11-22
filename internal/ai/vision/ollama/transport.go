package ollama

import (
	"errors"
	"fmt"
	"time"
)

// Response encapsulates the subset of the Ollama generate API response we care about.
type Response struct {
	ID                 string        `yaml:"Id,omitempty" json:"id,omitempty"`
	Code               int           `yaml:"Code,omitempty" json:"code,omitempty"`
	Error              string        `yaml:"Error,omitempty" json:"error,omitempty"`
	Model              string        `yaml:"Model,omitempty" json:"model,omitempty"`
	CreatedAt          time.Time     `yaml:"CreatedAt,omitempty" json:"created_at,omitempty"`
	Response           string        `yaml:"Response,omitempty" json:"response,omitempty"`
	Thinking           string        `yaml:"Thinking,omitempty" json:"thinking,omitempty"`
	Done               bool          `yaml:"Done,omitempty" json:"done,omitempty"`
	Context            []int         `yaml:"Context,omitempty" json:"context,omitempty"`
	TotalDuration      int64         `yaml:"TotalDuration,omitempty" json:"total_duration,omitempty"`
	LoadDuration       int           `yaml:"LoadDuration,omitempty" json:"load_duration,omitempty"`
	PromptEvalCount    int           `yaml:"PromptEvalCount,omitempty" json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int           `yaml:"PromptEvalDuration,omitempty" json:"prompt_eval_duration,omitempty"`
	EvalCount          int           `yaml:"EvalCount,omitempty" json:"eval_count,omitempty"`
	EvalDuration       int64         `yaml:"EvalDuration,omitempty" json:"eval_duration,omitempty"`
	Result             ResultPayload `yaml:"Result,omitempty" json:"result,omitempty"`
}

// Err returns an error if the request has failed.
func (r *Response) Err() error {
	if r == nil {
		return errors.New("response is nil")
	}

	if r.Code >= 400 {
		if r.Error != "" {
			return errors.New(r.Error)
		}

		return fmt.Errorf("error %d", r.Code)
	} else if len(r.Result.Labels) == 0 && r.Result.Caption == nil {
		return errors.New("no result")
	}

	return nil
}

// HasResult checks if there is at least one result in the response data.
func (r *Response) HasResult() bool {
	if r == nil {
		return false
	}

	return len(r.Result.Labels) > 0 || r.Result.Caption != nil
}

// ResultPayload mirrors the structure returned by Ollama for result data.
type ResultPayload struct {
	Labels  []LabelPayload  `json:"labels"`
	Caption *CaptionPayload `json:"caption,omitempty"`
}

// LabelPayload represents a single label object emitted by the Ollama adapter.
type LabelPayload struct {
	Name           string   `json:"name"`
	Source         string   `json:"source,omitempty"`
	Priority       int      `json:"priority,omitempty"`
	Confidence     float32  `json:"confidence,omitempty"`
	Topicality     float32  `json:"topicality,omitempty"`
	Categories     []string `json:"categories,omitempty"`
	NSFW           bool     `json:"nsfw,omitempty"`
	NSFWConfidence float32  `json:"nsfw_confidence,omitempty"`
}

// CaptionPayload represents the caption object emitted by the Ollama adapter.
type CaptionPayload struct {
	Text       string  `json:"text"`
	Source     string  `json:"source,omitempty"`
	Confidence float32 `json:"confidence,omitempty"`
}
