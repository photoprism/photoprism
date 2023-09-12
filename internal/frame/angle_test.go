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
	t.Run("MaxTooLow", func(t *testing.T) {
		e := float64(-35)
		a := RandomAngle(e)
		assert.GreaterOrEqual(t, a, e)
		assert.LessOrEqual(t, a, -1*e)
	})
	t.Run("MaxTooHigh", func(t *testing.T) {
		e := float64(200)
		a := RandomAngle(e)
		assert.LessOrEqual(t, a, e)
		assert.GreaterOrEqual(t, a, -1*e)
	})
}
