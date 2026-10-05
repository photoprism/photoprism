package entity

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCheckMakeModel verifies that only make and model names without control characters are accepted.
func TestCheckMakeModel(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		assert.NoError(t, checkMakeModel("Carl Zeiss", "Planar T* 50mm f/1.4 ZE"))
		assert.NoError(t, checkMakeModel("Pentax", "4 38"))
		assert.NoError(t, checkMakeModel("Ricoh", "GR III (ハイライト)"))
	})
	for name, values := range map[string][2]string{
		"TabInMake":      {"A\tB", "Model"},
		"NewlineInModel": {"Make", "Model\nX"},
		"CarriageReturn": {"Make\r", "Model"},
		"Escape":         {"Make", "\x1b[31mModel"},
		"Delete":         {"Make", "Model\x7f"},
		"C1Control":      {"Make\u0085", "Model"},
	} {
		t.Run(name, func(t *testing.T) {
			err := checkMakeModel(values[0], values[1])
			assert.True(t, errors.Is(err, ErrInvalidValue), err)
			assert.EqualError(t, err, "invalid value: make and model must not contain control characters")
		})
	}
}
