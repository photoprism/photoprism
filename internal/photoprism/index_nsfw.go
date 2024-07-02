package photoprism

import (
	"github.com/photoprism/photoprism/internal/tensorflow/nsfw"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/clean"
)

// NSFW returns true if media file might be offensive and detection is enabled.
func (ind *Index) NSFW(m *MediaFile) bool {
	filename, err := m.Thumbnail(Config().ThumbCachePath(), thumb.Fit720)

	if err != nil {
		log.Error(err)
		return false
	}

	if nsfwLabels, err := ind.nsfwDetector.File(filename); err != nil {
		log.Errorf("index: %s in %s (detect nsfw)", err, m.RootRelName())
		return false
	} else {
		if nsfwLabels.NSFW(nsfw.ThresholdHigh) {
			log.Warnf("index: %s might contain offensive content", clean.Log(m.RelName(Config().OriginalsPath())))
			return true
		}
	}

	return false
}
