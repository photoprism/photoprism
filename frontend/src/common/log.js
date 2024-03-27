/*

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

import Event from "pubsub-js";
import { config } from "app/session";

class Log {
  constructor() {
    this.cap = 150;
    this.created = new Date();
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

  debug(msg, data) {
    if (config.debug && msg) {
      if (data) {
        if (Array.isArray(data)) {
          data.forEach((val) => {
            msg += ", " + JSON.stringify(val);
          });
        } else {
          msg += ", " + JSON.stringify(data);
        }
      }

      this.onLog(
        {},
        {
          message: msg,
          level: "debug",
          time: new Date().toISOString(),
        }
      );
    }

    return this;
  }

  onLog(ev, data) {
    data.id = this.logId++;

    this.logs.unshift(data);

    if (this.logs.length > this.cap) {
      this.logs.splice(this.cap);
    }
  }
}

const log = new Log();

export default log;
