package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/thumb/crop"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestClusterSizeCond(t *testing.T) {
	t.Run("PrefersTheSampledExtent", func(t *testing.T) {
		cond, args := ClusterSizeCond("", 112)

		assert.Contains(t, cond, "thumb_size")
		assert.Contains(t, cond, "size")
		assert.Equal(t, []any{112}, args)
	})
	t.Run("Aliased", func(t *testing.T) {
		cond, _ := ClusterSizeCond("m2", 112)

		assert.Contains(t, cond, "m2.thumb_size")
		assert.Contains(t, cond, "m2.size")
	})
	t.Run("NoFloorSelectsEverything", func(t *testing.T) {
		// Zero asks for no size filter at all, which is what a caller counting every marker wants.
		cond, args := ClusterSizeCond("", 0)

		assert.Equal(t, "1 = 1", cond)
		assert.Nil(t, args)
	})
	t.Run("RejectsAnInjectedAlias", func(t *testing.T) {
		cond, _ := ClusterSizeCond("m2; DROP TABLE markers", 112)

		assert.NotContains(t, cond, "DROP")
	})
}

// TestMarker_ClusterSizeOf pins which extent the bar reads. The two columns measure different
// images, so a marker that clears one may not clear the other.
func TestMarker_ClusterSizeOf(t *testing.T) {
	t.Run("SampledExtentWins", func(t *testing.T) {
		m := &Marker{Size: 60, ThumbSize: 150}
		assert.Equal(t, 150, m.ClusterSizeOf())
	})
	t.Run("FallsBackToTheDetectionSize", func(t *testing.T) {
		// -1 is what a marker with no embedding carries, and 0 is what GORM leaves on insert.
		assert.Equal(t, 60, (&Marker{Size: 60, ThumbSize: -1}).ClusterSizeOf())
		assert.Equal(t, 60, (&Marker{Size: 60, ThumbSize: 0}).ClusterSizeOf())
	})
	t.Run("AMeasuredValueIsUsedHoweverSmall", func(t *testing.T) {
		// 1 is a measurement, not a sentinel, so it must not fall through to the larger size.
		assert.Equal(t, 5, (&Marker{Size: 900, ThumbSize: 5}).ClusterSizeOf())
	})
	t.Run("Nil", func(t *testing.T) {
		assert.Equal(t, -1, (*Marker)(nil).ClusterSizeOf())
	})
}

// TestMarker_ThumbSizeSettled covers the three states the column holds, only one of which is worth
// sampling again. All three read as absent to the size bar, which is what makes the difference easy
// to lose: it exists so a migration that re-embeds for a missing extent terminates.
func TestMarker_ThumbSizeSettled(t *testing.T) {
	t.Run("Measured", func(t *testing.T) {
		assert.True(t, (&Marker{ThumbSize: 112}).ThumbSizeSettled())
		assert.True(t, (&Marker{ThumbSize: 1}).ThumbSizeSettled(), "1 is a measurement, not a sentinel")
	})
	t.Run("TriedAndCouldNotMeasure", func(t *testing.T) {
		assert.True(t, (&Marker{ThumbSize: ThumbSizeUnmeasured}).ThumbSizeSettled())
		assert.Equal(t, 60, (&Marker{Size: 60, ThumbSize: ThumbSizeUnmeasured}).ClusterSizeOf(),
			"the bar still falls back, or recording the attempt would change what clusters")
	})
	t.Run("NeverSampled", func(t *testing.T) {
		// -1 is what a marker with no embedding carries and 0 is what GORM leaves on insert.
		assert.False(t, (&Marker{ThumbSize: -1}).ThumbSizeSettled())
		assert.False(t, (&Marker{ThumbSize: 0}).ThumbSizeSettled())
	})
	t.Run("Nil", func(t *testing.T) {
		assert.False(t, (*Marker)(nil).ThumbSizeSettled())
	})
}

// TestThumbSizeUnsettledCond pins that the SQL predicate selects the same rows as the Go one. The
// migration decides through Marker.ThumbSizeSettled and FaceMigrationRecropMarkers prices through
// this, so a divergence would quote work no run performs.
func TestThumbSizeUnsettledCond(t *testing.T) {
	cond := ThumbSizeUnsettledCond()

	assert.Contains(t, cond, "thumb_size IS NULL")
	assert.Contains(t, cond, "thumb_size <> -2")

	for _, size := range []int{-2, -1, 0, 1, 112} {
		uid := rnd.GenerateUID('m')
		m := &Marker{MarkerUID: uid, MarkerType: MarkerFace, MarkerSrc: SrcImage, ThumbSize: size}

		require.NoError(t, Db().Create(m).Error)
		t.Cleanup(func() { UnscopedDb().Delete(&Marker{}, "marker_uid = ?", uid) })

		var found int
		require.NoError(t, UnscopedDb().Model(&Marker{}).
			Where("marker_uid = ?", uid).Where(cond).Count(&found).Error)

		assert.Equal(t, m.ThumbSizeSettled(), found == 0, "thumb_size %d", size)
	}
}

// TestMarker_ClusterableByThumbSize is the case the column exists for: a face too small in the
// detection thumbnail, but sampled at full template size from a wider rendition.
func TestMarker_ClusterableByThumbSize(t *testing.T) {
	restore := face.ClusterSizeThreshold
	t.Cleanup(func() { face.ClusterSizeThreshold = restore })

	face.ClusterSizeThreshold = 112

	t.Run("AdmittedByTheSampledExtent", func(t *testing.T) {
		m := &Marker{MarkerType: MarkerFace, Size: 60, ThumbSize: 150, Score: 100}
		assert.True(t, m.Clusterable(), "a face sampled at 150 px is not invented by interpolation")
	})
	t.Run("RefusedWhenBothAreBelow", func(t *testing.T) {
		m := &Marker{MarkerType: MarkerFace, Size: 60, ThumbSize: 90, Score: 100}
		assert.False(t, m.Clusterable())
	})
	t.Run("UnknownFallsBackAndIsAdmittedOnSize", func(t *testing.T) {
		// size is a lower bound on the sampled extent, so clearing it is sound without provenance.
		m := &Marker{MarkerType: MarkerFace, Size: 120, ThumbSize: -1, Score: 100}
		assert.True(t, m.Clusterable())
	})
}

func TestMarkerThumbSize(t *testing.T) {
	file := File{FileWidth: 3000, FileHeight: 2000}
	area := crop.Area{Name: "face", X: 0.4, Y: 0.4, W: 0.1, H: 0.1}

	t.Run("ScalesTheDetectionSize", func(t *testing.T) {
		// Fit720 bounds a 3:2 original to 720x480, so the same area is 2.5x larger at 1800.
		assert.Equal(t, 2*MarkerSize(area, file)+MarkerSize(area, file)/2, MarkerThumbSize(area, file, 1800))
	})
	t.Run("UnknownSource", func(t *testing.T) {
		assert.Equal(t, -1, MarkerThumbSize(area, file, 0))
	})
	t.Run("UnknownFileSize", func(t *testing.T) {
		assert.Equal(t, -1, MarkerThumbSize(area, File{}, 1800))
	})
}
