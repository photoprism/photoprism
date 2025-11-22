package photoprism

import (
	"errors"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/ai/nsfw"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/media"
)

// GenerateCaption generates a caption for the provided media file using the active
// vision model. When captionSrc is SrcAuto the model's declared source is used;
// otherwise the explicit source is recorded on the returned caption.
func (m *MediaFile) GenerateCaption(captionSrc entity.Src) (caption *vision.CaptionResult, err error) {
	start := time.Now()

	model := vision.Config.Model(vision.ModelTypeCaption)

	// No caption generation model configured or usable.
	if model == nil {
		return caption, errors.New("no caption model configured")
	}

	if captionSrc == entity.SrcAuto {
		captionSrc = model.GetSource()
	}

	size := vision.Thumb(vision.ModelTypeCaption)

	// Get thumbnail filenames for the selected sizes.
	fileName, fileErr := m.Thumbnail(Config().ThumbCachePath(), size.Name)

	if fileErr != nil {
		return caption, err
	}

	// Get matching labels from computer vision model.
	// Generate a caption using the configured vision model.
	if caption, _, err = vision.GenerateCaption(vision.Files{fileName}, media.SrcLocal); err != nil {
		// Failed.
	} else if caption.Text != "" {
		if captionSrc != entity.SrcAuto {
			caption.Source = captionSrc
		}

		log.Infof("vision: generated caption for %s [%s]", clean.Log(m.RootRelName()), time.Since(start))
	}

	return caption, err
}

// GenerateLabels classifies the media file and returns matching labels. When labelSrc
// is SrcAuto the model's declared source is used; otherwise the provided source
// is applied to every returned label.
func (m *MediaFile) GenerateLabels(labelSrc entity.Src) (labels classify.Labels) {
	if m == nil {
		return labels
	}

	start := time.Now()

	var err error
	var sizes []thumb.Name
	var thumbnails []string

	model := vision.Config.Model(vision.ModelTypeLabels)

	// No label generation model configured or usable.
	if model == nil {
		return labels
	}

	if labelSrc == entity.SrcAuto {
		labelSrc = model.GetSource()
	}

	size := vision.Thumb(vision.ModelTypeLabels)

	// The thumbnail size may need to be adjusted to use other models.
	if size.Name != "" && size.Name != thumb.Tile224 {
		sizes = []thumb.Name{size.Name}
		thumbnails = make([]string, 0, 1)
	} else if m.Square() {
		// Only one thumbnail is required for square images.
		sizes = []thumb.Name{thumb.Tile224}
		thumbnails = make([]string, 0, 1)
	} else {
		// Use three thumbnails otherwise (center, left, right).
		sizes = []thumb.Name{thumb.Tile224, thumb.Left224, thumb.Right224}
		thumbnails = make([]string, 0, 3)
	}

	// Get thumbnail filenames for the selected sizes.
	for _, s := range sizes {
		if thumbnail, fileErr := m.Thumbnail(Config().ThumbCachePath(), s); fileErr != nil {
			log.Debugf("index: %s in %s", err, clean.Log(m.RootRelName()))
			continue
		} else {
			thumbnails = append(thumbnails, thumbnail)
		}
	}

	// Run the configured vision model to obtain labels for the generated thumbnails.
	if labels, err = vision.GenerateLabels(thumbnails, media.SrcLocal, labelSrc); err != nil {
		log.Debugf("labels: %s in %s", err, clean.Log(m.RootRelName()))
		return labels
	}

	// Log number and names of generated labels.
	if n := labels.Count(); n > 0 {
		log.Debugf("vision: %#v", labels)
		log.Infof("vision: generated %s for %s [%s]", english.Plural(n, "label", "labels"), clean.Log(m.RootRelName()), time.Since(start))
	}

	return labels
}

// DetectNSFW returns true if media file might be offensive and detection is enabled.
func (m *MediaFile) DetectNSFW() bool {
	filename, err := m.Thumbnail(Config().ThumbCachePath(), thumb.Fit720)

	if err != nil {
		log.Error(err)
		return false
	}

	if results, modelErr := vision.DetectNSFW([]string{filename}, media.SrcLocal); modelErr != nil {
		log.Errorf("vision: %s in %s (detect nsfw)", modelErr, clean.Log(m.RootRelName()))
		return false
	} else if len(results) < 1 {
		log.Errorf("vision: nsfw model returned no result for %s", clean.Log(m.RootRelName()))
		return false
	} else if results[0].IsNsfw(nsfw.ThresholdHigh) {
		log.Warnf("vision: detected offensive content in %s", clean.Log(m.RootRelName()))
		return true
	}

	return false
}
