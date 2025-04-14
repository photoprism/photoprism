package vision

import (
	"bytes"
	"errors"
	"fmt"
	"image/jpeg"

	"github.com/photoprism/photoprism/internal/ai/face"
)

// Face returns the embeddings for the specified face crop image.
func Face(imgData []byte) (embeddings face.Embeddings, err error) {
	if len(imgData) == 0 {
		return embeddings, errors.New("missing image")
	}

	if Config == nil {
		return embeddings, errors.New("vision service is not configured")
	} else if model := Config.Model(ModelTypeFace); model != nil {
		img, imgErr := jpeg.Decode(bytes.NewReader(imgData))

		if imgErr != nil {
			return embeddings, imgErr
		}

		if tf := model.FaceModel(); tf == nil {
			return embeddings, fmt.Errorf("invalid face model configuration")
		} else if embeddings = tf.Run(img); !embeddings.Empty() {
			return embeddings, nil
		} else {
			return face.Embeddings{}, nil
		}
	} else {
		return embeddings, fmt.Errorf("no face model configured")
	}
}
