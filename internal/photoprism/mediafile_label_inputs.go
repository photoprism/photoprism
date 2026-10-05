package photoprism

import (
	"fmt"
	"image"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/fs"
)

// labelInputGeometry records the oriented source and selected input region.
type labelInputGeometry struct {
	SourceBounds image.Rectangle `json:"sourceBounds"`
	CropBounds   image.Rectangle `json:"cropBounds"`
}

// PrepareLabelInputs builds the S2 center and capped whole-photo inputs.
func (m *MediaFile) PrepareLabelInputs() ([]classify.Input, error) {
	inputs, _, err := m.prepareLabelInputsWithGeometry()
	return inputs, err
}

// prepareLabelInputsWithGeometry returns inputs and their actual preparation regions.
func (m *MediaFile) prepareLabelInputsWithGeometry() ([]classify.Input, []labelInputGeometry, error) {
	if m == nil {
		return nil, nil, fmt.Errorf("missing media file")
	}
	cfg, err := m.DecodeConfig()
	if err != nil {
		return nil, nil, err
	}
	bounds := image.Rect(0, 0, cfg.Width, cfg.Height)
	inputs, geometry, inputErr := prepareLabelInputs(bounds, func() (*thumb.InputSource, error) {
		if min(cfg.Width, cfg.Height) < 224 || (cfg.Width == cfg.Height && cfg.Width <= 224) {
			return thumb.OpenInputSource(m.FileName(), m.Orientation())
		}
		name, thumbErr := m.Thumbnail(Config().ThumbCachePath(), thumb.Tile224)
		if thumbErr != nil {
			return nil, thumbErr
		}
		return thumb.OpenInputSource(name, 1)
	}, m.labelWholeSource)
	if inputErr == nil {
		return inputs, geometry, nil
	}
	original := func() (*thumb.InputSource, error) { return thumb.OpenInputSource(m.FileName(), m.Orientation()) }
	return prepareLabelInputs(bounds, original, original)
}

// labelWholeSource opens the smallest cached whole-photo rendition with a short side of at least 224 px,
// otherwise the largest cached rendition, and the original only if none is cached, so a large original is
// never decoded for a 224 px input.
func (m *MediaFile) labelWholeSource() (*thumb.InputSource, error) {
	var largest string

	// thumb.All is sorted by width, so the first adequate rendition is the smallest.
	for _, size := range thumb.All {
		if !size.Fit {
			continue
		}
		name, err := size.FileName(m.Hash(), Config().ThumbCachePath())
		if err != nil {
			continue
		}
		cfg, _, err := fs.DecodeImageConfigFile(name)
		if err != nil {
			continue
		}
		largest = name
		if min(cfg.Width, cfg.Height) < 224 {
			continue
		}
		if img, openErr := thumb.OpenInputSource(name, 1); openErr == nil {
			return img, nil
		} else if img != nil {
			img.Close()
		}
	}
	if largest != "" {
		if img, err := thumb.OpenInputSource(largest, 1); err == nil {
			return img, nil
		} else if img != nil {
			img.Close()
		}
	}
	return thumb.OpenInputSource(m.FileName(), m.Orientation())
}

// prepareLabelInputs applies the S2 geometry once to each selected source region.
func prepareLabelInputs(bounds image.Rectangle, center, whole func() (*thumb.InputSource, error)) ([]classify.Input, []labelInputGeometry, error) {
	if bounds.Empty() {
		return nil, nil, fmt.Errorf("invalid label source dimensions")
	}
	single := bounds.Dx() == bounds.Dy() || min(bounds.Dx(), bounds.Dy()) < 224
	count := 2
	if single {
		count = 1
	}
	inputs := make([]classify.Input, 0, count)
	geometry := make([]labelInputGeometry, 0, count)
	for i := 0; i < count; i++ {
		load := center
		if i == 1 {
			load = whole
		}
		source, err := load()
		if err != nil {
			return nil, nil, err
		}
		if source == nil || source.Bounds().Empty() {
			if source != nil {
				source.Close()
			}
			return nil, nil, fmt.Errorf("invalid label source")
		}
		region := source.Bounds()
		w, h := region.Dx(), region.Dy()
		if i == 0 {
			side := min(w, h)
			region.Min = region.Min.Add(image.Pt((w-side)/2, (h-side)/2))
			region.Max = region.Min.Add(image.Pt(side, side))
		} else {
			cropW, cropH := w, h
			if w > h {
				cropW = min(w, h*4/3)
			} else {
				cropH = min(h, w*4/3)
			}
			region.Min = region.Min.Add(image.Pt((w-cropW)/2, (h-cropH)/2))
			region.Max = region.Min.Add(image.Pt(cropW, cropH))
		}
		geometry = append(geometry, labelInputGeometry{SourceBounds: source.Bounds(), CropBounds: region})
		pixels, resizeErr := source.Resample(region, 224, 224)
		source.Close()
		if resizeErr != nil {
			return nil, nil, resizeErr
		}
		inputs = append(inputs, classify.Input{Image: pixels, Prepared: true})
	}
	return inputs, geometry, nil
}
