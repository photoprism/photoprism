package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPriorities_Report(t *testing.T) {
	t.Run("Len", func(t *testing.T) {
		rows, cols := SrcPriority.Report()
		assert.Len(t, cols, 3)
		assert.NotEmpty(t, rows)
	})
}

// TestSrcSubjects pins the sources a subject assignment can have, at their general priorities.
func TestSrcSubjects(t *testing.T) {
	t.Run("Sources", func(t *testing.T) {
		for _, src := range []Src{SrcAuto, SrcMarker, SrcMeta, SrcXmp, SrcBatch, SrcManual} {
			p, ok := SrcSubjects[src]
			assert.True(t, ok, src)
			assert.Equal(t, SrcPriority[src], p, src)
		}
		assert.Len(t, SrcSubjects, 6)
	})
	t.Run("Overrides", func(t *testing.T) {
		for _, src := range []Src{SrcAdmin, SrcVision} {
			_, ok := SrcSubjects[src]
			assert.False(t, ok, src)
		}
	})
}
