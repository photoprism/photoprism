/*
Package geo provides earth geometry functions and constants.

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
package geo

const (
	AverageEarthRadiusKm    = 6371.0                        // Global-average earth radius in km
	AverageEarthRadiusMeter = AverageEarthRadiusKm * 1000.0 // Global-average earth radius in m
	WGS84EarthRadiusKm      = 6378.137                      // WGS84 earth radius in km
	WGS84EarthRadiusMeter   = WGS84EarthRadiusKm * 1000.0   // WGS84 earth radius in m
)
