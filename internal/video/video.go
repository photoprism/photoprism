/*

Package video provides video file related types and functions.

Copyright (c) 2018 - 2022 Michael Mayer <hello@photoprism.org>

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as published
    by the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.

    PhotoPrism® is a registered trademark of Michael Mayer.  You may use it as required
    to describe our software, run your own server, for educational purposes, but not for
    offering commercial goods, products, or services without prior written permission.
    In other words, please ask.

Feel free to send an e-mail to hello@photoprism.org if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
https://docs.photoprism.app/developer-guide/

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
