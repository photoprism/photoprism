package photoprism

import (
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/media"
)

// Labels classifies a JPEG image and returns matching labels.
func (ind *Index) Labels(file *MediaFile) (labels classify.Labels) {
	start := time.Now()

	var err error
	var sizes []thumb.Name
	var thumbnails []string

	// The thumbnail size may need to be adjusted to use other models.
	if file.Square() {
		// Only one thumbnail is required for square images.
		sizes = []thumb.Name{thumb.Tile224}
		thumbnails = make([]string, 0, 1)
	} else {
		// Use three thumbnails otherwise (center, left, right).
		sizes = []thumb.Name{thumb.Tile224, thumb.Left224, thumb.Right224}
		thumbnails = make([]string, 0, 3)
	}

	// Get thumbnail filenames for the selected sizes.
	for _, size := range sizes {
		if thumbnail, fileErr := file.Thumbnail(Config().ThumbCachePath(), size); fileErr != nil {
			log.Debugf("index: %s in %s", err, clean.Log(file.BaseName()))
			continue
		} else {
			thumbnails = append(thumbnails, thumbnail)
		}
	}

	// Get matching labels from computer vision model.
	if labels, err = vision.Labels(thumbnails, media.SrcLocal); err != nil {
		log.Debugf("labels: %s in %s", err, clean.Log(file.BaseName()))
		return labels
	}

	// Log number of labels found and return results.
	if n := len(labels); n > 0 {
		log.Infof("index: found %s for %s [%s]", english.Plural(n, "label", "labels"), clean.Log(file.BaseName()), time.Since(start))
	}

	return labels
}
