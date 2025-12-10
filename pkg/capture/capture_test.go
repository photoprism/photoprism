package capture

import (
	"fmt"
	"os"
	"testing"
)

func TestOutputMergesStdoutAndStderr(t *testing.T) {
	got := Output(func() {
		fmt.Print("out")
		fmt.Fprint(stderrWriter(), "err") // write directly to stderr
	})
	if got != "outerr" {
		t.Fatalf("unexpected combined output: %q", got)
	}
}

// stderrWriter returns the current process stderr; split for test clarity.
func stderrWriter() *os.File {
	return os.Stderr
}
