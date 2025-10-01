package photoprism

import (
	"errors"
	"time"

	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/media"
)

// Caption generates a caption for the provided media file using the active
// vision model. When captionSrc is SrcAuto the model's declared source is used;
// otherwise the explicit source is recorded on the returned caption.
func (ind *Index) Caption(file *MediaFile, captionSrc entity.Src) (caption *vision.CaptionResult, err error) {
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
	fileName, fileErr := file.Thumbnail(Config().ThumbCachePath(), size.Name)

	if fileErr != nil {
		return caption, err
	}

	// Get matching labels from computer vision model.
	// Generate a caption using the configured vision model.
	if caption, _, err = vision.Caption(vision.Files{fileName}, media.SrcLocal); err != nil {
		// Failed.
	} else if caption.Text != "" {
		if captionSrc != entity.SrcAuto {
			caption.Source = captionSrc
		}

		log.Infof("vision: generated caption for %s [%s]", clean.Log(file.BaseName()), time.Since(start))
	}

	return caption, err
}
