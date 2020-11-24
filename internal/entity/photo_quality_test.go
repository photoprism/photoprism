package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhoto_QualityScore(t *testing.T) {
	t.Run("PhotoFixture19800101_000002_D640C559", func(t *testing.T) {
		assert.Equal(t, 4, PhotoFixtures.Pointer("19800101_000002_D640C559").QualityScore())
	})
	t.Run("PhotoFixturePhoto01 - favorite true - taken at before 2008", func(t *testing.T) {
		assert.Equal(t, 7, PhotoFixtures.Pointer("Photo01").QualityScore())
	})
	t.Run("PhotoFixturePhoto06 - taken at after 2012 - resolution 2", func(t *testing.T) {
		assert.Equal(t, 4, PhotoFixtures.Pointer("Photo06").QualityScore())
	})
	t.Run("PhotoFixturePhoto07 - score < 3 bit edited", func(t *testing.T) {
		assert.Equal(t, 3, PhotoFixtures.Pointer("Photo07").QualityScore())
	})
	t.Run("PhotoFixturePhoto15 - description with blacklist", func(t *testing.T) {
		assert.Equal(t, 2, PhotoFixtures.Pointer("Photo15").QualityScore())
	})
}
