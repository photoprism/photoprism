package header

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestID(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", ID(""))
	})
	t.Run("ID", func(t *testing.T) {
		assert.Equal(t, "8045b8926260372bb9afd65e451633ec4ab08817", ID("8045b8926260372bb9afd65e451633ec4ab08817"))
	})
	t.Run("BearerToken", func(t *testing.T) {
		assert.Equal(t, "Bearer S0VLU0UhIExFQ0tFUiEK", ID("Bearer S0VLU0UhIExFQ0tFUiEK"))
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.Equal(t, "@#$$#@#$_8oags JKHUGIUYG gsr\":", ID("  !@#$%^&*(*&^%$#@#$%^&*()_8oags JKHUGIUYG gsr<>?\":{}|{})"))
	})
}
