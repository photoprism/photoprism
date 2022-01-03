/*

Package s2 encapsulates Google's S2 library.

See https://s2geometry.io/

Copyright (c) 2018 - 2022 Michael Mayer <hello@photoprism.app>

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

Feel free to send an e-mail to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
https://docs.photoprism.app/developer-guide/

*/
package s2

import (
	gs2 "github.com/golang/geo/s2"
)

// Default cell level, see https://s2geometry.io/resources/s2cell_statistics.html.
var DefaultLevel = 21

// Token returns the S2 cell token for coordinates using the default level.
func Token(lat, lng float64) string {
	return TokenLevel(lat, lng, DefaultLevel)
}

// Token returns the S2 cell token for coordinates.
func TokenLevel(lat, lng float64, level int) string {
	if lat == 0.0 && lng == 0.0 {
		return ""
	}

	if lat < -90 || lat > 90 {
		return ""
	}

	if lng < -180 || lng > 180 {
		return ""
	}

	l := gs2.LatLngFromDegrees(lat, lng)
	return gs2.CellIDFromLatLng(l).Parent(level).ToToken()
}

// LatLng returns the coordinates for a S2 cell token.
func LatLng(token string) (lat, lng float64) {
	token = NormalizeToken(token)

	if len(token) < 3 {
		return 0.0, 0.0
	}

	c := gs2.CellIDFromToken(token)

	if !c.IsValid() {
		return 0.0, 0.0
	}

	l := c.LatLng()
	return l.Lat.Degrees(), l.Lng.Degrees()
}

// IsZero returns true if the coordinates are both empty.
func IsZero(lat, lng float64) bool {
	return lat == 0.0 && lng == 0.0
}

// Range returns a token range for finding nearby locations.
func Range(token string, levelUp int) (min, max string) {
	token = NormalizeToken(token)

	c := gs2.CellIDFromToken(token)

	if !c.IsValid() {
		return min, max
	}

	// See https://s2geometry.io/resources/s2cell_statistics.html
	lvl := c.Level()

	parent := c.Parent(lvl - levelUp)

	return parent.Prev().ChildBeginAtLevel(lvl).ToToken(), parent.Next().ChildBeginAtLevel(lvl).ToToken()
}
