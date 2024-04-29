package api

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewResponse(t *testing.T) {
	t.Run("404", func(t *testing.T) {
		r := NewResponse(404, errors.New("not found"), "details")
		assert.Equal(t, 404, r.Code)
		assert.Equal(t, "details", r.Details)
	})
}
