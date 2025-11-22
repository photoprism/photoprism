package vision

import (
	"fmt"
	"os"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/http/scheme"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/rnd"
)

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
