/*

Copyright (c) 2018 - 2020 Michael Mayer <hello@photoprism.org>

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

    PhotoPrism™ is a registered trademark of Michael Mayer.  You may use it as required
    to describe our software, run your own server, for educational purposes, but not for
    offering commercial goods, products, or services without prior written permission.
    In other words, please ask.

Feel free to send an e-mail to hello@photoprism.org if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
https://docs.photoprism.org/developer-guide/

*/

import Event from "pubsub-js";

class Log {
    constructor() {
        this.cap = 100;
        this.created = new Date;
        this.logs = [
            /* EXAMPLE LOG MESSAGE
            {
                "message": "waiting for events",
                "level": "debug",
                "time": this.created.toISOString(),
            },
            */
        ];

        this.logId = 0;

        Event.subscribe("log", this.onLog.bind(this));
    }

    onLog(ev, data) {
        data.id = this.logId++;

        this.logs.unshift(data);

        if(this.logs.length > this.cap) {
            this.logs.splice(this.cap);
        }
    }
}

const log = new Log;

export default log;
