package photoprism

import (
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/media"
)

// Labels classifies the media file and returns matching labels. When labelSrc
// is SrcAuto the model's declared source is used; otherwise the provided source
// is applied to every returned label.
func (ind *Index) Labels(file *MediaFile, labelSrc entity.Src) (labels classify.Labels) {
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
	} else if file.Square() {
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
		if thumbnail, fileErr := file.Thumbnail(Config().ThumbCachePath(), s); fileErr != nil {
			log.Debugf("index: %s in %s", err, clean.Log(file.BaseName()))
			continue
		} else {
			thumbnails = append(thumbnails, thumbnail)
		}
	}

	// Run the configured vision model to obtain labels for the generated thumbnails.
	if labels, err = vision.Labels(thumbnails, media.SrcLocal, labelSrc); err != nil {
		log.Debugf("labels: %s in %s", err, clean.Log(file.BaseName()))
		return labels
	}

	// Log number of labels found and return results.
	if n := len(labels); n > 0 {
		log.Infof("index: found %s for %s [%s]", english.Plural(n, "label", "labels"), clean.Log(file.BaseName()), time.Since(start))
	}

	return labels
}
