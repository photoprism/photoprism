package thumb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseQuality(t *testing.T) {
	t.Run("Worst", func(t *testing.T) {
		assert.Equal(t, QualityWorst, ParseQuality("worst"))
	})
	t.Run("Lowest", func(t *testing.T) {
		assert.Equal(t, QualityWorst, ParseQuality("lowest"))
	})
	t.Run("bad", func(t *testing.T) {
		assert.Equal(t, QualityBad, ParseQuality("bad"))
	})
	t.Run("low", func(t *testing.T) {
		assert.Equal(t, QualityLow, ParseQuality("low"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, QualityDefault, ParseQuality(""))
		assert.Equal(t, QualityDefault, ParseQuality("             "))
	})
	t.Run("Default", func(t *testing.T) {
		assert.Equal(t, QualityDefault, ParseQuality("default"))
	})
	t.Run("Medium", func(t *testing.T) {
		assert.Equal(t, QualityDefault, ParseQuality("medium"))
		assert.Equal(t, QualityDefault, ParseQuality("   \t       medium \n\r"))
		assert.Equal(t, QualityDefault, ParseQuality("MEDIUM"))
	})
	t.Run("Good", func(t *testing.T) {
		assert.Equal(t, QualityHigh, ParseQuality("Good"))
		assert.Equal(t, QualityHigh, ParseQuality("GOOD"))
	})
	t.Run("Best", func(t *testing.T) {
		assert.Equal(t, QualityBest, ParseQuality("Best"))
	})
	t.Run("Ultra", func(t *testing.T) {
		assert.Equal(t, QualityBest, ParseQuality("ultra"))
	})
	t.Run("0", func(t *testing.T) {
		assert.Equal(t, QualityWorst, ParseQuality("0"))
	})
	t.Run("1", func(t *testing.T) {
		assert.Equal(t, QualityBad, ParseQuality("1"))
	})
	t.Run("2", func(t *testing.T) {
		assert.Equal(t, QualityLow, ParseQuality("2"))
	})
	t.Run("3", func(t *testing.T) {
		assert.Equal(t, QualityDefault, ParseQuality("3"))
	})
	t.Run("4", func(t *testing.T) {
		assert.Equal(t, QualityHigh, ParseQuality("4"))
	})
	t.Run("5", func(t *testing.T) {
		assert.Equal(t, QualityBest, ParseQuality("5"))
	})
	t.Run("6", func(t *testing.T) {
		assert.Equal(t, QualityDefault, ParseQuality("6"))
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
	assert.Equal(t, "95", QualityBest.String())
	assert.Equal(t, "88", QualityHigh.String())
	assert.Equal(t, "82", QualityDefault.String())
	assert.Equal(t, "75", QualityBad.String())

}
