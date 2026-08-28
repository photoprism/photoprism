package face

import "strconv"

// String returns the name of a face kind, or its number when the value is not one this release
// knows. 2 and 3 name the retired children and background classifications rather than reading as
// unknown, because a library indexed before they were removed still holds them.
//
// Unclassified is blank rather than named: it is what every cluster is created with, so a column of
// them would say nothing, and a value that is there to be read is one that differs from it.
func (k Kind) String() string {
	switch k {
	case UnclassifiedFace:
		return ""
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
