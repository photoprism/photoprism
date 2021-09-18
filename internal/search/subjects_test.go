package search

import (
	"testing"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"

	"github.com/stretchr/testify/assert"
)

func TestSubjects(t *testing.T) {
	t.Run("FindAll", func(t *testing.T) {
		results, err := Subjects(form.SubjectSearch{Type: entity.SubjPerson})
		assert.NoError(t, err)
		// t.Logf("Subjects: %#v", results)
		assert.LessOrEqual(t, 3, len(results))
	})
}
