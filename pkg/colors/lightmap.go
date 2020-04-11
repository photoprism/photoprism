package colors

type LightMap []Luminance

// Hex returns all luminance value as a hex encoded string.
func (m LightMap) Hex() (result string) {
	for _, luminance := range m {
		result += luminance.Hex()
	}

	return result
}

type diffValue struct {
	a []int
	b []int
}

var diffValues = []diffValue{
	{a: []int{4, 4, 4, 4}, b: []int{1, 3, 5, 7}},
	{a: []int{0}, b: []int{1}},
	{a: []int{0}, b: []int{3}},
	{a: []int{2}, b: []int{1}},
	{a: []int{2}, b: []int{5}},
	{a: []int{6}, b: []int{3}},
	{a: []int{6}, b: []int{7}},
	{a: []int{8}, b: []int{5}},
	{a: []int{8}, b: []int{7}},
}

// Diff returns an integer that can be used to find similar images.
func (m LightMap) Diff() (result uint32) {
	if len(m) != 9 {
		return 0
	}

	result = 1

	for _, val := range diffValues {
		result = result << 1

		a := 0
		b := 0

		for _, i := range val.a {
			a += int(m[i])
		}

		for _, i := range val.b {
			b += int(m[i])
		}

		if a + 4 > b {
			result += 1
		}
	}

	return result
}
