package geo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDist(t *testing.T) {
	t.Run("BerlinShanghai", func(t *testing.T) {
		berlin := Position{52.5243700, 13.4105300}
		shanghai := Position{31.2222200, 121.4580600}

		result := Dist(berlin, shanghai)

		assert.Equal(t, 8396, int(result))
	})
}
