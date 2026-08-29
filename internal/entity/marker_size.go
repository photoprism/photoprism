package entity

import (
	"math"

	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/internal/thumb/crop"
	"github.com/photoprism/photoprism/pkg/clean"
)

// ClusterSizeCond builds the size bar automatic clustering applies, as one expression every query
// shares. It reads thumb_size, the extent of the image an embedding was sampled from, and falls
// back to size, which is a lower bound on it: Fit720 is the narrowest rendition a crop is drawn
// from, so a marker clearing the bar there cleared it in whatever rendition was actually used.
func ClusterSizeCond(alias string, floor int) (string, []any) {
	size, thumbSize := "size", "thumb_size"

	if alias = clean.SqlAlias(alias); alias != "" {
		size, thumbSize = alias+".size", alias+".thumb_size"
	}

	if floor < 1 {
		// No size filter at all, which is what a caller counting every marker asks for.
		return "1 = 1", nil
	}

	return "CASE WHEN " + thumbSize + " >= 1 THEN " + thumbSize + " ELSE " + size + " END >= ?", []any{floor}
}

// ClusterSizeOf returns the extent a marker is judged by, which is what its embedding was sampled
// from when that was recorded and the detection-thumbnail size otherwise.
func (m *Marker) ClusterSizeOf() int {
	if m == nil {
		return -1
	} else if m.ThumbSize >= 1 {
		return m.ThumbSize
	}

	return m.Size
}

// MarkerThumbSize returns the extent an area covers in an image of the given width, which is what
// an embedding sampled from that image saw. It scales the detection-thumbnail size the same way
// face.Face.SetThumbSize does, so a re-crop and a re-detection record the number in one unit.
func MarkerThumbSize(area crop.Area, file File, srcWidth int) int {
	size := MarkerSize(area, file)

	if srcWidth < 1 || size < 1 {
		return -1
	}

	w, _ := thumb.Sizes[thumb.Fit720].Fitted(file.FileWidth, file.FileHeight)

	if w < 1 {
		return -1
	}

	return max(1, int(math.Round(float64(size)*float64(srcWidth)/float64(w))))
}
