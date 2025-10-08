package face

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomDist(t *testing.T) {
	t.Run("Range", func(t *testing.T) {
		d := RandomDist()
		assert.GreaterOrEqual(t, d, 0.1)
		assert.LessOrEqual(t, d, 1.5)
	})
}

func TestRandomEmbeddings(t *testing.T) {
	t.Run("Regular", func(t *testing.T) {
		e := RandomEmbeddings(2, RegularFace)
		for i := range e {
			// t.Logf("embedding: %#v", e[i])
			assert.False(t, e[i].IsChild())
			assert.False(t, e[i].IsBackground())
		}
	})
	t.Run("Children", func(t *testing.T) {
		e := RandomEmbeddings(2, ChildrenFace)
		for i := range e {
			assert.False(t, e[i].IsBackground())
			assert.True(t, e[i].IsChild())
		}
	})
	t.Run("Background", func(t *testing.T) {
		e := RandomEmbeddings(2, BackgroundFace)
		for i := range e {
			assert.True(t, e[i].IsBackground())
			assert.False(t, e[i].IsChild())
		}
	})
}
