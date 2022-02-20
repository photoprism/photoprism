/*

Package i18n contains PhotoPrism status and error message strings.

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

Feel free to send an e-mail to hello@photoprism.org if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
https://docs.photoprism.app/developer-guide/

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

func Msg(id Message, params ...interface{}) string {
	msg := gotext.Get(Messages[id])

	if strings.Contains(msg, "%") {
		msg = fmt.Sprintf(msg, params...)
	}

	return msg
}

func Error(id Message, params ...interface{}) error {
	return errors.New(strings.ToLower(Msg(id, params...)))
}
