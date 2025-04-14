package photoprism

import (
	"time"

	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/media"
)

// Caption returns generated caption for the specified media file.
func (ind *Index) Caption(file *MediaFile) (caption vision.CaptionResult, err error) {
	start := time.Now()

	size := vision.Thumb(vision.ModelTypeCaption)

	// Get thumbnail filenames for the selected sizes.
	fileName, fileErr := file.Thumbnail(Config().ThumbCachePath(), size.Name)

	if fileErr != nil {
		return caption, err
	}

	// Get matching labels from computer vision model.
	if caption, err = vision.Caption(fileName, media.SrcLocal); err != nil {
	} else if caption.Text != "" {
		log.Infof("vision: generated caption for %s [%s]", clean.Log(file.BaseName()), time.Since(start))
	}

	return caption, err
}
