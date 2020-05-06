package form

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSelection_Empty(t *testing.T) {
	t.Run("not empty", func(t *testing.T) {
		sel := Selection{Photos: []string{"foo", "bar"}, Albums: []string{}, Labels: []string{}}
		assert.Equal(t, false, sel.Empty())
	})
	t.Run("empty", func(t *testing.T) {
		sel := Selection{Photos: []string{}, Albums: []string{}, Labels: []string{}}
		assert.Equal(t, true, sel.Empty())
	})
}
