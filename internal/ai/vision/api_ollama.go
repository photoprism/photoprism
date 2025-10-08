package vision

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/service/http/scheme"
)

// ApiResponseOllama represents a Ollama API service response.
type ApiResponseOllama struct {
	Id                 string    `yaml:"Id,omitempty" json:"id,omitempty"`
	Code               int       `yaml:"Code,omitempty" json:"code,omitempty"`
	Error              string    `yaml:"Error,omitempty" json:"error,omitempty"`
	Model              string    `yaml:"Model,omitempty" json:"model,omitempty"`
	CreatedAt          time.Time `yaml:"CreatedAt,omitempty" json:"created_at,omitempty"`
	Response           string    `yaml:"Response,omitempty" json:"response,omitempty"`
	Done               bool      `yaml:"Done,omitempty" json:"done,omitempty"`
	Context            []int     `yaml:"Context,omitempty" json:"context,omitempty"`
	TotalDuration      int64     `yaml:"TotalDuration,omitempty" json:"total_duration,omitempty"`
	LoadDuration       int       `yaml:"LoadDuration,omitempty" json:"load_duration,omitempty"`
	PromptEvalCount    int       `yaml:"PromptEvalCount,omitempty" json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int       `yaml:"PromptEvalDuration,omitempty" json:"prompt_eval_duration,omitempty"`
	EvalCount          int       `yaml:"EvalCount,omitempty" json:"eval_count,omitempty"`
	EvalDuration       int64     `yaml:"EvalDuration,omitempty" json:"eval_duration,omitempty"`
	Result             ApiResult `yaml:"Result,omitempty" json:"result,omitempty"`
}

// Err returns an error if the request has failed.
func (r *ApiResponseOllama) Err() error {
	if r == nil {
		return errors.New("response is nil")
	}

	if r.Code >= 400 {
		if r.Error != "" {
			return errors.New(r.Error)
		}

		return fmt.Errorf("error %d", r.Code)
	} else if r.Result.IsEmpty() {
		return errors.New("no result")
	}

	return nil
}

// HasResult checks if there is at least one result in the response data.
func (r *ApiResponseOllama) HasResult() bool {
	if r == nil {
		return false
	}

	return !r.Result.IsEmpty()
}

// NewApiRequestOllama returns a new Ollama API request with the specified images as payload.
func NewApiRequestOllama(images Files, fileScheme scheme.Type) (*ApiRequest, error) {
	imagesData := make(Files, len(images))

	for i := range images {
		switch fileScheme {
		case scheme.Data, scheme.Base64:
			if file, err := os.Open(images[i]); err != nil {
				return nil, fmt.Errorf("%s (create data url)", err)
			} else {
				imagesData[i] = media.DataBase64(file)
			}
		default:
			return nil, fmt.Errorf("unsupported file scheme %s", clean.Log(fileScheme))
		}
	}

	return &ApiRequest{
		Id:             rnd.UUID(),
		Model:          "",
		Images:         imagesData,
		ResponseFormat: ApiFormatOllama,
	}, nil
}
