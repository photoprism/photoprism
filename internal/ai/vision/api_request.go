package vision

import (
	"encoding/json"
	"path"

	"github.com/photoprism/photoprism/internal/api/download"
	"github.com/photoprism/photoprism/pkg/rnd"
)

const (
	LabelsEndpoint = "labels"
)

// ApiRequest represents a Vision API service request.
type ApiRequest struct {
	Id     string   `form:"id" yaml:"Id,omitempty" json:"id,omitempty"`
	Model  string   `form:"model" yaml:"Model,omitempty" json:"model,omitempty"`
	Images []string `form:"images" yaml:"Images,omitempty" json:"images,omitempty"`
	Videos []string `form:"videos" yaml:"Videos,omitempty" json:"videos,omitempty"`
}

func NewClientRequest(model string, images []string) *ApiRequest {
	imageUrls := make([]string, 0, len(images))

	for i := range images {
		if id, err := download.Register(images[i]); err != nil {
			log.Errorf("vision: %s (register download)", err)
		} else {
			imageUrls = append(imageUrls, path.Join(DownloadUrl, id))
		}
	}

	return &ApiRequest{
		Id:     rnd.UUID(),
		Images: imageUrls,
	}
}

// GetId returns the request ID string and generates a random ID if none was set.
func (r *ApiRequest) GetId() string {
	if r.Id == "" {
		r.Id = rnd.UUID()
	}

	return r.Id
}

// MarshalJSON returns request as JSON.
func (r *ApiRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(r)
}
