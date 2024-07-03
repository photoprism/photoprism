package colors

import "fmt"

type Luminance int16

func (l Luminance) Hex() string {
	return fmt.Sprintf("%X", l)
}
