package query

import (
	"testing"

	"github.com/photoprism/photoprism/internal/form"

	"github.com/photoprism/photoprism/internal/entity"

	"github.com/stretchr/testify/assert"
)

func TestSubjectSearch(t *testing.T) {
	t.Run("FindAll", func(t *testing.T) {
		results, err := SubjectSearch(form.SubjectSearch{Type: entity.SubjectPerson})
		assert.NoError(t, err)
		assert.LessOrEqual(t, 3, len(results))
	})

}
