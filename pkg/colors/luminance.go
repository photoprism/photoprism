package colors

import "fmt"

type Luminance uint8

func (l Luminance) Hex() string {
	return fmt.Sprintf("%X", l)
}
