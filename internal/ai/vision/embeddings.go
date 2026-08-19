package vision

import (
	"image"
	"path/filepath"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/thumb/crop"
	"github.com/photoprism/photoprism/pkg/clean"
)

// GenerateEmbeddings runs the embedding model on each detected face and assigns the result.
func GenerateEmbeddings(embedder face.Embedder, fileName string, faces face.Faces, cacheCrop bool) {
	if embedder == nil || len(faces) == 0 {
		return
	}

	width, height := embedder.CropSize()

	// Landmark alignment reads the source image directly, so it is decoded once for all
	// faces rather than per face. The smallest face decides the rendition, because it is
	// the one that would otherwise be upscaled onto the template.
	var srcImg image.Image

	if embedder.Aligned() {
		size := crop.Size{Width: width, Height: height, Options: crop.DefaultOptions}

		if img, err := crop.ImageFromIdealThumb(fileName, smallestFaceArea(faces), size); err != nil {
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

		// The name is recorded next to the vector because this is the last frame where the
		// producer is known: everything downstream would have to ask global configuration.
		if embeddings := embedder.Run(img); !embeddings.Empty() {
			f.Embeddings = embeddings
			f.EmbedModel = embedder.ModelName()
		}
	}
}

// smallestFaceArea returns the crop area of the smallest detected face, which is the one
// that needs the largest source image to fill a template without being upscaled.
func smallestFaceArea(faces face.Faces) crop.Area {
	var result crop.Area

	for i := range faces {
		if area := faces[i].CropArea(); result.W <= 0 || area.W < result.W {
			result = area
		}
	}

	return result
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
