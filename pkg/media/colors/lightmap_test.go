package colors

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLightMap_Hex(t *testing.T) {
	lum := []Luminance{1, 16, 2, 4, 15, 16, 1, 0, 8}
	lMap := LightMap(lum)
	assert.Equal(t, "1F24FF108", lMap.Hex())

}

func TestLightMap_Diff(t *testing.T) {
	t.Run("Random", func(t *testing.T) {
		lum := []Luminance{1, 16, 2, 4, 15, 16, 1, 0, 8}
		lMap := LightMap(lum)
		result := lMap.Diff()

		if result != 845 {
			t.Errorf("result should be 845: %d", result)
		}
	})
	t.Run("Empty", func(t *testing.T) {
		var lum []Luminance
		lMap := LightMap(lum)
		result := lMap.Diff()

		if result != 0 {
			t.Errorf("result should be 0: %d", result)
		}
	})
	t.Run("One", func(t *testing.T) {
		lum := []Luminance{0}
		lMap := LightMap(lum)
		result := lMap.Diff()

		if result != 0 {
			t.Errorf("result should be 0: %d", result)
		}
	})
	t.Run("Same", func(t *testing.T) {
		lum := []Luminance{1, 1, 1, 1, 1, 1, 1, 1, 1}
		lMap := LightMap(lum)
		result := lMap.Diff()

		if result != 1023 {
			t.Errorf("result should be 1023: %d", result)
		}
	})
	t.Run("Similar", func(t *testing.T) {
		lum := []Luminance{1, 1, 1, 1, 1, 1, 1, 1, 2}
		lMap := LightMap(lum)
		result := lMap.Diff()

		if result != 1023 {
			t.Errorf("result should be 1023: %d", result)
		}
	})
	t.Run("Happy", func(t *testing.T) {
		m1 := LightMap{8, 13, 7, 2, 2, 3, 6, 3, 4}
		d1 := m1.Diff()
		t.Log(strconv.FormatUint(uint64(uint16(d1)), 2)) //nolint:gosec // test logging
		m2 := LightMap{8, 13, 7, 3, 1, 3, 5, 3, 4}
		d2 := m2.Diff()
		t.Log(strconv.FormatUint(uint64(uint16(d2)), 2)) //nolint:gosec // test logging
		m3 := LightMap{9, 13, 7, 8, 2, 4, 5, 3, 4}
		d3 := m3.Diff()
		t.Log(strconv.FormatUint(uint64(uint16(d3)), 2)) //nolint:gosec // test logging
		m4 := LightMap{9, 13, 7, 7, 2, 4, 6, 2, 3}
		d4 := m4.Diff()
		t.Log(strconv.FormatUint(uint64(uint16(d4)), 2)) //nolint:gosec // test logging

		t.Logf("values: %d, %d, %d, %d", d1, d2, d3, d4)
	})
}
