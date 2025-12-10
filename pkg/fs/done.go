package fs

// Status indicates whether a path was seen or processed.
type Status int8

const (
	// Found marks a path as seen.
	Found Status = 1
	// Processed marks a path as fully handled.
	Processed Status = 2
)

// Done stores per-path processing state.
type Done map[string]Status

// Processed counts the number of processed files.
func (d Done) Processed() int {
	count := 0

	for _, s := range d {
		if s.Processed() {
			count++
		}
	}

	return count
}

// Exists reports whether any status is recorded.
func (s Status) Exists() bool {
	return s > 0
}

// Processed returns true if the path was marked as processed.
func (s Status) Processed() bool {
	return s >= Processed
}
