package query

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TODO test non empty case
func TestErrors(t *testing.T) {
	t.Run("NotExisting", func(t *testing.T) {
		errors, err := Errors(1000, 0, "notexistingErrorString")
		if err != nil {
			t.Fatal(err)
		}
		assert.Empty(t, errors)
	})
	t.Run("Error", func(t *testing.T) {
		errors, err := Errors(1000, 0, "errors")
		if err != nil {
			t.Fatal(err)
		}
		assert.Empty(t, errors)
	})
	t.Run("Warning", func(t *testing.T) {
		errors, err := Errors(1000, 0, "warnings")
		if err != nil {
			t.Fatal(err)
		}
		assert.Empty(t, errors)
	})

}

func TestDeleteErrors(t *testing.T) {
	t.Run("OneError", func(t *testing.T) {
		expected := "OneError Testing Message"
		if err := Db().Create(&entity.Error{ID: 999999, ErrorTime: time.Now(), ErrorLevel: "debug", ErrorMessage: expected}).Error; err != nil {
			t.Fatal(err)
		}

		err := DeleteErrors()
		assert.Empty(t, err)

		errors, err := Errors(1000, 0, "debug")
		if err != nil {
			t.Fatal(err)
		}
		assert.Empty(t, errors)
	})

}

func TestErrors_Literal(t *testing.T) {
	base := "zzl" + rnd.Base36(5)

	for _, msg := range []string{"message " + base + "_a", "message " + base + "Xa"} {
		m := &entity.Error{ErrorTime: time.Now().UTC(), ErrorLevel: "error", ErrorMessage: msg}
		require.NoError(t, entity.Db().Create(m).Error)
		t.Cleanup(func() { _ = entity.UnscopedDb().Delete(m).Error })
	}

	results, err := Errors(100, 0, base+"_a")
	require.NoError(t, err)

	if assert.Len(t, results, 1) {
		assert.Equal(t, "message "+base+"_a", results[0].ErrorMessage)
	}
}
