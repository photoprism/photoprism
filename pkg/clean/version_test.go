package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	t.Run("MariaDB", func(t *testing.T) {
		assert.Equal(t, "v10.5.1", Version("10.5.1"))
		assert.Equal(t, "v10.5.7", Version("v10.5.7"))
		assert.Equal(t, "v11.4.5", Version("11.4.5-MariaDB-ubu2404"))
		assert.Equal(t, "v10.5.5", Version("MariaDB-1:10.5.5+maria~focal"))
	})
	t.Run("Empty", func(t *testing.T) {
		result := Version("")
		assert.Equal(t, "", result)
	})
	t.Run("Invalid", func(t *testing.T) {
		result := Version("45345.356q636")
		assert.Equal(t, "", result)
	})
}
