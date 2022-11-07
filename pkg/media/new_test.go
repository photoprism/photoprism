package media

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	mediaType := New(" FOO-bar! FOO-bar! 0123 XXXXXXXXXXXXXXXXXXXXXXXXXXXX  YYYYY            YYYYYYYYY       YZ")

	assert.Equal(t, "foo-bar!", mediaType.String())
}
