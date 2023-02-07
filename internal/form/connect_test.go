package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnect_Invalid(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		f := Connect{}
		assert.True(t, f.Invalid())
	})
	t.Run("Invalid", func(t *testing.T) {
		f := Connect{Token: "1mna-2a4t-8729"}
		assert.True(t, f.Invalid())
	})
	t.Run("Valid", func(t *testing.T) {
		f := Connect{Token: "q85v-196o-7eb4"}
		assert.False(t, f.Invalid())
	})
}
