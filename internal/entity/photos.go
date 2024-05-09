package entity

// Photos represents a list of photos.
type Photos []Photo

// Photos returns the result as a slice of Photo.
func (m Photos) Photos() []PhotoInterface {
	result := make([]PhotoInterface, len(m))

	for i := range m {
		result[i] = &m[i]
	}

	return result
}

// UIDs returns tbe photo UIDs as string slice.
func (m Photos) UIDs() []string {
	result := make([]string, len(m))

	for i, photo := range m {
		result[i] = photo.GetUID()
	}

	return result
}
