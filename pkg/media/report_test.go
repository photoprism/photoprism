package media

import (
	"testing"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/stretchr/testify/assert"
)

func TestReport(t *testing.T) {
	m := fs.Extensions.Types(true)
	r, _ := Report(m, true, true, true)
	assert.GreaterOrEqual(t, len(r), 1)
	r2, _ := Report(m, false, true, true)
	assert.GreaterOrEqual(t, len(r2), 1)
}
