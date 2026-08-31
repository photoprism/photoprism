package capture

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOutput(t *testing.T) {
	t.Run("StdoutAndStderr", func(t *testing.T) {
		result := Output(func() {
			fmt.Fprint(os.Stdout, "foo")
			fmt.Fprint(os.Stderr, "bar")
		})

		assert.Equal(t, "foobar", result)
	})
	t.Run("LargerThanPipeBuffer", func(t *testing.T) {
		// Pipes hold 64 KiB on Linux, so a command that prints more than that blocks
		// until the buffer is drained. Reading only after the function returns would
		// deadlock here instead of failing.
		line := strings.Repeat("x", 1024)

		result := Output(func() {
			for range 256 {
				fmt.Fprintln(os.Stdout, line)
			}
		})

		assert.Len(t, result, 256*(len(line)+1))
	})
}
