package video

// Chunks represents a list of file chunks.
type Chunks []Chunk

// ContainsAny checks if at least one common chunk exists.
func (c Chunks) ContainsAny(b [][4]byte) bool {
	if len(c) == 0 || len(b) == 0 {
		return false
	}

	// Find matches.
	for i := range c {
		for j := range b {
			if b[j] == c[i] {
				return true
			}
		}
	}

	// Not found.
	return false
}
