package vision

import (
	"image"
	"path/filepath"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/thumb/crop"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// GenerateEmbeddings runs the embedding model on each detected face and assigns the result.
func GenerateEmbeddings(embedder face.Embedder, fileName string, faces face.Faces, cacheCrop bool) {
	if embedder == nil || len(faces) == 0 {
		return
	}

	width, height := embedder.CropSize()

	// Landmark alignment needs the image the detector measured the landmarks in,
	// so it is decoded once instead of per face.
	var srcImg image.Image

	if embedder.Aligned() {
		if img, _, err := fs.DecodeImageFile(fileName); err != nil {
			log.Warnf("vision: failed to decode %s (%s)", clean.Log(filepath.Base(fileName)), err)
		} else {
			srcImg = img
		}
	}

	for i := range faces {
		f := &faces[i]

		if f.Area.Col == 0 && f.Area.Row == 0 {
			continue
		}

		img, err := faceCropImage(embedder, srcImg, fileName, f, width, height, cacheCrop)

		if err != nil {
			log.Errorf("vision: failed to create face crop (%s)", err)
			continue
		}

		if embeddings := embedder.Run(img); !embeddings.Empty() {
			f.Embeddings = embeddings
		}
	}
}

// faceCropImage returns the image to run inference on, aligned on the detected landmarks
// when the model expects it, and a plain bounding box crop otherwise.
func faceCropImage(embedder face.Embedder, srcImg image.Image, fileName string, f *face.Face, width, height int, cacheCrop bool) (image.Image, error) {
	if embedder.Aligned() && srcImg != nil {
		if img, err := face.AlignedCrop(srcImg, f, width, height); err == nil {
			return img, nil
		} else {
			// Faces without a complete landmark set still get an embedding, at the cost
			// of the pose normalization the aligned models were trained with.
			log.Debugf("vision: %s, using unaligned face crop", err)
		}
	}

	img, _, err := crop.ImageFromThumb(fileName, f.CropArea(), face.CropSize, cacheCrop)

	return img, err
}
