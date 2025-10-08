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
