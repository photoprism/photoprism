package entity

// Photos represents a list of photos.
type Photos []*Photo

// Photos returns the result as a slice of Photo.
func (m Photos) Photos() []PhotoInterface {
	result := make([]PhotoInterface, len(m))

	for i := range m {
		result[i] = m[i]
	}

	return result
}

// UIDs returns the photo UIDs as string slice.
func (m Photos) UIDs() []string {
	result := make([]string, len(m))

	for i, photo := range m {
		result[i] = photo.GetUID()
	}

	return result
}

type PhotoSet struct {
	order []string
	items map[string]*Photo
	idx   map[string]int // UID -> index in order
}

func NewPhotoSet() *PhotoSet {
	return &PhotoSet{
		order: make([]string, 0),
		items: make(map[string]*Photo),
		idx:   make(map[string]int),
	}
}

func (s *PhotoSet) Add(r *Photo) {
	uid := r.GetUID()

	if _, exists := s.items[uid]; !exists {
		s.order = append(s.order, uid)
		s.idx[uid] = len(s.order) - 1
	}

	s.items[uid] = r
}

func (s *PhotoSet) Delete(uid string) {
	i, ok := s.idx[uid]

	if !ok {
		return
	}

	lastIdx := len(s.order) - 1
	lastUID := s.order[lastIdx]

	// swap removed element with last
	s.order[i] = lastUID
	s.idx[lastUID] = i

	// shrink slice
	s.order = s.order[:lastIdx]

	delete(s.idx, uid)
	delete(s.items, uid)
}
