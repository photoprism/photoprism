package photoprism

import (
	"cmp"
	"fmt"
	"image"
	"slices"

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
	bounds, err := m.labelSourceBounds()
	if err != nil {
		return nil, nil, err
	}
	centerFallback := false
	centerAttempted := make(map[string]bool)
	attempted := make(map[string]bool)
	center := func() (*thumb.InputSource, error) {
		if !centerFallback {
			name, pathErr := thumb.SizeTile224.FileName(m.Hash(), Config().ThumbCachePath())
			if pathErr != nil {
				return nil, pathErr
			}
			return thumb.OpenInputSource(name, 1)
		}
		return m.labelWholeSource(centerAttempted)
	}
	whole := func() (*thumb.InputSource, error) { return m.labelWholeSource(attempted) }
	retry := func(index int) bool {
		if index == 0 {
			if centerAttempted[m.FileName()] {
				return false
			}
			centerFallback = true
			return true
		}
		return !attempted[m.FileName()]
	}
	return prepareLabelInputs(bounds, center, whole, retry)

}

// labelSourceBounds reads decoded dimensions with the configured renderer as fallback.
func (m *MediaFile) labelSourceBounds() (image.Rectangle, error) {
	if m == nil {
		return image.Rectangle{}, fmt.Errorf("missing media file")
	}
	cfg, err := m.DecodeConfig()
	if err == nil {
		return image.Rect(0, 0, cfg.Width, cfg.Height), nil
	}
	if thumb.Library != thumb.LibVips {
		return image.Rectangle{}, err
	}
	source, sourceErr := thumb.OpenInputSource(m.FileName(), 1)
	if sourceErr != nil {
		return image.Rectangle{}, sourceErr
	}
	defer source.Close()
	return source.Bounds(), nil
}

// labelWholeSource opens cached fit inputs by actual size, then falls back to the original.
func (m *MediaFile) labelWholeSource(attempted map[string]bool) (*thumb.InputSource, error) {
	if attempted == nil {
		attempted = make(map[string]bool)
	}
	type candidate struct {
		name     string
		area     int64
		adequate bool
	}
	candidates := make([]candidate, 0, len(thumb.All))
	for _, size := range thumb.All {
		if !size.Fit {
			continue
		}
		name, err := size.FileName(m.Hash(), Config().ThumbCachePath())
		if err != nil || attempted[name] {
			continue
		}
		cfg, _, err := fs.DecodeImageConfigFile(name)
		if err != nil {
			continue
		}
		candidates = append(candidates, candidate{name: name, area: int64(cfg.Width) * int64(cfg.Height), adequate: min(cfg.Width, cfg.Height) >= 224})
	}
	// Fit boxes have different heights, so preset width is not a resolution order.
	slices.SortFunc(candidates, func(a, b candidate) int {
		if a.adequate != b.adequate {
			if a.adequate {
				return -1
			}
			return 1
		}
		order := cmp.Compare(a.area, b.area)
		if !a.adequate {
			order = -order
		}
		if order == 0 {
			return cmp.Compare(a.name, b.name)
		}
		return order
	})
	for _, item := range candidates {
		attempted[item.name] = true
		source, err := thumb.OpenInputSource(item.name, 1)
		if err == nil {
			return source, nil
		}
		if source != nil {
			source.Close()
		}
	}
	if attempted[m.FileName()] {
		return nil, fmt.Errorf("no usable label source")
	}
	attempted[m.FileName()] = true
	return thumb.OpenInputSource(m.FileName(), m.Orientation())
}

// prepareLabelInputs applies S2 geometry and optionally retries failed sources.
func prepareLabelInputs(bounds image.Rectangle, center, whole func() (*thumb.InputSource, error), retry ...func(int) bool) ([]classify.Input, []labelInputGeometry, error) {
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
		for {
			load := center
			if i == 1 {
				load = whole
			}
			source, err := load()
			if err == nil && (source == nil || source.Bounds().Empty()) {
				err = fmt.Errorf("invalid label source")
			}
			var pixels image.Image
			var selectedGeometry labelInputGeometry
			if err == nil {
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
				selectedGeometry = labelInputGeometry{SourceBounds: source.Bounds(), CropBounds: region}
				pixels, err = source.Resample(region, 224, 224)
			}
			if source != nil {
				source.Close()
			}
			if err != nil {
				if len(retry) > 0 && retry[0] != nil && retry[0](i) {
					continue
				}
				return nil, nil, err
			}
			geometry = append(geometry, selectedGeometry)
			inputs = append(inputs, classify.Input{Image: pixels, Prepared: true})
			break
		}
	}

	return inputs, geometry, nil
}
