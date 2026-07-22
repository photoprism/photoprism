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
//
// For media.SrcLocal the images are trusted local file paths read directly; for media.SrcRemote
// the caller-supplied references are resolved through remoteImageBase64 (https/data URLs only,
// never a local file).
func NewApiRequestOllama(images Files, fileScheme scheme.Type, mediaSrc media.Src) (*ApiRequest, error) {
	imagesData := make(Files, len(images))

	for i := range images {
		switch mediaSrc {
		case media.SrcRemote:
			b64, err := remoteImageBase64(images[i])
			if err != nil {
				return nil, err
			}
			imagesData[i] = b64
		case media.SrcLocal:
			switch fileScheme {
			case scheme.Data, scheme.Base64:
				file, err := os.Open(images[i])
				if err != nil {
					return nil, fmt.Errorf("%s (create data url)", err)
				}
				imagesData[i] = media.DataBase64(file)
				if err := file.Close(); err != nil {
					return nil, fmt.Errorf("%s (close data url)", err)
				}
			default:
				return nil, fmt.Errorf("unsupported file scheme %s", clean.Log(fileScheme))
			}
		default:
			return nil, fmt.Errorf("invalid media source %s", clean.Log(mediaSrc))
		}
	}

	return &ApiRequest{
		Id:             rnd.UUID(),
		Model:          "",
		Images:         imagesData,
		ResponseFormat: ApiFormatOllama,
	}, nil
}
