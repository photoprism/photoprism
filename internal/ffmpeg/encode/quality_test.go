package encode

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstantQuality(t *testing.T) {
	t.Run("Defaults", func(t *testing.T) {
		assert.Equal(t, "100", QvQuality(BestQuality))
		assert.Equal(t, "50", QvQuality(DefaultQuality))
		assert.Equal(t, "1", QvQuality(WorstQuality))
		assert.Equal(t, "50", QvQuality(-1))
		assert.Equal(t, "1", QvQuality(0))
		assert.Equal(t, "100", QvQuality(134))
	})
}

func TestGlobalQuality(t *testing.T) {
	t.Run("Defaults", func(t *testing.T) {
		assert.Equal(t, "25", GlobalQuality(0))
		assert.Equal(t, "1", GlobalQuality(BestQuality))
		assert.Equal(t, "25", GlobalQuality(DefaultQuality))
		assert.Equal(t, "49", GlobalQuality(WorstQuality))
		assert.Equal(t, "1", GlobalQuality(102))
	})
}

func TestCrfQuality(t *testing.T) {
	t.Run("Defaults", func(t *testing.T) {
		assert.Equal(t, "25", CrfQuality(0))
		assert.Equal(t, "0", CrfQuality(BestQuality))
		assert.Equal(t, "25", CrfQuality(DefaultQuality))
		assert.Equal(t, "49", CrfQuality(WorstQuality))
		assert.Equal(t, "0", CrfQuality(102))
	})
}

func TestQpQuality(t *testing.T) {
	t.Run("Defaults", func(t *testing.T) {
		assert.Equal(t, "25", QpQuality(0))
		assert.Equal(t, "0", QpQuality(BestQuality))
		assert.Equal(t, "25", QpQuality(DefaultQuality))
		assert.Equal(t, "49", QpQuality(WorstQuality))
		assert.Equal(t, "0", QpQuality(150))
	})
}

func TestCqQuality(t *testing.T) {
	t.Run("Defaults", func(t *testing.T) {
		assert.Equal(t, "31", CqQuality(0))
		assert.Equal(t, "1", CqQuality(BestQuality))
		assert.Equal(t, "31", CqQuality(DefaultQuality))
		assert.Equal(t, "50", CqQuality(WorstQuality))
		assert.Equal(t, "1", CqQuality(123))
	})
	t.Run("Scale", func(t *testing.T) {
		assert.Equal(t, "13", CqQuality(80))
		assert.Equal(t, "43", CqQuality(30))
		assert.Equal(t, "50", CqQuality(15))
		assert.Equal(t, "50", CqQuality(2))
	})
}
