package colors

import "fmt"

// Chroma represents colorfulness.
type Chroma int16

// Percent returns the colorfulness in percent.
func (c Chroma) Percent() int16 {
	if c > 100 {
		return 100
	} else if c < 0 {
		return 0
	}

	return int16(c)
}

// Value returns the chroma value stored in the index: the percentage plus one, clamped to 100,
// so that 1 means monochrome and is never confused with -1, which stands for unknown values.
func (c Chroma) Value() int16 {
	if p := c.Percent(); p < 100 {
		return p + 1
	}

	return 100
}

// Hex returns the colorfulness in percent has hex string.
func (c Chroma) Hex() string {
	return fmt.Sprintf("%X", c.Percent())
}

// Uint returns the colorfulness in percent as unsigned integer.
func (c Chroma) Uint() uint {
	return uint(c.Percent()) //nolint:gosec // Percent is bounded 0..100
}

// Int returns the colorfulness in percent as integer.
func (c Chroma) Int() int {
	return int(c.Percent())
}
