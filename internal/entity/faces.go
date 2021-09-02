package entity

// Faces represents a Face slice.
type Faces []Face

// Embeddings returns all face embeddings in this slice.
func (f Faces) Embeddings() (embeddings Embeddings) {
	for _, m := range f {
		embeddings = append(embeddings, m.Embedding())
	}

	return embeddings
}

// IDs returns all face IDs in this slice.
func (f Faces) IDs() (ids []string) {
	for _, m := range f {
		ids = append(ids, m.ID)
	}

	return ids
}
