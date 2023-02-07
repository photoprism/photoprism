package colors

import "fmt"

// Chroma represents colorfulness.
type Chroma int16

// Percent returns the colourfulness in percent.
func (c Chroma) Percent() int16 {
	if c > 100 {
		return 100
	} else if c < 0 {
		return 0
	}

	return int16(c)
}

// Hex returns the colourfulness in percent has hex string.
func (c Chroma) Hex() string {
	return fmt.Sprintf("%X", c.Percent())
}

// Uint returns the colourfulness in percent as unsigned integer.
func (c Chroma) Uint() uint {
	return uint(c.Percent())
}

// Int returns the colourfulness in percent as integer.
func (c Chroma) Int() int {
	return int(c.Percent())
}
