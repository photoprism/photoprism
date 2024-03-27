/*
Package pluscode provides an abstraction for Google's Open Location Code package.

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
package pluscode

import (
	"fmt"

	olc "github.com/google/open-location-code/go"
	"github.com/photoprism/photoprism/pkg/s2"
)

var defaultLen = 8

// Encode returns the plus code for the given coordinates using the default length.
func Encode(lat, lng float64) string {
	result, _ := EncodeLength(lat, lng, defaultLen)

	return result
}

// EncodeLength returns the plus code for the given coordinates.
func EncodeLength(lat, lng float64, length int) (plusCode string, err error) {
	if lat < -90 || lat > 90 {
		return "", fmt.Errorf("latitude out of range (%f)", lat)
	}

	if lng < -180 || lng > 180 {
		return "", fmt.Errorf("longitude out of range (%f)", lng)
	}

	return olc.Encode(lat, lng, length), nil
}

// LatLng returns the coordinates for a plus code token.
func LatLng(token string) (lat, lng float64) {
	if token == "" || token == "-" {
		return lat, lng
	}

	c, err := olc.Decode(token)

	if err != nil {
		return lat, lng
	}

	lat, lng = c.Center()

	return lat, lng
}

// S2 returns the S2 cell token for the plus code using the default cell level.
func S2(plusCode string) string {
	lat, lng := LatLng(plusCode)

	s2Token := s2.Token(lat, lng)

	return s2Token
}
