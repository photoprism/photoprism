package tz

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPosition(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", Position(0, 0))
	})
	t.Run("EuropeBerlin", func(t *testing.T) {
		assert.Equal(t, "Europe/Berlin", Position(52.472833, 13.407500))
	})
}
