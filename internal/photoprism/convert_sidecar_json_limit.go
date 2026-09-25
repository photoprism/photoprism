package photoprism

import "github.com/photoprism/photoprism/internal/meta"

// jsonOutputBuffer bounds command output while optionally refusing excess bytes.
type jsonOutputBuffer struct {
	data        []byte
	limit       int
	failOnLimit bool
	exceeded    bool
}

// Write captures at most the configured limit and reports or discards overflow.
func (b *jsonOutputBuffer) Write(p []byte) (int, error) {
	n := len(p)
	remaining := b.limit - len(b.data)

	if n > remaining {
		b.data = append(b.data, p[:remaining]...)
		b.exceeded = true

		if b.failOnLimit {
			return remaining, meta.ErrJSONFileTooLarge
		}

		return n, nil
	}

	b.data = append(b.data, p...)
	return n, nil
}
