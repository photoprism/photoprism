package capture

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStdout(t *testing.T) {
	result := Stdout(func() {
		fmt.Fprint(os.Stdout, "foo")
		fmt.Fprint(os.Stderr, "bar")
	})

	assert.Equal(t, "foo", result)
}
