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
			assert.False(t, e[i].KidsFace())
			assert.False(t, e[i].Ignored())
		}
	})
	t.Run("Kids", func(t *testing.T) {
		e := RandomEmbeddings(2, KidsFace)
		for i := range e {
			assert.False(t, e[i].Ignored())
			assert.True(t, e[i].KidsFace())
		}
	})
	t.Run("Ignored", func(t *testing.T) {
		e := RandomEmbeddings(2, IgnoredFace)
		for i := range e {
			assert.True(t, e[i].Ignored())
			assert.False(t, e[i].KidsFace())
		}
	})
}
