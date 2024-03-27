/*
Package i18n provides translatable notification and error messages.

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
package i18n

import (
	"errors"
	"fmt"
	"strings"

	"github.com/leonelquinteros/gotext"
)

//go:generate xgettext --no-wrap --language=c --from-code=UTF-8 --output=../../assets/locales/messages.pot messages.go

type Message int
type MessageMap map[Message]string

func gettext(s string) string {
	return gotext.Get(s)
}

// msgParams replaces message params with the actual values.
func msgParams(msg string, params ...interface{}) string {
	if strings.Contains(msg, "%") {
		msg = fmt.Sprintf(msg, params...)
	}

	return msg
}

// Msg returns a translated message string.
func Msg(id Message, params ...interface{}) string {
	return msgParams(gotext.Get(Messages[id]), params...)
}

// Error returns a translated error message.
func Error(id Message, params ...interface{}) error {
	return errors.New(Msg(id, params...))
}

// Lower returns the untranslated message as a lowercase string for use in logs.
func Lower(id Message, params ...interface{}) string {
	return strings.ToLower(msgParams(Messages[id], params...))
}
