package capture

import (
	"bytes"
	"io"
	"os"
)

// Stdout returns output to stdout for testing.
func Stdout(f func()) string {
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	stdout := os.Stdout
	os.Stdout = w
	defer func() {
		os.Stdout = stdout
	}()

	// Drain the pipe while f runs, since writing more than the pipe buffer holds would
	// otherwise block forever waiting for a reader that only starts afterwards.
	var buf bytes.Buffer
	done := make(chan struct{})

	go func() {
		defer close(done)
		_, _ = io.Copy(&buf, r)
	}()

	f()
	_ = w.Close()
	<-done
	_ = r.Close()

	return buf.String()
}
