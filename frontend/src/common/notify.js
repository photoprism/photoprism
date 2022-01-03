/*

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

import Event from "pubsub-js";
import { $gettext } from "./vm";

let ajaxPending = 0;
let ajaxCallbacks = [];

const Notify = {
  info: function (message) {
    Event.publish("notify.info", { message });
  },
  warn: function (message) {
    Event.publish("notify.warning", { message });
  },
  error: function (message) {
    Event.publish("notify.error", { message });
  },
  success: function (message) {
    Event.publish("notify.success", { message });
  },
  logout: function (message) {
    Event.publish("notify.error", { message });
    Event.publish("session.logout", { message });
  },
  ajaxStart: function () {
    ajaxPending++;
    Event.publish("ajax.start");
  },
  ajaxEnd: function () {
    ajaxPending--;
    Event.publish("ajax.end");

    if (!this.ajaxBusy()) {
      ajaxCallbacks.forEach((resolve) => {
        resolve();
      });
    }
  },
  ajaxBusy: function () {
    if (ajaxPending < 0) {
      ajaxPending = 0;
    }

    return ajaxPending > 0;
  },
  ajaxWait: function () {
    return new Promise((resolve) => {
      if (this.ajaxBusy()) {
        ajaxCallbacks.push(resolve);
      } else {
        resolve();
      }
    });
  },
  blockUI: function () {
    const el = document.getElementById("busy-overlay");

    if (el) {
      el.style.display = "block";
    }
  },
  unblockUI: function () {
    const el = document.getElementById("busy-overlay");

    if (el) {
      el.style.display = "none";
    }
  },
  wait: function () {
    this.warn($gettext("Busy, please wait…"));
  },
};

export default Notify;
