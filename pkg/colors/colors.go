/*
Package colors provides types and functions for color classification.

Copyright (c) 2018 - 2024 PhotoPrism UG. All rights reserved.

	This program is free software: you can redistribute it and/or modify
	it under Version 3 of the GNU Affero General Public License (the "AGPL"):
	<https://docs.photoprism.app/license/agpl>

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	The AGPL is supplemented by our Trademark and Brand Guidelines,
	which describe how our Brand Assets may be used:
	<https://www.photoprism.app/trademark>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>
*/
package colors

import (
	"fmt"

	"github.com/photoprism/photoprism/pkg/txt"
)

type Color int16
type Colors []Color

const (
	Black Color = iota
	Grey
	Brown
	Gold
	White
	Purple
	Blue
	Cyan
	Teal
	Green
	Lime
	Yellow
	Magenta
	Orange
	Red
	Pink
)

var All = Colors{
	Purple,
	Magenta,
	Pink,
	Red,
	Orange,
	Gold,
	Yellow,
	Lime,
	Green,
	Teal,
	Cyan,
	Blue,
	Brown,
	White,
	Grey,
	Black,
}

var Names = map[Color]string{
	Black:   "black",   // 0
	Grey:    "grey",    // 1
	Brown:   "brown",   // 2
	Gold:    "gold",    // 3
	White:   "white",   // 4
	Purple:  "purple",  // 5
	Blue:    "blue",    // 6
	Cyan:    "cyan",    // 7
	Teal:    "teal",    // 8
	Green:   "green",   // 9
	Lime:    "lime",    // A
	Yellow:  "yellow",  // B
	Magenta: "magenta", // C
	Orange:  "orange",  // D
	Red:     "red",     // E
	Pink:    "pink",    // F
}

var Weights = map[Color]uint16{
	Grey:    1,
	Black:   2,
	Brown:   2,
	White:   2,
	Blue:    3,
	Green:   3,
	Purple:  4,
	Gold:    4,
	Cyan:    4,
	Teal:    4,
	Orange:  4,
	Red:     4,
	Pink:    4,
	Lime:    5,
	Yellow:  5,
	Magenta: 5,
}

func (c Color) Name() string {
	return Names[c]
}

func (c Color) ID() int16 {
	return int16(c)
}

func (c Color) Hex() string {
	return fmt.Sprintf("%X", c)
}

func (c Colors) Hex() (result string) {
	for _, indexedColor := range c {
		result += indexedColor.Hex()
	}

	return result
}

func (c Colors) List() []map[string]string {
	result := make([]map[string]string, 0, len(c))

	for _, c := range c {
		result = append(result, map[string]string{"Slug": c.Name(), "Name": txt.UpperFirst(c.Name()), "Example": ColorExamples[c]})
	}

	return result
}
