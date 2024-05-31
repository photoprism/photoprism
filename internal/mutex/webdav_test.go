package mutex

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWebDAV(t *testing.T) {
	lockSystem := WebDAV("test")
	assert.NotNil(t, lockSystem)
}
