package projection

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTNew(t *testing.T) {
	proj := New(" FOO-bar! FOO-bar! 0123 XXXXXXXXXXXXXXXXXXXXXXXXXXXX  YYYYY            YYYYYYYYY       YZ")

	assert.Equal(t, "foo-bar! foo-bar! 0123 xxxxxxxxxxxxxxxxxxxxxxxxxxxx  yyyyy", proj.String())
}
