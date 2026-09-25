package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeprecatedForceFlag(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		f := DeprecatedForceFlag()

		assert.Equal(t, "force", f.Name)
		assert.Equal(t, []string{"f"}, f.Aliases)
		assert.True(t, f.Hidden, "the alias must not be advertised next to --yes")
		assert.Contains(t, f.Usage, "--yes")
	})
	t.Run("DistinctFromForceFlag", func(t *testing.T) {
		assert.False(t, ForceFlag("").Hidden, "the shared --force flag stays visible")
	})
}
