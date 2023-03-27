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

import Event from "pubsub-js";
import { $gettext } from "vm.js";

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
  blockUI: function (className) {
    const el = document.getElementById("busy-overlay");

    if (el) {
      el.style.display = "block";
      if (className) {
        el.className = className;
      }
    }
  },
  unblockUI: function () {
    const el = document.getElementById("busy-overlay");

    if (el) {
      el.style.display = "none";
      el.className = "";
    }
  },
  wait: function () {
    this.info($gettext("Please wait…"));
  },
  busy: function () {
    this.warn($gettext("Busy, please wait…"));
  },
};

export default Notify;
