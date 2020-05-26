/*
This package contains video related types and functions.

Additional information can be found in our Developer Guide:

https://github.com/photoprism/photoprism/wiki
*/
package video

import (
	"github.com/photoprism/photoprism/pkg/fs"
)

type Type struct {
	Format fs.FileType
	Width  int
	Height int
	Public bool
}

type TypeMap map[string]Type

var TypeMP4 = Type{
	Format: fs.TypeMP4,
	Width:  0,
	Height: 0,
	Public: true,
}

var Types = TypeMap{
	"":    TypeMP4,
	"mp4": TypeMP4,
}
