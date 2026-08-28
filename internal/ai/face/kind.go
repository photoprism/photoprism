package face

import "strconv"

// String returns the name of a face kind, or its number when the value is not one this release
// knows. 2 and 3 name the retired children and background classifications rather than reading as
// unknown, because a library indexed before they were removed still holds them.
func (k Kind) String() string {
	switch k {
	case 0:
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
