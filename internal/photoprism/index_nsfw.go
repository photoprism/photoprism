package photoprism

import (
	"github.com/photoprism/photoprism/internal/ai/nsfw"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/media"
)

// IsNsfw returns true if media file might be offensive and detection is enabled.
func (ind *Index) IsNsfw(m *MediaFile) bool {
	filename, err := m.Thumbnail(Config().ThumbCachePath(), thumb.Fit720)

	if err != nil {
		log.Error(err)
		return false
	}

	if results, modelErr := vision.Nsfw([]string{filename}, media.SrcLocal); modelErr != nil {
		log.Errorf("vision: %s in %s (detect nsfw)", modelErr, m.RootRelName())
		return false
	} else if len(results) < 1 {
		log.Errorf("vision: nsfw model returned no result for %s", m.RootRelName())
		return false
	} else if results[0].IsNsfw(nsfw.ThresholdHigh) {
		log.Warnf("vision: %s might contain offensive content", clean.Log(m.RelName(Config().OriginalsPath())))
		return true
	}

	return false
}
