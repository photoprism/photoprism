package face

import "strconv"

// Kind returns the classification of a single embedding.
//
// A usable vector is a regular face; nothing classifies faces further since the child and
// background filters were removed as inert. One with no magnitude describes no face, so a cluster
// built from it would sit one unit from every unit vector.
func (m Embedding) Kind() Kind {
	if len(m) == 0 || m.Zero() {
		return UnclassifiedFace
	}

	return RegularFace
}

// Kind returns the classification of a set of embeddings, which is the highest one any of them
// carries so a single uncertain member is not hidden by its neighbors.
func (embeddings Embeddings) Kind() (result Kind) {
	for _, e := range embeddings {
		if k := e.Kind(); k > result {
			result = k
		}
	}

	return result
}

// String returns the name of a face kind, or its number when the value is not one this release
// knows. 2 and 3 name the retired children and background classifications rather than reading as
// unknown, because a library indexed before they were removed still holds them.
func (k Kind) String() string {
	switch k {
	case UnclassifiedFace:
		return "unset"
	case RegularFace:
		return "regular"
	case 2:
		return "children"
	case 3:
		return "background"
	case AmbiguousFace:
		return "ambiguous"
	default:
		return strconv.Itoa(int(k))
	}
}
