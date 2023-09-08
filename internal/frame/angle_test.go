package frame

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomAngle(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		for i := 0; i < 50; i++ {
			e := float64(i)
			a := RandomAngle(e)
			t.Logf("%f => %f", e, a)
			assert.LessOrEqual(t, a, e)
			assert.GreaterOrEqual(t, a, -1*e)
		}
	})
}
