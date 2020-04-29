package classify

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLabel_NewLocationLabel(t *testing.T) {
	LocLabel := LocationLabel("locationtest", 23, 1)
	t.Log(LocLabel)
	assert.Equal(t, "location", LocLabel.Source)
	assert.Equal(t, 23, LocLabel.Uncertainty)
	assert.Equal(t, "locationtest", LocLabel.Name)

	t.Run("locationtest / slash", func(t *testing.T) {
		LocLabel := LocationLabel("locationtest / slash", 24, -2)
		t.Log(LocLabel)
		assert.Equal(t, "location", LocLabel.Source)
		assert.Equal(t, 24, LocLabel.Uncertainty)
		assert.Equal(t, "locationtest", LocLabel.Name)
	})

	t.Run("locationtest - minus", func(t *testing.T) {
		LocLabel := LocationLabel("locationtest - minus", 80, -2)
		t.Log(LocLabel)
		assert.Equal(t, "location", LocLabel.Source)
		assert.Equal(t, 80, LocLabel.Uncertainty)
		assert.Equal(t, "locationtest", LocLabel.Name)
	})
}

func TestLabel_Title(t *testing.T) {
	t.Run("locationtest123", func(t *testing.T) {
		LocLabel := LocationLabel("locationtest123", 23, 1)
		assert.Equal(t, "Locationtest123", LocLabel.Title())
	})

	t.Run("Berlin/Neukölln", func(t *testing.T) {
		LocLabel := LocationLabel("berlin/neukölln_hasenheide", 23, 1)
		assert.Equal(t, "Berlin/Neukölln_hasenheide", LocLabel.Title())
	})
}
