package entity

import (
	"math"
	"strconv"

	"github.com/photoprism/photoprism/internal/ai/face"

	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/internal/thumb/crop"
	"github.com/photoprism/photoprism/pkg/clean"
)

// ClusterSizeCond builds the bar automatic clustering applies to the pixels an embedding rests on,
// as one expression every query shares. It reads thumb_size, the extent of the image an embedding
// was sampled from, falling back to size, which is a lower bound on it for every marker detected
// on Fit720 - the only detection rendition this code writes, and the narrowest a crop is ever
// drawn from.
//
// The detail condition is part of it rather than beside it, so clustering, the counts faces status
// reports and the migration plan cannot disagree about what clusterable means.
func ClusterSizeCond(alias string, floor int) (string, []any) {
	size, thumbSize := "size", "thumb_size"

	if alias = clean.SqlAlias(alias); alias != "" {
		size, thumbSize = alias+".size", alias+".thumb_size"
	}

	if floor < 1 {
		// No size bar at all, which is what a caller counting every marker asks for. The detail
		// condition is not a size bar and is not configurable, so it stays.
		return EmbedDetailCond(alias), nil
	}

	return "(CASE WHEN " + thumbSize + " >= 1 THEN " + thumbSize + " ELSE " + size + " END >= ?) AND " +
		EmbedDetailCond(alias), []any{floor}
}

// EmbedDetailCond selects the markers whose embedding was drawn from a source that supplied the
// whole crop, and those no sampling has measured a share for.
//
// The size bar already keeps upscaled crops out at the shipped face-cluster-size - for an aligned
// marker full detail is arithmetically the same as thumb_size >= 112 - but an operator can lower
// that bar, and this holds whatever they set it to.
//
// ⚠ Not purely redundant even at the default: the unaligned fallback measures against the 160 px
// face.CropSize box rather than the model's 112, so this newly excludes unaligned markers in the
// 112-159 band. That is the weakest corner of the data - no pose normalization and an upscaled
// crop - and excluding it is deliberate.
//
// ⚠ NULL is named because NULL < 1 is NULL rather than true, and every marker written before the
// column existed is NULL or -1. Without that branch this would exclude an entire library.
func EmbedDetailCond(alias string) string {
	detail := "embed_detail"

	if alias = clean.SqlAlias(alias); alias != "" {
		detail = alias + ".embed_detail"
	}

	return "(" + detail + " IS NULL OR " + detail + " < 1 OR " + detail + " >= " + strconv.Itoa(face.EmbedDetailFull) + ")"
}

// ThumbSizeUnmeasured marks a marker a sampling already tried and could not measure an extent for,
// as distinct from one nothing has sampled yet. Both read as absent, since the bars compare against
// 1, and telling them apart is what lets a migration filling the column terminate.
const ThumbSizeUnmeasured = -2

// EmbedDetailUnknown marks a marker a migration sampled without the share of the crop its
// source supplied being measurable, beside the ThumbSizeUnmeasured its partner column records for
// the same pass. Negative rather than zero for the same reason as that one: GORM omits a zero
// field on insert where the column has a default, so a zero would read back as never sampled.
const EmbedDetailUnknown = -2

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
