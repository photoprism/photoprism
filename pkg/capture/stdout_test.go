package capture

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStdout(t *testing.T) {
	t.Run("StdoutOnly", func(t *testing.T) {
		result := Stdout(func() {
			fmt.Fprint(os.Stdout, "foo")
			fmt.Fprint(os.Stderr, "bar")
		})

		assert.Equal(t, "foo", result)
	})
	t.Run("LargerThanPipeBuffer", func(t *testing.T) {
		line := strings.Repeat("x", 1024)

		result := Stdout(func() {
			for range 256 {
				fmt.Fprintln(os.Stdout, line)
			}
		})

		assert.Len(t, result, 256*(len(line)+1))
	})
}
