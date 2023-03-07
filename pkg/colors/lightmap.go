package colors

type LightMap []Luminance

// Hex returns all luminance value as a hex encoded string.
func (m LightMap) Hex() (result string) {
	for _, luminance := range m {
		if luminance > 15 {
			result += "F"
		} else {
			result += luminance.Hex()
		}
	}

	return result
}

type diffValue struct {
	a []int
	b []int
}

// Pixel comparisons for calculating the diff value.
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

/*
	Alternative values to experiment with:

    {a: []int{4, 4, 4, 4}, b: []int{1, 3, 5, 7}},
	{a: []int{0}, b: []int{2}},
	{a: []int{8}, b: []int{6}},
	{a: []int{0}, b: []int{1}},
	{a: []int{1}, b: []int{2}},
	{a: []int{2}, b: []int{5}},
	{a: []int{8}, b: []int{7}},
	{a: []int{7}, b: []int{6}},
	{a: []int{5}, b: []int{8}},
	{a: []int{6}, b: []int{3}},

	Other ideas:
	- Use Lightness instead of luminance
    - Use more pixels (difficult as we only have 3x3 thumbs right now)
	- Iterative calculation using smaller offsets for each round (current offset is +4)

	For a more complex implementation, see
	https://github.com/EdjoLabs/image-match/blob/master/image_match/goldberg.py
    http://www.cs.cmu.edu/~hcwong/Pdfs/icip02.ps

	General blog post on Perceptual image hashing:
	https://jenssegers.com/perceptual-image-hashes
*/

// Diff returns an integer that can be used to find similar images.
func (m LightMap) Diff() (result int) {
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

		if a+4 > b {
			result += 1
		}
	}

	return result
}
