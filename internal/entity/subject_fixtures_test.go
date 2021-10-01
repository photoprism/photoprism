package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubjectMap_Get(t *testing.T) {
	t.Run("get existing subject", func(t *testing.T) {
		r := SubjectFixtures.Get("joe-biden")
		assert.Equal(t, "Joe Biden", r.SubjName)
		assert.IsType(t, Subject{}, r)
	})
	t.Run("get not existing subject", func(t *testing.T) {
		r := SubjectFixtures.Get("monstera")
		assert.Equal(t, "", r.SubjName)
		assert.IsType(t, Subject{}, r)
	})
}

func TestSubjectMap_Pointer(t *testing.T) {
	t.Run("get existing subject", func(t *testing.T) {
		r := SubjectFixtures.Pointer("joe-biden")
		assert.Equal(t, "Joe Biden", r.SubjName)
		assert.IsType(t, &Subject{}, r)
	})
	t.Run("get not existing subject", func(t *testing.T) {
		r := SubjectFixtures.Pointer("monstera")
		assert.Equal(t, "", r.SubjName)
		assert.IsType(t, &Subject{}, r)
	})
}
