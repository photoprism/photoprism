package photoprism

import (
	"image"
	"math"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/internal/thumb/crop"
	"github.com/photoprism/photoprism/pkg/clean"
)

// embedCropWidth returns the widest source a face crop may be requested at for an embedding model.
//
// The wider of the two a run can ask for: an aligned model warps the landmarks onto its own input
// geometry, but a face whose landmarks do not fit the template falls back to a face.CropSize box.
// Which of the two a face takes is not known until its landmarks have been fitted, so every file
// pays for the box - the alternative is rendering a second time for the few that fall back.
//
// The model's registered input rather than the loaded embedder's, which an ONNX graph may declare
// for itself: the two agree for every bundled model, and only this one is known before a crop.
func embedCropWidth(target string) int {
	width := face.CropSize.Width

	if model := face.FindEmbeddingModel(target); model != nil {
		if w, _ := model.InputSize(); w > width {
			width = w
		}
	}

	return width
}

// cropSourceWidth returns the source width a face covering relWidth of a picture needs, so a crop
// of cropWidth pixels is taken from it rather than upscaled onto that size.
func cropSourceWidth(relWidth float32, cropWidth int) int {
	if relWidth <= 0 || cropWidth < 1 {
		return 0
	}

	return int(math.Ceil(float64(cropWidth) / float64(relWidth)))
}

// markersCropSourceWidth returns the widest source these markers ask for, which is what the
// smallest face among them needs: a rendition that supplies it supplies every larger face too.
func markersCropSourceWidth(markers entity.Markers, cropWidth int) int {
	var required int

	for _, m := range markers {
		if width := cropSourceWidth(m.W, cropWidth); width > required {
			required = width
		}
	}

	return required
}

// facesCropSourceWidth returns the widest source these detections ask for, which is what the
// smallest face among them needs.
func facesCropSourceWidth(faces face.Faces, cropWidth int) int {
	var required int

	for i := range faces {
		if width := cropSourceWidth(faces[i].CropArea().W, cropWidth); width > required {
			required = width
		}
	}

	return required
}

// cropThumbSize returns the smallest rendition within limit that can supply a crop of the
// specified source width from a picture of these dimensions, and the one that holds the picture at
// its own resolution when the source is too small for any of them to.
//
// Never wider than that last one, which is also what indexing would have written: a rendition
// named for a box larger than the picture holds the same pixels under a name nothing else uses.
func cropThumbSize(fileWidth, fileHeight, width int, limit thumb.Size) thumb.Name {
	for _, size := range crop.UsableSizes() {
		if size.Width > limit.Width {
			break
		}

		if w, _ := size.Fitted(fileWidth, fileHeight); w >= width {
			return size.Name
		}
	}

	// Nothing within the bound supplies it, so the widest rendition this original is given does -
	// which for a picture larger than the bound is the bound itself.
	if native := thumb.FitBounds(image.Rect(0, 0, fileWidth, fileHeight)); native.Width < limit.Width {
		return native.Name
	}

	return limit.Name
}

// cachedCropWidth returns the widest source the renditions already on disk can supply for a file.
func cachedCropWidth(hash string, fileWidth, fileHeight int, thumbPath string) int {
	var widest int

	for _, size := range crop.UsableSizes() {
		if !crop.CachedSizeExists(size, hash, thumbPath) {
			continue
		}

		if w, _ := size.Fitted(fileWidth, fileHeight); w > widest {
			widest = w
		}
	}

	return widest
}

// faceCropSourceLimit returns the widest rendition a face crop may be rendered from on demand,
// which is the largest one a crop can be taken from within the configured ceiling. The zero size
// means on demand rendering is off, either by configuration or because nothing it allows is a
// rendition a crop is taken from.
//
// The ceiling never exceeds what this process renders at all, because thumb.FromFile refuses a
// larger size: an unclamped one would fail per file instead of rendering the widest it does allow.
func faceCropSourceLimit(maxWidth int) (limit thumb.Size) {
	maxWidth = min(maxWidth, thumb.MaxRenderSize())

	for _, size := range crop.UsableSizes() {
		if size.Width > maxWidth {
			break
		}

		limit = size
	}

	return limit
}

// cacheFaceCropSource renders the rendition these detections need to be cropped at full detail,
// when the cache holds none wide enough, and reports whether it asked for one.
//
// An embedding is drawn from a cached rendition, so without this what a crop can reach is decided
// by whatever the thumbnail limit happened to pre-generate, and a vector computed from upscaled
// pixels is indistinguishable from one that was not. The on demand limit the delivery paths obey
// is deliberately not inherited: that one prices a results page rendering hundreds of thumbnails
// at once, while indexing renders at most one rendition per file it reads anyway.
func cacheFaceCropSource(conf *config.Config, m *MediaFile, faces face.Faces) (rendered bool, err error) {
	if conf == nil || m == nil || len(faces) == 0 {
		return false, nil
	}

	limit := faceCropSourceLimit(conf.ThumbSizeFace())

	if limit.Width < 1 {
		return false, nil
	}

	fileWidth, fileHeight := m.Width(), m.Height()

	if fileWidth < 1 || fileHeight < 1 {
		return false, nil
	}

	required := facesCropSourceWidth(faces, embedCropWidth(face.EmbeddingModelName()))

	if required < 1 {
		return false, nil
	}

	thumbPath := conf.ThumbCachePath()
	size := thumb.Sizes[cropThumbSize(fileWidth, fileHeight, required, limit)]
	want, _ := size.Fitted(fileWidth, fileHeight)

	// Against what that rendition delivers rather than against what the faces asked for: a face
	// whose original holds too little detail is at its best there, and comparing with the request
	// would re-render for it on every pass.
	if want < 1 || cachedCropWidth(m.Hash(), fileWidth, fileHeight, thumbPath) >= want {
		return false, nil
	}

	if _, err = m.Thumbnail(thumbPath, size.Name); err != nil {
		return false, err
	}

	log.Debugf("faces: rendered %s of %s so its face crops are not upscaled",
		size.Name, clean.Log(m.BaseName()))

	return true, nil
}
