package s2

// DefaultLevel specifies the default S2 cell size.
var DefaultLevel = 21

// Level returns the S2 cell level based on the approximate cell size in km.
// see https://s2geometry.io/resources/s2cell_statistics.html
func Level(km uint) (level int) {
	switch {
	case km >= 7842:
		return 0
	case km >= 3921:
		return 1
	case km >= 1825:
		return 2
	case km >= 1130:
		return 3
	case km >= 579:
		return 4
	case km >= 287:
		return 5
	case km >= 143:
		return 6
	case km >= 72:
		return 7
	case km >= 36:
		return 8
	case km >= 18:
		return 9
	case km >= 9:
		return 10
	case km >= 4:
		return 11
	case km >= 2:
		return 12
	case km >= 1:
		return 13
	default:
		return 14
	}
}
