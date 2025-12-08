package colors

import "fmt"

// Luminance represents a luminance value.
type Luminance int16

// Hex returns the hex string for the luminance value.
func (l Luminance) Hex() string {
	return fmt.Sprintf("%X", l)
}
