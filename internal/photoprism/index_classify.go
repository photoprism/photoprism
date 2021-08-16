package photoprism

import (
	"sort"
	"time"

	"github.com/photoprism/photoprism/internal/classify"
	"github.com/photoprism/photoprism/pkg/txt"
)

// classifyImage classifies a JPEG image and returns matching labels.
func (ind *Index) classifyImage(jpeg *MediaFile) (results classify.Labels) {
	start := time.Now()

	var thumbs []string

	if jpeg.AspectRatio() == 1 {
		thumbs = []string{"tile_224"}
	} else {
		thumbs = []string{"tile_224", "left_224", "right_224"}
	}

	var labels classify.Labels

	for _, thumb := range thumbs {
		filename, err := jpeg.Thumbnail(Config().ThumbPath(), thumb)

		if err != nil {
			log.Debugf("%s in %s", err, txt.Quote(jpeg.BaseName()))
			continue
		}

		imageLabels, err := ind.tensorFlow.File(filename)

		if err != nil {
			log.Debugf("%s in %s", err, txt.Quote(jpeg.BaseName()))
			continue
		}

		labels = append(labels, imageLabels...)
	}

	// Sort by priority and uncertainty
	sort.Sort(labels)

	var confidence int

	for _, label := range labels {
		if confidence == 0 {
			confidence = 100 - label.Uncertainty
		}

		if (100 - label.Uncertainty) > (confidence / 3) {
			results = append(results, label)
		}
	}

	if len(labels) > 0 {
		log.Infof("index: found %d matching labels for %s [%s]", len(labels), txt.Quote(jpeg.BaseName()), time.Since(start))
	}

	return results
}
