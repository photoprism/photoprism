package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFeedback_Empty(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		f := Feedback{}
		assert.True(t, f.Empty())
	})
	t.Run("False", func(t *testing.T) {
		f := Feedback{Message: "I found a bug", Category: "Bug Report", UserEmail: "test@test.com"}
		assert.False(t, f.Empty())
	})
	t.Run("False", func(t *testing.T) {
		if f, err := NewFeedback(""); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, f.Empty())
		}
	})
}
