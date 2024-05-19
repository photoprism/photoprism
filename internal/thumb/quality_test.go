package thumb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJpegQuality(t *testing.T) {
	t.Run("Large", func(t *testing.T) {
		assert.Equal(t, JpegQualityDefault, JpegQuality(100, 500))
	})
	t.Run("Small", func(t *testing.T) {
		assert.Equal(t, JpegQualityDefault-5, JpegQuality(50, 150))
	})
}

func TestJpegQualitySmall(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		assert.Equal(t, JpegQualityDefault-5, JpegQualitySmall())
	})
}

func TestParseQuality(t *testing.T) {
	t.Run("Max", func(t *testing.T) {
		assert.Equal(t, QualityMax, ParseQuality("max"))
	})
	t.Run("Min", func(t *testing.T) {
		assert.Equal(t, QualityMin, ParseQuality("min"))
	})
	t.Run("bad", func(t *testing.T) {
		assert.Equal(t, QualityMedium, ParseQuality("bad"))
	})
	t.Run("low", func(t *testing.T) {
		assert.Equal(t, QualityLow, ParseQuality("low"))
	})
	t.Run("high", func(t *testing.T) {
		assert.Equal(t, QualityHigh, ParseQuality("high"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, QualityMedium, ParseQuality(""))
		assert.Equal(t, QualityMedium, ParseQuality("             "))
	})
	t.Run("Default", func(t *testing.T) {
		assert.Equal(t, QualityMedium, ParseQuality("default"))
	})
	t.Run("Medium", func(t *testing.T) {
		assert.Equal(t, QualityMedium, ParseQuality("medium"))
		assert.Equal(t, QualityMedium, ParseQuality("   \t       medium \n\r"))
		assert.Equal(t, QualityMedium, ParseQuality("MEDIUM"))
	})
	t.Run("Good", func(t *testing.T) {
		assert.Equal(t, QualityHigh, ParseQuality("Good"))
		assert.Equal(t, QualityHigh, ParseQuality("GOOD"))
	})
	t.Run("Best", func(t *testing.T) {
		assert.Equal(t, QualityMax, ParseQuality("Best"))
	})
	t.Run("Ultra", func(t *testing.T) {
		assert.Equal(t, QualityMax, ParseQuality("ultra"))
	})
	t.Run("0", func(t *testing.T) {
		assert.Equal(t, QualityMin, ParseQuality("0"))
	})
	t.Run("1", func(t *testing.T) {
		assert.Equal(t, QualityMin, ParseQuality("1"))
	})
	t.Run("2", func(t *testing.T) {
		assert.Equal(t, QualityLow, ParseQuality("2"))
	})
	t.Run("3", func(t *testing.T) {
		assert.Equal(t, QualityMedium, ParseQuality("3"))
	})
	t.Run("4", func(t *testing.T) {
		assert.Equal(t, QualityHigh, ParseQuality("4"))
	})
	t.Run("5", func(t *testing.T) {
		assert.Equal(t, QualityMax, ParseQuality("5"))
	})
	t.Run("6", func(t *testing.T) {
		assert.Equal(t, QualityMax, ParseQuality("6"))
	})
	t.Run("50", func(t *testing.T) {
		assert.Equal(t, Quality(50), ParseQuality("50"))
	})
	t.Run("66", func(t *testing.T) {
		assert.Equal(t, Quality(66), ParseQuality("66"))
	})
	t.Run("77", func(t *testing.T) {
		assert.Equal(t, Quality(77), ParseQuality("77"))
	})
	t.Run("89", func(t *testing.T) {
		assert.Equal(t, Quality(89), ParseQuality("89"))
	})
	t.Run("90", func(t *testing.T) {
		assert.Equal(t, Quality(90), ParseQuality("90"))
	})
	t.Run("100", func(t *testing.T) {
		assert.Equal(t, Quality(100), ParseQuality("100"))
	})
}

func TestQuality_String(t *testing.T) {
	assert.Equal(t, "91", QualityMax.String())
	assert.Equal(t, "87", QualityHigh.String())
	assert.Equal(t, "83", QualityMedium.String())
	assert.Equal(t, "79", QualityLow.String())
	assert.Equal(t, "70", QualityMin.String())
}
