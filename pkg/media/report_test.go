package media

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestReport(t *testing.T) {
	m := fs.Extensions.Types(true)
	r, _ := Report(m, true, true, true)
	assert.GreaterOrEqual(t, len(r), 1)
	r2, _ := Report(m, false, true, true)
	assert.GreaterOrEqual(t, len(r2), 1)
}
