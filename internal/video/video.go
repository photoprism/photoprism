/*

Package video provides video file related types and functions.

Copyright (c) 2018 - 2022 Michael Mayer <hello@photoprism.app>

    This program is free software: you can redistribute it and/or modify
    it under Version 3 of the GNU Affero General Public License (the "AGPL"):
    <https://docs.photoprism.app/license/agpl>

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    The AGPL is supplemented by our Trademark and Brand Guidelines,
    which describe how our Brand Assets may be used:
    <https://photoprism.app/trademark>

Feel free to send an e-mail to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>

*/
package video

import (
	"github.com/photoprism/photoprism/pkg/fs"
)

type Type struct {
	Format fs.FileFormat
	Codec  fs.FileCodec
	Width  int
	Height int
	Public bool
}

type TypeMap map[string]Type

var TypeMp4 = Type{
	Format: fs.FormatMp4,
	Codec:  fs.CodecAvc,
	Width:  0,
	Height: 0,
	Public: true,
}

var TypeAvc = Type{
	Format: fs.FormatAvc,
	Codec:  fs.CodecAvc,
	Width:  0,
	Height: 0,
	Public: true,
}

var Types = TypeMap{
	"":    TypeAvc,
	"mp4": TypeMp4,
	"avc": TypeAvc,
}
