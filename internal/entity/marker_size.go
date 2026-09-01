package entity

import (
	"math"
	"strconv"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/internal/thumb/crop"
	"github.com/photoprism/photoprism/pkg/clean"
)

// ClusterSizeCond builds the size bar automatic clustering applies, as one expression every query
// shares. It reads thumb_size, the extent of the image an embedding was sampled from, falling back
// to size, which is a lower bound on it for every marker detected on Fit720 - the only detection
// rendition this code writes, and the narrowest a crop is ever drawn from.
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

// ThumbSizeUnmeasured marks a marker a sampling already tried and could not measure an extent for,
// as distinct from one nothing has sampled yet. Both read as absent, since the bars compare against
// 1, and telling them apart is what lets a migration filling the column terminate.
const ThumbSizeUnmeasured = -2

// EmbedUpscaledUnknown marks a marker whose embedding was sampled without the share of the crop
// its source supplied being measurable, as distinct from the -1 of a marker nothing has sampled.
// Negative rather than zero, and the same value ThumbSizeUnmeasured uses: GORM omits a zero field
// on insert where the column has a default, so a zero would read back as never sampled.
const EmbedUpscaledUnknown = -2

// MarkerEmbedUpscaled returns what a marker records for a face an embedding was generated for:
// the percentage of the crop width its source supplied, or EmbedUpscaledUnknown where the
// sampling could not measure one. Stored rather than derived, because the crop width is
// face.CropSize or a model's own input today, so a later computation changes meaning with them.
func MarkerEmbedUpscaled(f face.Face) int {
	if f.EmbedUpscaled > 0 {
		return f.EmbedUpscaled
	}

	return EmbedUpscaledUnknown
}

// ThumbSizeSettled reports whether a sampling has answered for this marker's extent, either by
// measuring one or by trying and failing. Only an unsettled marker is worth sampling again.
func (m *Marker) ThumbSizeSettled() bool {
	return m != nil && (m.ThumbSize >= 1 || m.ThumbSize == ThumbSizeUnmeasured)
}

// ThumbSizeUnsettledCond selects the markers no sampling has answered for, as one expression the
// migration and the plan that prices it share.
func ThumbSizeUnsettledCond() string {
	return "thumb_size IS NULL OR (thumb_size < 1 AND thumb_size <> " + strconv.Itoa(ThumbSizeUnmeasured) + ")"
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
