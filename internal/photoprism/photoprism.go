/*
Package photoprism provides the core functionality of PhotoPrism®.

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
package photoprism

import (
	"github.com/photoprism/photoprism/internal/event"
)

var log = event.Log

type S []string

// logWarn logs an error as warning and keeps quiet otherwise.
func logWarn(prefix string, err error) {
	if err != nil {
		log.Warnf("%s: %s", prefix, err.Error())
	}
}

// logErr logs an error and keeps quiet otherwise.
func logErr(prefix, action string, err error) {
	if err != nil {
		log.Errorf("%s: %s (%s)", prefix, err, action)
	}
}
