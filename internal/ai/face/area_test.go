package face

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/thumb/crop"
)

var area1 = NewArea("face1", 400, 250, 200)
var area2 = NewArea("face2", 100, 100, 50)
var area3 = NewArea("face3", 900, 500, 25)
var area4 = NewArea("face4", 110, 110, 60)

func TestArea_TopLeft(t *testing.T) {
	t.Run("Area1", func(t *testing.T) {
		x, y := area1.TopLeft()
		assert.Equal(t, 300, x)
		assert.Equal(t, 150, y)
	})
	t.Run("Area2", func(t *testing.T) {
		x, y := area2.TopLeft()
		assert.Equal(t, 75, x)
		assert.Equal(t, 75, y)
	})
	t.Run("Area3", func(t *testing.T) {
		x, y := area3.TopLeft()
		assert.Equal(t, 888, x)
		assert.Equal(t, 488, y)
	})
	t.Run("Area4", func(t *testing.T) {
		x, y := area4.TopLeft()
		assert.Equal(t, 80, x)
		assert.Equal(t, 80, y)
	})
}

func TestArea_Relative(t *testing.T) {
	t.Parallel()

	base := NewArea("base", 100, 100, 20)
	a := NewArea("child", 150, 130, 40)

	rel := a.Relative(base, 200, 400)

	assert.Equal(t, "child", rel.Name)
	assert.InDelta(t, 0.075, rel.X, 0.0001)
	assert.InDelta(t, 0.25, rel.Y, 0.0001)
	assert.InDelta(t, 0.1, rel.W, 0.0001)
	assert.InDelta(t, 0.2, rel.H, 0.0001)
}

func TestArea_RelativeZeroDimensions(t *testing.T) {
	t.Parallel()

	base := NewArea("base", 0, 0, 0)
	a := NewArea("child", 10, 10, 5)

	rel := a.Relative(base, 0, 0)

	assert.Equal(t, "child", rel.Name)
	assert.InDelta(t, 1.0, rel.X, 0.0001)
	assert.InDelta(t, 1.0, rel.Y, 0.0001)
	assert.InDelta(t, 1.0, rel.W, 0.0001)
	assert.InDelta(t, 1.0, rel.H, 0.0001)
}

func TestAreas_Relative(t *testing.T) {
	t.Parallel()

	base := NewArea("base", 100, 100, 20)
	pts := Areas{
		NewArea("left-eye", 110, 120, 10),
		NewArea("right-eye", 110, 130, 10),
	}

	rel := pts.Relative(base, 200, 400)

	require.Len(t, rel, len(pts))

	assert.Equal(t, "left-eye", rel[0].Name)
	assert.InDelta(t, 0.05, rel[0].X, 0.0001)
	assert.InDelta(t, 0.05, rel[0].Y, 0.0001)

	assert.Equal(t, "right-eye", rel[1].Name)
	assert.InDelta(t, 0.075, rel[1].X, 0.0001)
	assert.InDelta(t, 0.05, rel[1].Y, 0.0001)
}

func TestAreas_RelativeEmpty(t *testing.T) {
	t.Parallel()

	var pts Areas

	rel := pts.Relative(NewArea("base", 0, 0, 0), 100, 100)

	assert.Nil(t, rel)
}

var benchmarkAreasResult crop.Areas

func BenchmarkAreasRelative(b *testing.B) {
	base := NewArea("base", 100, 100, 20)
	pts := make(Areas, 128)

	for i := range pts {
		pts[i] = Area{
			Name:  "landmark",
			Row:   base.Row + i%16,
			Col:   base.Col + i%16,
			Scale: 25,
		}
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		benchmarkAreasResult = pts.Relative(base, 200, 400)
	}
}
