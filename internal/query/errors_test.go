package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TODO test non empty case
func TestErrors(t *testing.T) {
	t.Run("not existing", func(t *testing.T) {
		errors, err := Errors(1000, 0, "notexistingErrorString")
		if err != nil {
			t.Fatal(err)
		}
		assert.Empty(t, errors)
	})
	t.Run("error", func(t *testing.T) {
		errors, err := Errors(1000, 0, "errors")
		if err != nil {
			t.Fatal(err)
		}
		assert.Empty(t, errors)
	})
	t.Run("warning", func(t *testing.T) {
		errors, err := Errors(1000, 0, "warnings")
		if err != nil {
			t.Fatal(err)
		}
		assert.Empty(t, errors)
	})

}
