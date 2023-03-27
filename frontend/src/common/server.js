/*

Copyright (c) 2018 - 2023 PhotoPrism UG. All rights reserved.

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

import Api from "api.js";
import { config } from "app/session";
import Notify from "notify.js";

function poll(interval, maxAttempts) {
  let attempts = 0;

  const executePoll = async (resolve, reject) => {
    attempts++;

    try {
      const xhr = new XMLHttpRequest();
      xhr.open("GET", config.apiUri + "/status", false);
      xhr.send();
      return resolve();
    } catch {
      if (maxAttempts && attempts === maxAttempts) {
        return reject(new Error("exceeded max attempts"));
      } else {
        setTimeout(executePoll, interval, resolve, reject);
      }
    }
  };

  return new Promise(executePoll);
}

export function restart() {
  Notify.blockUI("progress");
  Notify.wait();
  Notify.ajaxStart();

  return Api.post("server/stop")
    .then(() => {
      return poll(1000, 180)
        .then(() => {
          window.location.reload();
        })
        .catch(() => {
          Notify.ajaxEnd();
          Notify.error("Failed to restart server");
          Notify.unblockUI();
        });
    })
    .catch(() => {
      Notify.ajaxEnd();
      Notify.error("Failed to restart server");
      Notify.unblockUI();
    });
}
